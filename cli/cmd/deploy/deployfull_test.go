package deploy_test

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/shlex"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/rilldata/rill/admin/pkg/authtoken"
	"github.com/rilldata/rill/cli/cmd"
	"github.com/rilldata/rill/cli/cmd/devtool"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDeployE2E(t *testing.T) {
	testmode.Expensive(t)
	rillhome := t.TempDir()
	ch := mustNewHelper(t, rillhome)
	now := time.Now()
	cancelFn, closeChan := runDevtoolStartCloud(t, ch)
	t.Logf("services started in %v", time.Since(now))
	// Add test user
	addUser(t, ch)

	// Discover all test files in tests directory
	testFiles, err := discoverTestFiles("cli/cmd/deploy/tests")
	require.NoError(t, err)
	require.NotEmpty(t, testFiles, "No test files found in tests directory")

	// Run each test file
	for _, testFile := range testFiles {
		testName := filepath.Base(testFile)
		t.Run(testName, func(t *testing.T) {
			err := runTestFile(t, ch, testFile)
			require.NoError(t, err)
		})
	}

	// Stop devtool - cancel will send SIGINT and wait for graceful shutdown
	cancelFn()
	<-closeChan
	t.Log("Devtool shutdown complete")
}

// runDevtoolStartCloud runs `rill devtool start` as a subprocess with the specified preset and args.
func runDevtoolStartCloud(t *testing.T, ch *cmdutil.Helper) (context.CancelFunc, chan error) {
	// Change to repo root directory
	// Tests run from cli/cmd/deploy, need to go up to repo root to start devtool
	err := os.Chdir("../../..")
	require.NoError(t, err)

	// set a temporary RILL_DEVTOOL_STATE_DIRECTORY
	path, err := os.MkdirTemp(".", "rill-devtool-state-")
	require.NoError(t, err)
	path = filepath.Clean(path)
	t.Setenv("RILL_DEVTOOL_STATE_DIRECTORY", path)
	t.Cleanup(func() {
		err := os.RemoveAll(path)
		require.NoError(t, err)
	})

	ctx, cancel := context.WithCancel(t.Context())
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		err := devtool.Start(ctx, ch, "minimal", false, false, false, &devtool.ServicesCfg{
			Admin:   true,
			Runtime: true,
			Deps:    true,
		})
		if err != nil {
			errChan <- err
		}
	}()

	// Poll every 1 second to check if services are ready
	// The test will timeout after configured test timeout
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case err := <-errChan:
			t.Fatalf("`rill devtool start` failed with error: %s", err)
		case <-t.Context().Done():
			t.Fatal("timeout waiting for devtool to start")
		case <-ticker.C:
			// check for admin ready at `http://localhost:8080/v1/ping`
			resp, err := http.Get("http://localhost:8080/v1/ping")
			if err != nil || resp.StatusCode != http.StatusOK {
				if resp != nil {
					resp.Body.Close()
				}
				continue
			}
			resp.Body.Close()

			// check for runtime ready at `http://localhost:8081/v1/ping`
			resp, err = http.Get("http://localhost:8081/v1/ping")
			if err != nil || resp.StatusCode != http.StatusOK {
				if resp != nil {
					resp.Body.Close()
				}
				continue
			}
			resp.Body.Close()

			// Both services are ready
			t.Log("Admin and Runtime are ready")
			return cancel, errChan
		}
	}
}

// TestConfig represents the YAML test configuration
type TestConfig struct {
	Project              string        `yaml:"project"`
	InitialWaitReconcile time.Duration `yaml:"initial_wait_reconcile"`
	Tests                []TestCase    `yaml:"tests"`
}

// TestCase represents a single test case
type TestCase struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
	Output  string `yaml:"output"`
}

// discoverTestFiles finds all YAML test files in the specified directory
func discoverTestFiles(testsDir string) ([]string, error) {
	entries, err := os.ReadDir(testsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read tests directory: %w", err)
	}

	var testFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".yaml") || strings.HasSuffix(entry.Name(), ".yml") {
			testFiles = append(testFiles, filepath.Join(testsDir, entry.Name()))
		}
	}

	return testFiles, nil
}

