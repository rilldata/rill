package teststarrocks

import (
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcwait "github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/go-sql-driver/mysql" // MySQL driver for database/sql
)

const (
	// StarRocksVersion is the StarRocks version used for testing
	StarRocksVersion = "4.0.4"
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

// StarRocksInfo contains connection info for a StarRocks container
type StarRocksInfo struct {
	Host              string // Container host
	DSN               string // MySQL protocol DSN (port 9030)
	FEHTTPAddr        string // FE HTTP address for Stream Load (port 8030)
	BEHTTPAddr        string // BE HTTP address for Stream Load redirect (port 8040)
	FlightSQLPort     int    // Arrow Flight SQL port (mapped from FE's arrow_flight_port 9408)
	FlightSQLBEPort   int    // Arrow Flight SQL BE port (mapped from BE's arrow_flight_port 9419)
}

// Start starts a StarRocks all-in-one container for testing.
// It returns connection info for the container.
// The container is automatically terminated when the test ends.
func Start(t TestingT) StarRocksInfo {
	ctx := context.Background()

	// Mount custom be.conf and fe.conf to enable Arrow Flight SQL
	_, currentFile, _, _ := runtime.Caller(0)
	testdataDir := filepath.Join(filepath.Dir(currentFile), "testdata")
	beConfPath := filepath.Join(testdataDir, "be.conf")
	feConfPath := filepath.Join(testdataDir, "fe.conf")

	req := testcontainers.ContainerRequest{
		Image:        StarRocksImage,
		ExposedPorts: []string{"9030/tcp", "8030/tcp", "8040/tcp", "9408/tcp", "9419/tcp"},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      beConfPath,
				ContainerFilePath: "/data/deploy/starrocks/be/conf/be.conf",
				FileMode:          0644,
			},
			{
				HostFilePath:      feConfPath,
				ContainerFilePath: "/data/deploy/starrocks/fe/conf/fe.conf",
				FileMode:          0644,
			},
		},
		WaitingFor: tcwait.ForAll(
			tcwait.ForListeningPort("9030/tcp"),
			tcwait.ForListeningPort("8030/tcp"),
			tcwait.ForListeningPort("8040/tcp"),
			tcwait.ForListeningPort("9419/tcp"), // BE Arrow Flight SQL
			tcwait.ForLog("Enjoy the journey to StarRocks"),
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

	mysqlPort, err := container.MappedPort(ctx, nat.Port("9030/tcp"))
	require.NoError(t, err)

	feHTTPPort, err := container.MappedPort(ctx, nat.Port("8030/tcp"))
	require.NoError(t, err)

	beHTTPPort, err := container.MappedPort(ctx, nat.Port("8040/tcp"))
	require.NoError(t, err)

	flightSQLPort, err := container.MappedPort(ctx, nat.Port("9408/tcp"))
	require.NoError(t, err)

	flightSQLBEPort, err := container.MappedPort(ctx, nat.Port("9419/tcp"))
	require.NoError(t, err)

	return StarRocksInfo{
		Host:            host,
		DSN:             fmt.Sprintf("root:@tcp(%s:%s)/?parseTime=true&loc=UTC", host, mysqlPort.Port()),
		FEHTTPAddr:      fmt.Sprintf("%s:%s", host, feHTTPPort.Port()),
		BEHTTPAddr:      fmt.Sprintf("%s:%s", host, beHTTPPort.Port()),
		FlightSQLPort:   flightSQLPort.Int(),
		FlightSQLBEPort: flightSQLBEPort.Int(),
	}
}

// StartWithData starts a StarRocks container and initializes it with test tables.
// Returns DSN for connecting to the container.
func StartWithData(t TestingT) string {
	info := StartWithDataFull(t)
	return info.DSN
}

// StartWithDataFull starts a StarRocks container with test data and returns full connection info.
// Use this when you need access to Arrow Flight SQL port or other connection details.
func StartWithDataFull(t TestingT) StarRocksInfo {
	info := Start(t)

	// Wait for StarRocks to be fully ready
	waitForStarRocks(t, info.DSN)

	// Initialize test database and tables from init.sql
	initTestData(t, info.DSN)

	// Load ad_bids data from CSV via Stream Load
	loadAdBidsData(t, info.FEHTTPAddr, info.BEHTTPAddr)

	return info
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

// streamLoadResponse represents the JSON response from StarRocks Stream Load
type streamLoadResponse struct {
	Status string `json:"Status"`
	Msg    string `json:"Message"`
}

// loadAdBidsData loads ad_bids data from CSV file using StarRocks Stream Load API
func loadAdBidsData(t TestingT, feHTTPAddr, beHTTPAddr string) {
	// Find the AdBids.csv.gz file
	_, currentFile, _, _ := runtime.Caller(0)
	// Go up from teststarrocks -> starrocks -> drivers -> runtime -> testruntime/testdata/ad_bids/data
	csvGzPath := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "testruntime", "testdata", "ad_bids", "data", "AdBids.csv.gz")

	// Open and decompress the gzip file
	gzFile, err := os.Open(csvGzPath)
	require.NoError(t, err, "failed to open AdBids.csv.gz")
	defer gzFile.Close()

	gzReader, err := gzip.NewReader(gzFile)
	require.NoError(t, err, "failed to create gzip reader")
	defer gzReader.Close()

	// Read decompressed CSV data into memory
	csvData, err := io.ReadAll(gzReader)
	require.NoError(t, err, "failed to read CSV data")

	// Create HTTP request for Stream Load
	url := fmt.Sprintf("http://%s/api/test_db/ad_bids/_stream_load", feHTTPAddr)
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(string(csvData)))
	require.NoError(t, err, "failed to create Stream Load request")

	// Set required headers for Stream Load
	req.Header.Set("Expect", "100-continue")
	req.Header.Set("column_separator", ",")
	req.Header.Set("skip_header", "1") // Skip CSV header row
	// Use NULLIF to convert empty strings to NULL for publisher and domain columns (matches DuckDB behavior)
	req.Header.Set("columns", "id, timestamp, tmp_publisher, tmp_domain, bid_price, publisher=NULLIF(tmp_publisher, ''), domain=NULLIF(tmp_domain, '')")
	req.SetBasicAuth("root", "")

	// Create HTTP client with custom redirect policy
	// StarRocks FE redirects to BE for Stream Load, but the redirect URL contains
	// the internal container address. We need to rewrite it to the mapped host port.
	client := &http.Client{
		Timeout: 2 * time.Minute,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Rewrite the redirect URL to use the correct BE address
			req.URL.Host = beHTTPAddr
			// Preserve auth header on redirect (like curl --location-trusted)
			if len(via) > 0 {
				req.SetBasicAuth("root", "")
			}
			return nil
		},
	}
	resp, err := client.Do(req)
	require.NoError(t, err, "failed to execute Stream Load request")
	defer resp.Body.Close()

	// Parse response
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "failed to read Stream Load response")

	var result streamLoadResponse
	err = json.Unmarshal(body, &result)
	require.NoError(t, err, "failed to parse Stream Load response: %s", string(body))

	require.Equal(t, "Success", result.Status, "Stream Load failed: %s", result.Msg)
}
