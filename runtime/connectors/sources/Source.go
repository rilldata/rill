package sources

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
	Type        PropertyType // TODO: whats wrong with this?
	Required    bool
}

type PropertyType int

const (
	StringPropertyType  PropertyType = 1
	NumberPropertyType               = 2
	BooleanPropertyType              = 3
)

type SamplePolicy struct {
	Strategy string
	Sample   float32
	Limit    int
}
