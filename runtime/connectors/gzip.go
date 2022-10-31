package connectors

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

func extractGzipFile(fileName string) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	uncompressedStream, err := gzip.NewReader(f)
	defer uncompressedStream.Close()

	extension := getFileExtension(fileName)
	tempFile, err := os.CreateTemp(
		os.TempDir(),
		fmt.Sprintf(
			"%s*.%s",
			strings.Replace(fileName, "."+extension, "", 1),
			extension,
		),
	)

	_, err = io.Copy(tempFile, uncompressedStream)
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	return tempFile.Name(), nil
}
