package archive

import (
	"archive/tar"
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
)

func Download(ctx context.Context, downloadUrl, downloadDst, projPath string, clean bool) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadUrl, http.NoBody)
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
	err = untar(downloadDst, filepath.Clean(projPath))
	if err != nil {
		return err
	}
	return nil
}

func untar(src, dest string) error {
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
