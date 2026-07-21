package server

import (
	"context"
	"crypto/tls"

	"github.com/rilldata/rill/runtime/pkg/graceful"
	base "github.com/rilldata/rill/runtime/pkg/pgwire"
	"github.com/rilldata/rill/runtime/server/auth"
	pgwireserver "github.com/rilldata/rill/runtime/server/pgwire"
)

// ServePGWire starts the PostgreSQL wire-compatible Metrics SQL endpoint.
func (s *Server) ServePGWire(ctx context.Context, requirePassword bool) error {
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
		RequirePassword: requirePassword,
		Logger:          s.logger,
		NewSession: func(ctx context.Context, parameters map[string]string, password string) (base.Session, error) {
			ctx, err := auth.AuthenticateToken(ctx, s.aud, password)
			if err != nil {
				return nil, err
			}
			instanceID := parameters["database"]
			return pgwireserver.NewSession(s.runtime, instanceID, auth.GetClaims(ctx, instanceID))
		},
	})
	if err != nil {
		return err
	}
	return graceful.ServePGWire(ctx, server, s.opts.PSQLPort, s.logger)
}
