package fileutil

import (
	"embed"
	"fmt"
	"io"
	"os"
	"os/user"
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
func CopyToTempFile(r io.Reader, name, ext string) (string, int64, error) {
	// The * in the pattern will be replaced by a random string
	f, err := os.CreateTemp("", fmt.Sprintf("%s*%s", name, ext))
	if err != nil {
		return "", 0, fmt.Errorf("os.Create: %w", err)
	}

	written, err := io.Copy(f, r)
	if err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", 0, err
	}
	f.Close()

	return f.Name(), written, err
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

// IsGlob reports whether path contains any unescaped glob meta
// characters: '*', '?', '[', or '{'. A backslash '\' escapes the
// following character.
func IsGlob(path string) bool {
	for i := 0; i < len(path); i++ {
		c := path[i]
		// Skip escaped character
		if c == '\\' {
			i++
			continue
		}
		// Check unescaped meta characters
		if c == '*' || c == '?' || c == '[' || c == '{' {
			return true
		}
	}
	return false
}

// IsDoubleStarGlob reports whether path contains an unescaped "**".
//
// A backslash '\' escapes the following character, so "\**" does not
// count as a double-star glob.
func IsDoubleStarGlob(path string) bool {
	for i := 0; i < len(path); i++ {
		c := path[i]

		// Skip escaped character
		if c == '\\' {
			i++
			continue
		}

		// Check for unescaped "**"
		if c == '*' && i+1 < len(path) && path[i+1] == '*' {
			return true
		}
	}
	return false
}

// GlobPrefix returns the literal (non-glob) prefix of p.
//
// It returns the substring of p up to (but not including) the first
// unescaped glob meta character: '*', '?', '[', or '{'. Escaped meta
// characters (preceded by '\') are treated as literals. If p contains
// no unescaped meta characters, p is returned unchanged.
//
// Examples:
//
//	GlobPrefix("a/b/c/*.txt")    → "a/b/c/"
//	GlobPrefix("a/b/c/file*.go") → "a/b/c/file"
//	GlobPrefix("meta*/**")       → "meta"
func GlobPrefix(p string) (prefix string) {
	for i := 0; i < len(p); i++ {
		c := p[i]
		// Skip escaped characters
		if c == '\\' {
			i++
			continue
		}
		// Stop at first meta character
		if c == '*' || c == '?' || c == '[' || c == '{' {
			return p[:i]
		}
	}
	// No meta found
	return p
}

// PathLevel returns the number of path components in p separated
// by delim. Empty components (from repeated delimiters) are ignored.
// Leading and trailing delimiters do not create extra components.
//
// Examples (delim = '/'):
//
//	"a"          → 1
//	"a/b"        → 2
//	"a/b/c"      → 3
//	"a/b/c/"     → 3
//	"a/b//c"     → 3
//	"/a/b/c/"    → 3
func PathLevel(p string, delim byte) int {
	level := 0
	inSegment := false

	for i := 0; i < len(p); i++ {
		if p[i] == delim {
			if inSegment {
				level++
				inSegment = false
			}
			continue
		}
		inSegment = true
	}

	// Count final segment
	if inSegment {
		level++
	}

	return level
}

// PrefixUntilLevel returns the prefix of p that contains the first
// `level` path components separated by delim. Empty components
// (from repeated delimiters) are ignored.
//
// If p has fewer than `level` components, p is returned unchanged.
//
// Examples (delim = '/'):
//
//	"a/b/c/d.txt", level=1 → "a/"
//	"a/b/c/d.txt", level=2 → "a/b/"
//	"a/b/c/d.txt", level=3 → "a/b/c/"
//	"a/b/c/d.txt", level=4 → "a/b/c/d.txt"
//
//	"a/b//c/d.txt", level=3 → "a/b//c/"
//	"/a/b/c", level=2 → "/a/b/"
func PrefixUntilLevel(p string, level int, delim byte) string {
	if level <= 0 || p == "" {
		return ""
	}

	comp := 0
	inSegment := false

	for i := 0; i < len(p); i++ {
		if p[i] == delim {
			if inSegment {
				comp++
				inSegment = false

				if comp == level {
					return p[:i+1]
				}
			}
			continue
		}
		inSegment = true
	}

	// Last component (no trailing delimiter)
	if inSegment {
		comp++
		if comp == level {
			return p
		}
	}

	// Fewer components than requested
	return p
}

// EnsureTrailingDelim returns p with a trailing delimiter.
// If p already ends with delim, it is returned unchanged.
func EnsureTrailingDelim(p string, delim byte) string {
	if p == "" {
		return p
	}
	if p[len(p)-1] == delim {
		return p
	}
	return p + string(delim)
}

// ForceRemoveFiles deletes multiple files
// ignores path errors if any
func ForceRemoveFiles(paths []string) {
	for _, path := range paths {
		_ = os.Remove(path)
	}
}

func ExpandHome(path string) (string, error) {
	if path == "" || path[0] != '~' {
		return path, nil
	}
	if len(path) > 1 && path[1] != '/' {
		return path, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	if usr.HomeDir == "" {
		return "", fmt.Errorf("cannot expand '~' in path %q because the current user doesn't have a home directory", path)
	}

	return filepath.Join(usr.HomeDir, path[1:]), nil
}

func ResolveLocalPath(path, root string, allowHostAccess bool) (string, error) {
	path, err := ExpandHome(path)
	if err != nil {
		return "", err
	}

	finalPath := path
	if !filepath.IsAbs(path) {
		finalPath = filepath.Join(root, path)
	}

	if !allowHostAccess && !strings.HasPrefix(finalPath, root) {
		// path is outside the repo root
		return "", fmt.Errorf("path is outside repo root")
	}
	return finalPath, nil
}

// OpenTempFileInDir opens a temp file in given dir
// If dir doesn't exist it creates full dir path (recursively if required)
func OpenTempFileInDir(dir, filePath string) (*os.File, error) {
	// recursively create all directories
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}

	return os.CreateTemp(dir, fmt.Sprintf("%s*%s", Stem(filePath), FullExt(filePath)))
}
