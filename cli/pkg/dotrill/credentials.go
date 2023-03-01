package dotrill

import (
	"errors"
	"fmt"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

type Credentials struct {
	credentials string
}

func GetCredentials() (string, error) {
	credentials, err := loadCredentials()
	if err != nil {
		return "", err
	}

	cred := credentials.credentials
	if cred != "" {
		return cred, fmt.Errorf("no credentials found")
	}
	return "", nil
}

func SetCredentials(creds string) error {
	credentials := &Credentials{credentials: creds}

	err := WriteCredentials(credentials)
	if err != nil {
		return err
	}

	return nil
}

func WriteCredentials(c *Credentials) error {
	credPath, err := hostsCredentialsFile()
	if err != nil {
		return err
	}

	cred, err := yaml.Marshal(&c.credentials)
	if err != nil {
		return err
	}

	err = os.WriteFile(credPath, cred, 0o644)
	if err != nil {
		return err
	}
	return nil
}

func loadCredentials() (*Credentials, error) {
	filePath, err := hostsCredentialsFile()
	if err != nil {
		return nil, err
	}

	var c string
	creds, err := readFile(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		cred := &Credentials{credentials: c}
		err = WriteCredentials(cred)
		if err != nil {
			return cred, err
		}

		return cred, err
	}

	err = yaml.Unmarshal(creds, &c)
	if err != nil {
		return nil, err
	}

	return &Credentials{credentials: c}, nil
}

func hostsCredentialsFile() (string, error) {
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

	credFile := path.Join(confFolder, "credentials")
	_, err = os.Stat(credFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}

	return credFile, nil
}
