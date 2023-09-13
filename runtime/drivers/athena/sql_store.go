package athena

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime/drivers"
	rillblob "github.com/rilldata/rill/runtime/drivers/blob"
)

func (c *Connection) Query(ctx context.Context, props map[string]any) (drivers.RowIterator, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *Connection) QueryAsFiles(ctx context.Context, props map[string]any, opt *drivers.QueryOption, p drivers.Progress) (drivers.FileIterator, error) {
	conf, err := parseSourceProperties(props)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	cfg, err := c.newCfg(ctx)
	if err != nil {
		return nil, err
	}

	client := athena.NewFromConfig(cfg)
	outputLocation, err := resolveOutputLocation(ctx, client, conf)
	if err != nil {
		return nil, err
	}

	// ie
	// outputLocation s3://bucket-name/prefix
	// unloadLocation s3://bucket-name/prefix/rill-connector-parquet-output-<uuid>
	// unloadPath prefix/rill-connector-parquet-output-<uuid>
	unloadFolderName := "parquet_output_" + uuid.New().String()
	bucketName := strings.Split(strings.TrimPrefix(outputLocation, "s3://"), "/")[0]
	unloadLocation := strings.TrimRight(outputLocation, "/") + "/" + unloadFolderName
	unloadPath := strings.TrimPrefix(strings.TrimPrefix(unloadLocation, "s3://"+bucketName), "/")
	err = c.unload(ctx, client, cfg, conf, unloadLocation)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to unload: %w", err), cleanPath(ctx, cfg, bucketName, unloadPath))
	}

	bucketObj, err := c.openBucket(ctx, conf, bucketName)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("cannot open bucket %q: %w", bucketName, err), cleanPath(ctx, cfg, bucketName, unloadPath))
	}

	opts := rillblob.Options{
		GlobPattern: unloadPath + "/**",
	}

	it, err := rillblob.NewIterator(ctx, bucketObj, opts, c.logger)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("cannot download parquet output %q %w", opts.GlobPattern, err), cleanPath(ctx, cfg, bucketName, unloadPath))
	}

	return janitorIterator{
		FileIterator: it,
		ctx:          ctx,
		unloadPath:   unloadPath,
		bucketName:   bucketName,
		cfg:          cfg,
	}, nil
}
