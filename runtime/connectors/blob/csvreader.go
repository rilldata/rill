package blob

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"gocloud.dev/blob"
)

var _newLineSeparator = []byte("\n")

type csvExtractOption struct {
	extractOption *extractOption
	hasHeader     bool // set if first row is header
}

// downloadCSV copies partial data to fw with the assumption that rows are separated by \n
// the data format doesn't necessarily have to be csv
func downloadCSV(ctx context.Context, bucket *blob.Bucket, obj *blob.ListObject, option *csvExtractOption, fw *os.File) error {
	reader := NewBlobObjectReader(ctx, bucket, obj)

	rows, err := csvRows(reader, option)
	if err != nil {
		return err
	}

	// write rows
	for _, r := range rows {
		if _, err := fw.Write(r); err != nil {
			return err
		}
		if _, err := fw.Write(_newLineSeparator); err != nil {
			return err
		}
	}

	return nil
}

func csvRows(reader *ObjectReader, option *csvExtractOption) ([][]byte, error) {
	if option.extractOption.strategy == runtimev1.Source_ExtractPolicy_STRATEGY_TAIL {
		return csvRowsTail(reader, option)
	}
	return csvRowsHead(reader, option.extractOption)
}

func csvRowsTail(reader *ObjectReader, option *csvExtractOption) ([][]byte, error) {
	headerRow := make([]byte, 0)
	bytesToRead := option.extractOption.limitInBytes
	if option.hasHeader {
		// csv has header, need to read header first
		header, err := getHeader(reader)
		if err != nil {
			return nil, err
		}
		headerRow = header
		bytesToRead = option.extractOption.limitInBytes - uint64(len(header))
	}

	if _, err := reader.Seek(0-int64(bytesToRead), io.SeekEnd); err != nil {
		return nil, err
	}

	p := make([]byte, bytesToRead)
	if _, err := reader.Read(p); err != nil {
		return nil, err
	}

	rows := bytes.Split(p, _newLineSeparator)
	// remove first row (possibly incomplete)
	rows = rows[1:]
	// append header at start
	return append([][]byte{headerRow}, rows...), nil
}

func csvRowsHead(reader *ObjectReader, option *extractOption) ([][]byte, error) {
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	p := make([]byte, option.limitInBytes)
	if _, err := reader.Read(p); err != nil {
		return nil, err
	}

	rows := bytes.Split(p, _newLineSeparator)
	// remove last row (possibly incomplete)
	return rows[:len(rows)-1], nil
}

// tries to get csv header from reader by incrmentally reading 1KB bytes
func getHeader(r *ObjectReader) ([]byte, error) {
	fetchLength := 1024
	var p []byte
	for {
		temp := make([]byte, fetchLength)
		n, err := r.Read(temp)
		if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, err
		}

		p = append(p, temp...)
		rows := bytes.Split(p, _newLineSeparator)
		if len(rows) > 1 {
			// complete header found
			return rows[0], nil
		}

		if n < fetchLength {
			// end of csv
			return nil, io.EOF
		}
	}
}
