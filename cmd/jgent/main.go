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
	a, err := BuildAgent(modelsConf, agentConf)
	if err != nil {
		panic(err)
	}

	// Interact
	interactionLoop(a)
}

func BuildAgent(modelsConf ModelsConfig, agentConf AgentConfig) (*agent.Agent, error) {
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
	// Create tools
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
	tools = append(tools, &timeTool{})
	// Build agent
	return agent.NewAgent(modelBuilder, agent.WithTools(tools...)), nil
}
