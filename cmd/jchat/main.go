package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"jchat/ai"
	"jchat/ui"
	"os"
	"path/filepath"
	"time"

	"github.com/JoshPattman/agent"
	"github.com/JoshPattman/jpf"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	agentBuilder, agentConf, usageCounter, err := loadAndCreateAgentBuilder()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}
	chat := ui.NewChatPage(agentBuilder, ui.AgentSummary{
		Name:         agentConf.AgentName,
		Description:  agentConf.AgentDescription,
		NumMCP:       len(agentConf.MCPServers),
		NumSubAgents: len(agentConf.SubAgents),
		ModelName:    agentConf.ModelName,
	})
	p := tea.NewProgram(
		chat,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	go func() {
		for {
			p.Send(ui.UsageMessage{Usage: usageCounter.Get()})
			time.Sleep(time.Second / 2)
		}
	}()

	go p.Send(ui.SetConcurrentMessageSender{MsgSender: p.Send})

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}

var DefaultAgentConfig = ai.AgentConfig{
	AgentName: "craig",
	AgentDescription: []string{
		"A general purpose agent called craig",
		"Has no tool acsess",
	},
	Personality: "You are an agent called CRAIG",
	ModelName:   "default_model",
}

var DefaultModelsConfig = ai.ModelsConfig{
	Models: map[string]ai.ModelConfig{
		"default_model": {
			URL:  "",
			Name: "gpt-4.1",
			Key:  "Your API Key Here",
		},
	},
}

func loadAndCreateAgentBuilder() (func() (agent.Agent, error), ai.AgentConfig, *jpf.UsageCounter, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, ai.AgentConfig{}, nil, errors.Join(errors.New("could not load user home directory"), err)
	}
	dataPath := filepath.Join(homeDir, "jchat")
	agentFileName := filepath.Join(dataPath, "agent.json")
	modelsFileName := filepath.Join(dataPath, "models.json")

	// Load config files
	modelsConf, err := loadJSONFileButCreateIfNotExist(modelsFileName, DefaultModelsConfig)
	if err != nil {
		return nil, ai.AgentConfig{}, nil, err
	}
	agentConf, err := loadJSONFileButCreateIfNotExist(agentFileName, DefaultAgentConfig)
	if err != nil {
		return nil, ai.AgentConfig{}, nil, err
	}

	// Build the agent
	usageCounter := jpf.NewUsageCounter()
	builder, err := ai.BuildAgentBuilder(modelsConf, agentConf, usageCounter)
	if err != nil {
		return nil, ai.AgentConfig{}, nil, err
	}
	return func() (a agent.Agent, err error) { return builder(), nil }, agentConf, usageCounter, nil
}

func loadJSONFileButCreateIfNotExist[T any](filePath string, defaultVal T) (T, error) {
	result, err := loadJSONFile[T](filePath)
	if errors.Is(err, os.ErrNotExist) {
		folder := filepath.Dir(filePath)
		err := os.MkdirAll(folder, os.ModePerm)
		if err != nil {
			return *new(T), err
		}
		file, err := os.Create(filePath)
		if err != nil {
			return *new(T), err
		}
		defer file.Close()
		enc := json.NewEncoder(file)
		enc.SetIndent("", "    ")
		err = enc.Encode(defaultVal)
		if err != nil {
			return *new(T), err
		}
		return defaultVal, nil
	} else if err != nil {
		return *new(T), err
	} else {
		return result, nil
	}
}

func loadJSONFile[T any](filePath string) (T, error) {
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
