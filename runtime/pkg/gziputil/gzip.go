package gziputil

import (
	"bytes"
	"compress/gzip"
	"io"
)

// GZipCompress compress the input bytes using gzip.
func GZipCompress(v []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, err := w.Write(v)
	if err != nil {
		_ = w.Close()
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// GZipDecompress decompresses the input bytes using gzip.
func GZipDecompress(v []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(v))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}
