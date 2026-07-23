package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

type testTarEntry struct {
	name     string
	typeflag byte
	mode     int64
	linkname string
	body     []byte
}

func TestUntarRejectsPathTraversal(t *testing.T) {
	// Every archive member must remain within the destination after path normalization.
	t.Run("parent traversal", func(t *testing.T) {
		// A parent-relative member must be rejected before it can write beside the extraction root.
		root := t.TempDir()
		dest := filepath.Join(root, "project")
		archivePath := writeTestArchive(t, []testTarEntry{{
			name:     "../escaped.txt",
			typeflag: tar.TypeReg,
			mode:     0o644,
			body:     []byte("escaped"),
		}})

		if err := untar(archivePath, dest, false); err == nil {
			t.Error("untar accepted a parent-relative archive member")
		}
		assertPathDoesNotExist(t, filepath.Join(root, "escaped.txt"))
	})

	t.Run("sibling prefix traversal", func(t *testing.T) {
		// A shared prefix is not containment; executable mode keeps an unsafe parent searchable so the bug is not masked.
		root := t.TempDir()
		dest := filepath.Join(root, "project")
		archivePath := writeTestArchive(t, []testTarEntry{{
			name:     "../project-copy/escaped.txt",
			typeflag: tar.TypeReg,
			mode:     0o755,
			body:     []byte("escaped"),
		}})

		if err := untar(archivePath, dest, false); err == nil {
			t.Error("untar accepted traversal into a sibling with a shared path prefix")
		}
		assertPathDoesNotExist(t, filepath.Join(root, "project-copy", "escaped.txt"))
	})

	t.Run("absolute path", func(t *testing.T) {
		// Absolute names must be rejected; a temp target and executable mode make a vulnerable rebase safe to observe.
		root := t.TempDir()
		dest := filepath.Join(root, "project")
		absoluteTarget := filepath.Join(root, "absolute.txt")
		archivePath := writeTestArchive(t, []testTarEntry{{
			name:     filepath.ToSlash(absoluteTarget),
			typeflag: tar.TypeReg,
			mode:     0o755,
			body:     []byte("absolute"),
		}})

		if err := untar(archivePath, dest, false); err == nil {
			t.Error("untar accepted an absolute archive member")
		}
		assertPathDoesNotExist(t, absoluteTarget)
	})
}

func TestUntarRejectsLinkAndUnsupportedEntries(t *testing.T) {
	// Extraction only supports directories and regular files, so special entries must fail closed.
	t.Run("symbolic link", func(t *testing.T) {
		// Rejecting symlinks prevents later members from following a link outside the extraction root.
		root := t.TempDir()
		dest := filepath.Join(root, "project")
		archivePath := writeTestArchive(t, []testTarEntry{{
			name:     "link",
			typeflag: tar.TypeSymlink,
			mode:     0o777,
			linkname: "../outside",
		}})

		if err := untar(archivePath, dest, false); err == nil {
			t.Error("untar accepted a symbolic-link member")
		}
		assertPathDoesNotExist(t, filepath.Join(dest, "link"))
	})

	t.Run("hard link", func(t *testing.T) {
		// Hard links receive the same rejection as symlinks because their target is archive-controlled.
		root := t.TempDir()
		dest := filepath.Join(root, "project")
		archivePath := writeTestArchive(t, []testTarEntry{{
			name:     "hard-link",
			typeflag: tar.TypeLink,
			mode:     0o644,
			linkname: "../outside",
		}})

		if err := untar(archivePath, dest, false); err == nil {
			t.Error("untar accepted a hard-link member")
		}
		assertPathDoesNotExist(t, filepath.Join(dest, "hard-link"))
	})

	t.Run("unsupported type", func(t *testing.T) {
		// A FIFO represents a non-file/non-directory type and must fail closed rather than be ignored.
		root := t.TempDir()
		dest := filepath.Join(root, "project")
		archivePath := writeTestArchive(t, []testTarEntry{{
			name:     "pipe",
			typeflag: tar.TypeFifo,
			mode:     0o600,
		}})

		if err := untar(archivePath, dest, false); err == nil {
			t.Error("untar accepted an unsupported FIFO member")
		}
		assertPathDoesNotExist(t, filepath.Join(dest, "pipe"))
	})
}

func TestUntarRejectsTruncatedArchive(t *testing.T) {
	// This is a valid gzip stream containing a tar header that promises more bytes than it provides.
	root := t.TempDir()
	dest := filepath.Join(root, "project")
	archivePath := filepath.Join(root, "truncated.tar.gz")
	if err := os.WriteFile(archivePath, makeTruncatedTestArchive(t), 0o600); err != nil {
		t.Fatalf("write truncated archive: %v", err)
	}

	if err := untar(archivePath, dest, false); err == nil {
		t.Error("untar treated a truncated regular-file member as complete")
	}
}

