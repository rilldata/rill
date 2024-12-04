package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/pkg/gcputil"
	"gocloud.dev/blob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/gcp"
)

type Client struct {
	dataDirPath  string
	tempDirPath  string
	bucketConfig *gcsBucketConfig
	prefixes     []string
}

func New(dataDir string, bucketCfg map[string]any) (*Client, error) {
	tempDirPath, err := os.MkdirTemp("", "rill")
	if err != nil {
		return nil, err
	}
	c := &Client{
		dataDirPath: dataDir,
		tempDirPath: tempDirPath,
	}

	if len(bucketCfg) != 0 {
		gcsBucketConfig := &gcsBucketConfig{}
		err := mapstructure.WeakDecode(bucketCfg, gcsBucketConfig)
		if err != nil {
			return nil, err
		}
		c.bucketConfig = gcsBucketConfig
	}
	return c, nil
}

func MustNew(dataDir string, bucketCfg map[string]any) *Client {
	c, err := New(dataDir, bucketCfg)
	if err != nil {
		panic(err)
	}
	return c
}

func (c *Client) WithPrefix(prefix ...string) *Client {
	newClient := &Client{
		dataDirPath:  c.dataDirPath,
		bucketConfig: c.bucketConfig,
	}
	newClient.prefixes = append(newClient.prefixes, c.prefixes...)
	newClient.prefixes = append(newClient.prefixes, prefix...)
	return newClient
}

func (c *Client) RemovePrefix(ctx context.Context, prefix ...string) error {
	if c.prefixes != nil {
		return fmt.Errorf("storage: RemovePrefix is not supported for prefixed client")
	}

	// clean data dir
	removeErr := os.RemoveAll(c.path(c.dataDirPath, prefix...))

	// clean temp dir
	removeErr = errors.Join(removeErr, os.RemoveAll(c.path(c.tempDirPath, prefix...)))

	// clean bucket
	bkt, ok, err := c.OpenBucket(ctx, prefix...)
	if err != nil {
		return errors.Join(removeErr, err)
	}
	if !ok {
		return removeErr
	}
	defer bkt.Close()

	iter := bkt.List(&blob.ListOptions{})
	for {
		obj, err := iter.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return errors.Join(removeErr, err)
		}
		err = bkt.Delete(ctx, obj.Key)
		if err != nil {
			return errors.Join(removeErr, err)
		}
	}
	return removeErr
}

func (c *Client) DataDir(elem ...string) (string, error) {
	path := c.path(c.dataDirPath, elem...)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (c *Client) TempDir(elem ...string) (string, error) {
	path := c.path(c.tempDirPath, elem...)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (c *Client) RandomTempDir(pattern string, elem ...string) (string, error) {
	path, err := c.TempDir(elem...)
	if err != nil {
		return "", err
	}
	path, err = os.MkdirTemp(path, pattern)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (c *Client) OpenBucket(ctx context.Context, elem ...string) (*blob.Bucket, bool, error) {
	if c.bucketConfig == nil {
		return nil, false, nil
	}
	// Init dataBucket
	client, err := c.newGCPClient(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("could not create GCP client: %w", err)
	}

	bucket, err := gcsblob.OpenBucket(ctx, client, c.bucketConfig.Bucket, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to open bucket %q: %w", c.bucketConfig.Bucket, err)
	}
	var prefix string
	for _, p := range c.prefixes {
		prefix = prefix + p + "/"
	}
	for _, e := range elem {
		prefix = prefix + e + "/"
	}
	if prefix == "" {
		return bucket, true, nil
	}
	return blob.PrefixedBucket(bucket, prefix), true, nil
}

func (c *Client) path(base string, elem ...string) string {
	paths := []string{base}
	if c.prefixes != nil {
		paths = append(paths, c.prefixes...)
	}
	paths = append(paths, elem...)
	return filepath.Join(paths...)
}

func (c *Client) newGCPClient(ctx context.Context) (*gcp.HTTPClient, error) {
	creds, err := gcputil.Credentials(ctx, c.bucketConfig.GoogleApplicationCredentialsJSON, false)
	if err != nil {
		return nil, err
	}
	// the token source returned from credentials works for all kind of credentials like serviceAccountKey, credentialsKey etc.
	return gcp.NewHTTPClient(gcp.DefaultTransport(), gcp.CredentialsTokenSource(creds))
}

type gcsBucketConfig struct {
	Bucket                           string `mapstructure:"bucket"`
	GoogleApplicationCredentialsJSON string `mapstructure:"google_application_credentials_json"`
}
