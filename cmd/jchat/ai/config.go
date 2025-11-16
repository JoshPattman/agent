package ai

type ModelsConfig struct {
	Models map[string]ModelConfig `json:"models"`
}

type ModelConfig struct {
	URL  string `json:"url"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

type AgentConfig struct {
	AgentName        string            `json:"agent_name"`
	AgentDescription []string          `json:"agent_description"`
	Personality      string            `json:"personality"`
	ModelName        string            `json:"model_name"`
	MCPServers       []MCPServerConfig `json:"mcp_servers"`
	SubAgents        []AgentConfig     `json:"sub_agents"`
	ViewFiles        bool              `json:"view_files"`
}

type MCPServerConfig struct {
	Addr    string            `json:"addr"`
	Headers map[string]string `json:"headers"`
}
