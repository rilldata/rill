package fileutil

import (
	"os/user"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFullExt(t *testing.T) {
	variations := []struct {
		Path        string
		ExpectedExt string
	}{
		{"file.tar.gz", ".tar.gz"},
		{"/path/to/file.tar.gz", ".tar.gz"},
		{"/path/to/../file.tar.gz", ".tar.gz"},
		{"./file.tar.gz", ".tar.gz"},
		{"https://server.com/path/file.tar.gz", ".tar.gz"},
	}
	for _, tt := range variations {
		t.Run(tt.Path, func(t *testing.T) {
			ext := FullExt(tt.Path)
			require.Equal(t, ext, tt.ExpectedExt)
		})
	}
}

func TestGetFileName(t *testing.T) {
	variations := []struct {
		Path         string
		ExpectedName string
	}{
		{"file.yaml", "file"},
		{"file.tar.gz", "file"},
		{"/path/to/file.tar.gz", "file"},
		{"/path/to/../file.tar.gz", "file"},
		{"./file.tar.gz", "file"},
		{"https://server.com/path/file.tar.gz", "file"},
	}
	for _, tt := range variations {
		t.Run(tt.Path, func(t *testing.T) {
			ext := Stem(tt.Path)
			require.Equal(t, ext, tt.ExpectedName)
		})
	}
}

func TestExpandHome(t *testing.T) {
	usr, err := user.Current()
	require.NoError(t, err)
	home := usr.HomeDir

	variations := []struct {
		Path         string
		ExpectedPath string
	}{
		{"file.yaml", "file.yaml"},
		{"./file.tar.gz", "./file.tar.gz"},
		{"~", home},
		{"~/", home},
		{"~file.yaml", "~file.yaml"},
		{"~/path/to/file.tar.gz", home + "/path/to/file.tar.gz"},
		{"/path/to/file.tar.gz", "/path/to/file.tar.gz"},
	}

	for _, tt := range variations {
		t.Run(tt.Path, func(t *testing.T) {
			home, err := ExpandHome(tt.Path)
			require.NoError(t, err)
			require.Equal(t, tt.ExpectedPath, home)
		})
	}
}
