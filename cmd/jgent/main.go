package main

import (
	"flag"
	"fmt"

	"github.com/JoshPattman/agent"
	"github.com/JoshPattman/agent/agentmcp"
)

func main() {
	// Parse flags
	agentFileName := flag.String("agent", "./agent.json", "the path of the agent config file to use")
	modelsFileName := flag.String("models", "./models.json", "the path of the models config file to use")
	flag.Parse()

	// Load config files
	modelsConf, err := LoadJSONFile[ModelsConfig](*modelsFileName)
	if err != nil {
		panic(err)
	}
	agentConf, err := LoadJSONFile[AgentConfig](*agentFileName)
	if err != nil {
		panic(err)
	}

	// Build the agent
	ab, err := BuildAgentBuilder(modelsConf, agentConf)
	if err != nil {
		panic(err)
	}

	// Interact
	interactionLoop(ab())
}

func BuildAgentBuilder(modelsConf ModelsConfig, agentConf AgentConfig) (func() agent.Agent, error) {
	// Get model builder
	model, ok := modelsConf.Models[agentConf.ModelName]
	if !ok {
		return nil, fmt.Errorf("could not find model for %s", model)
	}
	modelBuilder := &ModelBuilder{
		model.Key,
		model.Name,
		model.URL,
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
		ab, err := BuildAgentBuilder(modelsConf, ac)
		if err != nil {
			return nil, err
		}
		tools = append(tools, agent.AgentAsTool(ab, ac.AgentName, ac.AgentDescription))
	}

	// Build agent
	ab := func() agent.Agent {
		return agent.NewCombinedReActAgent(modelBuilder, agent.WithTools(tools...))
	}
	return ab, nil
}