// runTestFile executes a single test file end-to-end
func runTestFile(t *testing.T, ch *cmdutil.Helper, testFilePath string) error {
	t.Logf("Running test file: %s", testFilePath)

	// Load test configuration
	testConfig, err := loadTestConfig(testFilePath)
	if err != nil {
		return err
	}

	// copy this to a temp directory to avoid creating state in the repo like setting git remote, creating tmp files etc.
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, testConfig.Project)
	err = os.MkdirAll(projectDir, 0o755)
	require.NoError(t, err)
	err = copyDir(projectDir, filepath.Join("cli/cmd/deploy/test-files", testConfig.Project))
	require.NoError(t, err)

	// create org
	org := uuid.New().String()
	res := runWithInput(t.Context(), ch, "", "org", "create", org)
	require.Equal(t, 0, res.ExitCode, "failed to create org: %s", res.Output)

	t.Cleanup(func() {
		// delete the org
		res := runWithInput(context.Background(), ch, "", "org", "delete", org, "--interactive=false")
		if res.ExitCode != 0 {
			t.Errorf("failed to delete org %s: %s", org, res.Output)
		}
	})

	// Deploy the project
	t.Logf("Deploying project: %s", projectDir)
	res = runWithInput(t.Context(), ch, "", "deploy", projectDir, "--project", testConfig.Project, "--managed", "true")
	require.Equal(t, 0, res.ExitCode, "failed to deploy project: %s", res.Output)
	t.Logf("Project deployed successfully: %s", res.Output)

	// Verify deployment status
	time.Sleep(testConfig.InitialWaitReconcile)
	err = checkDeploymentStatus(t, ch, testConfig.Project)
	if err != nil {
		return err
	}

	// Execute test cases
	if err := executeTestCases(t, ch, testConfig.Project, testConfig.Tests); err != nil {
		return err
	}
	return nil
}

// loadTestConfig loads the test configuration from a YAML file
func loadTestConfig(path string) (*TestConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read test config: %w", err)
	}

	var config TestConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse test config: %w", err)
	}

	return &config, nil
}

// verifyDeploymentIdle checks if all resources are in Idle state with retries
func verifyDeploymentIdle(t *testing.T, ch *cmdutil.Helper, projectName string) error {
	t.Log("Verifying deployment status...")
	return checkDeploymentStatus(t, ch, projectName)
}

// executeTestCases runs all test cases and validates their output
func executeTestCases(t *testing.T, ch *cmdutil.Helper, projectName string, tests []TestCase) error {
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			parts, err := shlex.Split(test.Command)
			if err != nil {
				t.Fatalf("failed to parse command: %v", err)
			}
			if len(parts) == 0 {
				t.Fatal("empty command")
			}

			// Verify it's a rill command
			if parts[0] != "rill" {
				t.Fatalf("command %q is not a valid rill command", test.Command)
			}

			parts = append(parts, "--project", projectName)
			res := runWithInput(t.Context(), ch, "", parts[1:]...)
			require.Equal(t, 0, res.ExitCode, "failed to run command: %s", res.Output)

			// Validate output matches expected
			require.Equal(t, strings.TrimSpace(res.Output), strings.TrimSpace(test.Output))
		})
	}
	return nil
}

// checkDeploymentStatus checks if all resources are in Idle state
func checkDeploymentStatus(t *testing.T, ch *cmdutil.Helper, projectName string) error {
	maxRetries := 5
	retryWait := 3 * time.Second

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(retryWait)
		}

		res := runWithInput(t.Context(), ch, "", "project", "status", projectName, "--format=csv")
		require.Equal(t, 0, res.ExitCode, "failed to get project status: %s", res.Output)
		allIdle, err := parseStatusCSV(res.Output)
		if err != nil {
			return fmt.Errorf("failed to parse status CSV: %w", err)
		}

		if allIdle {
			return nil
		}
		t.Logf("Not all resources are Idle yet (attempt %d/%d)", i+1, maxRetries)
	}

	return fmt.Errorf("timed out waiting for all resources to be Idle after %d retries", maxRetries)
}

