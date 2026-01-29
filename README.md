> For a better agent framework, take a look at [react](github.com/JoshPattman/react)

# Agent

A simple and flexible ReAct (Reasoning and Acting) agent for Go. This library implements the ReAct pattern, allowing AI agents to reason about problems and take actions using tools to solve complex tasks.

## Features

- **ReAct Pattern Implementation**: Agents can reason about problems and take actions in a structured loop
- **Tool System**: Easy-to-implement tool interface for extending agent capabilities
- **Concurrent Tool Execution**: Tools run in parallel for better performance
- **Customizable System Prompts**: Template-based system prompts with tool descriptions
- **Callback Support**: Hooks for monitoring agent reasoning and actions
- **Task History**: Built-in conversation and task history management

## Instalation
- For the agent package (for developers): `go get github.com/JoshPattman/agent`
- For the agent TUI (interactive way to chat with an agent): `go install github.com/JoshPattman/agent/cmd/jchat@latest`

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/JoshPattman/agent"
)

func main() {
    // Create a model builder (you'll need to implement this)
    builder := &YourModelBuilder{}
    
    // Create an agent with tools
    a := agent.NewAgent(
        builder,
        agent.WithTools(&yourTool1{}, &yourTool2{}),
        agent.WithSystemPromptTemplate("You are a helpful assistant..."),
    )
    
    // Ask the agent a question
    answer, err := a.Answer("What's the weather like today?")
    if err != nil {
        panic(err)
    }
    
    fmt.Println(answer)
}
```

## Core Concepts

### ReAct Pattern

The ReAct pattern combines reasoning and acting in a loop:

1. **Reason**: The agent thinks about what to do next
2. **Act**: The agent calls appropriate tools
3. **Observe**: The agent processes tool results
4. **Repeat**: Continue until the task is complete

### Tools

Tools are the actions an agent can take. Implement the `Tool` interface:

```go
// A function which can be described to and called by an agent.
type Tool interface {
	// The name of the tool to be used by the agent.
	Name() string
	// Describe the tool in short bullet points to the agent.
	Description() []string
	// Call the tool, providing a formatted response or an error if the tool call failed.
	Call(map[string]any) (string, error)
}
```

### Agent Options

Configure your agent with various options (when not specified, sensible defaults will be used):

- `WithTools(...Tool)`: Add tools for the agent to use
- `WithSystemPromptTemplate(string)`: Customize the system prompt
- `WithTaskPrefix(string)`: Change a prefix for starting a task
- `WithFinalAnswerMessage(string)`: Change the message to tell the agent to create a final answer

## Model Builder Interface

The agent requires a model builder to interface with language models. The interface creates a [jpf](github.com/JoshPattman/jpf) model, which allows you to easily compose retry logic, caching, and other useful features. Implement the `AgentModelBuilder` interface:

```go
// Specifies an object that can be called to create a model for the agent to use.
type AgentModelBuilder interface {
	// Create a model with a structured output schema for the object provided.
	BuildAgentModel(responseType any) jpf.Model
}
```

The model builder is responsible for:
- Configuring the language model (API keys, model selection, etc.)
- Creating appropriate prompts for reasoning vs. answering
- Handling the actual API calls to your chosen language model
- Managing conversation context and tool descriptions


## Callbacks

Monitor agent behavior with callbacks:

```go
agent.SetOnReActInitCallback(func(response ReActResponse) {
    fmt.Printf("Agent is has reasoned and called tools: %s\n", response)
})

agent.SetOnReActCompleteCallback(func(step ReActStep) {
    fmt.Printf("%d tool calls have been returned to agent\n", len(step.ActionObservations))
})
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
