package pgwire

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestExtendedErrorRecovery(t *testing.T) {
	address := startTestServer(t, Options{NewSession: func(context.Context, map[string]string, string) (Session, error) {
		return &errorSession{}, nil
	}})
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	conn, err := pgconn.Connect(ctx, fmt.Sprintf("postgres://user@%s/test?sslmode=disable", address))
	require.NoError(t, err)
	defer conn.Close(ctx)

	_, err = conn.Prepare(ctx, "", "parse error", nil)
	require.ErrorContains(t, err, "parse failed")
	_, err = conn.ExecParams(ctx, "execute error", nil, nil, nil, nil).Close()
	require.ErrorContains(t, err, "execution failed")
	result := conn.ExecParams(ctx, "select 7::int4", nil, nil, nil, nil).Read()
	require.NoError(t, result.Err)
	require.Equal(t, [][][]byte{{[]byte("7")}}, result.Rows)
}

type errorSession struct{ testSession }

func (s *errorSession) Describe(ctx context.Context, query string, oids []uint32) (*Description, error) {
	if query == "parse error" {
		return nil, errors.New("parse failed")
	}
	return s.testSession.Describe(ctx, query, oids)
}

func (s *errorSession) Query(ctx context.Context, query string, parameters []Parameter, formats []int16) (Rows, error) {
	if query == "execute error" {
		return nil, errors.New("execution failed")
	}
	return s.testSession.Query(ctx, query, parameters, formats)
}

func TestBindPreservesEmptyParameters(t *testing.T) {
	var output bytes.Buffer
	c := &connection{
		backend:    pgproto3.NewBackend(&output, &output),
		statements: map[string]*preparedStatement{"": {description: &Description{ParameterOIDs: []uint32{pgtype.TextOID, pgtype.ByteaOID, pgtype.TextOID, pgtype.TextOID}}}},
		portals:    make(map[string]*portal),
	}
	input := []byte("value")
	require.NoError(t, c.handleBind(&pgproto3.Bind{Parameters: [][]byte{{}, {}, nil, input}}))
	parameters := c.portals[""].parameters
	require.NotNil(t, parameters[0].Value)
	require.Empty(t, parameters[0].Value)
	require.NotNil(t, parameters[1].Value)
	require.Empty(t, parameters[1].Value)
	require.Nil(t, parameters[2].Value)
	input[0] = 'x'
	require.Equal(t, "value", string(parameters[3].Value))
}

func TestSendResultFlushesWhileReading(t *testing.T) {
	// Generate rows lazily. The writer records how far iteration has advanced
	// when it receives the first batch, rather than buffering a fixture result.
	rows := &generatedRows{value: bytes.Repeat([]byte("x"), 4096), count: 1000}
	writer := &batchWriter{rows: rows}
	c := &connection{backend: pgproto3.NewBackend(bytes.NewReader(nil), writer)}
	suspended, err := c.sendResult(rows, false, 0, nil)
	require.NoError(t, err)
	require.False(t, suspended)
	require.Positive(t, writer.firstRow)
	require.Less(t, writer.firstRow, rows.count)
	require.LessOrEqual(t, writer.largestBatch, 64<<10+len(rows.value)+11)
}

type generatedRows struct {
	value []byte
	count int
	index int
}

func (r *generatedRows) Fields() []pgproto3.FieldDescription { return nil }
func (r *generatedRows) Next() bool {
	if r.index == r.count {
		return false
	}
	r.index++
	return true
}
func (r *generatedRows) Values() [][]byte   { return [][]byte{r.value} }
func (r *generatedRows) Err() error         { return nil }
func (r *generatedRows) CommandTag() string { return "" }
func (r *generatedRows) Close() error       { return nil }

type batchWriter struct {
	rows         *generatedRows
	firstRow     int
	largestBatch int
}

func (w *batchWriter) Write(p []byte) (int, error) {
	if w.firstRow == 0 {
		w.firstRow = w.rows.index
	}
	w.largestBatch = max(w.largestBatch, len(p))
	return len(p), nil
}

func TestShutdownClosesConnectionsDuringStartup(t *testing.T) {
	server, err := NewServer(Options{NewSession: func(context.Context, map[string]string, string) (Session, error) {
		return &testSession{}, nil
	}})
	require.NoError(t, err)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- server.Serve(ctx, listener) }()
	conn, err := net.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)
	defer conn.Close()
	require.Eventually(t, func() bool {
		server.mu.Lock()
		defer server.mu.Unlock()
		return len(server.conns) == 1
	}, 5*time.Second, time.Millisecond)
	cancel()
	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))
	_, err = conn.Read(make([]byte, 1))
	require.ErrorIs(t, err, io.EOF)
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("server did not stop")
	}
}

