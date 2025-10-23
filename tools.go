package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"slices"

	"github.com/JoshPattman/jpf"
	"github.com/mitchellh/mapstructure"
)

func MapFuncTool[T any](mf jpf.MapFunc[T, string], name string, description []string) Tool {
	return &mapFuncTool[T]{
		mf:   mf,
		name: name,
		desc: description,
	}
}

type mapFuncTool[T any] struct {
	mf   jpf.MapFunc[T, string]
	name string
	desc []string
}

// Call implements Tool.
func (m *mapFuncTool[T]) Call(args map[string]any) (string, error) {
	var typedArgs T
	err := mapstructure.Decode(args, &typedArgs)
	if err != nil {
		return "", err
	}
	result, _, err := m.mf.Call(typedArgs)
	return result, err
}

// Description implements Tool.
func (m *mapFuncTool[T]) Description() []string {
	return slices.Clone(m.desc)
}

// Name implements Tool.
func (m *mapFuncTool[T]) Name() string {
	return m.name
}

func FunctionalTool(do func(map[string]any) (string, error), name string, description []string) Tool {
	return &functionalTool{
		do:   do,
		name: name,
		desc: description,
	}
}

type functionalTool struct {
	do   func(map[string]any) (string, error)
	name string
	desc []string
}

// Call implements Tool.
func (f *functionalTool) Call(args map[string]any) (string, error) {
	return f.do(args)
}

// Description implements Tool.
func (f *functionalTool) Description() []string {
	return slices.Clone(f.desc)
}

// Name implements Tool.
func (f *functionalTool) Name() string {
	return f.name
}

func AgentAsTool(buildAgent func() Agent, name string, description []string) Tool {
	return &agentAsTool{buildAgent, name, description}
}

type agentAsTool struct {
	buildAgent func() Agent
	name       string
	desc       []string
}

// Call implements Tool.
func (a *agentAsTool) Call(args map[string]any) (string, error) {
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(args)
	if err != nil {
		return "", err
	}
	return a.buildAgent().Answer(buf.String())
}

// Description implements Tool.
func (a *agentAsTool) Description() []string {
	return slices.Clone(a.desc)
}

// Name implements Tool.
func (a *agentAsTool) Name() string {
	return a.name
}

type SubAgentConstructor struct {
	Build func() Agent
	Desc  string
}

func CreateSubAgentTools(agentStorage map[string]Agent, createFuncs map[string]SubAgentConstructor) (Tool, Tool) {
	createTool := &createSubagentTool{
		existingAgents: agentStorage,
		agentFuncs:     createFuncs,
	}
	callTool := &continueSubagentTool{
		existingAgents: agentStorage,
	}
	return createTool, callTool
}

type createSubagentTool struct {
	existingAgents map[string]Agent
	agentFuncs     map[string]SubAgentConstructor
}

func (t *createSubagentTool) Name() string {
	return "create_subagent"
}

func (t *createSubagentTool) Description() []string {
	s := []string{
		"Create a new subagent of the given type.",
		"You may create as many agents (of the same or different types) as you wish.",
		"You must provide an `agent_type` string, and an `initial_query` string.",
		"Allowed types are:",
	}
	for key, af := range t.agentFuncs {
		s = append(s, fmt.Sprintf("`%s`: %s", key, af.Desc))
	}
	return s
}

func (t *createSubagentTool) Call(args map[string]any) (string, error) {
	// Extract and validate arguments
	agentTypeRaw, ok := args["agent_type"]
	if !ok {
		return "", fmt.Errorf("missing required argument: agent_type")
	}
	agentType, ok := agentTypeRaw.(string)
	if !ok {
		return "", fmt.Errorf("agent_type must be a string")
	}
	initialQueryRaw, ok := args["initial_query"]
	if !ok {
		return "", fmt.Errorf("missing required argument: initial_query")
	}
	initialQuery, ok := initialQueryRaw.(string)
	if !ok {
		return "", fmt.Errorf("initial_query must be a string")
	}

	// Find subagent constructor
	constructor, ok := t.agentFuncs[agentType]
	if !ok {
		return "", fmt.Errorf("unknown agent_type: %s", agentType)
	}

	// Build the subagent
	subagent := constructor.Build()

	// Answer the initial query
	answer, err := subagent.Answer(initialQuery)
	if err != nil {
		return "", err
	}

	key := randKey()
	t.existingAgents[key] = subagent

	return fmt.Sprintf("Created subagent with conversation key '%s'. It responded:\n\n%s", key, answer), nil
}

func randKey() string {
	alph := "abcdefghijklmnopqrstuvwxyz"
	s := []byte{}
	for range 10 {
		s = append(s, alph[rand.Intn(len(alph))])
	}
	return string(s)
}

type continueSubagentTool struct {
	existingAgents map[string]Agent
}

func (t *continueSubagentTool) Name() string {
	return "continue_subagent"
}

func (t *continueSubagentTool) Description() []string {
	return []string{
		"Continue a conversation with an existing subagent.",
		"Provide 'conversation_id' (string) and 'follow_up_query' (string).",
	}
}

func (t *continueSubagentTool) Call(args map[string]any) (string, error) {
	conversationIDRaw, ok := args["conversation_id"]
	if !ok {
		return "", fmt.Errorf("missing required argument: conversation_id")
	}
	conversationID, ok := conversationIDRaw.(string)
	if !ok {
		return "", fmt.Errorf("conversation_id must be a string")
	}
	followUpQueryRaw, ok := args["follow_up_query"]
	if !ok {
		return "", fmt.Errorf("missing required argument: follow_up_query")
	}
	followUpQuery, ok := followUpQueryRaw.(string)
	if !ok {
		return "", fmt.Errorf("follow_up_query must be a string")
	}

	subagent, ok := t.existingAgents[conversationID]
	if !ok {
		return "", fmt.Errorf("no subagent found for conversation_id: %s", conversationID)
	}

	answer, err := subagent.Answer(followUpQuery)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Subagent '%s' responded:\n\n%s", conversationID, answer), nil
}

// NewAgentQuickQuestionTool creates a tool that builds a fresh agent and asks a query.
// The tool's name will be agent_<name> and the description is formatted using agentDescription.
func NewAgentQuickQuestionTool(buildAgent func() Agent, name string, agentDescription string) Tool {
	return &newAgentQuickQuestionTool{
		buildAgent:       buildAgent,
		name:             name,
		agentDescription: agentDescription,
	}
}

type newAgentQuickQuestionTool struct {
	buildAgent       func() Agent
	name             string
	agentDescription string
}

func (t *newAgentQuickQuestionTool) Name() string {
	return "agent_" + t.name
}

func (t *newAgentQuickQuestionTool) Description() []string {
	return []string{
		"Agent: " + t.name + " — " + t.agentDescription,
		"Provide 'query' (string) - the question to ask the new agent.",
		"This creates a completely new agent instance each time, so there's no conversation history.",
	}
}

func (t *newAgentQuickQuestionTool) Call(args map[string]any) (string, error) {
	queryRaw, ok := args["query"]
	if !ok {
		return "", fmt.Errorf("missing required argument: query")
	}
	query, ok := queryRaw.(string)
	if !ok {
		return "", fmt.Errorf("query must be a string")
	}

	// Build a fresh agent
	agent := t.buildAgent()

	// Ask the query and return the result
	answer, err := agent.Answer(query)
	if err != nil {
		return "", err
	}

	return answer, nil
}
