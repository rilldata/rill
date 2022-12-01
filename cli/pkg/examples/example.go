package examples

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed all:embed
var exampleFS embed.FS

func List() ([]string, error) {
	entries, err := exampleFS.ReadDir("embed/dist")
	if err != nil {
		return nil, err
	}

	exampleList := make([]string, 0, len(entries))
	for _, entry := range entries {
		exampleList = append(exampleList, entry.Name())
	}

	return exampleList, nil
}

func Init(projectName string, projectDir string) error {
	examplePath, err := getExampleProject(projectName)
	if err != nil {
		return err
	}
	return copyDir(examplePath, projectDir)
}

func getExampleProject(projectName string) (string, error) {
	examplesPath := "embed/dist/" + projectName
	_, err := exampleFS.ReadDir(examplesPath)
	if err != nil {
		return "", err
	}
	return examplesPath, nil
}

func copyDir(origin string, dst string) (err error) {
	entries, err := exampleFS.ReadDir(origin)
	if err != nil {
		return err
	}

	for _, entry := range entries {

		srcPath := filepath.Join(origin, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			//Create dst dir if not exists
			err = os.MkdirAll(dstPath, os.ModePerm)
			if err != nil {
				return err
			}

			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}

		} else {
			fileContent, err := exampleFS.ReadFile(srcPath)
			if err != nil {
				return err
			}

			if err := os.WriteFile(dstPath, fileContent, os.ModePerm); err != nil {
				fmt.Printf("error os.WriteFile error: %v", err)
				return err
			}
		}

	}
	return nil

}
