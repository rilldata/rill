package blob

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"gocloud.dev/blob"
)

var _newLineSeparator = []byte("\n")

type textExtractOption struct {
	extractOption *extractOption
	hasCSVHeader  bool // set if first row is header
}

// downloadText copies partial data to fw with the assumption that rows are separated by \n
// the data format doesn't necessarily have to be csv
func downloadText(ctx context.Context, bucket *blob.Bucket, obj *blob.ListObject, option *textExtractOption, fw *os.File) error {
	reader := NewBlobObjectReader(ctx, bucket, obj)

	rows, err := rows(reader, option)
	if err != nil {
		return err
	}

	_, err = fw.Write(rows)
	return err
}

func rows(reader *ObjectReader, option *textExtractOption) ([]byte, error) {
	switch option.extractOption.strategy {
	case runtimev1.Source_ExtractPolicy_STRATEGY_HEAD:
		return rowsHead(reader, option.extractOption)
	case runtimev1.Source_ExtractPolicy_STRATEGY_TAIL:
		return rowsTail(reader, option)
	default:
		panic(fmt.Sprintf("unsupported strategy %s", option.extractOption.strategy))
	}
}

func rowsTail(reader *ObjectReader, option *textExtractOption) ([]byte, error) {
	header := make([]byte, 0)
	if option.hasCSVHeader {
		// csv has header, need to read header first
		headerRow, err := getHeader(reader)
		if err != nil {
			return nil, err
		}

		headerRow = append(headerRow, _newLineSeparator...)
		header = headerRow
	}

	bytesToRead := option.extractOption.limitInBytes - uint64(len(header))
	if _, err := reader.Seek(0-int64(bytesToRead), io.SeekEnd); err != nil {
		return nil, err
	}

	p := make([]byte, bytesToRead)
	_, err := reader.Read(p)
	if err := unsucessfullError(err); err != nil {
		return nil, err
	}

	lastLineIndex := bytes.Index(p, _newLineSeparator)
	// remove data before \n since its possibly incomplete
	// append header at start
	return append(header, p[lastLineIndex+1:]...), nil
}

func rowsHead(reader *ObjectReader, option *extractOption) ([]byte, error) {
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	p := make([]byte, option.limitInBytes)
	_, err := reader.Read(p)
	if err := unsucessfullError(err); err != nil {
		return nil, err
	}

	lastLineIndex := bytes.LastIndex(p, _newLineSeparator)
	if lastLineIndex == -1 {
		// data can still be complete in case there is a single row without any newline delimitter
		// let ingestion system decide
		return p, nil
	}
	// remove data after \n since its incomplete
	return p[:lastLineIndex+1], nil
}

// tries to get csv header from reader by incrmentally reading 1KB bytes
func getHeader(r *ObjectReader) ([]byte, error) {
	fetchLength := 1024
	var p []byte
	for {
		temp := make([]byte, fetchLength)
		n, err := r.Read(temp)
		if err := unsucessfullError(err); err != nil {
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

// unsucessfullError silents the io.EOF and io.ErrUnexpectedEOF
// the reader.Read can succeed as well as return the two errors in case more data is requested than what is present
func unsucessfullError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return nil
	}
	return err
}
