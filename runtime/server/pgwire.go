package server

import (
	"context"

	"github.com/rilldata/rill/runtime/pkg/graceful"
	base "github.com/rilldata/rill/runtime/pkg/pgwire"
	"github.com/rilldata/rill/runtime/server/auth"
	pgwireserver "github.com/rilldata/rill/runtime/server/pgwire"
)

// ServePGWire starts the PostgreSQL wire-compatible Metrics SQL endpoint.
func (s *Server) ServePGWire(ctx context.Context, requirePassword bool) error {
	server, err := base.NewServer(base.Options{
		TLSConfig:       base.NewTLSConfig(s.opts.TLSCertPath, s.opts.TLSKeyPath),
		RequirePassword: requirePassword,
		Logger:          s.logger,
		NewSession: func(ctx context.Context, parameters map[string]string, password string) (base.Session, error) {
			ctx, err := auth.AuthenticateToken(ctx, s.aud, password)
			if err != nil {
				return nil, &base.Error{Code: "28P01", Message: err.Error()}
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
