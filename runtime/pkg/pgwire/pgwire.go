// Package pgwire implements the server side of PostgreSQL protocol version 3.
package pgwire

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/jackc/pgx/v5/pgproto3"
	"go.uber.org/zap"
)

const maxMessageBodyLen = 64 << 20

// Parameter is a parameter supplied in an extended-protocol Bind message.
type Parameter struct {
	OID    uint32
	Format int16
	Value  []byte
}

// Description describes a prepared statement.
type Description struct {
	ParameterOIDs []uint32
	Fields        []pgproto3.FieldDescription
}

// Rows is a streaming query result. Values must already be encoded in the
// formats declared by Fields.
type Rows interface {
	Fields() []pgproto3.FieldDescription
	Next() bool
	Values() [][]byte
	Err() error
	CommandTag() string
	Close() error
}

// Session executes queries for one PostgreSQL client connection.
type Session interface {
	Describe(ctx context.Context, query string, parameterOIDs []uint32) (*Description, error)
	Query(ctx context.Context, query string, parameters []Parameter, resultFormats []int16) (Rows, error)
	Close() error
}

// SessionFactory authenticates a startup request and creates connection-local
// query state. Password is empty when RequirePassword is false.
type SessionFactory func(ctx context.Context, parameters map[string]string, password string) (Session, error)

// Options configures a Server.
type Options struct {
	TLSConfig       *tls.Config
	RequirePassword bool
	NewSession      SessionFactory
	Logger          *zap.Logger
}

// Server accepts PostgreSQL wire-compatible connections.
type Server struct {
	opts Options

	closed atomic.Bool
	nextID atomic.Uint32
	mu     sync.Mutex
	conns  map[net.Conn]context.CancelFunc
	cancel map[uint32]*cancelEntry
}

type cancelEntry struct {
	secret []byte
	mu     sync.Mutex
	active context.CancelFunc
}

func (e *cancelEntry) set(cancel context.CancelFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.active = cancel
}

func (e *cancelEntry) clear() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.active = nil
}

func (e *cancelEntry) cancel() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.active != nil {
		e.active()
	}
}

// NewTLSConfig creates a TLS configuration that reloads the certificate for each handshake.
func NewTLSConfig(certPath, keyPath string) *tls.Config {
	if certPath == "" || keyPath == "" {
		return nil
	}
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			certificate, err := tls.LoadX509KeyPair(certPath, keyPath)
			if err != nil {
				return nil, err
			}
			return &certificate, nil
		},
	}
}

// NewServer creates a PostgreSQL wire-compatible server.
func NewServer(opts Options) (*Server, error) {
	if opts.NewSession == nil {
		return nil, errors.New("pgwire: NewSession is required")
	}
	if opts.Logger == nil {
		opts.Logger = zap.NewNop()
	}
	return &Server{
		opts:   opts,
		conns:  make(map[net.Conn]context.CancelFunc),
		cancel: make(map[uint32]*cancelEntry),
	}, nil
}

// Serve accepts connections until ctx is cancelled or the listener fails.
func (s *Server) Serve(ctx context.Context, listener net.Listener) error {
	defer listener.Close()

	errCh := make(chan error, 1)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				errCh <- err
				return
			}
			connCtx, cancel := context.WithCancel(ctx)
			s.mu.Lock()
			if s.closed.Load() {
				s.mu.Unlock()
				cancel()
				_ = conn.Close()
				continue
			}
			s.conns[conn] = cancel
			s.mu.Unlock()
			go s.serveConn(connCtx, conn)
		}
	}()

	select {
	case <-ctx.Done():
		s.Close()
		return nil
	case err := <-errCh:
		if s.closed.Load() || errors.Is(err, net.ErrClosed) {
			return nil
		}
		s.Close()
		return err
	}
}

// Close closes the listener-independent connection state and all active clients.
func (s *Server) Close() {
	if !s.closed.CompareAndSwap(false, true) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for conn, cancel := range s.conns {
		cancel()
		_ = conn.Close()
	}
}

