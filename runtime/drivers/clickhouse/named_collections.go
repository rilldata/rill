package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

// NamedCollectionPrefix is the fixed prefix for ClickHouse named collections that Rill manages.
// The frontend connector-source templates emit SQL referencing identifiers of the form
// `rill_<connector_name>` (e.g. `s3(rill_my_bucket, url='...')`), so this prefix is part of
// the cross-component contract and must not be changed without coordination.
const NamedCollectionPrefix = "rill_"

// supportedNamedCollectionDrivers lists the connector driver types for which Rill creates a
// ClickHouse named collection on reconcile. Anything not in this list is silently ignored.
var supportedNamedCollectionDrivers = map[string]bool{
	"s3":       true,
	"gcs":      true,
	"azure":    true,
	"mysql":    true,
	"postgres": true,
}

// IsSupportedNamedCollectionDriver returns true if Rill manages a ClickHouse named collection
// for the given connector driver type.
func IsSupportedNamedCollectionDriver(driver string) bool {
	return supportedNamedCollectionDrivers[driver]
}

// NamedCollectionName returns the canonical named-collection identifier for a connector.
// The frontend templates rely on this exact format; do not change the convention.
func NamedCollectionName(connectorName string) string {
	return NamedCollectionPrefix + connectorName
}

// namedCollectionParam is a single key/value pair to be emitted into the
// `CREATE NAMED COLLECTION ... AS k1=v1, k2=v2` clause.
//
// Keys are emitted as bare identifiers (CH expects unquoted keys here) and must come
// from the static driverNamedCollectionFieldMap below — never from user input — to keep
// the parameter names stable for the frontend templates and prevent injection via keys.
// Values are escaped via drivers.EscapeStringValue.
type namedCollectionParam struct {
	Key   string
	Value string
}

