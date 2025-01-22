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

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/c2h5oh/datasize"
	"github.com/rilldata/rill/runtime/drivers"
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
	if resp.StatusCode != http.StatusOK {
		// return ghinstallation.HTTPError for outer retry to not retry on 404
		return &ghinstallation.HTTPError{Response: resp}
	}
	defer resp.Body.Close()

	out, err := os.Create(downloadDst)
	if err != nil {
		return err
	}
	defer os.Remove(downloadDst)

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		out.Close()
		return err
	}
	out.Close()

	// clean the projPath first to remove any files from previous download
	if clean {
		_ = os.RemoveAll(projPath)
	}

	// untar to the project path
	err = untar(downloadDst, filepath.Clean(projPath), ignorePaths)
	if err != nil {
		return err
	}
	return nil
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
	for {
		header, err := tarReader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break // End of tar archive
			}
			return err
		}

		// nolint:gosec // adding the '..' check here is still triggering gosec for GSC-G305
		if strings.Contains(header.Name, "..") ||
			(ignorePaths && drivers.IsIgnored(filepath.Join(string(filepath.Separator), header.Name), nil)) {
			continue
		}

		// Determine the proper path for the item
		target, err := sanitizeArchivePath(dest, header.Name)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Handle directory
			if err := os.MkdirAll(target, header.FileInfo().Mode()); err != nil {
				return err
			}
		case tar.TypeReg:
			// Handle regular file
			if err := os.MkdirAll(filepath.Dir(target), header.FileInfo().Mode()); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			// Setting a limit of 1GB to avoid G110: Potential DoS vulnerability via decompression bomb
			// The max file size allowed via upload path is 100MB. Assume that 100MB tar file can't be decompressed to more than 1GB.
			_, err = io.CopyN(outFile, tarReader, int64(datasize.GB))
			if err != nil && !errors.Is(err, io.EOF) {
				outFile.Close()
				return err
			}
			outFile.Close()
		default:
			return fmt.Errorf("unsupported header type: %c", header.Typeflag)
		}
	}
	return nil
}

func sanitizeArchivePath(dest, tarPath string) (v string, err error) {
	v = filepath.Join(dest, tarPath)
	if strings.HasPrefix(v, dest) {
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
