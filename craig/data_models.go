package craig

import "github.com/JoshPattman/agent"

type reActResponse struct {
	Reasoning string         `json:"reasoning"`
	Actions   []agent.Action `json:"actions"`
}

type answerResponse struct {
	Response string `json:"response"`
}

type reActStep struct {
	Reasoning          string                    `json:"reasoning"`
	ActionObservations []agent.ActionObservation `json:"action_observations"`
}

type executedTask struct {
	Task     string      `json:"task"`
	Steps    []reActStep `json:"steps"`
	Response string      `json:"response"`
}

type executingTask struct {
	Task  string      `json:"task"`
	Steps []reActStep `json:"steps"`
}

type executingState struct {
	History []executedTask
	Active  executingTask
}

type systemPromptData struct {
	Tools []agent.Tool
}
