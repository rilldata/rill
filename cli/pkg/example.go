package pkg

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed examples/*
var exampleFS embed.FS

func InitExample(projectName string, projectDir string) error {
	examplePath, err := getExampleProject(projectName)
	if err != nil {
		return err
	}
	return CopyDir(examplePath, projectDir)
}

func getExampleProject(projectName string) (string, error) {
	examplesPath := "examples/" + projectName
	_, err := exampleFS.ReadDir(examplesPath)
	if err != nil {
		return "", err
	}
	return examplesPath, nil
}

func CopyDir(origin string, dst string) (err error) {
	entries, err := exampleFS.ReadDir(origin)
	if err != nil {
		return err
	}

	for _, entry := range entries {

		srcPath := filepath.Join(origin, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			//Create dst dir if not exists
			err = os.MkdirAll(dstPath, 0777)
			if err != nil {
				return err
			}

			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return err
			}

		} else {
			fileContent, err := exampleFS.ReadFile(srcPath)
			if err != nil {
				return err
			}

			if err := os.WriteFile(dstPath, fileContent, 0666); err != nil {
				fmt.Printf("error os.WriteFile error: %v", err)
				return err
			}
		}

	}
	return nil

}
