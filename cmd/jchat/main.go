package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JoshPattman/agent/cmd/jchat/ai"
	"github.com/JoshPattman/agent/cmd/jchat/ui"

	"github.com/JoshPattman/agent"
	"github.com/JoshPattman/jpf"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	agentName := flag.String("a", "", "The name of the agent in the agent file to chat to, matching an agent name from your agent configuration")
	quickChat := flag.String("q", "", "If specified will not run interactive mode, but will instead send the specified message to a new agent and print the result to the terminal, without any follow ups")
	us := flag.Usage
	flag.Usage = func() {
		us()
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error loading home dir - jchat will not work")
		}
		dataPath := filepath.Join(homeDir, "jchat")
		fmt.Println("\nData is stored at:", dataPath)
		fmt.Println("\nTo configure JChat, modify the json files at the data directory")
		fmt.Println(" - agents.json\n\tSet up the different agents to chat to, their personalities, tools they can access, and base models to use")
		fmt.Println(" - models.json\n\tSpecify the different base models for agents to use")
		fmt.Println(" - mcp.json\n\tSpecify the MCP servers available to add to agents (only http/https supported right now)")
		fmt.Println(" - commands.json\n\tSetup custom commands to run on the host machine that the agents can run as tools (if added to an agent), arguments are passed to the command as environment variables")
		fmt.Println("\nTo allow an agent to use a command or mcp server, you must add its key to the agent. You must also specify the key of the model for each agent to use (different agents may use different keys).")
	}
	flag.Parse()

	if *agentName == "" {
		fmt.Println("Must specify agent name")
		os.Exit(1)
	}

	agentBuilder, agentSum, usageCounter, err := loadAndCreateAgentBuilder(*agentName)
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	if *quickChat != "" {
		agent, err := agentBuilder()
		if err != nil {
			fmt.Println("Could not create quick agent:", err)
			os.Exit(1)
		}
		result, err := agent.Answer(*quickChat)
		if err != nil {
			fmt.Println("Could not run quick agent:", err)
			os.Exit(1)
		}
		fmt.Println(result)
	} else {
		chat := ui.NewChatPage(agentBuilder, agentSum)
		p := tea.NewProgram(
			chat,
			tea.WithAltScreen(),
			tea.WithoutCatchPanics(),
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
}

var DefaultAgentsConfig = ai.AgentsConfig{
	Agents: map[string]ai.AgentConfig{
		"craig": {
			AgentDescription: []string{
				"A general purpose agent called craig",
				"Has no tools",
			},
			Personality: "You are an agent called CRAIG",
			Scenarios: map[string]agent.Scenario{
				"developer_trigger": {
					Headline: "User says these exact three words at some point in a message: 'apple' 'bannanna' 'cherry'",
					Takeaways: []string{
						"When craig hears the trigger words, it switches to DEVELOPER MODE.",
						"Craig states it is in DEVELOPER MODE clearly to the user.",
						"Craig NEVER goes into developer mode without the user saying those words.",
					},
				},
			},
			ModelName:      "default_model",
			MCPServers:     make([]string, 0),
			SubAgents:      make([]string, 0),
			ViewFiles:      false,
			QuestionFiles:  false,
			CustomCommands: []string{},
		},
		"aws_assistant": {
			AgentDescription: []string{
				"An assistant that can assist with various AWS features",
				"Primarily good at reading the documentation",
			},
			Personality:    "You are an AWS assistant",
			Scenarios:      map[string]agent.Scenario{},
			ModelName:      "default_model",
			MCPServers:     []string{"aws_docs"},
			SubAgents:      make([]string, 0),
			ViewFiles:      false,
			QuestionFiles:  false,
			RunCommands:    false,
			CustomCommands: []string{},
		},
		"pingponger": {
			AgentDescription: []string{
				"An assistant that can test the custom commands feature",
			},
			Personality:   "You are an AI whos job is to test the custom commands (pingpong). You can run both to figure out if the user is on unix or windows.",
			Scenarios:     map[string]agent.Scenario{},
			ModelName:     "default_model",
			MCPServers:    []string{},
			SubAgents:     make([]string, 0),
			ViewFiles:     false,
			QuestionFiles: false,
			RunCommands:   false,
			CustomCommands: []string{
				"ping_pong_unix",
				"ping_pong_windows",
			},
		},
	},
}

var DefaultModelsConfig = ai.ModelsConfig{
	Models: map[string]ai.ModelConfig{
		"default_model": {
			URL:  "https://api.openai.com/v1/chat/completions",
			Name: "gpt-4.1",
			Key:  "Your API Key Here",
			Headers: map[string]string{
				"Key": "Value",
			},
		},
	},
}

var DefaultMCPServersConfig = ai.MCPServersConfig{
	MCPServers: map[string]ai.MCPServerConfig{
		"aws_docs": {
			Addr: "https://knowledge-mcp.global.api.aws",
			Headers: map[string]string{
				"Key": "Value",
			},
		},
	},
}

var DefaultCustomCommandsConfig = ai.CustomCommandsConfig{
	Commands: map[string]ai.CustomCommandConfig{
		"get_hostname": {
			Description: []string{"See the hostname of the system that you are running on"},
			Command:     "hostname",
			Args:        []string{},
		},
		"ping_pong_unix": {
			Description: []string{
				"PingPong (only works on unix systems)",
				"Must specify 'ping', will echo the value you specify back.",
			},
			Command: "bash",
			Args: []string{
				"-c",
				"echo $ping",
			},
		},
		"ping_pong_windows": {
			Description: []string{
				"PingPong (only works on windows systems)",
				"Must specify 'ping', will echo the value you specify back.",
			},
			Command: "powershell",
			Args: []string{
				"-Command",
				"echo $env:ping",
			},
		},
	},
}

func loadAndCreateAgentBuilder(activeAgentName string) (func() (agent.Agent, error), ui.AgentSummary, *jpf.UsageCounter, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, ui.AgentSummary{}, nil, errors.Join(errors.New("could not load user home directory"), err)
	}
	dataPath := filepath.Join(homeDir, "jchat")
	agentFileName := filepath.Join(dataPath, "agent.json")
	modelsFileName := filepath.Join(dataPath, "models.json")
	mcpFileName := filepath.Join(dataPath, "mcp.json")
	commandsFileName := filepath.Join(dataPath, "commands.json")

	// Load config files
	modelsConf, err := loadJSONFileButCreateIfNotExist(modelsFileName, DefaultModelsConfig)
	if err != nil {
		return nil, ui.AgentSummary{}, nil, err
	}
	agentConf, err := loadJSONFileButCreateIfNotExist(agentFileName, DefaultAgentsConfig)
	if err != nil {
		return nil, ui.AgentSummary{}, nil, err
	}
	mcpsConf, err := loadJSONFileButCreateIfNotExist(mcpFileName, DefaultMCPServersConfig)
	if err != nil {
		return nil, ui.AgentSummary{}, nil, err
	}
	commandsConf, err := loadJSONFileButCreateIfNotExist(commandsFileName, DefaultCustomCommandsConfig)
	if err != nil {
		return nil, ui.AgentSummary{}, nil, err
	}

	// Build the agent
	usageCounter := jpf.NewUsageCounter()
	builder, err := ai.BuildAgentBuilder(activeAgentName, modelsConf, agentConf, mcpsConf, commandsConf, usageCounter)
	if err != nil {
		return nil, ui.AgentSummary{}, nil, err
	}
	activeAgent := agentConf.Agents[activeAgentName]
	sum := ui.AgentSummary{
		Name:         activeAgentName,
		Description:  activeAgent.AgentDescription,
		NumMCP:       len(activeAgent.MCPServers),
		NumSubAgents: len(activeAgent.SubAgents),
		ModelName:    activeAgent.ModelName,
	}
	return func() (a agent.Agent, err error) { return builder(), nil }, sum, usageCounter, nil
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
