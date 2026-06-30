// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package flowexit

import (
	"fmt"
	"iter"
	"log/slog"
	"strings"

	"github.com/google/cel-go/cel"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/session"

	toolsflowstate "github.com/achetronic/magec/server/agent/tools/flowstate"
)

// NewExitWhenAgent builds a synthetic ADK agent that, when run, evaluates
// the supplied CEL program against the current flow state and — if it
// returns true — emits a single event with Actions.Escalate=true. That
// event bubbles up to the enclosing loopagent (ADK's native one,
// untouched), which already reacts to Escalate by terminating the loop
// after the current iteration completes.
//
// Wiring strategy: the flow builder adds this agent as the last child of
// the loop step's sub-agent list. ADK runs it after every iteration of
// the user-defined sub-agents; if it triggers Escalate, the loop ends.
// This keeps the loop step a real loopagent.LoopAgent (so A2A agent-card
// generation, ADK introspection, etc. all keep working) without us having
// to reinvent loop bookkeeping.
//
// The agent ignores any LLM machinery — it is not an llmagent and never
// produces user-visible content. The synthetic event it emits when CEL
// is true carries no LLMResponse on purpose; ConversationRecorder filters
// such marker events out of user-perspective transcripts.
//
// On CEL runtime errors (missing key, type mismatch, etc.) the agent
// treats the result as false (continue iterating) and logs a warning.
// MaxIterations on the surrounding loopagent remains the hard safety cap.
func NewExitWhenAgent(name string, prog cel.Program, expr string) (adkagent.Agent, error) {
	if prog == nil {
		return nil, fmt.Errorf("CEL program is required")
	}
	return adkagent.New(adkagent.Config{
		Name:        name,
		Description: "Loop exit-when evaluator",
		Run: func(ctx adkagent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				stateMap := ExtractFlowState(ctx.Session().State())
				if !EvaluateExitWhen(prog, expr, stateMap) {
					return
				}
				ev := session.NewEvent(ctx.InvocationID())
				ev.Author = name
				ev.Actions.Escalate = true
				_ = yield(ev, nil)
			}
		},
	})
}

// EvaluateExitWhen runs the CEL program against the supplied flow-state
// snapshot and returns whether the loop should exit. It is exported (and
// extracted from the agent closure) so unit tests can exercise the
// boolean / error / non-bool paths without instantiating an ADK
// invocation context.
//
// expr is included only for log clarity when the program errors at
// evaluation time; it is not re-parsed.
func EvaluateExitWhen(prog cel.Program, expr string, stateMap map[string]any) bool {
	out, _, err := prog.Eval(map[string]any{"state": stateMap})
	if err != nil {
		slog.Warn("flowexit: CEL evaluation failed; treating as false",
			"expression", expr, "error", err, "state_keys", keysOf(stateMap))
		return false
	}
	b, ok := out.Value().(bool)
	if !ok {
		slog.Warn("flowexit: CEL did not return bool; treating as false",
			"expression", expr, "result", out.Value())
		return false
	}
	return b
}

// ExtractFlowState returns the subset of session state keys that live
// under the flow_state namespace, with the prefix stripped. CEL
// expressions reference these as `state.<key>`.
//
// Exported so tests and (potentially) other parts of the agent layer can
// reuse the same prefix-stripping logic without re-importing
// flowstate.StateKeyPrefix.
func ExtractFlowState(s session.State) map[string]any {
	out := map[string]any{}
	for k, v := range s.All() {
		if after, ok := strings.CutPrefix(k, toolsflowstate.StateKeyPrefix); ok {
			out[after] = v
		}
	}
	return out
}

func keysOf(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
