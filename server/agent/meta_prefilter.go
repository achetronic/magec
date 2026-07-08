// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"encoding/json"
	"regexp"
	"strings"

	adkagent "google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"

	toolsflowstate "github.com/achetronic/magec/server/agent/tools/flowstate"
)

// metaPrefilterNodeName is the synthetic node the builder wires between Start
// and the entry node. The double-underscore prefix is reserved for internal
// nodes by graph validation, so it can never collide with an operator node.
const metaPrefilterNodeName = "__meta__"

// metaStateKey is the flow-state key (without the flow: prefix) where the
// extracted client metadata lands, readable downstream as state.magec_meta.
const metaStateKey = "magec_meta"

// metaBlockRE matches the metadata block clients prepend to the user message:
// an HTML comment wrapping a JSON object, plus any trailing newline.
var metaBlockRE = regexp.MustCompile(`<!--MAGEC_META:(.*?):MAGEC_META-->\n?`)

// buildMetaPrefilterNode returns the synthetic node that strips the client
// metadata block from the raw input before it reaches the entry node, parking
// the parsed metadata in flow state instead. Flow nodes and agents downstream
// therefore receive only the user's words, and the metadata stays available
// as state.magec_meta (e.g. state.magec_meta.source in a router guard).
func buildMetaPrefilterNode() workflow.Node {
	fn := func(ctx adkagent.Context, input any, emit func(*session.Event) error) (any, error) {
		text, ok := input.(string)
		cleaned, meta := input, map[string]any(nil)
		if ok {
			cleanedText, parsed := extractMagecMeta(text)
			cleaned = cleanedText
			meta = parsed
		}

		ev := session.NewEvent(ctx, ctx.InvocationID())
		ev.Author = metaPrefilterNodeName
		ev.Output = cleaned
		if meta != nil {
			ev.Actions.StateDelta[toolsflowstate.StateKeyPrefix+metaStateKey] = meta
		}
		if err := emit(ev); err != nil {
			return nil, err
		}
		return nil, nil
	}
	return workflow.NewEmittingFunctionNode[any, any](metaPrefilterNodeName, fn, workflow.NodeConfig{})
}

// extractMagecMeta removes every metadata block from the text and returns the
// cleaned message plus the parsed metadata of the first block, or nil when the
// text carries none. A block whose JSON does not parse is still stripped, with
// the raw payload preserved under a "raw" key rather than silently dropped.
func extractMagecMeta(text string) (string, map[string]any) {
	matches := metaBlockRE.FindStringSubmatch(text)
	if matches == nil {
		return text, nil
	}

	meta := map[string]any{}
	if err := json.Unmarshal([]byte(matches[1]), &meta); err != nil {
		meta = map[string]any{"raw": matches[1]}
	}

	cleaned := metaBlockRE.ReplaceAllString(text, "")
	return strings.TrimSpace(cleaned), meta
}
