package ai

import (
	"fmt"

	"github.com/JoshPattman/agent"
	"github.com/JoshPattman/agent/agentmcp"
	"github.com/JoshPattman/agent/craig"
	"github.com/JoshPattman/jpf"
)

func BuildAgentBuilder(modelsConf ModelsConfig, agentConf AgentConfig, usageCounter *jpf.UsageCounter) (func() agent.Agent, error) {
	// Get model builder
	model, ok := modelsConf.Models[agentConf.ModelName]
	if !ok {
		return nil, fmt.Errorf("could not find model for %s", model)
	}
	modelBuilder := &ModelBuilder{
		model.Key,
		model.Name,
		model.URL,
		usageCounter,
	}

	// Create MCPtools
	tools := make([]agent.Tool, 0)
	for _, server := range agentConf.MCPServers {
		client, err := agentmcp.CreateClient(server.Addr, server.Headers)
		if err != nil {
			return nil, err
		}
		clientTools, err := agentmcp.CreateToolsFromMCP(client)
		if err != nil {
			return nil, err
		}
		tools = append(tools, clientTools...)
	}

	// Create built-in tools
	tools = append(tools, &timeTool{})

	// Create agent-as-tool tools
	for _, ac := range agentConf.SubAgents {
		ab, err := BuildAgentBuilder(modelsConf, ac, usageCounter)
		if err != nil {
			return nil, err
		}
		tools = append(tools, agent.AgentAsTool(ab, ac.AgentName, ac.AgentDescription))
	}

	// Build agent
	ab := func() agent.Agent {
		return craig.New(modelBuilder, craig.WithTools(tools...), craig.WithPersonality(agentConf.Personality))
	}
	return ab, nil
}
