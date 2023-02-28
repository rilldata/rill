package dotrill

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

type Credentials struct{}

type Config struct {
	entries map[string]any
}

func Write(c *Config) error {
	configPath, err := hostsConfigFile()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(&c.entries)
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, data, 0o644)
	if err != nil {
		return err
	}
	return nil
}

func hostsConfigFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	confFolder := path.Join(home, ".rill")
	_, err = os.Stat(confFolder)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// create folder if not exists
			err := os.MkdirAll(confFolder, os.ModePerm)
			if err != nil {
				return "", err
			}
		} else {
			// unknown error
			return "", err
		}
	}

	confFile := path.Join(confFolder, "config.yaml")
	_, err = os.Stat(confFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}

	return confFile, nil
}

func load(configPath string) (*Config, error) {
	m := make(map[string]any)
	data, err := readFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	err = yaml.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	return &Config{entries: m}, nil
}

func Set(key string, value any) error {
	configPath, err := hostsConfigFile()
	if err != nil {
		return err
	}

	config, err := load(configPath)
	if err != nil {
		return err
	}

	entries := config.entries
	entries[key] = value
	conf := &Config{entries: entries}

	err = Write(conf)
	if err != nil {
		return err
	}

	return nil
}

func Get(key string) (interface{}, error) {
	configPath, err := hostsConfigFile()
	if err != nil {
		return nil, err
	}

	config, err := load(configPath)
	if err != nil {
		return nil, err
	}

	data := config.entries
	value, ok := data[key]
	if !ok {
		return nil, fmt.Errorf("key Not found")
	}

	return value, nil
}

func readFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return data, nil
}
