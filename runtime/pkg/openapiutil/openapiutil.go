package openapiutil

import (
	"encoding/json"

	"github.com/go-openapi/spec"
)

func MapToParameters(params []map[string]any) ([]spec.Parameter, error) {
	var parameters []spec.Parameter
	for _, param := range params {
		var specParam spec.Parameter
		jsonData, err := json.Marshal(param)
		if err != nil {
			return nil, err
		}
		err = specParam.UnmarshalJSON(jsonData)
		if err != nil {
			return nil, err
		}
		parameters = append(parameters, specParam)
	}
	return parameters, nil
}

func MapToSchema(schema map[string]any) (*spec.Schema, error) {
	specSchema := spec.Schema{}
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
