package iceberg

import (
	"context"
	"github.com/apache/iceberg-go/catalog"
	"github.com/apache/iceberg-go/io"
	"github.com/apache/iceberg-go/table"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/ptr"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"net/url"
	"strings"
)

func (c Connection) AsFileStore() (drivers.FileStore, bool) {
	return c, true
}

func (c Connection) FilePaths(ctx context.Context, src map[string]any) ([]string, error) {

	srcProps, _ := parseSourceProperties(src)

	files, err := c.readIceberg(ctx, srcProps)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(*c.awsConfig())

	var out []string

	for _, file := range files {

		s3Url, _ := url.Parse(file)

		inp := s3.GetObjectInput{
			Bucket: ptr.String(s3Url.Host),
			Key:    ptr.String(strings.Replace(s3Url.Path, "/", "", 1)),
		}

		resp, err := client.GetObject(ctx, &inp)
		if err != nil {
			return nil, err
		}

		tmpFile, _, _ := fileutil.CopyToTempFile(resp.Body, "", ".parquet")

		out = append(out, tmpFile)
	}

	return out, nil
}

func parseSourceProperties(props map[string]any) (sourceProperties, error) {

	conf := &sourceProperties{}
	_ = mapstructure.WeakDecode(props, conf)

	// could imagine some validation

	return *conf, nil
}

type sourceProperties struct {
	Warehouse string `mapstructure:"warehouse"`
	Database  string `mapstructure:"database"`
	Table     string `mapstructure:"table"`
}

// Load the table via glue catalog
// Read the current snapshot
// Get the manifests for that snapshot
// Get all the paths for each manifest
func (c Connection) readIceberg(ctx context.Context, src sourceProperties) ([]string, error) {

	glue := catalog.NewGlueCatalog(catalog.WithAwsConfig(*c.awsConfig()))

	identifier := table.Identifier{src.Database, src.Table}

	// TODO does this need props
	icebergTable, err := glue.LoadTable(ctx, identifier, nil)
	if err != nil {
		return nil, err
	}

	// TODO does this need props
	s3io, err := io.LoadFS(nil, src.Warehouse)
	if err != nil {
		return nil, err
	}

	manifests, err := icebergTable.CurrentSnapshot().Manifests(s3io)
	if err != nil {
		return nil, err
	}

	var out []string

	for _, manifest := range manifests {

		// TODO double check on discard deleted
		entries, _ := manifest.FetchEntries(s3io, true)

		for _, entry := range entries {

			out = append(out, entry.DataFile().FilePath())
		}
	}

	return out, nil
}
