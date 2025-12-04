package clickhouse

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/gcs"
	"github.com/rilldata/rill/runtime/drivers/s3"
)

type selfToSelfExecutor struct {
	c *Connection
}

var _ drivers.ModelExecutor = &selfToSelfExecutor{}

func (e *selfToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return desired, true
	}
	return _defaultConcurrentInserts, true
}

func (e *selfToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	// Parse the input and output properties
	inputProps := &ModelInputProperties{}
	if err := mapstructure.WeakDecode(opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	outputProps := &ModelOutputProperties{}
	if err := mapstructure.WeakDecode(opts.OutputProperties, outputProps); err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}

	// Validate the output properties
	err := e.c.validateAndApplyDefaults(opts, inputProps, outputProps)
	if err != nil {
		return nil, fmt.Errorf("invalid model properties: %w", err)
	}

	usedModelName := false
	if outputProps.Table == "" {
		outputProps.Table = opts.ModelName
		usedModelName = true
	}

	asView := outputProps.Typ == "VIEW"
	tableName := outputProps.Table

	if !e.c.config.isClickhouseCloud() {
		connectors, autoDetected := connectorsForNameCollection(inputProps.CreateNamedCollectionsFromConnectors, e.c.config.CreateNamedCollectionsFromConnectors, opts.Env.Connectors)
		err = e.createNamedCollectionsForConnectors(ctx, connectors, autoDetected, opts.Env)
		if err != nil {
			return nil, err
		}
	}

	var metrics *tableWriteMetrics
	if !opts.IncrementalRun {
		stagingTableName := tableName
		if opts.Env.StageChanges {
			stagingTableName = stagingTableNameFor(tableName)
		}

		// Drop the staging view/table if it exists.
		// NOTE: This intentionally drops the end table if not staging changes.
		_ = e.c.dropTable(ctx, stagingTableName)

		// Create the table
		var err error
		metrics, err = e.c.createTableAsSelect(ctx, stagingTableName, inputProps.SQL, outputProps)
		if err != nil {
			_ = e.c.dropTable(ctx, stagingTableName)
			return nil, fmt.Errorf("failed to create model: %w", err)
		}

		// Rename the staging table to the final table name
		if stagingTableName != tableName {
			err = e.c.forceRenameTable(ctx, stagingTableName, asView, tableName)
			if err != nil {
				return nil, fmt.Errorf("failed to rename staged model: %w", err)
			}
		}
	} else {
		// Insert into the table
		var err error
		metrics, err = e.c.insertTableAsSelect(ctx, tableName, inputProps.SQL, &InsertTableOptions{
			Strategy: outputProps.IncrementalStrategy,
		}, outputProps)
		if err != nil {
			return nil, fmt.Errorf("failed to incrementally insert into table: %w", err)
		}
	}

	// Build result props
	resultProps := &ModelResultProperties{
		Table:         tableName,
		View:          asView,
		Typ:           outputProps.Typ,
		UsedModelName: usedModelName,
	}
	resultPropsMap := map[string]interface{}{}
	err = mapstructure.WeakDecode(resultProps, &resultPropsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to encode result properties: %w", err)
	}

	// Done
	return &drivers.ModelResult{
		Connector:    opts.OutputConnector,
		Properties:   resultPropsMap,
		Table:        tableName,
		ExecDuration: metrics.duration,
	}, nil
}

func (e *selfToSelfExecutor) createNamedCollectionsForConnectors(ctx context.Context, connectors []string, autoDetected bool, env *drivers.ModelEnv) error {
	if len(connectors) == 0 {
		return nil
	}

	connectorHashes, err := e.fetchConnectorHashes(ctx)
	if err != nil {
		return fmt.Errorf("failed reading stored connector hashes: %w", err)
	}

	for _, connector := range connectors {
		creds, err := getNamedCollectionCreds(ctx, connector, env)
		if err != nil {
			if autoDetected {
				continue
			}
			return err
		}
		if len(creds) == 0 {
			continue
		}
		// If prevHash and current hash is same not need to recreate the named collection
		hash, err := hashCreds(creds)
		if err != nil {
			return err
		}
		prevHash, hasPrev := connectorHashes[connector]
		if hasPrev && prevHash == hash {
			continue
		}

		// Acquire semaphore
		err = e.c.metaSem.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		if err := e.c.createNamedCollections(ctx, connector, creds); err != nil {
			e.c.metaSem.Release(1)
			return err
		}
		// update hash in named collection
		name := fmt.Sprintf("rill_hash_for_%s", connector)
		creds = map[string]string{hash: "true"}
		if err := e.c.createNamedCollections(ctx, name, creds); err != nil {
			e.c.metaSem.Release(1)
			return err
		}
		e.c.metaSem.Release(1)
	}
	return nil
}

