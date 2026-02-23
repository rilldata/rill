package google_analytics

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/rilldata/rill/runtime/drivers"
)

var _ drivers.FileIterator = &fileIterator{}

// fileIterator iterates over pre-fetched CSV temp files from the GA4 API.
type fileIterator struct {
	files              []string
	nextIndex          int
	done               bool
	keepFilesUntilClose bool
	tempFilePaths      []string
}

// Close implements drivers.FileIterator.
func (it *fileIterator) Close() error {
	for _, p := range it.tempFilePaths {
		os.Remove(p)
	}
	// Also clean up any remaining unread files
	for i := it.nextIndex; i < len(it.files); i++ {
		os.Remove(it.files[i])
	}
	it.tempFilePaths = nil
	it.files = nil
	return nil
}

// Format implements drivers.FileIterator.
func (it *fileIterator) Format() string {
	return "csv"
}

// SetKeepFilesUntilClose implements drivers.FileIterator.
func (it *fileIterator) SetKeepFilesUntilClose() {
	it.keepFilesUntilClose = true
}

// Next implements drivers.FileIterator.
func (it *fileIterator) Next(ctx context.Context) ([]string, error) {
	if it.done || it.nextIndex >= len(it.files) {
		return nil, io.EOF
	}

	// Clean up previously returned files unless keeping until close
	if !it.keepFilesUntilClose && len(it.tempFilePaths) > 0 {
		for _, p := range it.tempFilePaths {
			if err := os.Remove(p); err != nil {
				return nil, fmt.Errorf("failed to delete temp file: %w", err)
			}
		}
		it.tempFilePaths = nil
	}

	filePath := it.files[it.nextIndex]
	it.nextIndex++
	it.tempFilePaths = append(it.tempFilePaths, filePath)

	return []string{filePath}, nil
}
