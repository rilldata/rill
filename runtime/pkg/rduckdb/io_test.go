package rduckdb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopyDirEmptyDir(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()
	err := os.RemoveAll(dest)
	require.NoError(t, err)
	require.NoDirExists(t, dest)

	err = copyDir(dest, src)
	require.NoError(t, err)

	require.DirExists(t, dest)
	require.DirExists(t, src)
}

func TestCopyDirEmptyNestedDir(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()
	err := os.RemoveAll(dest)
	require.NoError(t, err)
	require.NoDirExists(t, dest)

	err = os.MkdirAll(filepath.Join(src, "nested1", "nested"), os.ModePerm)
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(src, "nested2"), os.ModePerm)
	require.NoError(t, err)

	err = copyDir(dest, src)
	require.NoError(t, err)

	require.DirExists(t, dest)
	require.DirExists(t, filepath.Join(dest, "nested1"))
	require.DirExists(t, filepath.Join(dest, "nested2"))
	require.DirExists(t, filepath.Join(dest, "nested1", "nested"))
}

func TestCopyDirWithFile(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dest, "existing"), os.ModePerm))

	err := os.MkdirAll(filepath.Join(src, "nested1", "nested"), os.ModePerm)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(src, "nested1", "file.txt"), []byte("nested1"), os.ModePerm))
	require.NoError(t, os.WriteFile(filepath.Join(src, "nested1", "nested", "file.txt"), []byte("nested1-nested"), os.ModePerm))

	err = os.MkdirAll(filepath.Join(src, "nested2"), os.ModePerm)
	require.NoError(t, os.WriteFile(filepath.Join(src, "nested2", "file.txt"), []byte("nested2"), os.ModePerm))
	require.NoError(t, err)

	err = copyDir(dest, src)
	require.NoError(t, err)

	contents, err := os.ReadFile(filepath.Join(dest, "nested1", "file.txt"))
	require.NoError(t, err)
	require.Equal(t, "nested1", string(contents))

	contents, err = os.ReadFile(filepath.Join(dest, "nested1", "nested", "file.txt"))
	require.NoError(t, err)
	require.Equal(t, "nested1-nested", string(contents))

	contents, err = os.ReadFile(filepath.Join(dest, "nested2", "file.txt"))
	require.NoError(t, err)
	require.Equal(t, "nested2", string(contents))

	require.DirExists(t, filepath.Join(dest, "existing"))
}
