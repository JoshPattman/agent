package ai

import "github.com/JoshPattman/agent"

type ModelsConfig struct {
	Models map[string]ModelConfig `json:"models"`
}

type ModelConfig struct {
	URL     string            `json:"url"`
	Name    string            `json:"name"`
	Key     string            `json:"key"`
	Headers map[string]string `json:"headers"`
}

type AgentsConfig struct {
	Agents map[string]AgentConfig `json:"agents"`
}

type AgentConfig struct {
	AgentDescription []string                  `json:"agent_description"`
	Personality      string                    `json:"personality"`
	Scenarios        map[string]agent.Scenario `json:"scenarios"`
	ModelName        string                    `json:"model_name"`
	MCPServers       []string                  `json:"mcp_servers"`
	SubAgents        []string                  `json:"sub_agents"`
	ViewFiles        bool                      `json:"view_files"`
	QuestionFiles    bool                      `json:"question_files"`
	RunCommands      bool                      `json:"run_commands"`
}

type MCPServersConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcp_servers"`
}

type MCPServerConfig struct {
	Addr    string            `json:"addr"`
	Headers map[string]string `json:"headers"`
}