// BuildNamedCollectionParams maps a Rill connector's resolved config to the field/value pairs
// that should populate a ClickHouse named collection for the given driver.
//
// The mapping is intentionally conservative: only fields ClickHouse table functions actually
// understand are emitted. Field names match what the corresponding CH table function expects
// (e.g. `s3` table function uses `url`, `access_key_id`, `secret_access_key`; `postgresql`
// uses `host`, `port`, `database`, `user`, `password`).
//
// The returned slice is sorted by key for deterministic output (test stability).
//
// Returns ErrUnsupportedNamedCollectionDriver if the driver is not in the supported set.
func BuildNamedCollectionParams(driver string, resolvedConfig map[string]any) ([]namedCollectionParam, error) {
	if !IsSupportedNamedCollectionDriver(driver) {
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedNamedCollectionDriver, driver)
	}

	params := make(map[string]string)
	switch driver {
	case "s3":
		var c struct {
			AccessKeyID     string `mapstructure:"aws_access_key_id"`
			SecretAccessKey string `mapstructure:"aws_secret_access_key"`
			SessionToken    string `mapstructure:"aws_access_token"`
			Region          string `mapstructure:"region"`
			Endpoint        string `mapstructure:"endpoint"`
		}
		if err := mapstructure.WeakDecode(resolvedConfig, &c); err != nil {
			return nil, fmt.Errorf("failed to parse s3 config: %w", err)
		}
		// CH `s3` table function field names. See:
		// https://clickhouse.com/docs/en/sql-reference/table-functions/s3
		// We omit `url` so each model can supply its own via the override syntax
		// `s3(rill_<conn>, url='...')`. This matches the frontend template convention.
		if c.AccessKeyID != "" {
			params["access_key_id"] = c.AccessKeyID
		}
		if c.SecretAccessKey != "" {
			params["secret_access_key"] = c.SecretAccessKey
		}
		if c.SessionToken != "" {
			params["session_token"] = c.SessionToken
		}
		if c.Region != "" {
			params["region"] = c.Region
		}
		if c.Endpoint != "" {
			params["endpoint"] = c.Endpoint
		}
	case "gcs":
		// GCS named collections in ClickHouse are accessed through the `s3` table function in
		// S3-compatibility mode. Rill only emits a named collection when HMAC creds are set —
		// native service-account JSON is not usable from ClickHouse without additional support.
		var c struct {
			KeyID  string `mapstructure:"key_id"`
			Secret string `mapstructure:"secret"`
		}
		if err := mapstructure.WeakDecode(resolvedConfig, &c); err != nil {
			return nil, fmt.Errorf("failed to parse gcs config: %w", err)
		}
		if c.KeyID == "" || c.Secret == "" {
			return nil, ErrGCSRequiresHMAC
		}
		params["access_key_id"] = c.KeyID
		params["secret_access_key"] = c.Secret
		// Default GCS S3-compatible endpoint. Users can override per-model with
		// `s3(rill_<conn>, endpoint='...')` if they need a different region endpoint.
		params["endpoint"] = "https://storage.googleapis.com"
	case "azure":
		// CH `azureBlobStorage` table function fields. We support `connection_string` (preferred)
		// or `account_name`/`account_key`. SAS-token-only auth is not exposed here because the
		// CH function takes the SAS as part of the connection string.
		var c struct {
			Account          string `mapstructure:"azure_storage_account"`
			Key              string `mapstructure:"azure_storage_key"`
			ConnectionString string `mapstructure:"azure_storage_connection_string"`
		}
		if err := mapstructure.WeakDecode(resolvedConfig, &c); err != nil {
			return nil, fmt.Errorf("failed to parse azure config: %w", err)
		}
		if c.ConnectionString != "" {
			params["connection_string"] = c.ConnectionString
		}
		if c.Account != "" {
			params["account_name"] = c.Account
		}
		if c.Key != "" {
			params["account_key"] = c.Key
		}
		if len(params) == 0 {
			return nil, fmt.Errorf("azure connector has no usable credentials for a ClickHouse named collection (set azure_storage_connection_string or azure_storage_account/key)")
		}
	case "mysql":
		// CH `mysql` table function fields. We deliberately do not parse the DSN form here —
		// users wanting to use a named collection with CH should set the structured fields.
		var c struct {
			Host     string `mapstructure:"host"`
			Port     int    `mapstructure:"port"`
			Database string `mapstructure:"database"`
			User     string `mapstructure:"user"`
			Password string `mapstructure:"password"`
		}
		if err := mapstructure.WeakDecode(resolvedConfig, &c); err != nil {
			return nil, fmt.Errorf("failed to parse mysql config: %w", err)
		}
		if c.Host == "" || c.User == "" {
			return nil, fmt.Errorf("mysql connector must have host and user set to create a ClickHouse named collection")
		}
		port := c.Port
		if port == 0 {
			port = 3306
		}
		params["host"] = fmt.Sprintf("%s:%d", c.Host, port)
		params["user"] = c.User
		if c.Password != "" {
			params["password"] = c.Password
		}
		if c.Database != "" {
			params["database"] = c.Database
		}
	case "postgres":
		// CH `postgresql` table function fields.
		var c struct {
			Host     string `mapstructure:"host"`
			Port     string `mapstructure:"port"`
			DBname   string `mapstructure:"dbname"`
			User     string `mapstructure:"user"`
			Password string `mapstructure:"password"`
		}
		if err := mapstructure.WeakDecode(resolvedConfig, &c); err != nil {
			return nil, fmt.Errorf("failed to parse postgres config: %w", err)
		}
		if c.Host == "" || c.User == "" {
			return nil, fmt.Errorf("postgres connector must have host and user set to create a ClickHouse named collection")
		}
		port := c.Port
		if port == "" {
			port = "5432"
		}
		params["host"] = fmt.Sprintf("%s:%s", c.Host, port)
		params["user"] = c.User
		if c.Password != "" {
			params["password"] = c.Password
		}
		if c.DBname != "" {
			params["database"] = c.DBname
		}
	}

	out := make([]namedCollectionParam, 0, len(params))
	for k, v := range params {
		out = append(out, namedCollectionParam{Key: k, Value: v})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out, nil
}

