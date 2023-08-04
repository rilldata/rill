package drivers

import "context"

type ObjectStore interface {
	// DownloadFiles provides an iterator for downloading and consuming files
	DownloadFiles(ctx context.Context, src *BucketSource) (FileIterator, error)
}

// FileIterator provides ways to iteratively download files from external sources
// Clients should call close once they are done with iterator to release any resources
type FileIterator interface {
	// Close do cleanup and release resources
	Close() error
	// NextBatch returns a list of file downloaded from external sources
	// and cleanups file created in previous batch
	NextBatch(limit int) ([]string, error)
	// HasNext can be utlisied to check if iterator has more elements left
	HasNext() bool
	// Size returns size of data downloaded in unit.
	// Returns 0,false if not able to compute size in given unit
	Size(unit ProgressUnit) (int64, bool)
	// KeepFilesUntilClose marks the iterator to keep the files until close is called.
	// This is used when the entire list of files is used at once in certain cases.
	KeepFilesUntilClose(keepFilesUntilClose bool)
}
