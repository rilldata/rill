package postgres

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
)

// ResolveLibpqDSN resolves a DSN for libpq-based clients, such as DuckDB's postgres extension.
// libpq reads certificates only from files, so certificate contents are written to dir.
// The caller must keep dir until the client stops opening connections.
func (c *ConfigProperties) ResolveLibpqDSN(dir string) (string, error) {
	dsn := c.ResolveDSN()
	if c.DSN != "" || c.DatabaseURL != "" {
		return dsn, nil
	}
	if c.hasCertificateContents() {
		// libpq skips TLS for Unix sockets: absolute paths, abstract names starting with "@", and the default when host is empty.
		for _, host := range strings.Split(c.Host, ",") {
			host = strings.TrimSpace(host)
			if host == "" || filepath.IsAbs(host) || strings.HasPrefix(host, "@") {
				return "", errors.New("postgres: certificate contents require verified TLS over TCP")
			}
		}
		// Fail with the same errors as pgx instead of libpq's.
		if _, _, err := c.parseCertificates(); err != nil {
			return "", err
		}
	}

	if c.SSLRootCertPEM != "" {
		path, err := writeCertificate(dir, "root.crt", c.SSLRootCertPEM)
		if err != nil {
			return "", err
		}
		dsn += " sslrootcert=" + quotedValue(path)
	} else if c.hasCertificateContents() || c.SSLMode == "verify-full" && !(c.AllowHostAccess && libpqDefaultRootCertExists()) {
		// Like pgx, verify against system roots; without sslrootcert, libpq fails instead.
		// Validate ensures certificate contents imply verify-full, which sslrootcert=system requires.
		dsn += " sslrootcert=system"
	}
	if c.SSLCertPEM != "" {
		certPath, err := writeCertificate(dir, "client.crt", c.SSLCertPEM)
		if err != nil {
			return "", err
		}
		keyPath, err := writeCertificate(dir, "client.key", c.SSLKeyPEM)
		if err != nil {
			return "", err
		}
		dsn += " sslcert=" + quotedValue(certPath) + " sslkey=" + quotedValue(keyPath)
	} else if c.hasCertificateContents() || !c.AllowHostAccess {
		// Don't send a client certificate from PGSSLCERT or the runtime user's home directory.
		dsn += " sslcertmode=disable"
	}
	return dsn, nil
}

func (c *ConfigProperties) hasCertificateContents() bool {
	return c.SSLRootCertPEM != "" || c.SSLCertPEM != "" || c.SSLKeyPEM != ""
}

// pgxConfig parses the connection config, using certificate contents in place of certificate files.
// Like libpq, pgx otherwise reads certificate files from PGSSL* variables and the runtime user's home directory;
// that requires host access, and is skipped when contents are provided.
func (c *ConfigProperties) pgxConfig() (*pgx.ConnConfig, error) {
	dsn := c.ResolveDSN()
	if c.hasCertificateContents() || !c.AllowHostAccess && c.DSN == "" && c.DatabaseURL == "" {
		dsn += " sslrootcert='' sslcert='' sslkey=''"
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	if !c.hasCertificateContents() {
		return cfg, nil
	}

	roots, certs, err := c.parseCertificates()
	if err != nil {
		return nil, err
	}
	configure := func(config *tls.Config) error {
		if config == nil || config.InsecureSkipVerify {
			return errors.New("postgres: certificate contents require verified TLS over TCP")
		}
		if roots != nil {
			config.RootCAs = roots
		}
		if certs != nil {
			config.Certificates = certs
		}
		return nil
	}
	if err := configure(cfg.TLSConfig); err != nil {
		return nil, err
	}
	// pgx creates separate TLS configurations for multiple hosts.
	for _, fallback := range cfg.Fallbacks {
		if err := configure(fallback.TLSConfig); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

func (c *ConfigProperties) parseCertificates() (*x509.CertPool, []tls.Certificate, error) {
	var roots *x509.CertPool
	if c.SSLRootCertPEM != "" {
		roots = x509.NewCertPool()
		if !roots.AppendCertsFromPEM(certificatePEM(c.SSLRootCertPEM)) {
			return nil, nil, errors.New("postgres: sslrootcert_pem must contain valid PEM certificates")
		}
	}
	var certs []tls.Certificate
	if c.SSLCertPEM != "" {
		cert, err := tls.X509KeyPair(certificatePEM(c.SSLCertPEM), certificatePEM(c.SSLKeyPEM))
		if err != nil {
			return nil, nil, errors.New("postgres: invalid client certificate or private key; provide matching PEM certificate and unencrypted private key")
		}
		certs = []tls.Certificate{cert}
	}
	return roots, certs, nil
}

// certificatePEM accepts raw PEM and the base64 encoding used by file uploads.
func certificatePEM(value string) []byte {
	value = strings.TrimSpace(value)
	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil {
		return decoded
	}
	return []byte(value)
}

// writeCertificate writes certificate contents to a file only the runtime user can read; libpq rejects keys with broader permissions.
func writeCertificate(dir, name, value string) (string, error) {
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, certificatePEM(value), 0o600); err != nil {
		return "", err
	}
	return path, nil
}

// libpqDefaultRootCertExists reports whether libpq finds a root certificate without sslrootcert.
func libpqDefaultRootCertExists() bool {
	if os.Getenv("PGSSLROOTCERT") != "" {
		return true
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(home, ".postgresql", "root.crt"))
	return err == nil
}
