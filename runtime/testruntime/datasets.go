package testruntime

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"github.com/stretchr/testify/require"
)

// Dataset represents a dataset for use in tests.
type Dataset string

const (
	DatasetAdbids          Dataset = "./testdata/adbids/data/AdBids.csv.gz"
	DatasetOpenRTBAuctions Dataset = "https://storage.googleapis.com/rilldata-public/auction_data.parquet"
	DatasetOpenRTBBids     Dataset = "https://storage.googleapis.com/rilldata-public/bids_data.parquet"
)

// DatasetPath returns a local path to the given dataset.
// For remote datasets, it downloads and caches them locally in ./testruntime/testdata/cache.
func DatasetPath(t TestingT, ds Dataset) string {
	// Parse the URL
	u, err := url.Parse(string(ds))
	require.NoError(t, err, "failed to parse dataset URL")

	// If dataset is a local file, return the absolute path
	if u.Scheme == "" {
		p, err := filepath.Abs(u.Path)
		require.NoError(t, err)
		require.FileExists(t, p)
		return p
	}

	// Only support HTTP(s) URLs
	require.True(t, u.Scheme == "http" || u.Scheme == "https")

	// Resolve path to ./testruntime/testdata/cache
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	cacheDir := filepath.Join(filename, "..", "testdata", "cache")
	cacheDir, err = filepath.Abs(cacheDir)
	require.NoError(t, err)
	err = os.MkdirAll(cacheDir, 0o755)
	require.NoError(t, err)

	// Resolve cached file path using the same filename as in the URL
	cachePath := filepath.Join(cacheDir, filepath.Base(u.Path))

	// If the cached file exists, return the cached path
	if _, err := os.Stat(cachePath); err == nil {
		return cachePath
	}

	// Download the dataset from its URL
	resp, err := http.DefaultClient.Get(string(ds))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Download into a temp file first
	tempPath := cachePath + ".tmp"
	out, err := os.Create(tempPath)
	require.NoError(t, err)
	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		os.Remove(tempPath)
		require.NoError(t, err)
	}

	// Move temp file to final location and return it
	err = os.Rename(tempPath, cachePath)
	require.NoError(t, err)
	return cachePath
}
