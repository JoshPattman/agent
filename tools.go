package agent

import (
	"bytes"
	"encoding/json"
	"slices"
)

func AgentAsTool(agent *Agent, name string, description []string) Tool {
	return &agentAsTool{agent, name, description}
}

type agentAsTool struct {
	agent *Agent
	name  string
	desc  []string
}

// Call implements Tool.
func (a *agentAsTool) Call(args map[string]any) (string, error) {
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(args)
	if err != nil {
		return "", err
	}
	return a.agent.Answer(buf.String())
}

// Description implements Tool.
func (a *agentAsTool) Description() []string {
	return slices.Clone(a.desc)
}

// Name implements Tool.
func (a *agentAsTool) Name() string {
	return a.name
}
