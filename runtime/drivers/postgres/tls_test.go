package postgres

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func testCertificate(t *testing.T) (string, string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "db.example.com"},
		DNSNames: []string{"db.example.com"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true,
		KeyUsage:    x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})), string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}))
}

func TestCertificateTLS(t *testing.T) {
	// Isolate pgx defaults from developer credentials and TLS environment settings.
	t.Setenv("HOME", t.TempDir())
	for _, name := range []string{"PGSSLMODE", "PGSSLROOTCERT", "PGSSLCERT", "PGSSLKEY", "PGSERVICE", "PGSERVICEFILE", "PGHOST"} {
		t.Setenv(name, "")
	}
	cert, key := testCertificate(t)
	otherCert, otherKey := testCertificate(t)
	serverCert, err := tls.X509KeyPair([]byte(cert), []byte(key))
	require.NoError(t, err)
	roots := x509.NewCertPool()
	require.True(t, roots.AppendCertsFromPEM([]byte(cert)))

	for _, tc := range []struct {
		name      string
		host      string
		root      string
		client    bool
		encoded   bool
		wantError bool
	}{
		{name: "mutual TLS", host: "db.example.com", root: cert, client: true},
		{name: "base64 mutual TLS", host: "db.example.com", root: cert, client: true, encoded: true},
		{name: "wrong hostname", host: "wrong.example.com", root: cert, client: true, wantError: true},
		{name: "untrusted server", host: "db.example.com", root: otherCert, client: true, wantError: true},
		{name: "missing client certificate", host: "db.example.com", root: cert, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := &ConfigProperties{Host: tc.host, User: "rill", SSLRootCertPEM: tc.root}
			if tc.client {
				c.SSLCertPEM = cert
				c.SSLKeyPEM = key
			}
			if tc.encoded {
				c.SSLRootCertPEM = base64.StdEncoding.EncodeToString([]byte(c.SSLRootCertPEM))
				c.SSLCertPEM = base64.StdEncoding.EncodeToString([]byte(c.SSLCertPEM))
				c.SSLKeyPEM = base64.StdEncoding.EncodeToString([]byte(c.SSLKeyPEM))
			}
			cfg, err := c.connectionConfig()
			require.NoError(t, err)
			require.False(t, cfg.TLSConfig.InsecureSkipVerify)
			left, right := net.Pipe()
			defer left.Close()
			defer right.Close()
			require.NoError(t, left.SetDeadline(time.Now().Add(5*time.Second)))
			require.NoError(t, right.SetDeadline(time.Now().Add(5*time.Second)))
			// TLS 1.2 makes rejection of the client certificate observable during both handshakes.
			server := tls.Server(left, &tls.Config{Certificates: []tls.Certificate{serverCert}, ClientCAs: roots, ClientAuth: tls.RequireAndVerifyClientCert, MaxVersion: tls.VersionTLS12})
			client := tls.Client(right, cfg.TLSConfig)
			done := make(chan error, 1)
			go func() { done <- server.Handshake(); left.Close() }()
			clientErr := client.Handshake()
			serverErr := <-done
			if tc.wantError {
				require.Error(t, clientErr)
				require.Error(t, serverErr)
			} else {
				require.NoError(t, clientErr)
				require.NoError(t, serverErr)
			}
		})
	}
	t.Run("invalid material", func(t *testing.T) {
		for _, c := range []*ConfigProperties{
			{Host: "db.example.com", SSLRootCertPEM: "invalid secret"},
			{Host: "db.example.com", SSLCertPEM: cert, SSLKeyPEM: otherKey},
			{Host: "db.example.com", SSLCertPEM: cert, SSLKeyPEM: "invalid secret"},
		} {
			_, err := c.connectionConfig()
			require.Error(t, err)
			require.NotContains(t, err.Error(), "invalid secret")
			require.NotContains(t, err.Error(), otherKey)
		}
	})
	t.Run("explicit contents override host certificate files", func(t *testing.T) {
		t.Setenv("PGSSLROOTCERT", "/missing/root.crt")
		t.Setenv("PGSSLCERT", "/missing/postgresql.crt")
		t.Setenv("PGSSLKEY", "/missing/postgresql.key")
		c := &ConfigProperties{Host: "db.example.com", SSLRootCertPEM: cert, SSLCertPEM: cert, SSLKeyPEM: key}
		_, err := c.connectionConfig()
		require.NoError(t, err)
	})
	t.Run("system roots with client certificate", func(t *testing.T) {
		c := &ConfigProperties{Host: "db.example.com", SSLCertPEM: cert, SSLKeyPEM: key}
		cfg, err := c.connectionConfig()
		require.NoError(t, err)
		require.Nil(t, cfg.TLSConfig.RootCAs) // crypto/tls uses system roots.
		require.Len(t, cfg.TLSConfig.Certificates, 1)
	})
	t.Run("reject Unix sockets", func(t *testing.T) {
		_, err := (&ConfigProperties{Host: "/tmp", SSLRootCertPEM: cert}).connectionConfig()
		require.Error(t, err)
	})
	t.Run("multiple hosts", func(t *testing.T) {
		c := &ConfigProperties{Host: "db.example.com,replica.example.com", SSLRootCertPEM: cert, SSLCertPEM: cert, SSLKeyPEM: key}
		cfg, err := c.connectionConfig()
		require.NoError(t, err)
		require.NotEmpty(t, cfg.Fallbacks)
		for _, f := range cfg.Fallbacks {
			require.NotNil(t, f.TLSConfig)
			require.False(t, f.TLSConfig.InsecureSkipVerify)
			require.Equal(t, cfg.TLSConfig.RootCAs, f.TLSConfig.RootCAs)
			require.Len(t, f.TLSConfig.Certificates, 1)
		}
	})
}

func TestCertificateValidation(t *testing.T) {
	for _, c := range []*ConfigProperties{
		{SSLCertPEM: "cert"}, {SSLKeyPEM: "key"},
		{DSN: "host=db.example.com", SSLRootCertPEM: "ca"},
		{DatabaseURL: "postgres://db.example.com", SSLRootCertPEM: "ca"},
	} {
		require.Error(t, c.Validate())
	}
	for _, mode := range []string{"disable", "allow", "prefer", "require", "verify-ca"} {
		require.Error(t, (&ConfigProperties{SSLRootCertPEM: "ca", SSLMode: mode}).Validate())
	}
	require.NoError(t, (&ConfigProperties{SSLRootCertPEM: "ca"}).Validate())
	require.NoError(t, (&ConfigProperties{SSLRootCertPEM: "ca", SSLMode: "verify-full"}).Validate())
	require.Equal(t, "host=db.example.com sslmode=verify-full", (&ConfigProperties{Host: "db.example.com", SSLRootCertPEM: "ca"}).ResolveDSN())
	require.Equal(t, "host=db.example.com", (&ConfigProperties{Host: "db.example.com"}).ResolveDSN())
	require.Equal(t, "postgres://db.example.com?sslmode=require", (&ConfigProperties{DSN: "postgres://db.example.com?sslmode=require"}).ResolveDSN())
}
