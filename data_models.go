package agent

type ReActResponse struct {
	Reasoning string   `json:"reasoning"`
	Actions   []Action `json:"actions"`
}

type AnswerResponse struct {
	Response string `json:"response"`
}

type Action struct {
	Name string      `json:"name"`
	Args []ActionArg `json:"args"`
}
type ActionArg struct {
	ArgName string `json:"arg_name"`
	ArgData any    `json:"arg_data"`
}

type Observation struct {
	Observed string `json:"observed"`
}

type ActionObservation struct {
	Action      Action      `json:"action"`
	Observation Observation `json:"observation"`
}

type ReActStep struct {
	Reasoning          string              `json:"reasoning"`
	ActionObservations []ActionObservation `json:"action_observations"`
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

// A function which can be described to and called by an agent.
type Tool interface {
	// The name of the tool to be used by the agent.
	Name() string
	// Describe the tool in short bullet points to the agent.
	Description() []string
	// Call the tool, providing a formatted response or an error if the tool call failed.
	Call(map[string]any) (string, error)
}

type SystemPromptData struct {
	Tools []Tool
}
