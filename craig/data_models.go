package craig

import "github.com/JoshPattman/agent"

type ReActResponse struct {
	Reasoning string         `json:"reasoning"`
	Actions   []agent.Action `json:"actions"`
}

type AnswerResponse struct {
	Response string `json:"response"`
}

type ReActStep struct {
	Reasoning          string                    `json:"reasoning"`
	ActionObservations []agent.ActionObservation `json:"action_observations"`
}

type ExecutedTask struct {
	Task     string      `json:"task"`
	Steps    []ReActStep `json:"steps"`
	Response string      `json:"response"`
}

type ExecutingTask struct {
	Task  string      `json:"task"`
	Steps []ReActStep `json:"steps"`
}

type ExecutingState struct {
	History []ExecutedTask
	Active  ExecutingTask
}

type SystemPromptData struct {
	Tools []agent.Tool
}
