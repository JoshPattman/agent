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
	AgentName        string        `json:"agent_name"`
	AgentDescription []string      `json:"agent_description"`
	Personality      string        `json:"personality"`
	ModelName        string        `json:"model_name"`
	MCPServers       []string      `json:"mcp_servers"`
	SubAgents        []AgentConfig `json:"sub_agents"`
	ViewFiles        bool          `json:"view_files"`
}

type MCPServersConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcp_servers"`
}

type MCPServerConfig struct {
	Addr    string            `json:"addr"`
	Headers map[string]string `json:"headers"`
}