func (e *selfToSelfExecutor) fetchConnectorHashes(ctx context.Context) (map[string]string, error) {
	rows, err := e.c.Query(ctx, &drivers.Statement{
		Query:    "SELECT name, collection FROM system.named_collections WHERE name LIKE 'rill_hash_for_%'",
		Priority: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("select named collections: %w", err)
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var name string
		var coll map[string]string
		if err := rows.Scan(&name, &coll); err != nil {
			return nil, err
		}

		// Extract connector
		prefix := "rill_hash_for_"
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		connector := name[len(prefix):]

		// Extract the one hash key
		for k := range coll {
			out[connector] = k
			break
		}
	}
	return out, nil
}

// connectorsForNamedCollections returns the list of connectors to be used for NamedCollection creation.
// Priority:
// 1. If the model configuration specifies connector names, use those.
// 2. if clickhouse connector configuration specifies connector names, use those
// 3. If neither is configured, automatically detect all connectors of type s3 and gcs.
// The boolean return value is true if the list of connectors was automatically detected.
func connectorsForNameCollection(modelNamedCollections, clickhouseNamedCollections []string, allConnectors []*runtimev1.Connector) ([]string, bool) {
	var configuredConnectorsForNamedCollections []string
	if len(modelNamedCollections) > 0 {
		configuredConnectorsForNamedCollections = append(configuredConnectorsForNamedCollections, modelNamedCollections...)
	} else if len(clickhouseNamedCollections) > 0 {
		configuredConnectorsForNamedCollections = append(configuredConnectorsForNamedCollections, clickhouseNamedCollections...)
	}

	// If no connectors are configured, automatically detect all connectors of type s3, gcs from the project.
	// If a single configured value contains a comma-separated list of connector names, split it into individual entries.
	// Otherwise, return the explicitly configured list of connectors.
	if len(configuredConnectorsForNamedCollections) == 0 {
		var res []string
		for _, c := range allConnectors {
			if c.Type == "s3" || c.Type == "gcs" {
				res = append(res, c.Name)
			}
		}
		return res, true
	} else if len(configuredConnectorsForNamedCollections) == 1 && strings.Contains(configuredConnectorsForNamedCollections[0], ",") {
		res := strings.Split(configuredConnectorsForNamedCollections[0], ",")
		for i, s := range res {
			res[i] = strings.TrimSpace(s)
		}
		return res, false
	}
	return configuredConnectorsForNamedCollections, false
}

// getNamedCollectionCreds extracts the credentials required to create a ClickHouse named collection
// for the given connector. If required credential fields are missing, it returns an explicit
// ErrMissingCredentials. Only supported connector types (e.g S3, GCS) are allowed.
func getNamedCollectionCreds(ctx context.Context, connector string, env *drivers.ModelEnv) (map[string]string, error) {
	handle, release, err := env.AcquireConnector(ctx, connector)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connector %q: %w", connector, err)
	}
	release()
	creds := map[string]string{}

	switch handle.Driver() {
	case "s3":
		conn, ok := handle.(*s3.Connection)
		if !ok {
			release()
			return nil, fmt.Errorf("internal error: expected s3 handle for %q", connector)
		}
		cfg := conn.ParsedConfig()

		if cfg.AccessKeyID == "" {
			return nil, fmt.Errorf("s3 aws_access_key_id is empty for connector %q", connector)
		}
		creds["access_key_id"] = cfg.AccessKeyID
		creds["secret_access_key"] = cfg.SecretAccessKey
		if cfg.SessionToken != "" {
			creds["session_token"] = cfg.SessionToken
		}

	case "gcs":
		prop := handle.Config()
		cfg, err := gcs.NewConfigProperties(prop)
		if err != nil {
			return nil, fmt.Errorf("failed gcs config for %q: %w", connector, err)
		}

		if cfg.KeyID == "" {
			return nil, fmt.Errorf("gcs key_id is empty for connector %q", connector)
		}
		creds["access_key_id"] = cfg.KeyID
		creds["secret_access_key"] = cfg.Secret

	default:
		return nil, fmt.Errorf("named collections not supported for connector type %q", handle.Driver())
	}
	return creds, nil
}

func hashCreds(m map[string]string) (string, error) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	hash := md5.New()
	for _, key := range keys {
		_, err := hash.Write([]byte(key))
		if err != nil {
			return "", err
		}
		_, err = hash.Write([]byte(m[key]))
		if err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
