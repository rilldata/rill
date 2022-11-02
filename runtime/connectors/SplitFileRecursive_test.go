package connectors

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplitFileRecursive(t *testing.T) {
	assertFilePath(t, "file.tar.gz", "file", ".tar.gz")
	assertFilePath(t, "/path/to/file.tar.gz", "/path/to/file", ".tar.gz")
	assertFilePath(t, "/path/to/../file.tar.gz", "/path/to/../file", ".tar.gz")
	assertFilePath(t, "./file.tar.gz", "./file", ".tar.gz")
	assertFilePath(t, "https://server.com/path/file.tar.gz", "https://server.com/path/file", ".tar.gz")
}

func assertFilePath(t *testing.T, path string, expectedName string, expectedExt string) {
	name, ext := SplitFileRecursive(path)
	require.Equal(t, name, expectedName)
	require.Equal(t, ext, expectedExt)
}
