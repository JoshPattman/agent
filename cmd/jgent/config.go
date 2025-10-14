package main

import (
	"encoding/json"
	"os"
)

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
}

type MCPServerConfig struct {
	Addr    string            `json:"addr"`
	Headers map[string]string `json:"headers"`
}

func LoadJSONFile[T any](filePath string) (T, error) {
	var t T
	file, err := os.Open(filePath)
	if err != nil {
		return t, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&t); err != nil {
		return t, err
	}
	return t, nil
}
