package server

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"
	base "github.com/rilldata/rill/runtime/pkg/pgwire"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestPGWireProxyLogsQueryLifecycle(t *testing.T) {
	server, err := base.NewServer(base.Options{NewSession: func(context.Context, map[string]string, string) (base.Session, error) {
		return &proxyTestSession{}, nil
	}})
	require.NoError(t, err)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- server.Serve(ctx, listener) }()
	t.Cleanup(func() {
		cancel()
		require.NoError(t, <-done)
	})
	ctx, timeout := context.WithTimeout(ctx, 5*time.Second)
	defer timeout()
	conn, err := pgconn.Connect(ctx, "postgres://user@"+listener.Addr().String()+"/test?sslmode=disable")
	require.NoError(t, err)
	core, logs := observer.New(zap.InfoLevel)
	session := &proxySession{conn: conn, logger: zap.New(core)}
	defer session.Close()

	_, err = session.Describe(ctx, "prepare error", nil)
	require.ErrorContains(t, err, "prepare failed")
	for _, query := range []string{"select empty", "execute error"} {
		rows, err := session.Query(ctx, query, nil, nil)
		require.NoError(t, err)
		if query == "select empty" {
			require.True(t, rows.Next())
			require.NotNil(t, rows.Values()[0])
			require.Empty(t, rows.Values()[0])
		}
		require.False(t, rows.Next())
		if query == "execute error" {
			require.ErrorContains(t, rows.Err(), "execution failed")
		} else {
			require.NoError(t, rows.Err())
		}
		_ = rows.Close()
		_ = rows.Close()
	}
	require.Equal(t, 1, logs.FilterMessage("pgwire proxy describe started").Len())
	require.Equal(t, 1, logs.FilterMessage("pgwire proxy describe completed").Len())
	require.Equal(t, 2, logs.FilterMessage("pgwire proxy query started").Len())
	completed := logs.FilterMessage("pgwire proxy query completed").All()
	require.Len(t, completed, 2)
	require.Contains(t, completed[0].ContextMap(), "duration")
	require.Contains(t, completed[1].ContextMap(), "error")
}

type proxyTestSession struct{}

func (s *proxyTestSession) Describe(_ context.Context, query string, _ []uint32) (*base.Description, error) {
	if query == "prepare error" {
		return nil, errors.New("prepare failed")
	}
	return &base.Description{Fields: []pgproto3.FieldDescription{{Name: []byte("value"), DataTypeOID: pgtype.TextOID, DataTypeSize: -1, TypeModifier: -1}}}, nil
}

func (s *proxyTestSession) Query(_ context.Context, query string, _ []base.Parameter, _ []int16) (base.Rows, error) {
	if query == "execute error" {
		return nil, errors.New("execution failed")
	}
	return &proxyTestRows{}, nil
}
func (s *proxyTestSession) Close() error { return nil }

type proxyTestRows struct{ sent bool }

func (r *proxyTestRows) Fields() []pgproto3.FieldDescription {
	return []pgproto3.FieldDescription{{Name: []byte("value"), DataTypeOID: pgtype.TextOID, DataTypeSize: -1, TypeModifier: -1}}
}
func (r *proxyTestRows) Next() bool {
	if r.sent {
		return false
	}
	r.sent = true
	return true
}
func (r *proxyTestRows) Values() [][]byte   { return [][]byte{{}} }
func (r *proxyTestRows) Err() error         { return nil }
func (r *proxyTestRows) CommandTag() string { return "SELECT 1" }
func (r *proxyTestRows) Close() error       { return nil }
