// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package admin

import (
	"strings"
	"testing"

	"github.com/achetronic/magec/server/store"
)

func TestValidateFlowStep_LoopWithMaxIterationsOnly(t *testing.T) {
	step := store.FlowStep{
		Type:          store.FlowStepLoop,
		MaxIterations: 5,
		Steps: []store.FlowStep{
			{Type: store.FlowStepAgent, AgentID: "agent-1"},
		},
	}
	if err := validateFlowStep(&step); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateFlowStep_LoopWithExitLoop(t *testing.T) {
	step := store.FlowStep{
		Type:     store.FlowStepLoop,
		ExitLoop: true,
		Steps: []store.FlowStep{
			{Type: store.FlowStepAgent, AgentID: "agent-1"},
		},
	}
	if err := validateFlowStep(&step); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateFlowStep_LoopWithValidExitWhen(t *testing.T) {
	step := store.FlowStep{
		Type:     store.FlowStepLoop,
		ExitWhen: `state.approved == true`,
		Steps: []store.FlowStep{
			{Type: store.FlowStepAgent, AgentID: "agent-1"},
		},
	}
	if err := validateFlowStep(&step); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateFlowStep_LoopExclusion(t *testing.T) {
	step := store.FlowStep{
		Type:     store.FlowStepLoop,
		ExitLoop: true,
		ExitWhen: `state.approved == true`,
		Steps: []store.FlowStep{
			{Type: store.FlowStepAgent, AgentID: "agent-1"},
		},
	}
	err := validateFlowStep(&step)
	if err == nil {
		t.Fatal("expected exclusion error")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("error should mention mutual exclusion, got %v", err)
	}
}

func TestValidateFlowStep_LoopRejectsBadCEL(t *testing.T) {
	step := store.FlowStep{
		Type:     store.FlowStepLoop,
		ExitWhen: `state.x ===`,
		Steps: []store.FlowStep{
			{Type: store.FlowStepAgent, AgentID: "agent-1"},
		},
	}
	if err := validateFlowStep(&step); err == nil {
		t.Fatal("expected CEL parse error")
	}
}

func TestValidateFlowStep_LoopRejectsNonBoolCEL(t *testing.T) {
	step := store.FlowStep{
		Type:     store.FlowStepLoop,
		ExitWhen: `state.score + 1`,
		Steps: []store.FlowStep{
			{Type: store.FlowStepAgent, AgentID: "agent-1"},
		},
	}
	if err := validateFlowStep(&step); err == nil {
		t.Fatal("expected non-bool error")
	}
}

func TestValidateFlowStep_NonLoopRejectsExitLoop(t *testing.T) {
	step := store.FlowStep{
		Type:     store.FlowStepSequential,
		ExitLoop: true,
		Steps: []store.FlowStep{
			{Type: store.FlowStepAgent, AgentID: "agent-1"},
		},
	}
	err := validateFlowStep(&step)
	if err == nil {
		t.Fatal("expected error: exitLoop only valid on loop steps")
	}
	if !strings.Contains(err.Error(), "loop") {
		t.Fatalf("error should mention loop, got %v", err)
	}
}

func TestValidateFlowStep_NonLoopRejectsExitWhen(t *testing.T) {
	step := store.FlowStep{
		Type:     store.FlowStepParallel,
		ExitWhen: `state.x == true`,
		Steps: []store.FlowStep{
			{Type: store.FlowStepAgent, AgentID: "agent-1"},
		},
	}
	err := validateFlowStep(&step)
	if err == nil {
		t.Fatal("expected error: exitWhen only valid on loop steps")
	}
	if !strings.Contains(err.Error(), "loop") {
		t.Fatalf("error should mention loop, got %v", err)
	}
}

func TestValidateFlowStep_NestedExclusionInChildLoop(t *testing.T) {
	step := store.FlowStep{
		Type: store.FlowStepSequential,
		Steps: []store.FlowStep{
			{
				Type:     store.FlowStepLoop,
				ExitLoop: true,
				ExitWhen: `state.x == true`,
				Steps: []store.FlowStep{
					{Type: store.FlowStepAgent, AgentID: "agent-1"},
				},
			},
		},
	}
	if err := validateFlowStep(&step); err == nil {
		t.Fatal("expected validation to recurse into nested loop and fail")
	}
}
