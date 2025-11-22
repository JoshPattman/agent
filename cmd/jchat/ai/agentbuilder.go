package ai

import (
	"fmt"
	"strings"

	"github.com/JoshPattman/agent"
	"github.com/JoshPattman/agent/agentmcp"
	"github.com/JoshPattman/agent/craig"
	"github.com/JoshPattman/jpf"
)

func BuildAgentBuilder(activeAgentName string, modelsConf ModelsConfig, agentsConf AgentsConfig, mcpsConf MCPServersConfig, usageCounter *jpf.UsageCounter) (func() agent.Agent, error) {
	agentConf, ok := agentsConf.Agents[activeAgentName]
	if !ok {
		return nil, fmt.Errorf("could not find a configured agent called '%s'", activeAgentName)
	}
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
		model.Headers,
	}

	// Create MCPtools
	tools := make([]agent.Tool, 0)
	for _, serverName := range agentConf.MCPServers {
		server, ok := mcpsConf.MCPServers[serverName]
		if !ok {
			return nil, fmt.Errorf("could not find mcp server '%s'", serverName)
		}
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
	tools = append(tools, agent.NewTimeTool())
	if agentConf.ViewFiles {
		tools = append(tools, agent.NewListDirectoryTool())
		tools = append(tools, agent.NewReadFileTool())
	}
	if agentConf.QuestionFiles {
		tools = append(tools, NewFileQATool(modelBuilder))
	}

	// Create agent-as-tool tools
	for _, ac := range agentConf.SubAgents {
		ab, err := BuildAgentBuilder(ac, modelsConf, agentsConf, mcpsConf, usageCounter)
		if err != nil {
			return nil, err
		}
		subAgentConfig := agentsConf.Agents[ac] // This is safe to not check as we already checked above recursively
		tools = append(tools, agent.NewAgentQuickQuestionTool(ab, ac, strings.Join(subAgentConfig.AgentDescription, ". ")))
	}

	// Build agent
	ab := func() agent.Agent {
		return craig.New(
			modelBuilder,
			craig.WithTools(tools...),
			craig.WithPersonality(agentConf.Personality),
			craig.WithScenarios(agentConf.Scenarios),
		)
	}
	return ab, nil
}
