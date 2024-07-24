package drivers

import "context"

type ObjectStore interface {
	// DownloadFiles provides an iterator for downloading and consuming files
	DownloadFiles(ctx context.Context, src map[string]any) (FileIterator, error)
}

// FileIterator provides ways to iteratively download files from external sources
// Clients should call close once they are done with iterator to release any resources
type FileIterator interface {
	// Close do cleanup and release resources
	Close() error
	// Next returns a list of file downloaded from external sources
	// and cleanups file created in previous batch
	Next() ([]string, error)
	// Size returns size of data downloaded in unit.
	// Returns 0,false if not able to compute size in given unit
	Size(unit ProgressUnit) (int64, bool)
	// Format returns general file format (json, csv, parquet, etc)
	// Returns an empty string if there is no general format
	Format() string
}

type ObjectStoreModelOutputProperties struct {
	Path   string     `mapstructure:"path"`
	Format FileFormat `mapstructure:"format"`
}

type ObjectStoreModelResultProperties struct {
	Path   string     `mapstructure:"path"`
	Format FileFormat `mapstructure:"format"`
}
