package agent

import (
	"bytes"
	"encoding/json"
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

func AgentAsTool(buildAgent func() *Agent, name string, description []string) Tool {
	return &agentAsTool{buildAgent, name, description}
}

type agentAsTool struct {
	buildAgent func() *Agent
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
