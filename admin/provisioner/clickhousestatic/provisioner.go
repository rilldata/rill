package clickhousestatic

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
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

	opts, err := clickhouse.ParseDSN(spec.DSN)
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

func (p *Provisioner) Close() error {
	return p.ch.Close()
}

func (p *Provisioner) Provision(ctx context.Context, r *provisioner.Resource, opts *provisioner.ResourceOptions) (*provisioner.Resource, error) {
	// Can only provision clickhouse resources
	if r.Type != provisioner.ResourceTypeClickHouse {
		return nil, provisioner.ErrResourceTypeNotSupported
	}

	// Parse the resource's config (in case it's an update/check)
	cfg, err := provisioner.NewClickhouseConfig(r.Config)
	if err != nil {
		return nil, err
	}

	// If the config has already been populated, do a health check and exit early (currently there's nothing to update).
	if cfg.DSN != "" {
		err := p.pingWithResourceDSN(ctx, cfg.DSN)
		if err != nil {
			return nil, fmt.Errorf("failed to ping clickhouse resource: %w", err)
		}

		return r, nil
	}

	// Prepare for creating the schema and user.
	id := strings.ReplaceAll(r.ID, "-", "")
	dbName := fmt.Sprintf("rill_%s", id)
	user := fmt.Sprintf("rill_%s", id)
	password := newPassword()
	annotationsJSON, err := json.Marshal(opts.Annotations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal annotations: %w", err)
	}

	// Idempotently create the schema
	_, err = p.ch.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s COMMENT ?", dbName), string(annotationsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create clickhouse database: %w", err)
	}

	// Idempotently create the user.
	_, err = p.ch.ExecContext(ctx, fmt.Sprintf("CREATE USER IF NOT EXISTS %s IDENTIFIED WITH sha256_password BY ? DEFAULT DATABASE %s GRANTEES NONE", user, dbName), password)
	if err != nil {
		return nil, fmt.Errorf("failed to create clickhouse user: %w", err)
	}

	// When creating the user, the password assignment is not idempotent (if there are two concurrent invocations, we don't know which password was used).
	// By adding the password separately, we ensure all passwords will work.
	// NOTE: Requires ClickHouse 24.9 or later.
	_, err = p.ch.ExecContext(ctx, fmt.Sprintf("ALTER USER %s ADD IDENTIFIED WITH sha256_password BY ?", user), password)
	if err != nil {
		return nil, fmt.Errorf("failed to add password for clickhouse user: %w", err)
	}

	// Grant privileges on the database to the user
	_, err = p.ch.ExecContext(ctx, fmt.Sprintf(`
		GRANT
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
	`, dbName, user))
	if err != nil {
		return nil, fmt.Errorf("failed to grant privileges to clickhouse user: %w", err)
	}

	// Grant some additional global privileges to the user
	_, err = p.ch.ExecContext(ctx, fmt.Sprintf(`
		GRANT
			URL,
			REMOTE,
			MONGO,
			MYSQL,
			POSTGRES,
			S3,
			AZURE
		ON *.* TO %s
	`, user))
	if err != nil {
		return nil, fmt.Errorf("failed to grant global privileges to clickhouse user: %w", err)
	}

	// Build DSN for the resource and return it
	dsn, err := url.Parse(p.spec.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base DSN: %w", err)
	}
	dsn.User = url.UserPassword(user, password)
	dsn.Path = dbName
	cfg = &provisioner.ClickhouseConfig{
		DSN: dsn.String(),
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

	// Parse the DSN
	opts, err := clickhouse.ParseDSN(cfg.DSN)
	if err != nil {
		return fmt.Errorf("failed to parse DSN during deprovisioning: %w", err)
	}

	// Drop the database
	_, err = p.ch.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", opts.Auth.Database))
	if err != nil {
		return fmt.Errorf("failed to drop clickhouse database: %w", err)
	}

	// Drop the user
	_, err = p.ch.ExecContext(ctx, fmt.Sprintf("DROP USER IF EXISTS %s", opts.Auth.Username))
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
		return fmt.Errorf("failed to execute query on tenant: %w", err)
	}

	return nil
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
