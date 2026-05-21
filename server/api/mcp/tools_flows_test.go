package mcp

import (
	"context"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/achetronic/magec/server/store"
)

func TestCreateFlow_ValidatesRootStep(t *testing.T) {
	h := newTestHandler(t)
	ctx := context.Background()

	// Loop step with both exitLoop and exitWhen → admin validator rejects it.
	bad := store.FlowDefinition{
		Name: "bad",
		Root: store.FlowStep{
			Type:     store.FlowStepLoop,
			ExitLoop: true,
			ExitWhen: "state.done == true",
			Steps: []store.FlowStep{
				{Type: store.FlowStepAgent, AgentID: "agent-1"},
			},
		},
	}
	if _, _, err := h.createFlow(ctx, &sdk.CallToolRequest{}, createFlowInput{Flow: bad}); err == nil {
		t.Fatal("expected validation error for mutually exclusive exitLoop/exitWhen")
	}

	good := store.FlowDefinition{
		Name: "good",
		Root: store.FlowStep{
			Type: store.FlowStepSequential,
			Steps: []store.FlowStep{
				{Type: store.FlowStepAgent, AgentID: "agent-1"},
			},
		},
	}
	_, created, err := h.createFlow(ctx, &sdk.CallToolRequest{}, createFlowInput{Flow: good})
	if err != nil {
		t.Fatalf("create good flow: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty flow id")
	}
}
