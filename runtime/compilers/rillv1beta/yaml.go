package rillv1beta

import (
	"regexp"
	"strings"
)

var alphaNumericRegex = regexp.MustCompile("[^A-Za-z0-9]+")

type Source struct {
	Type   string
	URI    string `yaml:"uri,omitempty"`
	Path   string `yaml:"path,omitempty"`
	Region string `yaml:"region,omitempty"`
}

type ProjectConfig struct {
	// Project variables
	Variables map[string]string `yaml:"env,omitempty"`
	Name      string            `yaml:"name,omitempty"`
}

func (p *ProjectConfig) SanitizedName() string {
	name := alphaNumericRegex.ReplaceAllString(strings.TrimSpace(p.Name), "-")
	if name == "-" { // no alphanumeric characters
		return ""
	}
	return name
}
