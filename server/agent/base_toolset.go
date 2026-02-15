package agent

import (
	"google.golang.org/adk/agent"
	"google.golang.org/adk/tool"
)

// baseToolset provides tools that are available to every agent regardless of
// configuration.
//
// TODO: Explore injecting exit_loop only to agents inside a loopagent (option 3).
// This would require cloning agents when building flow steps so the same agent
// definition can participate in a loop (with exit_loop) and outside one (without).
type baseToolset struct {
	tools []tool.Tool
}

func newBaseToolset() (*baseToolset, error) {
	return &baseToolset{tools: []tool.Tool{}}, nil
}

func (b *baseToolset) Name() string {
	return "base_toolset"
}

func (b *baseToolset) Tools(_ agent.ReadonlyContext) ([]tool.Tool, error) {
	return b.tools, nil
}
