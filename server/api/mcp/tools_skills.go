package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/achetronic/magec/server/store"
)

type listSkillsOutput struct {
	Skills []store.Skill `json:"skills"`
}

func (h *Handler) listSkills(_ context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, listSkillsOutput, error) {
	return nil, listSkillsOutput{Skills: h.store.ListRawSkills()}, nil
}

func (h *Handler) getSkill(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, store.Skill, error) {
	sk, ok := h.store.GetRawSkill(in.ID)
	if !ok {
		return nil, store.Skill{}, fmt.Errorf("get skill: %w", errValidation("not found: "+in.ID))
	}
	return nil, sk, nil
}

func (h *Handler) deleteSkill(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, emptyOutput, error) {
	if err := h.store.DeleteSkill(in.ID); err != nil {
		return nil, emptyOutput{}, fmt.Errorf("delete skill: %w", err)
	}
	return nil, okOutput, nil
}

func (h *Handler) registerSkillTools() {
	destructive := true
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_list_skills", Title: "List skills",
		Description: "List every registered skill package (id and slug only). Use the admin UI to read SKILL.md content.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.listSkills)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_get_skill", Title: "Get skill",
		Description: "Return one skill stub by id. Content lives on disk at data/skills/{slug}/.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.getSkill)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_delete_skill", Title: "Delete skill",
		Description: "Delete a skill by id (store record and the on-disk package).",
		Annotations: &sdk.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, h.deleteSkill)
	h.toolCount++
}
