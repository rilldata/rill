package mysql

import (
	"testing"

	gomysql "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func TestResolveDSN_WithOnlyDSN(t *testing.T) {
	c := &ConfigProperties{
		DSN: "mysql://user:pass@localhost:3306/dbname?ssl-mode=disable",
	}

	dsn, err := c.ResolveDSN()
	require.NoError(t, err)
	require.Equal(t, c.DSN, dsn)
}

func TestResolveDSN_WithIndividualFields(t *testing.T) {
	c := &ConfigProperties{
		Host:     "db.example.com",
		Port:     3306,
		User:     "admin",
		Password: "secret",
		Database: "analytics",
		SSLMode:  "prefer",
	}

	dsn, err := c.ResolveDSN()
	require.NoError(t, err)

	expected := "mysql://admin:secret@db.example.com:3306/analytics?ssl-mode=prefer"
	require.Equal(t, expected, dsn)
}

func TestResolveDSN_WithDSNAndIndividualFields_ShouldError(t *testing.T) {
	c := &ConfigProperties{
		DSN:      "mysql://some:dsn@localhost/db",
		Host:     "localhost",
		Port:     3306,
		User:     "user",
		Database: "testdb",
	}

	_, err := c.ResolveDSN()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid config")
}

func TestResolveGoFormatDSN_Basic(t *testing.T) {
	c := &ConfigProperties{
		User:     "root",
		Password: "pass",
		Host:     "127.0.0.1",
		Port:     3306,
		Database: "mydb",
		SSLMode:  "preferred",
	}

	goDSN, err := c.resolveGoFormatDSN()
	require.NoError(t, err)

	cfg := gomysql.Config{
		User:      "root",
		Passwd:    "pass",
		Addr:      "127.0.0.1:3306",
		DBName:    "mydb",
		TLSConfig: "preferred",
	}
	expected := cfg.FormatDSN()
	require.Equal(t, expected, goDSN)
}

func TestResolveGoFormatDSN_InvalidSSLMode(t *testing.T) {
	c := &ConfigProperties{
		User:     "root",
		Password: "pass",
		Host:     "localhost",
		Port:     3306,
		Database: "mydb",
		SSLMode:  "invalid-mode",
	}

	_, err := c.resolveGoFormatDSN()
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported ssl-mode")
}
