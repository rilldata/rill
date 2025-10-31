package testmode

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/joho/godotenv"
)

// modeEnviromentVariable is the environment variable that controls the test mode.
const modeEnvironmentVariable = "RILL_RUNTIME_TEST_MODE"

// validModes is the set of valid test modes.
var validModes = map[string]bool{
	"":          true,
	"expensive": true,
}

// Expensive marks the test as an expensive operation.
// Expensive tests only run when the environment variable RILL_RUNTIME_TEST_MODE=expensive is set.
// They will error if the environment variable is not set, unless the -short flag is used, in which case they are skipped.
func Expensive(t TestingT) {
	// Skip expensive tests in short mode.
	if testing.Short() {
		t.SkipNow()
	}

	// Fail if running an expensive test without the env var set.
	if Mode(t) != "expensive" {
		t.Errorf("expensive tests are disabled; set %s=expensive to enable", modeEnvironmentVariable)
	}
}

// Mode returns the current test mode set in the RILL_RUNTIME_TEST_MODE environment variable.
// Currently valid values are "" (default) and "expensive".
func Mode(t TestingT) string {
	// Load variable
	loadDotEnv(t)
	val := os.Getenv(modeEnvironmentVariable)

	// Validate value
	if !validModes[val] {
		t.Errorf("invalid value for environment variable %s: %s", modeEnvironmentVariable, val)
	}

	return val
}

// TestingT is an interface that matches *testing.T and *testing.B.
type TestingT interface {
	SkipNow()
	Errorf(format string, args ...any)
}

// loadDotEnv loads the .env file at the repo root (if any).
func loadDotEnv(t TestingT) {
	_, currentFile, _, _ := runtime.Caller(0)
	envPath := filepath.Join(currentFile, "..", "..", "..", "..", ".env")
	_, err := os.Stat(envPath)
	if err == nil {
		err = godotenv.Load(envPath)
		if err != nil {
			t.Errorf("error loading .env file: %w", err)
		}
	}
}
