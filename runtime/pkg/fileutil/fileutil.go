package fileutil

import (
	"embed"
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
func CopyToTempFile(r io.Reader, name, ext string) (string, error) {
	// The * in the pattern will be replaced by a random string
	f, err := os.CreateTemp("", fmt.Sprintf("%s*%s", name, ext))
	if err != nil {
		return "", fmt.Errorf("os.Create: %w", err)
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

// CopyEmbedDir copies an embedded directory to the local file system.
func CopyEmbedDir(fs embed.FS, src, dst string) error {
	// Get items in src
	entries, err := fs.ReadDir(src)
	if err != nil {
		return err
	}

	// Create dst dir if not exists
	err = os.MkdirAll(dst, os.ModePerm)
	if err != nil {
		return err
	}

	// Check dst is a directory
	stat, err := os.Stat(dst)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("destination '%s' is not a directory", dst)
	}

	// Copy items recursively
	for _, entry := range entries {
		entrySrc := filepath.Join(src, entry.Name())
		entryDst := filepath.Join(dst, entry.Name())

		// If it's a directory, recurse
		if entry.IsDir() {
			err := CopyEmbedDir(fs, entrySrc, entryDst)
			if err != nil {
				return err
			}
			continue
		}

		// It's a file, copy it

		data, err := fs.ReadFile(entrySrc)
		if err != nil {
			return err
		}

		if err := os.WriteFile(entryDst, data, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}
