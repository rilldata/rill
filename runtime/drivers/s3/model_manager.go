package s3

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pathutil"
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
	client, err := getS3Client(ctx, c.config, u.Host)
	if err != nil {
		return err
	}

	base, _ := doublestar.SplitPattern(strings.TrimPrefix(u.Path, "/"))
	return deleteObjectsInPrefix(ctx, client, u.Host, base)
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

func deleteObjectsInPrefix(ctx context.Context, client *s3.Client, bucketName, prefix string) error {
	deleteBatch := func(objects []types.ObjectIdentifier) error {
		_, err := client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &types.Delete{
				Objects: objects,
			},
		})
		return err
	}

	var continuationToken *string
	for {
		out, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucketName),
			Prefix:            aws.String(prefix),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return err
		}

		ids := make([]types.ObjectIdentifier, 0, len(out.Contents))
		for _, o := range out.Contents {
			ids = append(ids, types.ObjectIdentifier{
				Key: o.Key,
			})
		}

		if len(ids) > 0 {
			if err := deleteBatch(ids); err != nil {
				return err
			}
		}

		if out.IsTruncated != nil && *out.IsTruncated && out.NextContinuationToken != nil {
			continuationToken = out.NextContinuationToken
		} else {
			break
		}
	}
	return nil
}