func TestPortalSuspensionAndCleanup(t *testing.T) {
	for _, closeEarly := range []bool{false, true} {
		t.Run(fmt.Sprintf("close_early_%t", closeEarly), func(t *testing.T) {
			session := &streamTestSession{}
			address := startTestServer(t, Options{NewSession: func(context.Context, map[string]string, string) (Session, error) {
				return session, nil
			}})
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			conn, err := pgconn.Connect(ctx, fmt.Sprintf("postgres://user@%s/test?sslmode=disable", address))
			require.NoError(t, err)
			defer conn.Close(ctx)
			require.NoError(t, conn.Conn().SetDeadline(time.Now().Add(5*time.Second)))
			frontend := conn.Frontend()
			frontend.Send(&pgproto3.Parse{Name: "statement", Query: "stream"})
			frontend.Send(&pgproto3.Bind{PreparedStatement: "statement", DestinationPortal: "portal"})
			frontend.Send(&pgproto3.Execute{Portal: "portal", MaxRows: 2})
			frontend.Send(&pgproto3.Flush{})
			require.NoError(t, frontend.Flush())
			for _, expected := range []pgproto3.BackendMessage{
				&pgproto3.ParseComplete{}, &pgproto3.BindComplete{},
				&pgproto3.DataRow{Values: [][]byte{[]byte("1")}},
				&pgproto3.DataRow{Values: [][]byte{[]byte("2")}},
				&pgproto3.PortalSuspended{},
			} {
				message, err := frontend.Receive()
				require.NoError(t, err)
				require.Equal(t, expected, message)
			}
			require.Equal(t, int32(1), session.opened.Load())
			require.Zero(t, session.closed.Load())

			if closeEarly {
				frontend.Send(&pgproto3.Close{ObjectType: 'P', Name: "portal"})
			} else {
				frontend.Send(&pgproto3.Execute{Portal: "portal"})
			}
			frontend.Send(&pgproto3.Sync{})
			require.NoError(t, frontend.Flush())
			var expected []pgproto3.BackendMessage
			if closeEarly {
				expected = []pgproto3.BackendMessage{&pgproto3.CloseComplete{}}
			} else {
				expected = []pgproto3.BackendMessage{
					&pgproto3.DataRow{Values: [][]byte{[]byte("3")}},
					&pgproto3.CommandComplete{CommandTag: []byte("SELECT 3")},
				}
			}
			expected = append(expected, &pgproto3.ReadyForQuery{TxStatus: 'I'})
			for _, expected := range expected {
				message, err := frontend.Receive()
				require.NoError(t, err)
				require.Equal(t, expected, message)
			}
			require.Equal(t, int32(1), session.opened.Load(), "resuming must reuse the result stream")
			require.Equal(t, int32(1), session.closed.Load())
		})
	}
}

func TestStreamingErrorRecovery(t *testing.T) {
	session := &streamTestSession{}
	address := startTestServer(t, Options{NewSession: func(context.Context, map[string]string, string) (Session, error) {
		return session, nil
	}})
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	conn, err := pgconn.Connect(ctx, fmt.Sprintf("postgres://user@%s/test?sslmode=disable", address))
	require.NoError(t, err)
	defer conn.Close(ctx)
	for _, simple := range []bool{false, true} {
		if simple {
			results, err := conn.Exec(ctx, "stream error").ReadAll()
			require.ErrorContains(t, err, "stream failed")
			require.Len(t, results, 1)
			require.Equal(t, [][][]byte{{[]byte("1")}, {[]byte("2")}, {[]byte("3")}}, results[0].Rows)
		} else {
			result := conn.ExecParams(ctx, "stream error", nil, nil, nil, nil).Read()
			require.ErrorContains(t, result.Err, "stream failed")
			require.Equal(t, [][][]byte{{[]byte("1")}, {[]byte("2")}, {[]byte("3")}}, result.Rows)
		}
		result := conn.ExecParams(ctx, "select 7::int4", nil, nil, nil, nil).Read()
		require.NoError(t, result.Err)
		require.Equal(t, [][][]byte{{[]byte("7")}}, result.Rows)
	}
	require.Equal(t, int32(2), session.opened.Load())
	require.Eventually(t, func() bool { return session.closed.Load() == 2 }, time.Second, time.Millisecond)
}

type streamTestSession struct {
	testSession
	opened atomic.Int32
	closed atomic.Int32
}

func (s *streamTestSession) Query(ctx context.Context, query string, parameters []Parameter, formats []int16) (Rows, error) {
	if query != "stream" && query != "stream error" {
		return s.testSession.Query(ctx, query, parameters, formats)
	}
	description, err := s.Describe(ctx, query, nil)
	if err != nil {
		return nil, err
	}
	s.opened.Add(1)
	var streamErr error
	if query == "stream error" {
		streamErr = errors.New("stream failed")
	}
	return &streamTestRows{
		testRows: testRows{fields: description.Fields, values: [][][]byte{{[]byte("1")}, {[]byte("2")}, {[]byte("3")}}},
		closed:   &s.closed,
		err:      streamErr,
	}, nil
}

type streamTestRows struct {
	testRows
	closed *atomic.Int32
	err    error
}

func (r *streamTestRows) Err() error         { return r.err }
func (r *streamTestRows) CommandTag() string { return "" }
func (r *streamTestRows) Close() error {
	r.closed.Add(1)
	return nil
}
