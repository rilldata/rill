package rillv1

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type ProjectConfig struct {
	// Project variables
	Variables   map[string]string `yaml:"env,omitempty"`
	Title       string            `yaml:"title,omitempty"`
	Description string            `yaml:"description,omitempty"`
}

func HasRillProject(dir string) bool {
	_, err := os.Open(filepath.Join(dir, "rill.yaml"))
	return err == nil
}

func ParseProjectConfig(content []byte) (*ProjectConfig, error) {
	c := &ProjectConfig{Variables: make(map[string]string)}
	if err := yaml.Unmarshal(content, c); err != nil {
		return nil, err
	}

	return c, nil
}
