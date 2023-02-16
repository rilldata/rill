package drivers

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-yaml/yaml"
)

type EnviornmentVariables map[string]map[string]string

func NewEnvVariables(ctx context.Context, yamlFile, envString string) (EnviornmentVariables, error) {
	// default env variables from rill.yaml
	e := &rillYaml{env: make(map[string]string)}
	if err := yaml.Unmarshal([]byte(yamlFile), e); err != nil {
		return nil, err
	}

	// env variables from rill start command
	m, err := parse(envString)
	if err != nil {
		return nil, err
	}

	// override defaults
	for key, value := range m {
		e.env[key] = value
	}

	return EnviornmentVariables{"env": e.env}, nil
}

func (e *EnviornmentVariables) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func (e *EnviornmentVariables) Scan(val interface{}) error {
	return json.Unmarshal(val.([]byte), e)
}

func (e EnviornmentVariables) String() string {
	val, err := json.Marshal(e)
	if err != nil {
		_ = err
	}

	return string(val)
}

func (e EnviornmentVariables) Get(key string) string {
	env, ok := e["env"]
	if !ok {
		return ""
	}

	return env["key"]
}

func parse(envString string) (map[string]string, error) {
	if envString == "" {
		return make(map[string]string), nil
	}

	envs := strings.Split(envString, ";")
	vars := make(map[string]string, len(envs))
	for _, env := range envs {
		keyvalue := strings.Split(env, "=")
		if len(keyvalue) != 2 {
			return nil, fmt.Errorf("invalid env string %q", env)
		}
		vars[keyvalue[0]] = keyvalue[1]
	}
	return vars, nil
}

type rillYaml struct {
	env map[string]string `yaml:"env"`
}
