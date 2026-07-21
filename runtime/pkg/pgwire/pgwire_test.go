package pgwire

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestServerSimpleAndExtendedProtocol(t *testing.T) {
	var sessions atomic.Int32
	address := startTestServer(t, Options{
		RequirePassword: true,
		NewSession: func(ctx context.Context, parameters map[string]string, password string) (Session, error) {
			require.Equal(t, "secret", password)
			require.Equal(t, "test", parameters["database"])
			sessions.Add(1)
			return &testSession{}, nil
		},
	})

	config, err := pgx.ParseConfig(fmt.Sprintf("postgres://user:secret@%s/test?sslmode=disable", address))
	require.NoError(t, err)
	conn, err := pgx.ConnectConfig(t.Context(), config)
	require.NoError(t, err)
	defer conn.Close(t.Context())

	var value int32
	err = conn.QueryRow(t.Context(), "select $1::int4", int32(42)).Scan(&value)
	require.NoError(t, err)
	require.Equal(t, int32(42), value)

	config.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	simpleConn, err := pgx.ConnectConfig(t.Context(), config)
	require.NoError(t, err)
	defer simpleConn.Close(t.Context())
	err = simpleConn.QueryRow(t.Context(), "select 7::int4").Scan(&value)
	require.NoError(t, err)
	require.Equal(t, int32(7), value)
	require.Equal(t, int32(2), sessions.Load())
}

func TestPreparedStatementsAreConnectionScoped(t *testing.T) {
	address := startTestServer(t, Options{
		NewSession: func(context.Context, map[string]string, string) (Session, error) {
			return &testSession{}, nil
		},
	})
	config, err := pgx.ParseConfig(fmt.Sprintf("postgres://user@%s/test?sslmode=disable", address))
	require.NoError(t, err)
	first, err := pgx.ConnectConfig(t.Context(), config)
	require.NoError(t, err)
	defer first.Close(t.Context())
	second, err := pgx.ConnectConfig(t.Context(), config)
	require.NoError(t, err)
	defer second.Close(t.Context())

	_, err = first.Prepare(t.Context(), "shared_name", "select $1::int4")
	require.NoError(t, err)
	_, err = second.Prepare(t.Context(), "shared_name", "select $1::int4")
	require.NoError(t, err)
	for _, conn := range []*pgx.Conn{first, second} {
		var value int32
		err := conn.QueryRow(t.Context(), "shared_name", int32(9)).Scan(&value)
		require.NoError(t, err)
		require.Equal(t, int32(9), value)
	}
}

func TestServerTLSAndAuthenticationFailure(t *testing.T) {
	certificate := selfSignedCertificate(t)
	address := startTestServer(t, Options{
		TLSConfig:       &tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{certificate}},
		RequirePassword: true,
		NewSession: func(ctx context.Context, parameters map[string]string, password string) (Session, error) {
			if password != "secret" {
				return nil, &Error{Code: "28P01", Message: "invalid password"}
			}
			return &testSession{}, nil
		},
	})

	config, err := pgx.ParseConfig(fmt.Sprintf("postgres://user:wrong@%s/test?sslmode=require", address))
	require.NoError(t, err)
	config.TLSConfig.InsecureSkipVerify = true //nolint:gosec // Self-signed test certificate.
	_, err = pgx.ConnectConfig(t.Context(), config)
	require.Error(t, err)

	config.Password = "secret"
	conn, err := pgx.ConnectConfig(t.Context(), config)
	require.NoError(t, err)
	require.NoError(t, conn.Close(t.Context()))
}

func TestServerCancelRequest(t *testing.T) {
	started := make(chan struct{})
	address := startTestServer(t, Options{
		NewSession: func(context.Context, map[string]string, string) (Session, error) {
			return &testSession{block: started}, nil
		},
	})
	config, err := pgx.ParseConfig(fmt.Sprintf("postgres://user@%s/test?sslmode=disable", address))
	require.NoError(t, err)
	config.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	conn, err := pgx.ConnectConfig(t.Context(), config)
	require.NoError(t, err)
	defer conn.Close(t.Context())

	errCh := make(chan error, 1)
	go func() {
		_, err := conn.Exec(context.Background(), "select blocked")
		errCh <- err
	}()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("query did not start")
	}
	require.NoError(t, conn.PgConn().CancelRequest(t.Context()))
	select {
	case err := <-errCh:
		require.Error(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("query was not cancelled")
	}
}

func selfSignedCertificate(t *testing.T) tls.Certificate {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certificate, err := tls.X509KeyPair(certPEM, keyPEM)
	require.NoError(t, err)
	return certificate
}

func startTestServer(t *testing.T, opts Options) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	server, err := NewServer(opts)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- server.Serve(ctx, listener) }()
	t.Cleanup(func() {
		cancel()
		server.Close()
		require.NoError(t, <-done)
	})
	return listener.Addr().String()
}

type testSession struct {
	block chan struct{}
}

func (s *testSession) Describe(_ context.Context, query string, parameterOIDs []uint32) (*Description, error) {
	if len(parameterOIDs) == 0 && query == "select $1::int4" {
		parameterOIDs = []uint32{pgtype.Int4OID}
	}
	return &Description{
		ParameterOIDs: parameterOIDs,
		Fields: []pgproto3.FieldDescription{{
			Name:         []byte("value"),
			DataTypeOID:  pgtype.Int4OID,
			DataTypeSize: 4,
			TypeModifier: -1,
		}},
	}, nil
}

func (s *testSession) Query(ctx context.Context, query string, parameters []Parameter, resultFormats []int16) (Rows, error) {
	if s.block != nil && query == "select blocked" {
		close(s.block)
		<-ctx.Done()
		return nil, ctx.Err()
	}
	value := int32(7)
	if len(parameters) != 0 {
		if err := pgtype.NewMap().Scan(parameters[0].OID, parameters[0].Format, parameters[0].Value, &value); err != nil {
			return nil, err
		}
	}
	format := int16(0)
	if len(resultFormats) != 0 {
		format = resultFormats[0]
	}
	encoded, err := pgtype.NewMap().Encode(pgtype.Int4OID, format, value, nil)
	if err != nil {
		return nil, err
	}
	return &testRows{
		fields: []pgproto3.FieldDescription{{Name: []byte("value"), DataTypeOID: pgtype.Int4OID, DataTypeSize: 4, TypeModifier: -1, Format: format}},
		values: [][][]byte{{encoded}},
	}, nil
}

func (s *testSession) Close() error { return nil }

type testRows struct {
	fields []pgproto3.FieldDescription
	values [][][]byte
	index  int
}

func (r *testRows) Fields() []pgproto3.FieldDescription { return r.fields }
func (r *testRows) Next() bool {
	if r.index >= len(r.values) {
		return false
	}
	r.index++
	return true
}
func (r *testRows) Values() [][]byte   { return r.values[r.index-1] }
func (r *testRows) Err() error         { return nil }
func (r *testRows) CommandTag() string { return fmt.Sprintf("SELECT %d", len(r.values)) }
func (r *testRows) Close() error       { return nil }
