package blob

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"

	"gocloud.dev/blob"
)

func getHeader(r *blobObjectReader) (string, error) {
	fetchLength := 1024
	p := make([]byte, 0)
	for {
		temp := make([]byte, fetchLength)
		n, err := r.Read(temp)
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}

		p = append(p, temp...)
		rows := strings.Split(string(p), "\n")
		if len(rows) > 1 {
			// complete header found
			return rows[0], nil
		}

		if n < fetchLength {
			// end of csv
			return "", io.EOF
		}
	}
}

// todo :: check if string conversions can be avoided
func DownloadCSV(ctx context.Context, bucket *blob.Bucket, obj *blob.ListObject, option ExtractConfig, fw *os.File) error {
	reader := newBlobObjectReader(ctx, bucket, obj)
	header, err := getHeader(reader)
	if err != nil {
		return err
	}

	// need to write header first in case strategy is tail
	if _, err := fw.WriteString(header); err != nil {
		return err
	}

	if option.Strategy == TAIL {
		_, err = reader.Seek(0-option.Size, os.SEEK_END)
		if err != nil {
			return err
		}
	}

	p := make([]byte, option.Size)
	if _, err := reader.Read(p); err != nil {
		return err
	}

	rows := strings.Split(string(p), "\n")

	// remove first row (header) and last row(possibly incomplete)
	rows = rows[1 : len(rows)-1]

	for _, r := range rows {
		if _, err := fw.WriteString(r); err != nil {
			return err
		}
	}

	return nil
}
