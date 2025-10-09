package agent

import "github.com/JoshPattman/jpf"

// Specifies an object that can be called to create a model for the agent to use.
type AgentModelBuilder interface {
	// Create a model with a structured output schema for the object provided.
	BuildAgentModel(responseType any) jpf.Model
}
