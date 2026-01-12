package teststarrocks

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// StarRocksVersion is the StarRocks version used for testing
	StarRocksVersion = "4.0.3"
	// StarRocksImage is the Docker image for StarRocks all-in-one container
	StarRocksImage = "starrocks/allin1-ubuntu:" + StarRocksVersion
)

// TestingT satisfies both *testing.T and *testing.B.
type TestingT interface {
	Name() string
	TempDir() string
	FailNow()
	Errorf(format string, args ...interface{})
	Cleanup(f func())
}

// Start starts a StarRocks all-in-one container for testing.
// It returns the DSN for connecting to the container via MySQL protocol (port 9030).
// The container is automatically terminated when the test ends.
func Start(t TestingT) string {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        StarRocksImage,
		ExposedPorts: []string{"9030/tcp", "8030/tcp", "8040/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("9030/tcp"),
			wait.ForLog("Enjoy the journey to StarRocks"),
		).WithDeadline(5 * time.Minute),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		err := container.Terminate(ctx)
		require.NoError(t, err)
	})

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, nat.Port("9030/tcp"))
	require.NoError(t, err)

	dsn := fmt.Sprintf("root:@tcp(%s:%s)/", host, port.Port())
	return dsn
}

// StartWithData starts a StarRocks container and initializes it with test tables.
// Returns DSN for connecting to the container.
func StartWithData(t TestingT) string {
	dsn := Start(t)

	// Wait for StarRocks to be fully ready
	waitForStarRocks(t, dsn)

	// Initialize test database and tables from init.sql
	initTestData(t, dsn)

	return dsn
}

// waitForStarRocks waits for StarRocks to be ready to accept queries
func waitForStarRocks(t TestingT, dsn string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	defer db.Close()

	// Wait until we can execute a simple query
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			require.Fail(t, "timeout waiting for StarRocks to be ready")
			return
		case <-ticker.C:
			_, err := db.ExecContext(ctx, "SELECT 1")
			if err == nil {
				return
			}
		}
	}
}

// initTestData initializes test database and tables from init.sql
func initTestData(t TestingT, dsn string) {
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	defer db.Close()

	// Read init.sql from testdata
	_, currentFile, _, _ := runtime.Caller(0)
	initSQLPath := filepath.Join(filepath.Dir(currentFile), "testdata", "init.sql")

	content, err := os.ReadFile(initSQLPath)
	require.NoError(t, err, "failed to read init.sql")

	// Parse and execute SQL statements
	statements := parseSQLStatements(string(content))
	for _, stmt := range statements {
		_, err := db.Exec(stmt)
		if err != nil {
			// DDL with "IF NOT EXISTS" may have benign failures
			// DML (INSERT) failures are more serious but may happen on re-run
			isDDL := strings.HasPrefix(strings.ToUpper(stmt), "CREATE") ||
				strings.HasPrefix(strings.ToUpper(stmt), "USE")
			if isDDL {
				// DDL errors are usually benign (already exists, etc.)
				continue
			}
			// Log DML errors but continue - data may already exist
			stmtPreview := stmt
			if len(stmtPreview) > 100 {
				stmtPreview = stmtPreview[:100] + "..."
			}
			t.Errorf("Warning executing statement: %v\nStatement: %s", err, stmtPreview)
		}
	}
}

// parseSQLStatements parses SQL file content into individual statements.
// Handles comments and multi-line statements.
func parseSQLStatements(content string) []string {
	var statements []string
	var current strings.Builder

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}

		current.WriteString(line)
		current.WriteString("\n")

		// Check if statement ends with semicolon
		if strings.HasSuffix(trimmed, ";") {
			stmt := strings.TrimSpace(current.String())
			// Remove trailing semicolon for execution
			stmt = strings.TrimSuffix(stmt, ";")
			stmt = strings.TrimSpace(stmt)
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		}
	}

	// Handle any remaining statement without semicolon
	if remaining := strings.TrimSpace(current.String()); remaining != "" {
		statements = append(statements, remaining)
	}

	return statements
}
