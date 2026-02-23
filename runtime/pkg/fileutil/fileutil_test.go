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

func TestIsGlob(t *testing.T) {
	variations := []struct {
		Path     string
		Expected bool
	}{
		// No glob
		{"plain/path/file.txt", false},
		{"file.txt", false},
		{"C:\\Users\\file.txt", false},

		// Simple globs
		{"file*.txt", true},
		{"file?.txt", true},
		{"file[0-9].txt", true},
		{"file{a,b}.txt", true},
		{"a/b/*", true},
		{"*/test", true},

		// Escaped meta characters (should NOT be glob)
		{`escaped\*star.txt`, false},
		{`escaped\?mark.txt`, false},
		{`escaped\[abc].txt`, false},
		{`escaped\{a,b}.txt`, false},

		// Mixed escaped and unescaped
		{`dir/\*/file*.txt`, true}, // escaped *, but later real *
		{`dir/\[abc]/file`, false}, // only escaped meta
		{`dir/\{x}/file?`, true},   // escaped { }, but unescaped ?

		// Escape at end (no following char)
		{`endswithslash\`, false},

		// Meta in first segment
		{"*file.txt", true},
		{"?file.txt", true},
		{"[a]file.txt", true},
		{"{x}file.txt", true},
	}

	for _, tt := range variations {
		t.Run(tt.Path, func(t *testing.T) {
			got := IsGlob(tt.Path)
			require.Equal(t, tt.Expected, got)
		})
	}
}

func TestIsDoubleStartGlob(t *testing.T) {
	variations := []struct {
		Path     string
		Expected bool
	}{
		// No double-star
		{"plain/path/file.txt", false},
		{"a/b/*/c", false},
		{"file*.txt", false},
		{"*", false},

		// Simple double-star
		{"**", true},
		{"**/file.txt", true},
		{"a/**/b", true},
		{"a/b/**", true},
		{"a/**", true},

		// Multiple stars but not double-star
		{"***", true}, // contains "**"
		{"a/*/**/c", true},

		// Escaped double-star (should NOT count)
		{`a/\**/b`, false},
		{`\**`, false},

		// Mixed escaped and unescaped
		{`a/\**/**/b`, true}, // second ** is real
		{`a/**/\**/b`, true}, // first ** is real

		// Edge cases
		{"", false},
		{`endswithslash\`, false},
	}

	for _, tt := range variations {
		t.Run(tt.Path, func(t *testing.T) {
			got := IsDoubleStarGlob(tt.Path)
			require.Equal(t, tt.Expected, got)
		})
	}
}

func TestGlobPrefix(t *testing.T) {
	variations := []struct {
		Path           string
		ExpectedPrefix string
	}{
		{"a/b/c/*.txt", "a/b/c/"},
		{"a/b/c/file*.txt", "a/b/c/file"},
		{"meta*/**", "meta"},
		{"/var/log/*", "/var/log/"},
		{"../../path/to/meta*/**", "../../path/to/meta"},
		{"plain/path/no_glob", "plain/path/no_glob"},
		{"file?.go", "file"},
		{`escaped\*star.txt`, `escaped\*star.txt`}, // escaped '*' is literal
		{`dir/\[abc]/file`, `dir/\[abc]/file`},     // escaped '[' is literal
	}

	for _, tt := range variations {
		t.Run(tt.Path, func(t *testing.T) {
			prefix := GlobPrefix(tt.Path)
			require.Equal(t, tt.ExpectedPrefix, prefix)
		})
	}
}

func TestPathLevel(t *testing.T) {
	variations := []struct {
		Path     string
		Delim    byte
		Expected int
	}{
		// Basic paths
		{"a/b/c/", '/', 3},
		{"a/b/c", '/', 3},
		{"a/b/c/txt", '/', 4},
		{"a//b///c", '/', 3},

		// Windows-style delimiter
		{`a\b\c\`, '\\', 3},

		// Empty path
		{"", '/', 0},

		// Glob paths (treated as plain text)
		{"a/b/*.txt", '/', 3},
		{"a/**/b", '/', 3},

		// Escaped glob characters
		{`a/b/\*.txt`, '/', 3},
	}

	for _, tt := range variations {
		t.Run(tt.Path, func(t *testing.T) {
			got := PathLevel(tt.Path, tt.Delim)
			require.Equal(t, tt.Expected, got)
		})
	}
}

func TestPrefixUntilLevel(t *testing.T) {
	variations := []struct {
		Path     string
		Level    int
		Delim    byte
		Expected string
	}{
		// Basic paths
		{"a/b/c/d.txt", 1, '/', "a/"},
		{"a/b/c/d.txt", 2, '/', "a/b/"},
		{"a/b/c/d.txt", 3, '/', "a/b/c/"},
		{"a/b/c/d.txt", 4, '/', "a/b/c/d.txt"},

		// Trailing slash
		{"a/b/c/", 2, '/', "a/b/"},

		// Repeated delimiters (empty components ignored)
		{"a/b//c/d.txt", 3, '/', "a/b//c/"},

		// Leading delimiter
		{"/a/b/c", 2, '/', "/a/b/"},

		// Single component
		{"a", 1, '/', "a"},

		// Level greater than components → whole path as dir
		{"a/b", 5, '/', "a/b"},

		// Empty / zero level
		{"a/b/c", 0, '/', ""},
		{"", 2, '/', ""},
	}

	for _, tt := range variations {
		t.Run(tt.Path, func(t *testing.T) {
			got := PrefixUntilLevel(tt.Path, tt.Level, tt.Delim)
			require.Equal(t, tt.Expected, got)
		})
	}
}
