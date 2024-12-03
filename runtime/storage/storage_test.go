package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_DataDir(t *testing.T) {
	tempDir := os.TempDir()
	client := &Client{
		dataDirPath: tempDir,
	}

	client = client.WithPrefix("testprefix")

	tests := []struct {
		name string
		elem []string
	}{
		{
			name: "create single directory",
			elem: []string{"testdir"},
		},
		{
			name: "create nested directories",
			elem: []string{"testdir", "nested"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.DataDir(tt.elem...)
			require.NoError(t, err)
			if _, err := os.Stat(got); os.IsNotExist(err) {
				t.Errorf("Client.DataDir() path = %v, directory does not exist", got)
			}
			require.Equal(t, filepath.Join(append([]string{tempDir, "testprefix"}, tt.elem...)...), got)
		})
	}
}

func TestClient_TempDir(t *testing.T) {
	tempDir := os.TempDir()
	client := &Client{
		dataDirPath: tempDir,
	}
	client = client.WithPrefix("testprefix", "testtempdir")

	tests := []struct {
		name string
		elem []string
	}{
		{
			name: "create single temp directory",
			elem: []string{"testtempdir"},
		},
		{
			name: "create nested temp directories",
			elem: []string{"testtempdir", "nested"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.TempDir(tt.elem...)
			require.NoError(t, err)
			if _, err := os.Stat(got); os.IsNotExist(err) {
				t.Errorf("Client.TempDir() path = %v, directory does not exist", got)
			}
			require.Equal(t, filepath.Join(append([]string{tempDir, "testprefix", "testtempdir"}, tt.elem...)...), got)
		})
	}
}

func TestClient_RandomTempDir(t *testing.T) {
	tempDir := os.TempDir()
	client := &Client{
		dataDirPath: tempDir,
	}

	tests := []struct {
		name    string
		pattern string
		elem    []string
	}{
		{
			name:    "create single random temp directory",
			pattern: "testtempdir-*",
			elem:    []string{"random"},
		},
		{
			name:    "create nested random temp directories",
			pattern: "testtempdir-*",
			elem:    []string{"random", "nested"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.RandomTempDir(tt.pattern, tt.elem...)
			require.NoError(t, err)
			if _, err := os.Stat(got); os.IsNotExist(err) {
				t.Errorf("Client.RandomTempDir() path = %v, directory does not exist", got)
			}
			require.Equal(t, filepath.Join(append([]string{tempDir}, tt.elem...)...), filepath.Dir(got))
		})
	}
}
