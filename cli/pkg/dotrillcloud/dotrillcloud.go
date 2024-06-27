package dotrillcloud

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rilldata/rill/admin/pkg/adminenv"
	"gopkg.in/yaml.v3"
)

type RillCloud struct {
	ProjectID string `yaml:"project_id"`
}

func GetAll(localProjectPath, adminURL string) (*RillCloud, error) {
	confPath, err := getConfPath(localProjectPath, adminURL)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(confPath)
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

func SetAll(localProjectPath, adminURL string, conf *RillCloud) error {
	err := os.MkdirAll(filepath.Join(localProjectPath, ".rillcloud"), os.ModePerm)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}

	confPath, err := getConfPath(localProjectPath, adminURL)
	if err != nil {
		return err
	}

	return os.WriteFile(confPath, data, 0o644)
}

func Delete(localProjectPath, adminURL string) error {
	confPath, err := getConfPath(localProjectPath, adminURL)
	if err != nil {
		return err
	}

	err = os.Remove(confPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return nil
}

func getConfPath(localProjectPath, adminURL string) (string, error) {
	env, err := adminenv.Infer(adminURL)
	if err != nil {
		return "", err
	}

	if env != "prod" {
		return filepath.Join(localProjectPath, ".rillcloud", fmt.Sprintf("project_%s.yaml", env)), nil
	}
	return filepath.Join(localProjectPath, ".rillcloud", "project.yaml"), nil
}
