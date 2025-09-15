package mysql

import (
	"fmt"
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

	cfg := gomysql.NewConfig()
	cfg.User = "root"
	cfg.Passwd = "pass"
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"
	cfg.DBName = "mydb"
	cfg.TLSConfig = "preferred"
	expected := "root:pass@tcp(127.0.0.1:3306)/mydb?tls=preferred"

	fmt.Println(cfg.FormatDSN())
	require.Equal(t, expected, goDSN)
}

func TestResolveGoFormatDSN_SpecialPassword(t *testing.T) {
	cp := &ConfigProperties{
		User:     "test_user",
		Password: "Aa1 ~`!@#$%^&*()_+-={}[]|\\;:'<>\",./?",
		Host:     "localhost",
		Port:     3306,
		Database: "test",
		SSLMode:  "required",
	}
	dsn, err := cp.ResolveDSN()
	require.NoError(t, err)
	expected := "mysql://test_user:Aa1%20~%60%21%40%23$%25%5E%26%2A%28%29_+-%3D%7B%7D%5B%5D%7C%5C;%3A%27%3C%3E%22,.%2F%3F@localhost:3306/test?ssl-mode=required"
	require.Equal(t, expected, dsn)

	expectedGo := "test_user:Aa1 ~`!@#$%^&*()_+-={}[]|\\;:'<>\",./?@tcp(localhost:3306)/test?tls=skip-verify"
	dsnGo, err := cp.resolveGoFormatDSN()
	require.NoError(t, err)
	require.Equal(t, expectedGo, dsnGo)

	cfg := gomysql.NewConfig()
	cfg.User = "test_user"
	cfg.Passwd = "Aa1 ~`!@#$%^&*()_+-={}[]|\\;:'<>\",./?"
	cfg.Net = "tcp"
	cfg.Addr = "localhost:3306"
	cfg.DBName = "test"
	cfg.TLSConfig = "skip-verify"
	require.Equal(t, expectedGo, cfg.FormatDSN())

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
