package local

import (
	"encoding/json"
	"errors"
	"os"
	"path"

	"github.com/google/uuid"
)

type localConfig struct {
	InstallID        string `json:"installId"`
	AnalyticsEnabled bool   `json:"analyticsEnabled"`
}

func newLocalConfig() (*localConfig, error) {
	conf := &localConfig{
		AnalyticsEnabled: true,
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return conf, err
	}

	confFolder := path.Join(home, ".rill")
	_, err = os.Stat(confFolder)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// create folder if not exists
			err := os.MkdirAll(confFolder, os.ModePerm)
			if err != nil {
				return conf, err
			}
		} else {
			// unknown error
			return conf, err
		}
	}

	confFile := path.Join(confFolder, "local.json")
	_, err = os.Stat(confFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			// return if unknown error
			return conf, err
		}
	} else {
		// read file if exists
		confStr, err := os.ReadFile(confFile)
		if err != nil {
			return conf, err
		}
		err = json.Unmarshal(confStr, &conf)
		if err != nil {
			return conf, err
		}
	}

	// installId was used in nodejs.
	// keeping it as is to retain the same ID for existing users
	if conf.InstallID == "" {
		// create install id if not exists
		conf.InstallID = uuid.New().String()
		confJSON, err := json.Marshal(&conf)
		if err != nil {
			return conf, err
		}
		err = os.WriteFile(confFile, confJSON, 0o644)
		if err != nil {
			return conf, err
		}
	}

	return conf, nil
}
