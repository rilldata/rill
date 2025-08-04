package clickhousestatic

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/provisioner"
	"go.uber.org/zap"
)

func init() {
	provisioner.Register("clickhouse-static", New)
}

type Spec struct {
	// DSN with admin permissions for a Clickhouse service.
	// This will be used to create a new (virtual) database and access-restricted user for each provisioned resource.
	DSN string `json:"dsn"`
	// DSNEnv variable that contains the clickhouse DSN.
	// This is an alternative to specifying the DSN directly, which can be useful for injecting secrets
	DSNEnv string `json:"dsn_env"`
	// WriteDSNEnv with permissions for write operations (optional)
	WriteDSNEnv string `json:"write_dsn_env,omitempty"`

	// Cluster name for ClickHouse cluster operations.
	// If specified, all DDL operations will include an ON CLUSTER clause.
	Cluster string `json:"cluster,omitempty"`
}

// Provisioner provisions Clickhouse resources using a static, multi-tenant Clickhouse service.
// It creates a new (virtual) database and user with access restricted to that database for each resource.
type Provisioner struct {
	spec   *Spec
	logger *zap.Logger
	ch     *sql.DB
}

var _ provisioner.Provisioner = (*Provisioner)(nil)

func New(specJSON []byte, _ database.DB, logger *zap.Logger) (provisioner.Provisioner, error) {
	spec := &Spec{}
	err := json.Unmarshal(specJSON, spec)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provisioner spec: %w", err)
	}

	var dsn string
	if spec.DSN != "" {
		dsn = spec.DSN
	} else if spec.DSNEnv != "" {
		dsn = os.Getenv(spec.DSNEnv)
		if dsn == "" {
			return nil, fmt.Errorf("environment variable %q is not set or empty", spec.DSNEnv)
		}
	} else {
		return nil, fmt.Errorf("either dsn or dsn_env must be specified")
	}

	// Get optional write DSN
	var writeDSN string
	if spec.WriteDSNEnv != "" {
		writeDSN = os.Getenv(spec.WriteDSNEnv)
		if writeDSN == "" {
			return nil, fmt.Errorf("environment variable %q is not set or empty", spec.WriteDSNEnv)
		}
	}

	// Use writeDSN for provisioning operations if available, otherwise use the primary DSN
	provisioningDSN := dsn
	if writeDSN != "" {
		provisioningDSN = writeDSN
	}

	opts, err := clickhouse.ParseDSN(provisioningDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}
	ch := clickhouse.OpenDB(opts)

	return &Provisioner{
		spec:   spec,
		logger: logger,
		ch:     ch,
	}, nil
}

func (p *Provisioner) Type() string {
	return "clickhouse-static"
}

func (p *Provisioner) Supports(rt provisioner.ResourceType) bool {
	return rt == provisioner.ResourceTypeClickHouse
}

func (p *Provisioner) Close() error {
	var err error

	if closeErr := p.ch.Close(); closeErr != nil {
		err = closeErr
	}

	return err
}