// Errors returned by the named-collections subsystem.
var (
	// ErrUnsupportedNamedCollectionDriver indicates the connector driver type is not in the
	// list of drivers that map to a ClickHouse named collection.
	ErrUnsupportedNamedCollectionDriver = errors.New("clickhouse: connector driver does not support named collections")

	// ErrGCSRequiresHMAC indicates a GCS connector is configured with native service-account
	// credentials only. Rill cannot create a CH named collection from these — HMAC keys are required.
	ErrGCSRequiresHMAC = errors.New("clickhouse: gcs connector requires HMAC credentials (key_id + secret) to create a named collection")

	// ErrNamedCollectionAdminMissing is returned when the ClickHouse user lacks the
	// `named_collection_admin` access right (or equivalent CREATE / DROP NAMED COLLECTION grants).
	ErrNamedCollectionAdminMissing = errors.New("ClickHouse user lacks `named_collection_admin` permissions; named collections cannot be created. Either grant the permission or remove the connector resource")
)

// CreateOrReplaceNamedCollection creates or replaces a named collection on the ClickHouse server.
//
// The collection name is `rill_<connectorName>` and is emitted as a backtick-quoted identifier.
// Field keys come from the static driverNamedCollectionFieldMap and are emitted unquoted (CH grammar).
// Field values are escaped as SQL string literals to prevent injection.
//
// If the connection has a cluster configured, `ON CLUSTER <cluster>` is appended.
func (c *Connection) CreateOrReplaceNamedCollection(ctx context.Context, connectorName string, params []namedCollectionParam) error {
	if len(params) == 0 {
		return fmt.Errorf("clickhouse: refusing to create empty named collection for connector %q", connectorName)
	}

	stmt, err := buildCreateNamedCollectionSQL(connectorName, params, c.config.Cluster)
	if err != nil {
		return err
	}
	return c.Exec(ctx, &drivers.Statement{Query: stmt, Priority: 100})
}

// DropNamedCollection drops a named collection if it exists.
func (c *Connection) DropNamedCollection(ctx context.Context, connectorName string) error {
	stmt := buildDropNamedCollectionSQL(connectorName, c.config.Cluster)
	return c.Exec(ctx, &drivers.Statement{Query: stmt, Priority: 100})
}

// CheckNamedCollectionAdmin verifies that the current ClickHouse user has the privilege to
// create and drop named collections. The cheapest reliable check is to perform a no-op create
// + drop on a uniquely-named probe collection. This is more accurate than reading
// `system.users` because the user's effective grants depend on roles, default roles, and
// `access_management` settings that aren't trivially queryable in a portable way.
//
// Returns ErrNamedCollectionAdminMissing wrapping the underlying error if the privilege check
// fails. Returns nil if the probe succeeds.
func (c *Connection) CheckNamedCollectionAdmin(ctx context.Context) error {
	probeName := "rill_probe_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
	createStmt, err := buildCreateNamedCollectionSQL(strings.TrimPrefix(probeName, NamedCollectionPrefix), []namedCollectionParam{{Key: "dummy", Value: "1"}}, c.config.Cluster)
	if err != nil {
		return err
	}
	if err := c.Exec(ctx, &drivers.Statement{Query: createStmt, Priority: 100}); err != nil {
		return fmt.Errorf("%w: %v", ErrNamedCollectionAdminMissing, err)
	}
	dropStmt := buildDropNamedCollectionSQL(strings.TrimPrefix(probeName, NamedCollectionPrefix), c.config.Cluster)
	if err := c.Exec(ctx, &drivers.Statement{Query: dropStmt, Priority: 100}); err != nil {
		// If the create succeeded but drop failed, log via returned error — leaving the probe
		// behind is undesirable but the user clearly does have create rights, so we still
		// surface this as a permission-related issue to be safe.
		return fmt.Errorf("%w: failed to drop probe named collection: %v", ErrNamedCollectionAdminMissing, err)
	}
	return nil
}

