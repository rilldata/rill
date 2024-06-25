package dotrillcloud

import (
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

type RillCloud struct {
	ProjectID string `json:"project_id"`
}

const confPath = ".rillcloud/project.yaml"

func GetAll(localProjectPath string) (*RillCloud, bool, error) {
	data, err := os.ReadFile(path.Join(localProjectPath, confPath))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	conf := &RillCloud{}
	err = yaml.Unmarshal(data, conf)
	if err != nil {
		return nil, false, err
	}
	return conf, false, nil
}

func SetAll(localProjectPath string, conf *RillCloud) error {
	err := os.MkdirAll(path.Join(localProjectPath, ".rillcloud"), os.ModePerm)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}

	return os.WriteFile(path.Join(localProjectPath, confPath), data, 0o644)
}
