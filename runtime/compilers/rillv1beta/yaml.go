package rillv1beta

type Source struct {
	Type   string
	URI    string `yaml:"uri,omitempty"`
	Path   string `yaml:"path,omitempty"`
	Region string `yaml:"region,omitempty"`
}

type ProjectConfig struct {
	// Project variables
	Variables   map[string]string `yaml:"env,omitempty"`
	Title       string            `yaml:"title,omitempty"`
	Description string            `yaml:"description,omitempty"`
}
