package fileutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FullExt recursively strips file extension returns it
func FullExt(path string) string {
	fullExt := filepath.Ext(path)
	fullName := strings.TrimSuffix(path, fullExt)

	for {
		ext := filepath.Ext(fullName)
		if ext == "" {
			break
		}
		fullExt = ext + fullExt
		fullName = strings.TrimSuffix(path, fullExt)
	}

	return fullExt
}

func CopyToTempFile(r io.Reader, name string, ext string) (string, error) {
	// CreateTemp adds a random string at the end.
	// But we need an extension at the end so that duckdb uses the correct loader.
	// Hence adding <name>*<extension> so that CreateTemp adds the random strings before the extension.
	f, err := os.CreateTemp(
		os.TempDir(),
		fmt.Sprintf("%s*%s", name, ext),
	)
	if err != nil {
		return "", fmt.Errorf("os.Create: %v", err)
	}

	_, err = io.Copy(f, r)
	if err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", err
	}
	f.Close()
	return f.Name(), err
}
