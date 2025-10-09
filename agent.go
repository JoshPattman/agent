package agent

import (
	"fmt"
	"sync"
)

func NewAgent(modelBuilder AgentModelBuilder, opts ...AgentOpt) *Agent {
	params := &agentParams{
		systemPrompt:       defaultSystemPrompt,
		taskPrefix:         defaultReActModePrefix,
		finalAnswerMessage: defaultAnswerModeContent,
	}
	for _, o := range opts {
		o(params)
	}
	return &Agent{
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
	tools              []Tool
}

type AgentOpt func(*agentParams)

func WithSystemPromptTemplate(tpl string) AgentOpt {
	return func(a *agentParams) {
		a.systemPrompt = tpl
	}
}

func WithTools(tools ...Tool) AgentOpt {
	return func(a *agentParams) {
		a.tools = append(a.tools, tools...)
	}
}

func WithTaskPrefix(prefix string) AgentOpt {
	return func(a *agentParams) {
		a.taskPrefix = prefix
	}
}

func WithFinalAnswerMessage(msg string) AgentOpt {
	return func(a *agentParams) {
		a.finalAnswerMessage = msg
	}
}

type Agent struct {
	history         []ExecutedTask
	reActStepper    reActStepper
	answerStepper   responseStepper
	tools           []Tool
	onReActInit     func(ReActResponse)
	onReActComplete func(ReActStep)
}

func (a *Agent) Answer(query string) (string, error) {
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

func (a *Agent) SetOnReActInitCallback(callback func(ReActResponse)) { a.onReActInit = callback }
func (a *Agent) SetOnReActCompleteCallback(callback func(ReActStep)) { a.onReActComplete = callback }

func (a *Agent) stepTaskState(state ExecutingState) (ExecutingState, bool, error) {
	resp, _, err := a.reActStepper.Call(state)
	if err != nil {
		return ExecutingState{}, false, err
	}
	if a.onReActInit != nil {
		a.onReActInit(resp)
	}
	actionObservations := a.observeActions(resp.Actions)
	step := ReActStep{
		Reasoning:          resp.Reasoning,
		ActionObservations: actionObservations,
	}
	state.Active.Steps = append(state.Active.Steps, step)
	if a.onReActComplete != nil {
		a.onReActComplete(step)
	}
	if len(actionObservations) == 0 {
		return state, false, nil
	} else {
		return state, true, nil
	}
}

func (a *Agent) observeActions(actions []Action) []ActionObservation {
	actionObservations := make([]ActionObservation, len(actions))
	wg := &sync.WaitGroup{}
	wg.Add(len(actions))
	for i, action := range actions {
		go func(i int, action Action) {
			defer wg.Done()
			var tool Tool
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
			actionObservations[i] = ActionObservation{
				Action: action,
				Observation: Observation{
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
