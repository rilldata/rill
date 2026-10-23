package postgres

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (c *ConfigProperties) hasCertificateContents() bool {
	return c.SSLRootCertPEM != "" || c.SSLCertPEM != "" || c.SSLKeyPEM != ""
}

// certificatePEM accepts raw PEM and the base64 encoding used by file uploads.
func certificatePEM(value string) []byte {
	value = strings.TrimSpace(value)
	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil {
		return decoded
	}
	return []byte(value)
}

func (c *ConfigProperties) connectionConfig() (*pgx.ConnConfig, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	// Explicit contents must not cause pgx to read unrelated credential files
	// from PGSSL* variables or the runtime user's home directory.
	cfg, err := pgx.ParseConfig(c.ResolveDSN() + " sslrootcert='' sslcert='' sslkey=''")
	if err != nil {
		// pgx errors can include the connection string and its password.
		return nil, errors.New("postgres: invalid connection configuration")
	}
	var roots *x509.CertPool
	if c.SSLRootCertPEM != "" {
		roots = x509.NewCertPool()
		if !roots.AppendCertsFromPEM(certificatePEM(c.SSLRootCertPEM)) {
			return nil, errors.New("postgres: sslrootcert_pem must contain valid PEM certificates")
		}
	}
	var certs []tls.Certificate
	if c.SSLCertPEM != "" {
		cert, err := tls.X509KeyPair(certificatePEM(c.SSLCertPEM), certificatePEM(c.SSLKeyPEM))
		if err != nil {
			return nil, errors.New("postgres: invalid client certificate or private key; provide matching PEM certificate and unencrypted private key")
		}
		certs = []tls.Certificate{cert}
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
			config.GetClientCertificate = nil
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
