package openapiutil

import (
	"encoding/json"

	"github.com/getkin/kin-openapi/openapi3"
)

func MapToParameters(params []map[string]any) (openapi3.Parameters, error) {
	var parameters openapi3.Parameters

	jsonData, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonData, &parameters)
	if err != nil {
		return nil, err
	}

	return parameters, nil
}

func MapToSchema(schema map[string]any) (*openapi3.Schema, error) {
	specSchema := openapi3.Schema{}

	jsonData, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}

	err = specSchema.UnmarshalJSON(jsonData)
	if err != nil {
		return nil, err
	}

	return &specSchema, nil
}
