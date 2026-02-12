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
func BuildFlowAgent(flow store.FlowDefinition, agentMap map[string]adkagent.Agent) (adkagent.Agent, error) {
	return buildStep(flow.Name, &flow.Root, agentMap, 0)
}

func buildStep(name string, step *store.FlowStep, agentMap map[string]adkagent.Agent, depth int) (adkagent.Agent, error) {
	stepName := fmt.Sprintf("%s_%d", name, depth)

	switch step.Type {
	case store.FlowStepAgent:
		a, ok := agentMap[step.AgentID]
		if !ok {
			return nil, fmt.Errorf("agent %q not found in agent map", step.AgentID)
		}
		return a, nil

	case store.FlowStepSequential:
		children, err := buildChildren(name, step.Steps, agentMap, depth)
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
		children, err := buildChildren(name, step.Steps, agentMap, depth)
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
		children, err := buildChildren(name, step.Steps, agentMap, depth)
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

func buildChildren(name string, steps []store.FlowStep, agentMap map[string]adkagent.Agent, depth int) ([]adkagent.Agent, error) {
	children := make([]adkagent.Agent, 0, len(steps))
	for i := range steps {
		child, err := buildStep(name, &steps[i], agentMap, depth+1+i)
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}
	return children, nil
}
