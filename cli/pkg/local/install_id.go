package local

import (
	"encoding/json"
	"errors"
	"os"
	"path"

	"github.com/google/uuid"
)

func InstallID() (string, error) {
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

	globalConf := map[string]any{}
	var installId string

	confFile := path.Join(confFolder, "local.json")
	_, err = os.Stat(confFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			// return if unknown error
			return "", err
		}
	} else {
		// read file if exists
		conf, err := os.ReadFile(confFile)
		if err != nil {
			return "", err
		}
		err = json.Unmarshal(conf, &globalConf)
		if err != nil {
			return "", err
		}
	}

	// installId was used in nodejs.
	// keeping it as is to retain the same ID for existing users
	installIdAny, ok := globalConf["installId"]
	if !ok {
		// create install id if not exists
		installId = uuid.New().String()
		globalConf["installId"] = installId
		globalConfJson, err := json.Marshal(&globalConf)
		if err != nil {
			return "", err
		}
		err = os.WriteFile(confFile, globalConfJson, 0644)
		if err != nil {
			return "", err
		}
	} else {
		installId = installIdAny.(string)
	}

	return installId, nil
}
