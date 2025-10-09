package main

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
)

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
