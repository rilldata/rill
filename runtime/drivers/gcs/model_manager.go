package gcs

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/gcputil"
	"github.com/rilldata/rill/runtime/pkg/pathutil"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var _ drivers.ModelManager = &Connection{}

func (c *Connection) Rename(ctx context.Context, res *drivers.ModelResult, newName string, env *drivers.ModelEnv) (*drivers.ModelResult, error) {
	return nil, nil
}

func (c *Connection) Exists(ctx context.Context, res *drivers.ModelResult) (bool, error) {
	return true, nil
}

func (c *Connection) Delete(ctx context.Context, res *drivers.ModelResult) error {
	p := &drivers.ObjectStoreModelResultProperties{}
	if err := mapstructure.Decode(res.Properties, p); err != nil {
		return err
	}
	u, err := url.Parse(p.Path)
	if err != nil {
		return err
	}

	base, _ := doublestar.SplitPattern(strings.TrimPrefix(u.Path, "/"))
	return deleteObjectsInPrefix(ctx, c, u.Host, base)
}

func (c *Connection) MergePartitionResults(a, b *drivers.ModelResult) (*drivers.ModelResult, error) {
	propsA := &drivers.ObjectStoreModelResultProperties{}
	if err := mapstructure.Decode(a.Properties, propsA); err != nil {
		return nil, err
	}

	propsB := &drivers.ObjectStoreModelResultProperties{}
	if err := mapstructure.Decode(b.Properties, propsB); err != nil {
		return nil, err
	}

	if propsA.Format != propsB.Format {
		return nil, fmt.Errorf("cannot merge partitioned results that output to different file formats (format %q is not %q)", propsA.Format, propsB.Format)
	}

	// NOTE: This makes an assumption that the common path of the individual partition results only contains data for the model.
	// This is a convenient assumption, but may cause data loss if the common path contains other data.
	// To protect against the most obvious error case, we check that the common path is not the bucket root.

	commonPath := pathutil.CommonPrefix(propsA.Path, propsB.Path)
	if commonPath == "" {
		return nil, fmt.Errorf("cannot merge partitioned results that do not share a common subpath (%q vs. %q)", propsA.Path, propsB.Path)
	}

	p := &drivers.ObjectStoreModelResultProperties{
		Path:   commonPath,
		Format: propsA.Format,
	}

	pm := map[string]any{}
	if err := mapstructure.Decode(p, &pm); err != nil {
		return nil, err
	}

	return &drivers.ModelResult{
		Connector:  a.Connector,
		Properties: pm,
		Table:      "",
	}, nil
}

func deleteObjectsInPrefix(ctx context.Context, c *Connection, bucketName, prefix string) error {
	cred, err := gcputil.Credentials(ctx, c.config.SecretJSON, c.config.AllowHostAccess)
	if err != nil {
		return err
	}
	client, err := storage.NewClient(ctx, option.WithCredentials(cred))
	if err != nil {
		return fmt.Errorf("failed to create GCS client: %w", err)
	}
	defer client.Close()
	bucket := client.Bucket(bucketName)
	it := bucket.Objects(ctx, &storage.Query{Prefix: prefix})
	for {
		objAttrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return fmt.Errorf("error listing objects: %w", err)
		}

		if err := bucket.Object(objAttrs.Name).Delete(ctx); err != nil {
			return fmt.Errorf("failed to delete object %s: %w", objAttrs.Name, err)
		}
	}
	return nil
}