// NamedCollectionExists returns true if a named collection with the given Rill-managed name
// exists on the server. This is intended for tests; production code should not branch on it.
func (c *Connection) NamedCollectionExists(ctx context.Context, connectorName string) (bool, error) {
	name := NamedCollectionName(connectorName)
	var count uint64
	row := c.readDB.QueryRowxContext(ctx, "SELECT count() FROM system.named_collections WHERE name = ?", name)
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("failed to query system.named_collections: %w", err)
	}
	return count > 0, nil
}

// buildCreateNamedCollectionSQL builds the CREATE OR REPLACE statement. Extracted for testability.
func buildCreateNamedCollectionSQL(connectorName string, params []namedCollectionParam, cluster string) (string, error) {
	if connectorName == "" {
		return "", errors.New("clickhouse: connector name is required")
	}
	name := NamedCollectionName(connectorName)
	var b strings.Builder
	b.WriteString("CREATE OR REPLACE NAMED COLLECTION ")
	b.WriteString(safeSQLName(name))
	if cluster != "" {
		b.WriteString(" ON CLUSTER ")
		b.WriteString(safeSQLName(cluster))
	}
	b.WriteString(" AS ")
	for i, p := range params {
		if i > 0 {
			b.WriteString(", ")
		}
		// Keys come from the static map above; they are safe ASCII identifiers and the CH
		// grammar requires them unquoted in this position. Values are escaped as string literals.
		b.WriteString(p.Key)
		b.WriteString(" = ")
		b.WriteString(drivers.EscapeStringValue(p.Value))
	}
	return b.String(), nil
}

// buildDropNamedCollectionSQL builds the DROP IF EXISTS statement. Extracted for testability.
func buildDropNamedCollectionSQL(connectorName, cluster string) string {
	name := NamedCollectionName(connectorName)
	var b strings.Builder
	b.WriteString("DROP NAMED COLLECTION IF EXISTS ")
	b.WriteString(safeSQLName(name))
	if cluster != "" {
		b.WriteString(" ON CLUSTER ")
		b.WriteString(safeSQLName(cluster))
	}
	return b.String()
}

// namedCollectionRefRegexp matches references to Rill-managed named collections inside model SQL,
// e.g. `s3(rill_my_bucket, url='...')` or `postgresql(rill_my_db, table='foo')`. The capture group
// returns the full identifier (`rill_my_bucket`).
//
// We deliberately use a regex rather than a SQL parser: ClickHouse SQL is dialect-specific and
// we only need to detect identifier references for the auto-detection feature. False positives
// are harmless — we just verify the connector name actually exists before doing anything with it.
//
// The regex accommodates cluster-aware variants like `s3Cluster('cluster_name', rill_xxx, ...)`,
// where a string-literal cluster name precedes the named-collection reference. We do that with
// an optional `('...',\s*)?` group rather than enumerating both forms.
var namedCollectionRefRegexp = regexp.MustCompile(`(?i)\b(?:s3|s3Cluster|gcs|azureBlobStorage|mysql|postgresql|url)\s*\(\s*(?:'[^']*'\s*,\s*)?(rill_[A-Za-z0-9_]+)\b`)

// DetectNamedCollectionRefs returns the set of Rill connector names referenced via named
// collections in the given SQL. The returned names exclude the `rill_` prefix.
//
// This is analogous to the auto-detection path in DuckDB's `connectorsForSecrets`: callers can
// use it to figure out which connectors a model depends on without requiring an explicit list.
// Note: detection is best-effort. The actual creation/deletion of named collections is driven
// by connector-resource lifecycle, not by model SQL.
func DetectNamedCollectionRefs(sql string) []string {
	matches := namedCollectionRefRegexp.FindAllStringSubmatch(sql, -1)
	seen := make(map[string]struct{}, len(matches))
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		name := strings.TrimPrefix(m[1], NamedCollectionPrefix)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}
