// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

// Package flowexit owns the loop-exit-by-expression mechanism. It compiles
// the operator-supplied CEL expression once (at flow build time, after the
// admin API has already validated it), and evaluates it after each loop
// iteration against a snapshot of the shared flow state.
//
// CEL was picked over a hand-rolled DSL because it is:
//   - the de-facto Google evaluation language (Kubernetes, IAM, Envoy);
//   - explicitly designed for compile-once / evaluate-many pipelines;
//   - thread-safe and side-effect free;
//   - schema-aware enough to reject expressions that don't return a bool;
//   - expressive (==, !=, <, >, &&, ||, !, has(), size(), .contains())
//     without inviting Turing-complete computation.
//
// See decision #28 in .agents/DECISIONS.md.
package flowexit

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/ext"
)

// celEnv declares the variables available to router guard expressions:
//   - `input`: the output of the previous node, typed dyn so any shape is
//     accepted. Lets a router branch on upstream content directly, e.g.
//     `input.contains("error")`.
//   - `state`: a map<string, dyn> populated by the flow runner from session
//     state values written through set_state (the "flow:" namespace).
//   - `iterations`: an int, how many times the evaluating router node has been
//     activated in the current run. Lets operators cap loops in CEL, e.g.
//     `state.done == true || iterations >= 5`.
//
// The variable is package-level because cel.Env construction is non-trivial
// and the env is stateless and safe to share.
var celEnv *cel.Env

// celValueEnv is the environment for expression (transform) nodes. Like
// celEnv it exposes `input` and `state`, and its expressions may return any
// type, not just bool. There is no `iterations` here: transform nodes are not
// loop guards.
var celValueEnv *cel.Env

func init() {
	env, err := cel.NewEnv(
		cel.Variable("input", cel.DynType),
		cel.Variable("state", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("iterations", cel.IntType),
	)
	if err != nil {
		// init-time failure means the CEL declaration is wrong, which is
		// strictly a programming bug. Better to fail at startup than to
		// surface a confusing runtime error per flow build.
		panic(fmt.Errorf("flowexit: cel.NewEnv failed: %w", err))
	}
	celEnv = env

	venv, err := cel.NewEnv(
		cel.Variable("input", cel.DynType),
		cel.Variable("state", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("secret", cel.MapType(cel.StringType, cel.StringType)),
		// String and list extensions give transform expressions the tools they
		// exist for: split/join/substring/replace and list map/filter helpers.
		// e.g. input.split(",") is what feeds a downstream For-Each node.
		ext.Strings(),
		ext.Lists(),
	)
	if err != nil {
		panic(fmt.Errorf("flowexit: cel.NewEnv (value) failed: %w", err))
	}
	celValueEnv = venv
}

// Compile parses, type-checks and lowers the supplied CEL expression. The
// returned cel.Program is stateless and can be shared across goroutines.
//
// Compile rejects expressions that fail parsing, type-checking, or whose
// inferred output type is not bool. The admin API calls Compile during
// flow validation so malformed expressions are caught at save time; the
// flow builder calls Compile again at startup to obtain the runnable
// program (programs are not serialisable, so we keep only the source
// string in the store).
func Compile(expr string) (cel.Program, error) {
	if expr == "" {
		return nil, fmt.Errorf("expression is empty")
	}
	ast, issues := celEnv.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("invalid CEL expression: %w", issues.Err())
	}
	if ast.OutputType() != types.BoolType {
		return nil, fmt.Errorf("expression must return bool, got %s", ast.OutputType().String())
	}
	prog, err := celEnv.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("failed to lower CEL program: %w", err)
	}
	return prog, nil
}

// CompileValue parses, type-checks and lowers a CEL expression for a transform
// (expression) node. It exposes `input` and `state` and, unlike Compile,
// accepts any output type (the expression's job is to produce a value, not a
// boolean). Called by admin validation at save time and by the flow builder
// at startup.
func CompileValue(expr string) (cel.Program, error) {
	if expr == "" {
		return nil, fmt.Errorf("expression is empty")
	}
	ast, issues := celValueEnv.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("invalid CEL expression: %w", issues.Err())
	}
	prog, err := celValueEnv.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("failed to lower CEL program: %w", err)
	}
	return prog, nil
}
