package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/achetronic/magec/server/store"
)

type listCommandsOutput struct {
	Commands []store.Command `json:"commands"`
}

func (h *Handler) listCommands(_ context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, listCommandsOutput, error) {
	return nil, listCommandsOutput{Commands: h.store.ListRawCommands()}, nil
}

func (h *Handler) getCommand(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, store.Command, error) {
	c, ok := h.store.GetRawCommand(in.ID)
	if !ok {
		return nil, store.Command{}, fmt.Errorf("get command: %w", errValidation("not found: "+in.ID))
	}
	return nil, c, nil
}

type createCommandInput struct {
	Command store.Command `json:"command" jsonschema:"command definition"`
}

func (h *Handler) createCommand(_ context.Context, _ *sdk.CallToolRequest, in createCommandInput) (*sdk.CallToolResult, store.Command, error) {
	if in.Command.Name == "" {
		return nil, store.Command{}, fmt.Errorf("create command: %w", errValidation("name is required"))
	}
	if in.Command.Prompt == "" {
		return nil, store.Command{}, fmt.Errorf("create command: %w", errValidation("prompt is required"))
	}
	created, err := h.store.CreateCommand(in.Command)
	if err != nil {
		return nil, store.Command{}, fmt.Errorf("create command: %w", err)
	}
	return nil, created, nil
}

type updateCommandInput struct {
	ID      string        `json:"id" jsonschema:"command id"`
	Command store.Command `json:"command" jsonschema:"new command definition"`
}

func (h *Handler) updateCommand(_ context.Context, _ *sdk.CallToolRequest, in updateCommandInput) (*sdk.CallToolResult, store.Command, error) {
	if err := h.store.UpdateCommand(in.ID, in.Command); err != nil {
		return nil, store.Command{}, fmt.Errorf("update command: %w", err)
	}
	updated, _ := h.store.GetRawCommand(in.ID)
	return nil, updated, nil
}

func (h *Handler) deleteCommand(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, emptyOutput, error) {
	if err := h.store.DeleteCommand(in.ID); err != nil {
		return nil, emptyOutput{}, fmt.Errorf("delete command: %w", err)
	}
	return nil, okOutput, nil
}

func (h *Handler) registerCommandTools() {
	destructive := true
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_list_commands", Title: "List commands",
		Description: "List reusable prompts that can be invoked via cron or webhook clients.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.listCommands)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_get_command", Title: "Get command",
		Description: "Return one command by id.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.getCommand)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_create_command", Title: "Create command",
		Description: "Create a new reusable command.",
	}, h.createCommand)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_update_command", Title: "Update command",
		Description: "Replace the command identified by id.",
		Annotations: &sdk.ToolAnnotations{IdempotentHint: true},
	}, h.updateCommand)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_delete_command", Title: "Delete command",
		Description: "Delete a command by id.",
		Annotations: &sdk.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, h.deleteCommand)
	h.toolCount++
}
