package duckdb

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/connectors/file"
	"github.com/rilldata/rill/runtime/connectors/gcs"
	"github.com/rilldata/rill/runtime/connectors/s3"
	"github.com/rilldata/rill/runtime/drivers"
)

func (c *connection) Ingest(ctx context.Context, source *connectors.Source) error {
	err := source.Validate()
	if err != nil {
		return err
	}

	// Driver-specific overrides
	switch source.Connector {
	case "file":
		return c.ingestFile(ctx, source)
	case "s3":
		return c.ingestS3(ctx, source)
	case "gcs":
		return c.ingestGCS(ctx, source)
	}

	// TODO: Use generic connectors.Consume when it's implemented
	return drivers.ErrUnsupportedConnector
}

func (c *connection) ingestFile(ctx context.Context, source *connectors.Source) error {
	conf, err := file.ParseConfig(source.Properties)
	if err != nil {
		return err
	}

	// Not using query args since not quite sure about behaviour of injecting table names that way.
	// Also, it's a source, so the caller can be trusted.

	from := fmt.Sprintf("'%s'", conf.Path)
	if conf.Format == "csv" && conf.CSVDelimiter != "" {
		from = fmt.Sprintf("read_csv_auto('%s', delim='%s')", conf.Path, conf.CSVDelimiter)
	}

	qry := fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM %s)", source.Name, from)

	rows, err := c.Execute(ctx, &drivers.Statement{Query: qry, Priority: 1})
	if err != nil {
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}

	return nil
}

func (c *connection) ingestS3(ctx context.Context, source *connectors.Source) error {
	fmt.Println("ingestS3 called")
	conf, err := s3.ParseConfig(source.Properties)
	if err != nil {
		return err
	}

	// TODO: set AWS settings for the transaction only

	qry := fmt.Sprintf("SET s3_endpoint='s3.amazonaws.com';SET s3_region='%s';", conf.AWSRegion)

	if conf.AWSKey != "" && conf.AWSSecret != "" {
		qry += fmt.Sprintf("SET s3_access_key_id='%s'; SET s3_secret_access_key='%s';", conf.AWSKey, conf.AWSSecret)
	} else if conf.AWSSession != "" {
		qry += fmt.Sprintf("SET s3_session_token='%s';", conf.AWSSession)
	}

	rows, err := c.Execute(ctx, &drivers.Statement{
		Query:    qry,
		Priority: 1,
	})

	if err != nil {
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}

	// TODO: we need to fix the issue of no error returned for the last query in a multi query request
	rows, err = c.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM '%s');", source.Name, conf.Path),
		Priority: 1,
	})
	if err != nil {
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}

	return nil
}

func (c *connection) ingestGCS(ctx context.Context, source *connectors.Source) error {
	conf, err := gcs.ParseConfig(source.Properties)
	if err != nil {
		return err
	}

	// TODO: set AWS settings for the transaction only

	qry := fmt.Sprintf("SET s3_endpoint='storage.googleapis.com';SET s3_region='%s';", conf.GCPRegion)

	if conf.GCPKey != "" && conf.GCPSecret != "" {
		qry += fmt.Sprintf("SET s3_access_key_id='%s'; SET s3_secret_access_key='%s';", conf.GCPKey, conf.GCPSecret)
	}

	qry += fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM '%s');", source.Name, conf.Path)

	rows, err := c.Execute(ctx, &drivers.Statement{
		Query:    qry,
		Priority: 1,
	})
	if err != nil {
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}

	// TODO: we need to fix the issue of no error returned for the last query in a multi query request
	rows, err = c.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM '%s');", source.Name, conf.Path),
		Priority: 1,
	})
	if err != nil {
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}

	return nil
}
