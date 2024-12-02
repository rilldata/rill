package storage

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
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
	bucketConfig *gcsBucketConfig
	prefixes     []string
}

func New(dataDir string, bucketCfg map[string]any) (*Client, error) {
	c := &Client{
		dataDirPath: dataDir,
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

func (c *Client) DataDir(elem ...string) string {
	paths := []string{c.dataDirPath}
	if c.prefixes != nil {
		paths = append(paths, c.prefixes...)
	}
	paths = append(paths, elem...)
	return filepath.Join(paths...)
}

func (c *Client) TempDir(elem ...string) string {
	paths := []string{c.dataDirPath}
	if c.prefixes != nil {
		paths = append(paths, c.prefixes...)
	}
	paths = append(paths, "tmp")
	paths = append(paths, elem...)
	return filepath.Join(paths...)
}

func (c *Client) OpenBucket(ctx context.Context, elem ...string) (*blob.Bucket, bool, error) {
	if len(c.bucketConfig) == 0 {
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

func AddInstance(c *Client, instanceID string) error {
	if c.prefixes != nil {
		return fmt.Errorf("storage: should not call AddInstance with prefixed client")
	}

	c = c.WithPrefix(instanceID)
	err := os.Mkdir(c.DataDir(), os.ModePerm)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("could not create instance directory: %w", err)
	}

	// recreate instance's tmp directory
	tmpDir := c.TempDir()
	if err := os.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("could not remove instance tmp directory: %w", err)
	}
	if err := os.Mkdir(tmpDir, os.ModePerm); err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}

	return nil
}

func RemoveInstance(c *Client, instanceID string) error {
	if c.prefixes != nil {
		return fmt.Errorf("storage: should not call RemoveInstance with prefixed client")
	}

	err := os.RemoveAll(c.DataDir())
	if err != nil {
		return fmt.Errorf("could not remove instance directory: %w", err)
	}
	return nil
}

func (c *Client) newGCPClient(ctx context.Context) (*gcp.HTTPClient, error) {
	creds, err := gcputil.Credentials(ctx, c.bucketConfig.SecretJSON, false)
	if err != nil {
		return nil, err
	}
	// the token source returned from credentials works for all kind of credentials like serviceAccountKey, credentialsKey etc.
	return gcp.NewHTTPClient(gcp.DefaultTransport(), gcp.CredentialsTokenSource(creds))
}

type gcsBucketConfig struct {
	Bucket     string `mapstructure:"bucket"`
	GoogleApplicationCredentialsJSON string `mapstructure:"google_application_credentials_json"`
}
