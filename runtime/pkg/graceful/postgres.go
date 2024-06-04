package graceful

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"strings"

	wire "github.com/jeroenrinzema/psql-wire"
	"go.uber.org/zap"
)

type ServePSQLOptions struct {
	QueryHandler wire.ParseFn
	AuthHandler  func(ctx context.Context, username, password string) (context.Context, bool, error)
	Port         int
	TLSCertPath  string
	TLSKeyPath   string
	Logger       *zap.Logger
}

// ServePSQL serves a Postgres server and performs a graceful shutdown if/when ctx is cancelled.
func ServePSQL(ctx context.Context, serveOpts *ServePSQLOptions) error {
	// Calling net.Listen("tcp", ...) will succeed if the port is blocked on IPv4 but not on IPv6.
	// This workaround ensures we get the port on IPv4 (and most likely also on IPv6).
	lis, err := net.Listen("tcp4", fmt.Sprintf(":%d", serveOpts.Port))
	if err == nil {
		lis.Close()
		lis, err = net.Listen("tcp", fmt.Sprintf(":%d", serveOpts.Port))
	}
	if err != nil {
		if strings.Contains(err.Error(), "address already in use") {
			return fmt.Errorf("psql port %d is in use by another process. Either kill that process or pass `--port-psql PORT` to run Rill on another port", serveOpts.Port)
		}
		return err
	}

	opts := []wire.OptionFn{
		wire.Version("16.3"),
		wire.Logger(slog.New(discard)),
		wire.GlobalParameters(wire.Parameters{"standard_conforming_strings": "on"}),
	}
	if serveOpts.AuthHandler != nil {
		opts = append(opts, wire.SessionAuthStrategy(wire.ClearTextPassword(serveOpts.AuthHandler)))
	}
	if serveOpts.TLSCertPath != "" && serveOpts.TLSKeyPath != "" {
		certificates := make([]tls.Certificate, 1)
		certificates[0], err = tls.LoadX509KeyPair(serveOpts.TLSCertPath, serveOpts.TLSKeyPath)
		if err != nil {
			return err
		}
		opts = append(opts, wire.Certificates(certificates))
	}

	server, err := wire.NewServer(serveOpts.QueryHandler, opts...)
	if err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(ctx)
	var serveErr error
	go func() {
		serveErr = server.Serve(lis)
		cancel()
	}()

	<-cctx.Done()
	if serveErr == nil {
		server.Close()
	}
	return serveErr
}

// discard is a slog.Handler which is always disabled and therefore logs nothing.
var discard slog.Handler = discardHandler{}

type discardHandler struct{}

func (discardHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (discardHandler) Handle(context.Context, slog.Record) error { return nil }
func (d discardHandler) WithAttrs([]slog.Attr) slog.Handler      { return d }
func (d discardHandler) WithGroup(string) slog.Handler           { return d }