func (p *Provisioner) Provision(ctx context.Context, r *provisioner.Resource, opts *provisioner.ResourceOptions) (*provisioner.Resource, error) {
	// Parse the resource's config (in case it's an update/check)
	cfg, err := provisioner.NewClickhouseConfig(r.Config)
	if err != nil {
		return nil, err
	}

	// If the config has already been populated, do a health check and exit early
	if cfg.DSN != "" {
		err := p.pingWithResourceDSN(ctx, cfg.DSN)
		if err != nil {
			return nil, fmt.Errorf("failed to ping clickhouse resource: %w", err)
		}

		// Ping the write DSN if it exists
		if cfg.WriteDSN != "" {
			err := p.pingWithResourceDSN(ctx, cfg.WriteDSN)
			if err != nil {
				return nil, fmt.Errorf("failed to ping clickhouse write resource: %w", err)
			}
		}

		return r, nil
	}

	// Prepare for creating the schema and user.
	id := strings.ReplaceAll(r.ID, "-", "")
	user := fmt.Sprintf("rill_%s", id)
	dbName := fmt.Sprintf("rill_%s", id)

	// Use org and project names to create a more human-readable database name.
	orgName := sanitizeName(getAnnotationValue(opts.Annotations, "org"))
	projectName := sanitizeName(getAnnotationValue(opts.Annotations, "project"))
	if orgName != "" && projectName != "" {
		dbName = fmt.Sprintf("rill_%s_%s_%s", orgName, projectName, id)
	}

	password := newPassword()
	annotationsJSON, err := json.Marshal(opts.Annotations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal annotations: %w", err)
	}

	// Idempotently create the schema
	_, err = p.ch.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s %s COMMENT ?", dbName, p.onCluster()), string(annotationsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create clickhouse database: %w", err)
	}

	// Idempotently create the user.
	_, err = p.ch.ExecContext(ctx, fmt.Sprintf("CREATE USER IF NOT EXISTS %s %s IDENTIFIED WITH sha256_password BY ? DEFAULT DATABASE %s GRANTEES NONE", user, p.onCluster(), dbName), password)
	if err != nil {
		return nil, fmt.Errorf("failed to create clickhouse user: %w", err)
	}

	// When creating the user, the password assignment is not idempotent (if there are two concurrent invocations, we don't know which password was used).
	// By adding the password separately, we ensure all passwords will work.
	// NOTE: Requires ClickHouse 24.9 or later.
	_, err = p.ch.ExecContext(ctx, fmt.Sprintf("ALTER USER %s %s ADD IDENTIFIED WITH sha256_password BY ?", user, p.onCluster()), password)
	if err != nil {
		return nil, fmt.Errorf("failed to add password for clickhouse user: %w", err)
	}

	// Grant privileges on the database to the user
	_, err = p.ch.ExecContext(ctx, fmt.Sprintf(`
		GRANT %s
			SELECT,
			INSERT,
			ALTER,
			CREATE TABLE,
			CREATE DICTIONARY,
			CREATE VIEW,
			DROP TABLE,
			DROP DICTIONARY,
			DROP VIEW,
			TRUNCATE,
			OPTIMIZE,
			SHOW DICTIONARIES,
			dictGet
		ON %s.* TO %s
	`, p.onCluster(), dbName, user))
	if err != nil {
		return nil, fmt.Errorf("failed to grant privileges to clickhouse user: %w", err)
	}

	// Grant access to system.parts for reporting disk usage.
	// NOTE 1: ClickHouse automatically adds row filters to restrict result to tables the user has access to.
	// NOTE 2: We do not need to explicitly grant access to system.tables and system.columns because ClickHouse adds those implicitly.
	_, err = p.ch.ExecContext(ctx, fmt.Sprintf("GRANT %s SELECT ON system.parts TO %s", p.onCluster(), user))
	if err != nil {
		return nil, fmt.Errorf("failed to grant system privileges to clickhouse user: %w", err)
	}

	// Grant some additional global privileges to the user
	_, err = p.ch.ExecContext(ctx, fmt.Sprintf(`
		GRANT %s
			URL,
			REMOTE,
			MONGO,
			MYSQL,
			POSTGRES,
			S3,
			AZURE
		ON *.* TO %s
	`, p.onCluster(), user))
	if err != nil {
		return nil, fmt.Errorf("failed to grant global privileges to clickhouse user: %w", err)
	}

	// Create the resource DSN - always use the primary DSN as the base
	// The client will use this for reads, and can optionally use writeDSN for writes
	resourceDSN, err := url.Parse(p.spec.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base DSN: %w", err)
	}
	resourceDSN.User = url.UserPassword(user, password)
	resourceDSN.Path = "/" + dbName

	cfg = &provisioner.ClickhouseConfig{
		DSN: resourceDSN.String(),
	}

	// If we have a separate write DSN, include it in the config
	if p.spec.WriteDSNEnv != "" {
		writeDSN := os.Getenv(p.spec.WriteDSNEnv)
		if writeDSN == "" {
			return nil, fmt.Errorf("environment variable %q is not set or empty", p.spec.WriteDSNEnv)
		}

		writeResourceDSN, err := url.Parse(writeDSN)
		if err != nil {
			return nil, fmt.Errorf("failed to parse write DSN: %w", err)
		}
		writeResourceDSN.User = url.UserPassword(user, password)
		writeResourceDSN.Path = "/" + dbName

		cfg.WriteDSN = writeResourceDSN.String()
	}

	return &provisioner.Resource{
		ID:     r.ID,
		Type:   r.Type,
		State:  nil,
		Config: cfg.AsMap(),
	}, nil
}

