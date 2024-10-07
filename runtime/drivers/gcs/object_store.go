package gcs

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"

	"github.com/c2h5oh/datasize"
	"github.com/rilldata/rill/runtime/drivers"
	rillblob "github.com/rilldata/rill/runtime/drivers/blob"
	"gocloud.dev/blob/gcsblob"
	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi"
)

// ListObjects implements drivers.ObjectStore.
func (c *Connection) ListObjects(ctx context.Context, propsMap map[string]any) ([]drivers.ObjectStoreEntry, error) {
	props, err := parseSourceProperties(propsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}

	gcsBucket, err := gcsblob.OpenBucket(ctx, client, props.url.Host, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket %q, %w", props.url.Host, err)
	}

	bucket, err := rillblob.NewBucket(gcsBucket, c.logger)
	if err != nil {
		return nil, err
	}
	defer bucket.Close()

	return bucket.ListObjects(ctx, props.url.Path)
}

// DownloadFiles returns a file iterator over objects stored in gcs.
// The credential json is read from config google_application_credentials.
// Additionally in case `allow_host_credentials` is true it looks for "Application Default Credentials" as well
func (c *Connection) DownloadFiles(ctx context.Context, props map[string]any) (drivers.FileIterator, error) {
	conf, err := parseSourceProperties(props)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}

	bucketObj, err := gcsblob.OpenBucket(ctx, client, conf.url.Host, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket %q, %w", conf.url.Host, err)
	}

	var batchSize datasize.ByteSize
	if conf.BatchSize == "-1" {
		batchSize = math.MaxInt64 // download everything in one batch
	} else {
		batchSize, err = datasize.ParseString(conf.BatchSize)
		if err != nil {
			return nil, err
		}
	}
	// prepare fetch configs
	opts := rillblob.Options{
		GlobMaxTotalSize:      conf.GlobMaxTotalSize,
		GlobMaxObjectsMatched: conf.GlobMaxObjectsMatched,
		GlobMaxObjectsListed:  conf.GlobMaxObjectsListed,
		GlobPageSize:          conf.GlobPageSize,
		GlobPattern:           conf.url.Path,
		ExtractPolicy:         conf.extractPolicy,
		BatchSizeBytes:        int64(batchSize.Bytes()),
		KeepFilesUntilClose:   conf.BatchSize == "-1",
		TempDir:               c.config.TempDir,
	}

	iter, err := rillblob.NewIterator(ctx, bucketObj, opts, c.logger)
	if err != nil {
		// in cases when no creds are passed
		apiError := &googleapi.Error{}
		if errors.As(err, &apiError) && apiError.Code == http.StatusUnauthorized {
			return nil, drivers.NewPermissionDeniedError(fmt.Sprintf("can't access remote err: %v", apiError))
		}

		// StatusUnauthorized when incorrect key is passsed
		// StatusBadRequest when key doesn't have a valid credentials file
		retrieveError := &oauth2.RetrieveError{}
		if errors.As(err, &retrieveError) && (retrieveError.Response.StatusCode == http.StatusUnauthorized || retrieveError.Response.StatusCode == http.StatusBadRequest) {
			return nil, drivers.NewPermissionDeniedError(fmt.Sprintf("can't access remote err: %v", retrieveError))
		}

		return nil, err
	}

	return iter, nil
}
