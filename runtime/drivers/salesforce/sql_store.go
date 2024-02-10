package salesforce

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

const defaultClientID = "3MVG9KsVczVNcM8y6w3Kjszy.DW9gMzcYDHT97WIX3NYNYA35UvITypEhtYc6FDY8qqcDEIQc_qJgZErv6Q_d"

// Query implements drivers.SQLStore
func (c *connection) Query(ctx context.Context, props map[string]any) (drivers.RowIterator, error) {
	return nil, drivers.ErrNotImplemented
}

// QueryAsFiles implements drivers.SQLStore
func (c *connection) QueryAsFiles(ctx context.Context, props map[string]any, opt *drivers.QueryOption, p drivers.Progress) (drivers.FileIterator, error) {
	srcProps, err := parseSourceProperties(props)
	if err != nil {
		return nil, err
	}

	var username, password, endpoint, key, clientID string
	if srcProps.Username != "" { // get from src properties
		username = srcProps.Username
	} else if u, ok := c.config["username"].(string); ok && u != "" { // get from driver configs
		username = u
	} else {
		return nil, fmt.Errorf("the property 'username' is required for Salesforce. Provide 'username' in the YAML properties or pass '--variable connector.salesforce.username=...' to 'rill start'")
	}

	if srcProps.Endpoint != "" { // get from src properties
		endpoint = srcProps.Endpoint
	} else if e, ok := c.config["endpoint"].(string); ok && e != "" { // get from driver configs
		endpoint = e
	} else {
		return nil, fmt.Errorf("the property 'endpoint' is required for Salesforce. Provide 'endpoint' in the YAML properties or pass '--variable connector.salesforce.endpoint=...' to 'rill start'")
	}

	if srcProps.ClientID != "" { // get from src properties
		clientID = srcProps.ClientID
	} else if c, ok := c.config["client_id"].(string); ok && c != "" { // get from driver configs
		clientID = c
	} else {
		clientID = defaultClientID
	}

	if srcProps.Password != "" { // get from src properties
		password = srcProps.Password
	} else if p, ok := c.config["password"].(string); ok && p != "" { // get from driver configs
		password = p
	}

	if srcProps.Key != "" { // get from src properties
		key = srcProps.Key
	} else if k, ok := c.config["key"].(string); ok && k != "" { // get from driver configs
		key = k
	}

	if password == "" && key == "" {
		return nil, fmt.Errorf("the property 'password' or property 'key' is required for Salesforce. Provide 'password' or 'key' in the YAML properties or pass '--variable connector.salesforce.password=...' or '--variable connector.salesforce.key=...' to 'rill start'")
	}

	authOptions := authenticationOptions{
		Username:     username,
		Password:     password,
		JWT:          key,
		Endpoint:     endpoint,
		ConnectedApp: clientID,
	}

	session, err := authenticate(authOptions)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	job := makeBulkJob(session, srcProps.SObject, srcProps.SOQL, srcProps.QueryAll, c.logger)

	err = c.startJob(ctx, job)
	if err != nil {
		return nil, err
	}

	err = job.getBatches(ctx)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (j *bulkJob) Format() string {
	return "csv"
}

// Close implements drivers.RowIterator.
func (j *bulkJob) Close() error {
	if j.tempFilePath != "" {
		err := os.Remove(j.tempFilePath)
		j.tempFilePath = ""
		if err != nil {
			return fmt.Errorf("failed to delete temp file: %w", err)
		}
	}
	return nil
}

// Next implements drivers.RowIterator.
func (j *bulkJob) Next() ([]string, error) {
	if j.jobID == "" {
		return nil, fmt.Errorf("invalid job: no job id")
	}
	if j.job.NumberRecordsProcessed == 0 {
		return nil, io.EOF
	}
	if j.tempFilePath != "" {
		err := os.Remove(j.tempFilePath)
		j.tempFilePath = ""
		if err != nil {
			return nil, fmt.Errorf("failed to delete temp file: %w", err)
		}
	}
	if j.nextResult == len(j.results) {
		return nil, io.EOF
	}
	tempFile, err := j.retrieveJobResult(context.Background(), j.nextResult)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve batch: %w", err)
	}
	j.tempFilePath = tempFile
	j.nextResult++
	return []string{j.tempFilePath}, nil
}

// Size implements drivers.RowIterator.
func (j *bulkJob) Size(unit drivers.ProgressUnit) (int64, bool) {
	switch unit {
	case drivers.ProgressUnitRecord:
		return int64(j.job.NumberRecordsProcessed), true
	case drivers.ProgressUnitFile:
		return int64(len(j.results)), true
	default:
		return 0, false
	}
}

type sourceProperties struct {
	SOQL     string `mapstructure:"soql"`
	SObject  string `mapstructure:"sobject"`
	QueryAll bool   `mapstructure:"queryAll"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Key      string `mapstructure:"key"`
	Endpoint string `mapstructure:"endpoint"`
	ClientID string `mapstructure:"client_id"`
}

func parseSourceProperties(props map[string]any) (*sourceProperties, error) {
	conf := &sourceProperties{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}
	if conf.SOQL == "" {
		return nil, fmt.Errorf("property 'soql' is mandatory for connector \"salesforce\"")
	}
	if conf.SObject == "" {
		return nil, fmt.Errorf("property 'sobject' is mandatory for connector \"salesforce\"")
	}
	return conf, err
}
