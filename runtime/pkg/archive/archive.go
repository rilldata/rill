package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/c2h5oh/datasize"
	"github.com/rilldata/rill/runtime/drivers"
)

const (
	maxArchiveBytes       = int64(100 * 1024 * 1024)
	maxExtractedFileBytes = int64(datasize.GB)
)

var ignoreFileList = []string{
	"/.env",
	"/.git",
}

func Download(ctx context.Context, downloadURL, downloadDst, projPath string, clean, ignorePaths bool) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, http.NoBody)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: status code %d", resp.StatusCode)
	}
	if resp.ContentLength > maxArchiveBytes {
		return fmt.Errorf("failed to download file: archive exceeds %d byte limit", maxArchiveBytes)
	}

	out, err := os.Create(downloadDst)
	if err != nil {
		return err
	}
	defer os.Remove(downloadDst)

	// Read one byte beyond the limit so chunked responses cannot bypass the
	// Content-Length check.
	written, err := io.Copy(out, io.LimitReader(resp.Body, maxArchiveBytes+1))
	if err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if written > maxArchiveBytes {
		return fmt.Errorf("failed to download file: archive exceeds %d byte limit", maxArchiveBytes)
	}

	if !clean {
		return untar(downloadDst, filepath.Clean(projPath), ignorePaths)
	}

	// Validate and extract beside the destination before touching the live
	// project. Keeping staging on the same filesystem also makes the final rename
	// atomic for readers that observe the project path.
	projPath = filepath.Clean(projPath)
	parent := filepath.Dir(projPath)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}
	staging, err := os.MkdirTemp(parent, ".rill-project-staging-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staging)
	if err := untar(downloadDst, staging, ignorePaths); err != nil {
		return err
	}
	return replaceDirectory(staging, projPath)
}

func replaceDirectory(staging, dest string) error {
	_, err := os.Lstat(dest)
	if errors.Is(err, os.ErrNotExist) {
		return os.Rename(staging, dest)
	}
	if err != nil {
		return err
	}

	// Reserve a unique sibling name, then remove the placeholder so Rename can
	// move the existing project there without overwriting another path.
	backupFile, err := os.CreateTemp(filepath.Dir(dest), ".rill-project-backup-*")
	if err != nil {
		return err
	}
	backup := backupFile.Name()
	if err := backupFile.Close(); err != nil {
		return err
	}
	if err := os.Remove(backup); err != nil {
		return err
	}
	if err := os.Rename(dest, backup); err != nil {
		return err
	}
	if err := os.Rename(staging, dest); err != nil {
		rollbackErr := os.Rename(backup, dest)
		return errors.Join(err, rollbackErr)
	}
	return os.RemoveAll(backup)
}

func CreateAndUpload(ctx context.Context, files []drivers.DirEntry, root, url string, headers map[string]string) error {
	// generate a tar ball
	b := &bytes.Buffer{}

	err := createTar(b, files, root)
	if err != nil {
		return err
	}

	err = uploadTarBall(ctx, url, b, headers)
	if err != nil {
		return err
	}

	return nil
}

func Create(ctx context.Context, files []drivers.DirEntry, root string) (*bytes.Buffer, error) {
	b := &bytes.Buffer{}

	err := createTar(b, files, root)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func Upload(ctx context.Context, url string, body io.Reader, headers map[string]string) error {
	err := uploadTarBall(ctx, url, body, headers)
	if err != nil {
		return err
	}

	return nil
}

// borrowed from https://github.com/goreleaser/goreleaser/blob/main/pkg/archive/tar/tar.go with minor changes
func createTar(writer io.Writer, files []drivers.DirEntry, root string) error {
	gw, err := gzip.NewWriterLevel(writer, gzip.BestCompression)
	if err != nil {
		return err
	}
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	for _, entry := range files {
		if drivers.IsIgnored(entry.Path, ignoreFileList) {
			continue
		}
		fullPath := filepath.Join(root, entry.Path)
		info, err := os.Lstat(fullPath)
		if err != nil {
			return fmt.Errorf("%s: %w", fullPath, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%s: repo contains symlinks", entry.Path)
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("%s: %w", fullPath, err)
		}
		header.Name = entry.Path
		if err = tw.WriteHeader(header); err != nil {
			return fmt.Errorf("%s: %w", fullPath, err)
		}
		if info.IsDir() {
			continue
		}

		file, err := os.Open(fullPath)
		if err != nil {
			return fmt.Errorf("%s: %w", fullPath, err)
		}
		if _, err := io.Copy(tw, file); err != nil {
			file.Close()
			return fmt.Errorf("%s: %w", fullPath, err)
		}
		file.Close()
	}
	return nil
}

func untar(src, dest string, ignorePaths bool) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()
	tarReader := tar.NewReader(gz)
	var directoryModes []struct {
		path string
		mode os.FileMode
	}
	for {
		header, err := tarReader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break // End of tar archive
			}
			return err
		}

		// Determine the proper path for the item
		target, err := sanitizeArchivePath(dest, header.Name)
		if err != nil {
			return err
		}

		// nolint:gosec // sanitizeArchivePath checks for GSC-G305 and throws error but linter cannot know this
		if ignorePaths && drivers.IsIgnored(filepath.Join(string(filepath.Separator), header.Name), nil) {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Keep directories writable during extraction, then restore their
			// archived modes after all children have been created.
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			directoryModes = append(directoryModes, struct {
				path string
				mode os.FileMode
			}{path: target, mode: header.FileInfo().Mode()})
		case tar.TypeReg:
			if header.Size > maxExtractedFileBytes {
				return fmt.Errorf("archive entry %q exceeds %d byte extraction limit", header.Name, maxExtractedFileBytes)
			}
			// Implicit parent directories need execute bits independent of the
			// file's archived mode, otherwise a 0644 file creates an unusable 0644 directory.
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, header.FileInfo().Mode())
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(outFile, tarReader)
			closeErr := outFile.Close()
			if copyErr != nil || closeErr != nil {
				return errors.Join(copyErr, closeErr)
			}
			// Chmod after creation so the process umask does not silently strip
			// executable or group permission bits stored in the archive.
			if err := os.Chmod(target, header.FileInfo().Mode()); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported header type: %c", header.Typeflag)
		}
	}
	for i := len(directoryModes) - 1; i >= 0; i-- {
		if err := os.Chmod(directoryModes[i].path, directoryModes[i].mode); err != nil {
			return err
		}
	}
	return nil
}

func sanitizeArchivePath(dest, tarPath string) (v string, err error) {
	if tarPath == "" || filepath.IsAbs(tarPath) || filepath.VolumeName(tarPath) != "" {
		return "", fmt.Errorf("%s: %s", "content filepath is tainted", tarPath)
	}

	dest, err = filepath.Abs(filepath.Clean(dest))
	if err != nil {
		return "", err
	}
	v = filepath.Join(dest, filepath.Clean(tarPath))
	rel, err := filepath.Rel(dest, v)
	if err != nil {
		return "", err
	}
	if rel != ".." && !filepath.IsAbs(rel) && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return v, nil
	}

	return "", fmt.Errorf("%s: %s", "content filepath is tainted", tarPath)
}

func uploadTarBall(ctx context.Context, url string, body io.Reader, headers map[string]string) error {
	// Create a put request
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to upload file: status code %d, response %s", resp.StatusCode, string(body))
	}
	return nil
}
