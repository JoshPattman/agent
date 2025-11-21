package craig

import (
	_ "embed"

	"github.com/JoshPattman/agent"
	"github.com/JoshPattman/jpf"
)

// Given a state, create the following react step.
type reActStepper jpf.MapFunc[executingState, reActResponse]

// Given a state, create the final response.
type responseStepper jpf.MapFunc[executingState, answerResponse]

//go:embed system.gtpl
var defaultSystemPrompt string
var defaultReActModePrefix = "You are now in reason-action mode. Your next task / query to respond to is as follows:\n"
var defaultAnswerModeContent = "You are now in final answer mode, create your final answer."

func newReActStepper(personality string, modelBuilder agent.AgentModelBuilder, tools []agent.Tool, systemPrompt string, taskPrefix string, answerModeContent string, scenarios map[string]agent.Scenario) reActStepper {
	return jpf.NewOneShotMapFunc(
		&stateHistoryMessageEncoder{
			personality,
			systemPrompt,
			taskPrefix,
			answerModeContent,
			reActState,
			tools,
			scenarios,
		},
		jpf.NewJsonResponseDecoder[reActResponse](), //jpf.NewValidatingResponseDecoder(, func(resp reActResponse) error { return fmt.Errorf("%v", resp) }),
		modelBuilder.BuildAgentModel(reActResponse{}),
	)
}

func newAnswerStepper(personality string, modelBuilder agent.AgentModelBuilder, tools []agent.Tool, systemPrompt string, taskPrefix string, answerModeContent string, scenarios map[string]agent.Scenario) responseStepper {
	return jpf.NewOneShotMapFunc(
		&stateHistoryMessageEncoder{
			personality,
			systemPrompt,
			taskPrefix,
			answerModeContent,
			answerState,
			tools,
			scenarios,
		},
		jpf.NewJsonResponseDecoder[answerResponse](),
		modelBuilder.BuildAgentModel(answerResponse{}),
	)
}
