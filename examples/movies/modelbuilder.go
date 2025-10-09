package main

import (
	"os"

	"github.com/JoshPattman/jpf"
)

type ModelBuilder struct {
	Key       string
	ModelName string
}

func (b *ModelBuilder) BuildAgentModel(responseType any) jpf.Model {
	rformat, err := getSchema(responseType)
	if err != nil {
		panic(err)
	}
	return jpf.NewOpenAIModel(
		os.Getenv("OPENAI_KEY"),
		"gpt-4.1",
		jpf.WithJsonSchema{X: rformat},
	)
}

// type mdl struct {
// 	model jpf.Model
// }

// // Respond implements jpf.Model.
// func (m *mdl) Respond(msgs []jpf.Message) (jpf.ModelResponse, error) {
// 	resp, err := m.model.Respond(msgs)
// 	fmt.Println(">>", resp.PrimaryMessage.Content, "<<")
// 	return resp, err
// }