// parseStatusCSV parses the CSV status output and checks if all resources are Idle
func parseStatusCSV(output string) (bool, error) {
	// Find the CSV header line
	lines := strings.Split(output, "\n")
	csvStartIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "Type,Name,Status,Error") {
			csvStartIdx = i
			break
		}
	}

	if csvStartIdx == -1 {
		return false, fmt.Errorf("CSV header 'Type,Name,Status,Error' not found in output")
	}

	// Parse CSV starting from the header
	csvData := strings.Join(lines[csvStartIdx:], "\n")
	reader := csv.NewReader(strings.NewReader(csvData))

	records, err := reader.ReadAll()
	if err != nil {
		return false, fmt.Errorf("failed to parse CSV: %w", err)
	}

	// Skip header row and check each record
	if len(records) <= 1 {
		// No data rows, consider it as all idle
		return true, nil
	}

	for i, record := range records[1:] {
		if len(record) < 3 {
			return false, fmt.Errorf("invalid CSV record at line %d: expected at least 3 columns, got %d", i+2, len(record))
		}

		status := strings.TrimSpace(record[2])
		if status != "Idle" {
			return false, nil
		}
	}

	return true, nil
}

// mustNewHelper creates a command helper or fails the test
func mustNewHelper(t *testing.T, homeDir string) *cmdutil.Helper {
	ch, err := cmdutil.NewHelper(version.Version{}, homeDir)
	ch.Interactive = false
	if err != nil {
		t.Fatalf("failed to create helper: %v", err)
	}
	return ch
}

// runWithInput executes the CLI with the given input and arguments and captures the output.
func runWithInput(ctx context.Context, ch *cmdutil.Helper, input string, args ...string) result {
	// Buffer for capturing output
	var out bytes.Buffer

	ch.Printer.OverrideDataOutput(&out)
	ch.Printer.OverrideHumanOutput(&out)

	// Configure the root command for testing
	root := cmd.RootCmd(ch)
	ch.Interactive = false
	root.SetOut(&out)
	root.SetErr(&out)
	if input != "" {
		root.SetIn(bytes.NewBufferString(input + "\n"))
	}
	root.SetArgs(args)

	// Execute the command (mirrors the logic in cli/cmd.Run)
	err := root.ExecuteContext(ctx)
	code := cmd.HandleExecuteError(ch, err)
	return result{
		ExitCode: code,
		Output:   out.String(),
	}
}

// Result represents the output of a CLI invocation.
type result struct {
	ExitCode int
	Output   string
}

// copyDir copies a directory from source to destination
// It recursively copies all the contents of the source directory to the destination directory.
// Files with the same name in the destination directory will be overwritten.
func copyDir(dst, src string) error {
	// Create the destination directory
	err := os.MkdirAll(dst, os.ModePerm)
	if err != nil {
		return err
	}
	// Read the contents of the source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy the contents of the source directory
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(dstPath, srcPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(dstPath, srcPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(dst, src string) error {
	// Open the source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy the content from source to destination
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	return nil
}

func addUser(t *testing.T, ch *cmdutil.Helper) {
	err := godotenv.Load(".env")
	require.NoError(t, err)

	val, ok := os.LookupEnv("RILL_ADMIN_DATABASE_URL")
	require.True(t, ok, "RILL_ADMIN_DATABASE_URL not set in .env")

	pgx, err := pgx.Connect(t.Context(), val)
	require.NoError(t, err)
	defer pgx.Close(t.Context())

	res, err := pgx.Query(t.Context(), `
		INSERT INTO "users" (email, display_name, photo_url, quota_trial_orgs, quota_singleuser_orgs, superuser)
		VALUES ('test@rilldata.com', 'Test User', '', 100, 100, true) RETURNING id
	`)
	require.NoError(t, err)
	require.True(t, res.Next())
	var userID string
	err = res.Scan(&userID)
	require.NoError(t, err)
	t.Logf("Created user with ID: %v", userID)
	res.Close()

	// insert auth token for the user
	token := authtoken.NewRandom(authtoken.TypeUser)

	_, err = pgx.Exec(t.Context(), `INSERT INTO user_auth_tokens (id, secret_hash, user_id, display_name, auth_client_id, representing_user_id, refresh, expires_on)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING (id, secret_hash)`, token.ID.String(), token.SecretHash(), userID, "Test Token", "12345678-0000-0000-0000-000000000002", nil, false, nil)
	require.NoError(t, err)

	t.Logf("Created auth token for user: %s", token.String())

	err = ch.DotRill.SetAccessToken(token.String())
	require.NoError(t, err)

	err = ch.ReloadAdminConfig()
	require.NoError(t, err)
}