func (s *Server) serveConn(ctx context.Context, conn net.Conn) {
	rawConn := conn
	defer func() {
		s.mu.Lock()
		if cancel, ok := s.conns[rawConn]; ok {
			cancel()
			delete(s.conns, rawConn)
		}
		s.mu.Unlock()
		_ = rawConn.Close()
	}()

	backend := pgproto3.NewBackend(conn, conn)
	backend.SetMaxBodyLen(maxMessageBodyLen)
	startup, password, backend, done, err := s.startup(ctx, backend, conn)
	if done {
		return
	}
	if err != nil {
		s.opts.Logger.Debug("pgwire startup failed", zap.Error(err), zap.Stringer("remote", conn.RemoteAddr()))
		return
	}

	session, err := s.opts.NewSession(ctx, startup.Parameters, password)
	if err != nil {
		sendError(backend, err, "28P01")
		_ = backend.Flush()
		return
	}
	defer session.Close()

	pid := s.nextID.Add(1)
	if pid == 0 {
		pid = s.nextID.Add(1)
	}
	secret := make([]byte, 4)
	if _, err := rand.Read(secret); err != nil {
		s.opts.Logger.Error("failed to generate pgwire cancellation secret", zap.Error(err))
		return
	}
	cancelEntry := &cancelEntry{secret: secret}
	s.mu.Lock()
	s.cancel[pid] = cancelEntry
	s.mu.Unlock()
	defer func() {
		cancelEntry.cancel()
		s.mu.Lock()
		delete(s.cancel, pid)
		s.mu.Unlock()
	}()

	backend.Send(&pgproto3.AuthenticationOk{})
	for key, value := range map[string]string{
		"server_version":              "16.3",
		"server_encoding":             "UTF8",
		"client_encoding":             "UTF8",
		"DateStyle":                   "ISO, MDY",
		"integer_datetimes":           "on",
		"standard_conforming_strings": "on",
		"TimeZone":                    "Etc/UTC",
	} {
		backend.Send(&pgproto3.ParameterStatus{Name: key, Value: value})
	}
	backend.Send(&pgproto3.BackendKeyData{ProcessID: pid, SecretKey: secret})
	backend.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
	if err := backend.Flush(); err != nil {
		return
	}

	c := &connection{
		backend:    backend,
		session:    session,
		ctx:        ctx,
		cancel:     cancelEntry,
		statements: make(map[string]*preparedStatement),
		portals:    make(map[string]*portal),
		txStatus:   'I',
	}
	if err := c.run(); err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) && !errors.Is(err, context.Canceled) {
		s.opts.Logger.Debug("pgwire connection closed with error", zap.Error(err), zap.Stringer("remote", conn.RemoteAddr()))
	}
}

func (s *Server) startup(ctx context.Context, backend *pgproto3.Backend, conn net.Conn) (*pgproto3.StartupMessage, string, *pgproto3.Backend, bool, error) {
	for {
		message, err := backend.ReceiveStartupMessage()
		if err != nil {
			return nil, "", backend, false, err
		}
		switch message := message.(type) {
		case *pgproto3.SSLRequest:
			if s.opts.TLSConfig == nil {
				if _, err := conn.Write([]byte("N")); err != nil {
					return nil, "", backend, false, err
				}
				continue
			}
			if _, err := conn.Write([]byte("S")); err != nil {
				return nil, "", backend, false, err
			}
			tlsConn := tls.Server(conn, s.opts.TLSConfig.Clone())
			if err := tlsConn.HandshakeContext(ctx); err != nil {
				return nil, "", backend, false, err
			}
			conn = tlsConn
			backend = pgproto3.NewBackend(conn, conn)
			backend.SetMaxBodyLen(maxMessageBodyLen)
		case *pgproto3.GSSEncRequest:
			if _, err := conn.Write([]byte("N")); err != nil {
				return nil, "", backend, false, err
			}
		case *pgproto3.CancelRequest:
			s.mu.Lock()
			entry, ok := s.cancel[message.ProcessID]
			s.mu.Unlock()
			if ok && bytes.Equal(entry.secret, message.SecretKey) {
				entry.cancel()
			}
			return nil, "", backend, true, nil
		case *pgproto3.StartupMessage:
			var password string
			if s.opts.RequirePassword {
				backend.Send(&pgproto3.AuthenticationCleartextPassword{})
				if err := backend.SetAuthType(pgproto3.AuthTypeCleartextPassword); err != nil {
					return nil, "", backend, false, err
				}
				if err := backend.Flush(); err != nil {
					return nil, "", backend, false, err
				}
				response, err := backend.Receive()
				if err != nil {
					return nil, "", backend, false, err
				}
				passwordMessage, ok := response.(*pgproto3.PasswordMessage)
				if !ok {
					return nil, "", backend, false, fmt.Errorf("expected password message, got %T", response)
				}
				password = passwordMessage.Password
			}
			return message, password, backend, false, nil
		default:
			return nil, "", backend, false, fmt.Errorf("unsupported startup message %T", message)
		}
	}
}