func TestUntarRejectsOversizedEntryBeforeWriting(t *testing.T) {
	// The tar header is enough to reject a decompression bomb; extraction must
	// not write a truncated 1 GiB prefix and then continue as if it succeeded.
	root := t.TempDir()
	dest := filepath.Join(root, "project")
	archivePath := filepath.Join(root, "oversized.tar.gz")
	if err := os.WriteFile(archivePath, makeHeaderOnlyTestArchive(t, maxExtractedFileBytes+1), 0o600); err != nil {
		t.Fatalf("write oversized archive: %v", err)
	}

	if err := untar(archivePath, dest, false); err == nil {
		t.Error("untar accepted an entry larger than the extraction limit")
	}
	assertPathDoesNotExist(t, filepath.Join(dest, "oversized.bin"))
}

func TestUntarCreatesSafeImplicitParentsAndPreservesFileMode(t *testing.T) {
	// Archives do not have to contain directory headers. A regular file's 0644
	// mode must not be reused for its parent, which would remove directory search access.
	if os.PathSeparator == '\\' {
		t.Skip("Unix permission bits are not portable to Windows")
	}
	dest := filepath.Join(t.TempDir(), "project")
	archivePath := writeTestArchive(t, []testTarEntry{{
		name:     "nested/config.yaml",
		typeflag: tar.TypeReg,
		mode:     0o640,
		body:     []byte("version: 1\n"),
	}})

	if err := untar(archivePath, dest, false); err != nil {
		t.Fatalf("untar regular file with implicit parent: %v", err)
	}
	info, err := os.Stat(filepath.Join(dest, "nested"))
	if err != nil {
		t.Fatalf("stat implicit parent: %v", err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("implicit parent mode %#o is not searchable", info.Mode().Perm())
	}
	info, err = os.Stat(filepath.Join(dest, "nested", "config.yaml"))
	if err != nil {
		t.Fatalf("stat extracted file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o640 {
		t.Errorf("extracted file mode %#o, want %#o", got, os.FileMode(0o640))
	}
}

func makeTestArchive(t *testing.T, entries []testTarEntry) []byte {
	t.Helper()

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzipWriter)
	for _, entry := range entries {
		header := &tar.Header{
			Name:     entry.name,
			Mode:     entry.mode,
			Size:     int64(len(entry.body)),
			Typeflag: entry.typeflag,
			Linkname: entry.linkname,
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatalf("write tar header %q: %v", entry.name, err)
		}
		if len(entry.body) > 0 {
			if _, err := tarWriter.Write(entry.body); err != nil {
				t.Fatalf("write tar body %q: %v", entry.name, err)
			}
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}
	return buf.Bytes()
}

func writeTestArchive(t *testing.T, entries []testTarEntry) string {
	t.Helper()

	archivePath := filepath.Join(t.TempDir(), "fixture.tar.gz")
	if err := os.WriteFile(archivePath, makeTestArchive(t, entries), 0o600); err != nil {
		t.Fatalf("write archive fixture: %v", err)
	}
	return archivePath
}

func makeTruncatedTestArchive(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzipWriter)
	header := &tar.Header{
		Name:     "partial.txt",
		Mode:     0o644,
		Size:     1024,
		Typeflag: tar.TypeReg,
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatalf("write truncated tar header: %v", err)
	}
	if _, err := tarWriter.Write([]byte("partial")); err != nil {
		t.Fatalf("write truncated tar body: %v", err)
	}
	// Intentionally leave the tar writer open so the promised body and trailer are absent.
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("close truncated gzip stream: %v", err)
	}
	return buf.Bytes()
}

func makeHeaderOnlyTestArchive(t *testing.T, size int64) []byte {
	t.Helper()

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzipWriter)
	if err := tarWriter.WriteHeader(&tar.Header{
		Name:     "oversized.bin",
		Mode:     0o600,
		Size:     size,
		Typeflag: tar.TypeReg,
	}); err != nil {
		t.Fatalf("write oversized tar header: %v", err)
	}
	// The extractor should reject from the header, so no enormous body is needed.
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("close oversized gzip stream: %v", err)
	}
	return buf.Bytes()
}

func assertPathDoesNotExist(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Lstat(path); err == nil {
		t.Errorf("unexpected archive output at %q", path)
	} else if !os.IsNotExist(err) {
		t.Errorf("stat unexpected archive output %q: %v", path, err)
	}
}
