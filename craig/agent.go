// craig implements a Combined ReAct Intelligent aGent, where the agent reasons and acts in a single message.
package craig

import (
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
		reActStepper: newReActStepper(
			params.personality,
			modelBuilder,
			params.tools,
			params.systemPrompt,
			params.taskPrefix,
			params.finalAnswerMessage,
			params.scenarios,
		),
		answerStepper: newAnswerStepper(
			params.personality,
			modelBuilder,
			params.tools,
			params.systemPrompt,
			params.taskPrefix,
			params.finalAnswerMessage,
			params.scenarios,
		),
		tools: params.tools,
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
	a.history = append(a.history, executedTask{
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

func (a *combineReActAgent) stepTaskState(state executingState) (executingState, bool, error) {
	resp, _, err := a.reActStepper.Call(state)
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

func newTaskState(query string, history []executedTask) executingState {
	return executingState{
		History: history,
		Active: executingTask{
			Task: query,
		},
	}
}
