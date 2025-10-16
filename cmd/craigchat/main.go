package main

import (
	"craigchat/ai"
	"craigchat/ui"
	"encoding/json"
	"flag"
	"fmt"
	"os"
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

func loadAndCreateAgentBuilder() (func() (agent.Agent, error), ai.AgentConfig, *jpf.UsageCounter, error) {
	// Parse flags
	agentFileName := flag.String("agent", "./agent.json", "the path of the agent config file to use")
	modelsFileName := flag.String("models", "./models.json", "the path of the models config file to use")
	flag.Parse()

	// Load config files
	modelsConf, err := loadJSONFile[ai.ModelsConfig](*modelsFileName)
	if err != nil {
		return nil, ai.AgentConfig{}, nil, err
	}
	agentConf, err := loadJSONFile[ai.AgentConfig](*agentFileName)
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
