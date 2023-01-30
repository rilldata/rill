package blob

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/c2h5oh/datasize"
	"gocloud.dev/blob"
)

// ObjectReader reads range of bytes from cloud objects
// implements io.ReaderAt and io.Seeker interfaces
type ObjectReader struct {
	ctx    context.Context
	bucket *blob.Bucket
	index  int64
	obj    *blob.ListObject

	// debug data
	debugMode bool
	call      int64
	bytes     int64
}

func (f *ObjectReader) ReadAt(p []byte, off int64) (int, error) {
	if f.debugMode {
		fmt.Printf("reading %v bytes at offset %v\n", len(p), off)
		atomic.AddInt64(&f.call, 1)
	}

	reader, err := f.bucket.NewRangeReader(f.ctx, f.obj.Key, off, int64(len(p)), nil)
	if err != nil {
		return 0, err
	}
	defer reader.Close()

	n, err := io.ReadFull(reader, p)
	if err != nil {
		return n, err
	}
	if f.debugMode {
		atomic.AddInt64(&f.bytes, int64(n))
	}
	return n, nil
}

func (f *ObjectReader) Read(p []byte) (int, error) {
	n, err := f.ReadAt(p, f.index)
	f.index += int64(n)
	return n, err
}

func (f *ObjectReader) Size() int64 {
	return f.obj.Size
}

func (f *ObjectReader) Close() error {
	if f.debugMode {
		bytes := datasize.ByteSize(f.bytes)
		fmt.Printf("made %v calls data fetched %v \n", f.call, bytes.HumanReadable())
	}
	return nil
}

func (f *ObjectReader) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = f.index + offset
	case io.SeekEnd:
		abs = f.Size() + offset
	default:
		return 0, errors.New("bytes.Reader.Seek: invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("bytes.Reader.Seek: negative position")
	}
	f.index = abs

	return abs, nil
}

func NewBlobObjectReader(ctx context.Context, bucket *blob.Bucket, obj *blob.ListObject) *ObjectReader {
	return &ObjectReader{
		ctx:       ctx,
		bucket:    bucket,
		obj:       obj,
		debugMode: true,
	}
}
