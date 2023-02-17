package drivers

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-yaml/yaml"
)

// EnviornmentVariables holds the env variables set at an instance level
// first level in nested map will correspond to kind of env variable like user defined env, secrets, global etc
// second level will be name of the key
// Using a map in place of struct at first level, in order to be able to refer keys by small letters
// for eg : {{.env.timeout}} instead of {{.Env.timeout}}
// This places a constraint that env variables need to be flattened
type EnviornmentVariables map[string]map[string]string

// NewEnvVariables creates a new EnviornmentVariables.
// Defaults are picked from yamlFile
// envString correspond to user set values
func NewEnvVariables(ctx context.Context, yamlFile, envString string) (EnviornmentVariables, error) {
	// default env variables from rill.yaml
	defaultEnv := &rillYaml{Env: make(map[string]string)}
	if err := yaml.Unmarshal([]byte(yamlFile), defaultEnv); err != nil {
		return nil, err
	}

	// env variables from rill start command
	parsedEnv, err := parse(envString)
	if err != nil {
		return nil, err
	}

	// override defaults
	for key, value := range parsedEnv {
		defaultEnv.Env[key] = value
	}

	return EnviornmentVariables{"env": defaultEnv.Env}, nil
}

// Value implements driver.Valuer interface
func (e *EnviornmentVariables) Value() (driver.Value, error) {
	return json.Marshal(e)
}

// Scan implements sql.Scanner interface
func (e *EnviornmentVariables) Scan(val interface{}) error {
	return json.Unmarshal(val.([]byte), e)
}

// String returns string representation of EnviornmentVariables
func (e EnviornmentVariables) String() string {
	// ignoring the error since db values will always be serializable
	val, err := json.Marshal(e)
	_ = err

	return string(val)
}

// Get fetches the value from env as per key and kind
// At present only env kind is present which is a straightfwd map lookup
// Some kind like secrets may require fetching data from some key vault
func (e EnviornmentVariables) Get(kind, key string) string {
	env, ok := e[kind]
	if !ok {
		return ""
	}

	return env[key]
}

func parse(envString string) (map[string]string, error) {
	if envString == "" {
		return make(map[string]string), nil
	}

	// split each env variable
	envs := strings.Split(envString, ";")
	vars := make(map[string]string, len(envs))
	for _, env := range envs {
		if env == "" {
			// extra semi colon
			continue
		}
		// split into key value pairs
		key, value, found := strings.Cut(env, "=")
		// key can't be empty value can be
		if !found || key == "" {
			return nil, fmt.Errorf("invalid env token %q", env)
		}
		vars[key] = value
	}
	return vars, nil
}

type rillYaml struct {
	Env map[string]string `yaml:"env,omitempty"`
}
