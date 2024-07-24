package s3

import (
	"context"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

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

	creds, err := c.getCredentials()
	if err != nil {
		return err
	}

	session, err := c.getAwsSessionConfig(ctx, &sourceProperties{}, u.Host, creds)
	if err != nil {
		return err
	}
	base, _ := doublestar.SplitPattern(strings.TrimPrefix(u.Path, "/"))
	return deleteObjectsInPrefix(ctx, session, u.Host, base)
}

func deleteObjectsInPrefix(ctx context.Context, sess *session.Session, bucketName, prefix string) error {
	s3client := s3.New(sess)
	deleteBatch := func(objects []*s3.ObjectIdentifier) error {
		_, err := s3client.DeleteObjectsWithContext(ctx, &s3.DeleteObjectsInput{
			Bucket: &bucketName,
			Delete: &s3.Delete{
				Objects: objects,
			},
		})
		return err
	}

	var continuationToken *string
	for {
		out, err := s3client.ListObjectsV2WithContext(ctx, &s3.ListObjectsV2Input{
			Bucket:            &bucketName,
			Prefix:            &prefix,
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return err
		}

		ids := make([]*s3.ObjectIdentifier, 0, len(out.Contents))
		for _, o := range out.Contents {
			ids = append(ids, &s3.ObjectIdentifier{
				Key: o.Key,
			})
		}

		if len(ids) > 0 {
			if err := deleteBatch(ids); err != nil {
				return err
			}
		}

		if *out.IsTruncated && out.NextContinuationToken != nil {
			continuationToken = out.NextContinuationToken
		} else {
			break
		}
	}
	return nil
}
