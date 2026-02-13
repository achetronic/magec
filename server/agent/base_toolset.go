package agent

import (
	"google.golang.org/adk/agent"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/exitlooptool"
)

// baseToolset provides tools that are available to every agent regardless of
// configuration. Currently includes exit_loop so that any agent participating
// in a loop with escalate enabled can signal it wants to stop iterating.
type baseToolset struct {
	tools []tool.Tool
}

func newBaseToolset() (*baseToolset, error) {
	exitTool, err := exitlooptool.New()
	if err != nil {
		return nil, err
	}
	return &baseToolset{tools: []tool.Tool{exitTool}}, nil
}

func (b *baseToolset) Name() string {
	return "base_toolset"
}

func (b *baseToolset) Tools(_ agent.ReadonlyContext) ([]tool.Tool, error) {
	return b.tools, nil
}
