package sources

import "github.com/rilldata/rill/runtime/api"

type Source struct {
	Name         string
	Connector    string
	SamplePolicy SamplePolicy
	Properties   map[string]any
}

type Property struct {
	Key         string
	DisplayName string
	Description string
	Placeholder string
	Type        api.Connector_Property_Type // TODO: whats wrong with this?
	Required    bool
}

type SamplePolicy struct {
	Strategy string
	Sample   float32
	Limit    int
}
