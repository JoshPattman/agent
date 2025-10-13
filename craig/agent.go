// craig implements a Combined ReAct Intelligent aGent, where the agent reasons and acts in a single message.
package craig

import (
	"fmt"
	"sync"

	"github.com/JoshPattman/agent"
)

func NewCombinedReActAgent(modelBuilder agent.AgentModelBuilder, opts ...CombinedReActAgentOpt) agent.Agent {
	params := &agentParams{
		systemPrompt:       defaultSystemPrompt,
		taskPrefix:         defaultReActModePrefix,
		finalAnswerMessage: defaultAnswerModeContent,
	}
	for _, o := range opts {
		o(params)
	}
	return &combineReActAgent{
		reActStepper: newReActStepper(
			modelBuilder,
			params.tools,
			params.systemPrompt,
			params.taskPrefix,
			params.finalAnswerMessage,
		),
		answerStepper: newAnswerStepper(
			modelBuilder,
			params.tools,
			params.systemPrompt,
			params.taskPrefix,
			params.finalAnswerMessage,
		),
		tools: params.tools,
	}
}

type agentParams struct {
	systemPrompt       string
	taskPrefix         string
	finalAnswerMessage string
	tools              []agent.Tool
}

type CombinedReActAgentOpt func(*agentParams)

func WithSystemPromptTemplate(tpl string) CombinedReActAgentOpt {
	return func(a *agentParams) {
		a.systemPrompt = tpl
	}
}

func WithTools(tools ...agent.Tool) CombinedReActAgentOpt {
	return func(a *agentParams) {
		a.tools = append(a.tools, tools...)
	}
}

func WithTaskPrefix(prefix string) CombinedReActAgentOpt {
	return func(a *agentParams) {
		a.taskPrefix = prefix
	}
}

func WithFinalAnswerMessage(msg string) CombinedReActAgentOpt {
	return func(a *agentParams) {
		a.finalAnswerMessage = msg
	}
}

type combineReActAgent struct {
	history         []ExecutedTask
	reActStepper    reActStepper
	answerStepper   responseStepper
	tools           []agent.Tool
	onReActInit     func(string, []agent.Action)
	onReActComplete func(string, []agent.ActionObservation)
}

func (a *combineReActAgent) Answer(query string) (string, error) {
	state := newTaskState(query, a.history)
	// Do reasoning and acting loop
	for {
		newState, ok, err := a.stepTaskState(state)
		if err != nil {
			return "", err
		}
		state = newState
		if !ok {
			break
		}
	}
	// Finalise output
	finalResponse, _, err := a.answerStepper.Call(state)
	if err != nil {
		return "", err
	}
	a.history = append(a.history, ExecutedTask{
		Task:     query,
		Steps:    state.Active.Steps,
		Response: finalResponse.Response,
	})
	return finalResponse.Response, nil
}

func (a *combineReActAgent) SetOnReActInitCallback(callback func(reasoning string, actions []agent.Action)) {
	a.onReActInit = callback
}
func (a *combineReActAgent) SetOnReActCompleteCallback(callback func(reasoning string, actionObs []agent.ActionObservation)) {
	a.onReActComplete = callback
}

func (a *combineReActAgent) stepTaskState(state ExecutingState) (ExecutingState, bool, error) {
	resp, _, err := a.reActStepper.Call(state)
	if err != nil {
		return ExecutingState{}, false, err
	}
	if a.onReActInit != nil {
		a.onReActInit(resp.Reasoning, resp.Actions)
	}
	actionObservations := a.observeActions(resp.Actions)
	step := ReActStep{
		Reasoning:          resp.Reasoning,
		ActionObservations: actionObservations,
	}
	state.Active.Steps = append(state.Active.Steps, step)
	if a.onReActComplete != nil {
		a.onReActComplete(step.Reasoning, step.ActionObservations)
	}
	if len(actionObservations) == 0 {
		return state, false, nil
	} else {
		return state, true, nil
	}
}

func (a *combineReActAgent) observeActions(actions []agent.Action) []agent.ActionObservation {
	actionObservations := make([]agent.ActionObservation, len(actions))
	wg := &sync.WaitGroup{}
	wg.Add(len(actions))
	for i, action := range actions {
		go func(i int, action agent.Action) {
			defer wg.Done()
			var tool agent.Tool
			for _, t := range a.tools {
				if t.Name() == action.Name {
					tool = t
					break
				}
			}
			var response string
			if tool == nil {
				response = "error: there were no tools available with that name."
			} else {
				args := convertActionArgsToMap(action.Args)
				resp, err := tool.Call(args)
				if err != nil {
					response = fmt.Sprintf("error: %s", err.Error())
				} else {
					response = resp
				}
			}
			actionObservations[i] = agent.ActionObservation{
				Action: action,
				Observation: agent.Observation{
					Observed: response,
				},
			}
		}(i, action)
	}
	wg.Wait()
	return actionObservations
}

func newTaskState(query string, history []ExecutedTask) ExecutingState {
	return ExecutingState{
		History: history,
		Active: ExecutingTask{
			Task: query,
		},
	}
}