func (p *Provisioner) Deprovision(ctx context.Context, r *provisioner.Resource) error {
	// Check it's a clickhouse resource
	if r.Type != provisioner.ResourceTypeClickHouse {
		return fmt.Errorf("unexpected resource type %q", r.Type)
	}

	// Parse the resource's config
	cfg, err := provisioner.NewClickhouseConfig(r.Config)
	if err != nil {
		return err
	}

	// Exit early if the config is empty (nothing to deprovision)
	if cfg.DSN == "" {
		return nil
	}

	// Parse the DSN to get database and user info
	opts, err := clickhouse.ParseDSN(cfg.DSN)
	if err != nil {
		return fmt.Errorf("failed to parse DSN during deprovisioning: %w", err)
	}

	// Drop the database
	_, err = p.ch.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s %s", opts.Auth.Database, p.onCluster()))
	if err != nil {
		return fmt.Errorf("failed to drop clickhouse database: %w", err)
	}

	// Drop the user
	_, err = p.ch.ExecContext(ctx, fmt.Sprintf("DROP USER IF EXISTS %s %s", opts.Auth.Username, p.onCluster()))
	if err != nil {
		return fmt.Errorf("failed to drop clickhouse user: %w", err)
	}

	return nil
}

func (p *Provisioner) AwaitReady(ctx context.Context, r *provisioner.Resource) error {
	return nil
}

func (p *Provisioner) Check(ctx context.Context) error {
	return nil
}

func (p *Provisioner) CheckResource(ctx context.Context, r *provisioner.Resource, opts *provisioner.ResourceOptions) (*provisioner.Resource, error) {
	// Provision is idempotent and will do nothing if the resource is already provisioned.
	return p.Provision(ctx, r, opts)
}

func (p *Provisioner) pingWithResourceDSN(ctx context.Context, dsn string) error {
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return fmt.Errorf("failed to open tenant connection: %w", err)
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, "SELECT 1")
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

// onCluster returns the ON CLUSTER clause if a cluster is configured, otherwise returns an empty string.
func (p *Provisioner) onCluster() string {
	if p.spec.Cluster != "" {
		return fmt.Sprintf("ON CLUSTER %s", p.spec.Cluster)
	}
	return ""
}

func newPassword() string {
	var b [16]byte
	_, err := io.ReadFull(rand.Reader, b[:])
	if err != nil {
		panic(err)
	}
	// Ensure all of digits/letters/uppercase/lowercase/special characters
	return fmt.Sprintf("1Rr!%x", b[:])
}

// Helper function to get annotation value
func getAnnotationValue(annotations map[string]string, key string) string {
	if annotations == nil {
		return ""
	}
	return annotations[key]
}

// Helper function to sanitize names for ClickHouse identifiers
func sanitizeName(name string) string {
	if name == "" {
		return ""
	}

	// Replace invalid characters with underscores and convert to lowercase
	sanitized := strings.ToLower(name)
	sanitized = strings.ReplaceAll(sanitized, "-", "_")
	sanitized = strings.ReplaceAll(sanitized, " ", "_")

	// Remove any characters that aren't alphanumeric or underscore
	var result strings.Builder
	for _, r := range sanitized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}

	return result.String()
}
