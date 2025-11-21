package craig

import (
	"encoding/json"
	"slices"

	"github.com/JoshPattman/agent"
	"github.com/JoshPattman/jpf"
)

type agentState uint8

const (
	reActState agentState = iota
	answerState
)

// Encodes the state into messages for the LLM.
// These messages will look very similar to the messages the LLM generates
// (however we parsed them and threw away the raw text),
// BUT they will be perfectly formatted, encoragin the LLM to do the same in the future.
// Thi can optionally add a message to the end which sets the agent into final answer mode.
type stateHistoryMessageEncoder struct {
	personality            string
	systemPrompt           string
	reactModePrefix        string
	finalAnswerModeMessage string
	state                  agentState
	tools                  []agent.Tool
	scenarios              map[string]agent.Scenario
}

func (enc *stateHistoryMessageEncoder) BuildInputMessages(state executingState) ([]jpf.Message, error) {
	messages := make([]jpf.Message, 0)
	// System prompt
	sys, err := enc.makeSysMessage()
	if err != nil {
		return nil, err
	}
	messages = append(messages, sys)
	// Previous tasks
	for _, group := range state.History {
		messages = append(messages, enc.makeBeginTaskMessage(group.Task))
		messages = append(messages, enc.makeMessagesForReActSteps(group.Steps)...)
		messages = append(messages, enc.makeAnswerTaskMessage())
		messages = append(messages, enc.makeTaskAnsweredMessage(group.Response))
	}
	// Current task
	messages = append(messages, enc.makeBeginTaskMessage(state.Active.Task))
	messages = append(messages, enc.makeMessagesForReActSteps(state.Active.Steps)...)
	if enc.state == answerState {
		messages = append(messages, enc.makeAnswerTaskMessage())
	}
	return messages, nil
}

func (enc *stateHistoryMessageEncoder) makeSysMessage() (jpf.Message, error) {
	scens := make([]systemPromptScenario, 0)
	for k, v := range enc.scenarios {
		scens = append(scens, systemPromptScenario{
			Key:      k,
			Scenario: v,
		})
	}
	slices.SortFunc(scens, func(scenA, scenB systemPromptScenario) int {
		if scenA.Key > scenB.Key {
			return 1
		} else if scenA.Key < scenB.Key {
			return -1
		} else {
			return 0
		}
	})
	sysPrompt, err := formatTemplate(enc.systemPrompt, systemPromptData{
		Personality: enc.personality,
		Tools:       enc.tools,
		Scenarios:   scens,
	})
	if err != nil {
		return jpf.Message{}, err
	}
	return jpf.Message{
		Role:    jpf.UserRole,
		Content: sysPrompt,
	}, nil
}

func (enc *stateHistoryMessageEncoder) makeBeginTaskMessage(task string) jpf.Message {
	return jpf.Message{
		Role:    jpf.UserRole,
		Content: enc.reactModePrefix + task,
	}
}
func (enc *stateHistoryMessageEncoder) makeAnswerTaskMessage() jpf.Message {
	return jpf.Message{
		Role:    jpf.UserRole,
		Content: enc.finalAnswerModeMessage,
	}
}
func (enc *stateHistoryMessageEncoder) makeTaskAnsweredMessage(answer string) jpf.Message {
	resp := answerResponse{
		Response: answer,
	}
	bs, _ := json.Marshal(resp)
	return jpf.Message{
		Role:    jpf.AssistantRole,
		Content: string(bs),
	}
}

func (enc *stateHistoryMessageEncoder) makeMessagesForReActSteps(steps []reActStep) []jpf.Message {
	messages := make([]jpf.Message, 0)
	for _, item := range steps {
		messages = append(
			messages,
			jpf.Message{
				Role:    jpf.AssistantRole,
				Content: formatStepForAIMessage(item),
			},
		)
		if len(item.ActionObservations) > 0 {
			messages = append(
				messages,
				jpf.Message{
					Role:    jpf.UserRole,
					Content: formatStepForUserMessage(item),
				},
			)
		}
	}
	return messages
}

func formatStepForAIMessage(item reActStep) string {
	resp := reActResponse{
		Reasoning: item.Reasoning,
	}
	for _, ao := range item.ActionObservations {
		resp.Actions = append(resp.Actions, ao.Action)
	}
	bs, _ := json.Marshal(resp)
	return string(bs)
}

func formatStepForUserMessage(item reActStep) string {
	var resps []agent.Observation
	for _, ao := range item.ActionObservations {
		resps = append(resps, ao.Observation)
	}
	bs, _ := json.Marshal(resps)
	return string(bs)
}
