package dotgit

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	FileName = ".git"
)

// Constants for keys
const (
	RemoteKey         = "remote"
	UsernameKey       = "username"
	PasswordKey       = "password"
	PasswordExpiryKey = "password_expiry"
)

type DotGit struct {
	projectDir string
}

func New(projectDir string) DotGit {
	return DotGit{projectDir: projectDir}
}

func (d DotGit) GetAll() (map[string]string, error) {
	data, err := os.ReadFile(d.filePath())
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}

	conf := map[string]string{}
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func (d DotGit) Get(key string) (string, error) {
	conf, err := d.GetAll()
	if err != nil {
		return "", err
	}

	return conf[key], nil
}

func (d DotGit) Set(key, value string) error {
	conf, err := d.GetAll()
	if err != nil {
		return err
	}
	conf[key] = value

	data, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}

	return os.WriteFile(d.filePath(), data, 0o644)
}

func (d DotGit) filePath() string {
	return filepath.Join(d.projectDir, "tmp", FileName)
}
