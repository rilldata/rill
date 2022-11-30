package fileutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FullExt returns all of path's extensions. For example, for "foo.csv.zip"
// it returns ".csv.zip", not just ".zip" as filepath.Ext from the standard
// library does.
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

// Stem returns the file name after removing directory and all extensions.
// Uses FullExt to strip extensions.
func Stem(path string) string {
	return strings.TrimSuffix(filepath.Base(path), FullExt(path))
}

// CopyToTempFile pipes a reader to a temporary file. The caller must delete
// the temporary file when it's no longer needed.
func CopyToTempFile(r io.Reader, name string, ext string) (string, error) {
	// The * in the pattern will be replaced by a random string
	f, err := os.CreateTemp("", fmt.Sprintf("%s*%s", name, ext))
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
