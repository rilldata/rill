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
	"os"
	"path/filepath"
	"strings"
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
			cfg, err := c.pgxConfig()
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
			_, err := c.pgxConfig()
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
		_, err := c.pgxConfig()
		require.NoError(t, err)
	})
	t.Run("system roots with client certificate", func(t *testing.T) {
		c := &ConfigProperties{Host: "db.example.com", SSLCertPEM: cert, SSLKeyPEM: key}
		cfg, err := c.pgxConfig()
		require.NoError(t, err)
		require.Nil(t, cfg.TLSConfig.RootCAs) // crypto/tls uses system roots.
		require.Len(t, cfg.TLSConfig.Certificates, 1)
	})
	t.Run("reject Unix sockets", func(t *testing.T) {
		_, err := (&ConfigProperties{Host: "/tmp", SSLRootCertPEM: cert}).pgxConfig()
		require.Error(t, err)
	})
	t.Run("multiple hosts", func(t *testing.T) {
		c := &ConfigProperties{Host: "db.example.com,replica.example.com", SSLRootCertPEM: cert, SSLCertPEM: cert, SSLKeyPEM: key}
		cfg, err := c.pgxConfig()
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

func TestResolveLibpqDSN(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PGSSLROOTCERT", "")
	cert, key := testCertificate(t)

	t.Run("certificate files", func(t *testing.T) {
		dir := t.TempDir()
		c := &ConfigProperties{
			Host:           "db.example.com",
			SSLRootCertPEM: base64.StdEncoding.EncodeToString([]byte(cert)),
			SSLCertPEM:     cert,
			SSLKeyPEM:      base64.StdEncoding.EncodeToString([]byte(key)),
		}
		dsn, err := c.ResolveLibpqDSN(dir)
		require.NoError(t, err)
		require.Equal(t, "host=db.example.com sslmode=verify-full sslrootcert="+filepath.Join(dir, "root.crt")+" sslcert="+filepath.Join(dir, "client.crt")+" sslkey="+filepath.Join(dir, "client.key"), dsn)
		for name, want := range map[string]string{"root.crt": cert, "client.crt": cert, "client.key": key} {
			got, err := os.ReadFile(filepath.Join(dir, name))
			require.NoError(t, err)
			require.Equal(t, strings.TrimSpace(want), strings.TrimSpace(string(got)))
			info, err := os.Stat(filepath.Join(dir, name))
			require.NoError(t, err)
			require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
		}
	})

	t.Run("invalid material", func(t *testing.T) {
		dir := t.TempDir()
		_, err := (&ConfigProperties{Host: "db.example.com", SSLCertPEM: cert, SSLKeyPEM: "invalid secret"}).ResolveLibpqDSN(dir)
		require.Error(t, err)
		require.NotContains(t, err.Error(), "invalid secret")
		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		require.Empty(t, entries)
	})

	t.Run("reject Unix sockets", func(t *testing.T) {
		for _, host := range []string{"", "/tmp", "@pg", "db.example.com,/tmp", "db.example.com,"} {
			_, err := (&ConfigProperties{Host: host, SSLRootCertPEM: cert}).ResolveLibpqDSN(t.TempDir())
			require.Error(t, err, host)
		}
	})

	t.Run("host certificates", func(t *testing.T) {
		dir := t.TempDir()
		rootCert := filepath.Join(home, ".postgresql", "root.crt")
		for _, tc := range []struct {
			name         string
			config       *ConfigProperties
			hostRootCert bool
			want         string
		}{
			{name: "client certificate uses system roots", config: &ConfigProperties{Host: "db", SSLCertPEM: cert, SSLKeyPEM: key, AllowHostAccess: true}, hostRootCert: true, want: "host=db sslmode=verify-full sslrootcert=system sslcert=" + filepath.Join(dir, "client.crt") + " sslkey=" + filepath.Join(dir, "client.key")},
			{name: "root certificate disables host client certificate", config: &ConfigProperties{Host: "db", SSLRootCertPEM: cert, AllowHostAccess: true}, want: "host=db sslmode=verify-full sslrootcert=" + filepath.Join(dir, "root.crt") + " sslcertmode=disable"},
			{name: "verify-full uses system roots", config: &ConfigProperties{Host: "db", SSLMode: "verify-full", AllowHostAccess: true}, want: "host=db sslmode=verify-full sslrootcert=system"},
			{name: "verify-full uses host root certificate", config: &ConfigProperties{Host: "db", SSLMode: "verify-full", AllowHostAccess: true}, hostRootCert: true, want: "host=db sslmode=verify-full"},
			{name: "verify-full without host access", config: &ConfigProperties{Host: "db", SSLMode: "verify-full"}, hostRootCert: true, want: "host=db sslmode=verify-full sslrootcert=system sslcertmode=disable"},
			{name: "require without host access", config: &ConfigProperties{Host: "db", SSLMode: "require"}, want: "host=db sslmode=require sslcertmode=disable"},
			{name: "require with host access", config: &ConfigProperties{Host: "db", SSLMode: "require", AllowHostAccess: true}, want: "host=db sslmode=require"},
			{name: "dsn", config: &ConfigProperties{DSN: "host=db sslmode=verify-full"}, want: "host=db sslmode=verify-full"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				require.NoError(t, os.RemoveAll(filepath.Dir(rootCert)))
				if tc.hostRootCert {
					require.NoError(t, os.MkdirAll(filepath.Dir(rootCert), 0o700))
					require.NoError(t, os.WriteFile(rootCert, []byte(cert), 0o600))
				}
				dsn, err := tc.config.ResolveLibpqDSN(dir)
				require.NoError(t, err)
				require.Equal(t, tc.want, dsn)
			})
		}
	})
}

func TestPgxConfigHostAccess(t *testing.T) {
	cert, key := testCertificate(t)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "root.crt"), []byte(cert), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "client.crt"), []byte(cert), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "client.key"), []byte(key), 0o600))
	t.Setenv("PGSSLROOTCERT", filepath.Join(dir, "root.crt"))
	t.Setenv("PGSSLCERT", filepath.Join(dir, "client.crt"))
	t.Setenv("PGSSLKEY", filepath.Join(dir, "client.key"))

	cfg, err := (&ConfigProperties{Host: "db.example.com", SSLMode: "verify-full", AllowHostAccess: true}).pgxConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg.TLSConfig.RootCAs)
	require.Len(t, cfg.TLSConfig.Certificates, 1)

	cfg, err = (&ConfigProperties{Host: "db.example.com", SSLMode: "verify-full"}).pgxConfig()
	require.NoError(t, err)
	require.Nil(t, cfg.TLSConfig.RootCAs)
	require.Empty(t, cfg.TLSConfig.Certificates)
}
