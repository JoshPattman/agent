package ai

import (
	"encoding/json"

	"github.com/JoshPattman/jpf"
	"github.com/invopop/jsonschema"
)

type ModelBuilder struct {
	Key          string
	ModelName    string
	URL          string
	UsageCounter *jpf.UsageCounter
}

func (b *ModelBuilder) BuildAgentModel(responseType any) jpf.Model {
	rformat, err := getSchema(responseType)
	if err != nil {
		panic(err)
	}
	model := jpf.NewOpenAIModel(
		b.Key,
		b.ModelName,
		jpf.WithJsonSchema{X: rformat},
		jpf.WithURL{X: b.URL},
	)
	model = jpf.NewRetryModel(model, 5)
	model = jpf.NewUsageCountingModel(model, b.UsageCounter)
	return model
}

func getSchema(obj any) (map[string]any, error) {
	r := &jsonschema.Reflector{
		BaseSchemaID:   "Anonymous",
		Anonymous:      true,
		DoNotReference: true,
	}
	s := r.Reflect(obj)
	schemaBs, err := s.MarshalJSON()
	if err != nil {
		return nil, err
	}
	schema := make(map[string]any)
	err = json.Unmarshal(schemaBs, &schema)
	if err != nil {
		return nil, err
	}
	return schema, nil
}
