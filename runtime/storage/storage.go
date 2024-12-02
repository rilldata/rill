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
	instanceID   string
}

type gcsBucketConfig struct {
	Bucket          string `mapstructure:"bucket"`
	SecretJSON      string `mapstructure:"google_application_credentials"`
	AllowHostAccess bool   `mapstructure:"allow_host_access"`
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
	c := &Client{
		dataDirPath: dataDir,
	}

	if len(bucketCfg) != 0 {
		gcsBucketConfig := &gcsBucketConfig{}
		if err := mapstructure.WeakDecode(bucketCfg, gcsBucketConfig); err != nil {
			panic(err)
		}
		c.bucketConfig = gcsBucketConfig
	}
	return c
}

func (c *Client) AddInstance(instanceID string) error {
	err := os.Mkdir(filepath.Join(c.dataDirPath, instanceID), os.ModePerm)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("could not create instance directory: %w", err)
	}

	// recreate instance's tmp directory
	tmpDir := filepath.Join(c.dataDirPath, instanceID, "tmp")
	if err := os.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("could not remove instance tmp directory: %w", err)
	}
	if err := os.Mkdir(tmpDir, os.ModePerm); err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}

	return nil
}

func (c *Client) RemoveInstance(instanceID string) error {
	err := os.RemoveAll(filepath.Join(c.dataDirPath, instanceID))
	if err != nil {
		return fmt.Errorf("could not remove instance directory: %w", err)
	}
	return nil
}

func (c *Client) WithPrefix(prefix string) *Client {
	c.instanceID = prefix
	return c
}

func (c *Client) DataDir(elem ...string) string {
	paths := []string{c.dataDirPath}
	if c.instanceID != "" {
		paths = append(paths, c.instanceID)
	}
	paths = append(paths, elem...)
	return filepath.Join(paths...)
}

func (c *Client) TempDir(elem ...string) string {
	paths := []string{c.dataDirPath}
	if c.instanceID != "" {
		paths = append(paths, c.instanceID)
	}
	paths = append(paths, "tmp")
	paths = append(paths, elem...)
	return filepath.Join(paths...)
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
	if c.instanceID != "" {
		prefix = c.instanceID + "/"
	}
	for _, e := range elem {
		prefix = prefix + e + "/"
	}
	if prefix == "" {
		return bucket, true, nil
	}
	return blob.PrefixedBucket(bucket, prefix), true, nil
}

func (c *Client) newGCPClient(ctx context.Context) (*gcp.HTTPClient, error) {
	creds, err := gcputil.Credentials(ctx, c.bucketConfig.SecretJSON, c.bucketConfig.AllowHostAccess)
	if err != nil {
		if !errors.Is(err, gcputil.ErrNoCredentials) {
			return nil, err
		}

		// no credentials set, we try with a anonymous client in case user is trying to access public buckets
		return gcp.NewAnonymousHTTPClient(gcp.DefaultTransport()), nil
	}
	// the token source returned from credentials works for all kind of credentials like serviceAccountKey, credentialsKey etc.
	return gcp.NewHTTPClient(gcp.DefaultTransport(), gcp.CredentialsTokenSource(creds))
}
