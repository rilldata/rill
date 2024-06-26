package dotrillcloud

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type RillCloud struct {
	ProjectID string `json:"project_id"`
}

var confPath = filepath.Join(".rillcloud", "project.yaml")

func GetAll(localProjectPath string) (*RillCloud, error) {
	data, err := os.ReadFile(filepath.Join(localProjectPath, confPath))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	conf := &RillCloud{}
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func SetAll(localProjectPath string, conf *RillCloud) error {
	err := os.MkdirAll(filepath.Join(localProjectPath, ".rillcloud"), os.ModePerm)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(localProjectPath, confPath), data, 0o644)
}
