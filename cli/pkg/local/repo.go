package local

import (
	"path/filepath"
)

func PathToProjectName(path string) string {
	name := filepath.Base(path)
	if name == "" || name == "." || name == ".." {
		return "untitled"
	}
	return name
}
