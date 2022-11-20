package fileutil

import (
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
