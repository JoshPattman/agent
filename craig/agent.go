// craig implements a Combined ReAct Intelligent aGent, where the agent reasons and acts in a single message.
package craig

import (
	"context"
	"fmt"
	"sync"

	"github.com/JoshPattman/agent"
)

func New(modelBuilder agent.AgentModelBuilder, opts ...NewOpt) agent.Agent {
	params := &agentParams{
		systemPrompt:       defaultSystemPrompt,
		taskPrefix:         defaultReActModePrefix,
		finalAnswerMessage: defaultAnswerModeContent,
	}
	for _, o := range opts {
		o(params)
	}
	if len(params.scenarios) > 0 {
		params.tools = append(params.tools, agent.NewScenarioRetrieverTool(params.scenarios))
	}
	return &combineReActAgent{
		params:       *params,
		modelBuilder: modelBuilder,
	}
}

type agentParams struct {
	personality        string
	systemPrompt       string
	taskPrefix         string
	finalAnswerMessage string
	tools              []agent.Tool
	scenarios          map[string]agent.Scenario
}

type NewOpt func(*agentParams)

func WithSystemPromptTemplate(tpl string) NewOpt {
	return func(a *agentParams) {
		a.systemPrompt = tpl
	}
}

func WithTools(tools ...agent.Tool) NewOpt {
	return func(a *agentParams) {
		a.tools = append(a.tools, tools...)
	}
}

func WithTaskPrefix(prefix string) NewOpt {
	return func(a *agentParams) {
		a.taskPrefix = prefix
	}
}

func WithFinalAnswerMessage(msg string) NewOpt {
	return func(a *agentParams) {
		a.finalAnswerMessage = msg
	}
}

func WithPersonality(personality string) NewOpt {
	return func(a *agentParams) {
		a.personality = personality
	}
}

func WithScenarios(scenarios map[string]agent.Scenario) NewOpt {
	return func(ap *agentParams) {
		ap.scenarios = scenarios
	}
}

type combineReActAgent struct {
	history         []executedTask
	params          agentParams
	modelBuilder    agent.AgentModelBuilder
	onReActInit     func(string, []agent.Action)
	onReActComplete func(string, []agent.ActionObservation)
	onStreamBegin   func()
	onStreamChunk   func(string)
}

func (a *combineReActAgent) Answer(query string) (string, error) {
	reactStepper, answerStepper := a.buildSteppers()
	state := newTaskState(query, a.history)
	// Do reasoning and acting loop
	for {
		newState, ok, err := a.stepTaskState(reactStepper, state)
		if err != nil {
			return "", err
		}
		state = newState
		if !ok {
			break
		}
	}
	// Finalise output
	finalResponse, _, err := answerStepper.Call(context.Background(), state)
	if err != nil {
		return "", err
	}
	a.history = append(a.history, executedTask{
		Task:     query,
		Steps:    state.Active.Steps,
		Response: finalResponse,
	})
	return finalResponse, nil
}

func (a *combineReActAgent) buildSteppers() (reActStepper, responseStepper) {
	rs := newReActStepper(
		a.params.personality,
		a.modelBuilder,
		a.params.tools,
		a.params.systemPrompt,
		a.params.taskPrefix,
		a.params.finalAnswerMessage,
		a.params.scenarios,
	)
	as := newAnswerStepper(
		a.params.personality,
		a.modelBuilder,
		a.params.tools,
		a.params.systemPrompt,
		a.params.taskPrefix,
		a.params.finalAnswerMessage,
		a.params.scenarios,
		a.onStreamBegin,
		a.onStreamChunk,
	)
	return rs, as
}

func (a *combineReActAgent) SetOnReActInitCallback(callback func(reasoning string, actions []agent.Action)) {
	a.onReActInit = callback
}
func (a *combineReActAgent) SetOnReActCompleteCallback(callback func(reasoning string, actionObs []agent.ActionObservation)) {
	a.onReActComplete = callback
}

func (a *combineReActAgent) SetOnBeginStreamAnswerCallback(callback func()) {
	a.onStreamBegin = callback
}

func (a *combineReActAgent) SetOnStreamAnswerChunkCallback(callback func(string)) {
	a.onStreamChunk = callback
}

func (a *combineReActAgent) stepTaskState(stepper reActStepper, state executingState) (executingState, bool, error) {
	resp, _, err := stepper.Call(context.Background(), state)
	if err != nil {
		return executingState{}, false, err
	}
	if a.onReActInit != nil {
		a.onReActInit(resp.Reasoning, resp.Actions)
	}
	actionObservations := a.observeActions(resp.Actions)
	step := reActStep{
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
			for _, t := range a.params.tools {
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

func newTaskState(query string, history []executedTask) executingState {
	return executingState{
		History: history,
		Active: executingTask{
			Task: query,
		},
	}
}
