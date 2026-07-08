// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	adkagent "google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"

	"github.com/achetronic/magec/server/agent/flowexit"
	toolsflowstate "github.com/achetronic/magec/server/agent/tools/flowstate"
	"github.com/achetronic/magec/server/store"
)

// buildExpressionNode builds a transform node whose body is a CEL expression
// over `input` (the previous node's output) and `state` (the shared flow
// state). The expression's result becomes the node's output; when OutputKey is
// set it is also written into flow state under that key (readable downstream as
// state.<key>).
func buildExpressionNode(node store.FlowNode, secretMap map[string]string) (workflow.Node, error) {
	prog, err := flowexit.CompileValue(node.Expression)
	if err != nil {
		// Should not happen: the admin API validates at save time. Recompiling
		// here is necessary because programs are not serialisable.
		return nil, fmt.Errorf("expression node %q: %w", node.ID, err)
	}
	expr := node.Expression
	outputKey := node.OutputKey

	fn := func(ctx adkagent.Context, input any, emit func(*session.Event) error) (any, error) {
		state := flowexit.ExtractFlowState(ctx.State())
		out, err := flowexit.EvaluateValue(prog, expr, input, state, secretMap)
		if err != nil {
			return nil, fmt.Errorf("expression node %q: %w", node.ID, err)
		}
		ev := session.NewEvent(ctx, ctx.InvocationID())
		ev.Author = node.ID
		ev.Output = out
		if outputKey != "" {
			ev.Actions.StateDelta[toolsflowstate.StateKeyPrefix+outputKey] = out
		}
		if err := emit(ev); err != nil {
			return nil, err
		}
		return nil, nil
	}
	return workflow.NewEmittingFunctionNode[any, any](node.ID, fn, workflow.NodeConfig{}), nil
}

// buildTemplateNode builds a transform node that renders a text template. The
// rendered string becomes the node's output; when OutputKey is set it is also
// written into flow state under that key.
func buildTemplateNode(node store.FlowNode, secretMap map[string]string) (workflow.Node, error) {
	tmpl := node.Template
	outputKey := node.OutputKey

	fn := func(ctx adkagent.Context, input any, emit func(*session.Event) error) (any, error) {
		state := flowexit.ExtractFlowState(ctx.State())
		out := renderTemplate(tmpl, input, state, secretMap)
		ev := session.NewEvent(ctx, ctx.InvocationID())
		ev.Author = node.ID
		ev.Output = out
		if outputKey != "" {
			ev.Actions.StateDelta[toolsflowstate.StateKeyPrefix+outputKey] = out
		}
		if err := emit(ev); err != nil {
			return nil, err
		}
		return nil, nil
	}
	return workflow.NewEmittingFunctionNode[any, any](node.ID, fn, workflow.NodeConfig{}), nil
}

// placeholderRE matches a template placeholder: {{ input }}, {{ input.field }},
// {{ state.key }}, {{ secret.KEY }}. The captured group is a dot path starting
// with input, state or secret. Whitespace inside the braces is tolerated.
var placeholderRE = regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_.]*)\s*\}\}`)

// renderTemplate resolves each placeholder against input and state by walking
// the dot path into nested maps. An unknown root or a path that does not
// resolve renders as an empty string, so a template never fails the flow.
func renderTemplate(tmpl string, input any, state map[string]any, secret map[string]string) string {
	return placeholderRE.ReplaceAllStringFunc(tmpl, func(match string) string {
		path := placeholderRE.FindStringSubmatch(match)[1]
		segs := strings.Split(path, ".")

		var cur any
		switch segs[0] {
		case "input":
			cur = input
		case "state":
			cur = state
		case "secret":
			if len(segs) == 2 {
				if v, ok := secret[segs[1]]; ok {
					return v
				}
			}
			return ""
		default:
			return ""
		}
		for _, s := range segs[1:] {
			m, ok := cur.(map[string]any)
			if !ok {
				return ""
			}
			cur = m[s]
		}
		return stringify(cur)
	})
}

// stringify renders a resolved value for template output: strings verbatim,
// nil as empty, maps and slices as compact JSON, everything else via %v.
func stringify(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case map[string]any, []any:
		if b, err := json.Marshal(x); err == nil {
			return string(b)
		}
		return fmt.Sprintf("%v", x)
	default:
		return fmt.Sprintf("%v", x)
	}
}
