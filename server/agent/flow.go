package agent

import (
	"fmt"

	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/workflowagents/loopagent"
	"google.golang.org/adk/agent/workflowagents/parallelagent"
	"google.golang.org/adk/agent/workflowagents/sequentialagent"

	"github.com/achetronic/magec/server/store"
)

// BuildFlowAgent recursively translates a FlowDefinition into an ADK agent tree.
// The agentMap must contain pre-built ADK agents keyed by their store ID.
// The root step uses the flow ID as its ADK agent name so flows are addressable
// by ID, consistent with how individual agents are addressed.
func BuildFlowAgent(flow store.FlowDefinition, agentMap map[string]adkagent.Agent) (adkagent.Agent, error) {
	return buildStep(flow.ID, &flow.Root, agentMap, "")
}

func buildStep(flowID string, step *store.FlowStep, agentMap map[string]adkagent.Agent, path string) (adkagent.Agent, error) {
	stepName := flowID
	if path != "" {
		stepName = fmt.Sprintf("%s_%s", flowID, path)
	}

	switch step.Type {
	case store.FlowStepAgent:
		a, ok := agentMap[step.AgentID]
		if !ok {
			return nil, fmt.Errorf("agent %q not found in agent map", step.AgentID)
		}
		return a, nil

	case store.FlowStepSequential:
		children, err := buildChildren(flowID, step.Steps, agentMap, path)
		if err != nil {
			return nil, err
		}
		return sequentialagent.New(sequentialagent.Config{
			AgentConfig: adkagent.Config{
				Name:      stepName,
				SubAgents: children,
			},
		})

	case store.FlowStepParallel:
		children, err := buildChildren(flowID, step.Steps, agentMap, path)
		if err != nil {
			return nil, err
		}
		return parallelagent.New(parallelagent.Config{
			AgentConfig: adkagent.Config{
				Name:      stepName,
				SubAgents: children,
			},
		})

	case store.FlowStepLoop:
		children, err := buildChildren(flowID, step.Steps, agentMap, path)
		if err != nil {
			return nil, err
		}
		return loopagent.New(loopagent.Config{
			AgentConfig: adkagent.Config{
				Name:      stepName,
				SubAgents: children,
			},
			MaxIterations: step.MaxIterations,
		})

	default:
		return nil, fmt.Errorf("unknown flow step type %q", step.Type)
	}
}

func buildChildren(flowID string, steps []store.FlowStep, agentMap map[string]adkagent.Agent, parentPath string) ([]adkagent.Agent, error) {
	children := make([]adkagent.Agent, 0, len(steps))
	for i := range steps {
		childPath := fmt.Sprintf("%d", i)
		if parentPath != "" {
			childPath = fmt.Sprintf("%s_%d", parentPath, i)
		}
		child, err := buildStep(flowID, &steps[i], agentMap, childPath)
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}
	return children, nil
}
