package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgconn/ctxwatch"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/rilldata/rill/admin/server/auth"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	base "github.com/rilldata/rill/runtime/pkg/pgwire"
)

// ServePGWire starts the admin PostgreSQL wire-compatible endpoint. Each client
// session is proxied to the selected production runtime over one dedicated pgx
// connection.
func (s *Server) ServePGWire(ctx context.Context) error {
	var tlsConfig *tls.Config
	if s.opts.TLSCertPath != "" && s.opts.TLSKeyPath != "" {
		tlsConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
				certificate, err := tls.LoadX509KeyPair(s.opts.TLSCertPath, s.opts.TLSKeyPath)
				if err != nil {
					return nil, err
				}
				return &certificate, nil
			},
		}
	}
	server, err := base.NewServer(base.Options{
		TLSConfig:       tlsConfig,
		RequirePassword: true,
		Logger:          s.logger,
		NewSession:      s.newPGWireProxySession,
	})
	if err != nil {
		return err
	}
	return graceful.ServePGWire(ctx, server, s.opts.PSQLPort, s.logger)
}

func (s *Server) newPGWireProxySession(ctx context.Context, parameters map[string]string, password string) (base.Session, error) {
	ctx, err := s.authenticator.AuthenticateToken(ctx, password)
	if err != nil {
		return nil, err
	}

	org, project, ok := strings.Cut(parameters["database"], ".")
	if !ok || org == "" || project == "" || strings.Contains(project, ".") {
		return nil, &base.Error{Code: "3D000", Message: "database must have the form org.project"}
	}
	proj, err := s.admin.DB.FindProjectByName(ctx, org, project)
	if err != nil {
		return nil, &base.Error{Code: "3D000", Message: "invalid organization or project"}
	}
	if proj.PrimaryDeploymentID == nil {
		return nil, &base.Error{Code: "3D000", Message: "project has no production deployment"}
	}
	deployment, err := s.admin.DB.FindDeployment(ctx, *proj.PrimaryDeploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find production deployment: %w", err)
	}

	claims := auth.GetClaims(ctx)
	permissions := claims.ProjectPermissions(ctx, proj.OrganizationID, deployment.ProjectID)
	if !permissions.ReadProd {
		return nil, &base.Error{Code: "42501", Message: "permission denied for production deployment"}
	}
	jwt, err := s.issueRuntimeToken(ctx, &issueRuntimeTokenOptions{
		project:            proj,
		deployment:         deployment,
		projectPermissions: permissions,
		forOwner:           true,
		ttl:                runtimeProxyAccessTokenTTL,
	})
	if err != nil {
		return nil, err
	}

	runtimeURL, err := url.Parse(deployment.RuntimeHost)
	if err != nil {
		return nil, fmt.Errorf("invalid runtime host: %w", err)
	}
	host := runtimeURL.Hostname()
	if host == "" {
		host = deployment.RuntimeHost
	}
	sslMode := "disable"
	if strings.EqualFold(runtimeURL.Scheme, "https") {
		sslMode = "require"
	}
	connectionURL := &url.URL{
		Scheme:   "postgres",
		User:     url.User("rill"),
		Host:     net.JoinHostPort(host, strconv.Itoa(s.opts.RuntimePSQLPort)),
		Path:     deployment.RuntimeInstanceID,
		RawQuery: "sslmode=" + sslMode,
	}
	config, err := pgx.ParseConfig(connectionURL.String())
	if err != nil {
		return nil, err
	}
	config.Password = jwt
	config.BuildContextWatcherHandler = func(conn *pgconn.PgConn) ctxwatch.Handler {
		return &pgconn.CancelRequestContextWatcherHandler{
			Conn:               conn,
			CancelRequestDelay: 100 * time.Millisecond,
			DeadlineDelay:      5 * time.Second,
		}
	}
	if config.TLSConfig != nil {
		config.TLSConfig.MinVersion = tls.VersionTLS12
		config.TLSConfig.ServerName = host
	}
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to runtime pgwire endpoint: %w", err)
	}
	s.admin.Used.Deployment(deployment.ID)
	return &proxySession{conn: conn}, nil
}

type proxySession struct {
	conn *pgx.Conn
}

func (s *proxySession) Close() error {
	return s.conn.Close(context.Background())
}

func (s *proxySession) Describe(ctx context.Context, query string, parameterOIDs []uint32) (*base.Description, error) {
	description, err := s.conn.PgConn().Prepare(ctx, "", query, parameterOIDs)
	if err != nil {
		return nil, err
	}
	return &base.Description{
		ParameterOIDs: append([]uint32(nil), description.ParamOIDs...),
		Fields:        proxyFields(description.Fields),
	}, nil
}

func (s *proxySession) Query(ctx context.Context, query string, parameters []base.Parameter, resultFormats []int16) (base.Rows, error) {
	values := make([][]byte, len(parameters))
	oids := make([]uint32, len(parameters))
	formats := make([]int16, len(parameters))
	for i, parameter := range parameters {
		values[i] = parameter.Value
		oids[i] = parameter.OID
		formats[i] = parameter.Format
	}
	reader := s.conn.PgConn().ExecParams(ctx, query, values, oids, formats, resultFormats)
	return &proxyRows{reader: reader, fields: proxyFields(reader.FieldDescriptions())}, nil
}

func proxyFields(fields []pgconn.FieldDescription) []pgproto3.FieldDescription {
	result := make([]pgproto3.FieldDescription, len(fields))
	for i, field := range fields {
		result[i] = pgproto3.FieldDescription{
			Name:                 []byte(field.Name),
			TableOID:             field.TableOID,
			TableAttributeNumber: field.TableAttributeNumber,
			DataTypeOID:          field.DataTypeOID,
			DataTypeSize:         field.DataTypeSize,
			TypeModifier:         field.TypeModifier,
			Format:               field.Format,
		}
	}
	return result
}

type proxyRows struct {
	reader *pgconn.ResultReader
	fields []pgproto3.FieldDescription
	tag    string
	err    error
	closed bool
}

func (r *proxyRows) Fields() []pgproto3.FieldDescription { return r.fields }
func (r *proxyRows) Next() bool {
	if r.reader.NextRow() {
		return true
	}
	_ = r.Close()
	return false
}
func (r *proxyRows) Values() [][]byte   { return r.reader.Values() }
func (r *proxyRows) Err() error         { return r.err }
func (r *proxyRows) CommandTag() string { return r.tag }
func (r *proxyRows) Close() error {
	if r.closed {
		return r.err
	}
	r.closed = true
	tag, err := r.reader.Close()
	r.tag = tag.String()
	r.err = err
	return err
}
