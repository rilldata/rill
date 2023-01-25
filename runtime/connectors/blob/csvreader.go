package blob

import (
	"context"
	"io"
	"os"
	"strings"

	"gocloud.dev/blob"
)

// tries to get csv header from reader by incrmentally reading 1KB bytes
func getHeader(r *blobObjectReader) (string, error) {
	fetchLength := 1024
	p := make([]byte, 0)
	for {
		temp := make([]byte, fetchLength)
		n, err := r.Read(temp)
		if err != nil && !strings.Contains(err.Error(), "EOF") {
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

	var seekError error
	var rows []string
	if option.Strategy == TAIL {
		header, err := getHeader(reader)
		if err != nil {
			return err
		}

		// need to write header first in case strategy is tail
		if _, err := fw.WriteString(header); err != nil {
			return err
		}

		remBytes := option.Size - int64(len([]byte(header)))
		_, seekError = reader.Seek(0-remBytes, io.SeekEnd)
		if seekError != nil {
			return seekError
		}

		p := make([]byte, remBytes)
		if _, err := reader.Read(p); err != nil {
			return err
		}

		rows = strings.Split(string(p), "\n")
		// remove first row (possibly incomplete)
		rows = rows[1:]
	} else { // HEAD strategy
		_, seekError = reader.Seek(0, io.SeekStart)
		if seekError != nil {
			return seekError
		}

		p := make([]byte, option.Size)
		if _, err := reader.Read(p); err != nil {
			return err
		}

		rows = strings.Split(string(p), "\n")
		// remove last row (possibly incomplete)
		rows = rows[:len(rows)-1]
	}

	// write remaining rows
	for _, r := range rows {
		if _, err := fw.WriteString("\n"); err != nil {
			return err
		}

		if _, err := fw.WriteString(r); err != nil {
			return err
		}
	}

	return nil
}
