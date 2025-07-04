package testcli

import (
	"bytes"
	"testing"

	"github.com/rilldata/rill/admin/testadmin"
	"github.com/rilldata/rill/cli/cmd"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/rilldata/rill/cli/pkg/version"
	"github.com/stretchr/testify/require"
)

// Fixture represents a (possibly authenticated) CLI installation.
type Fixture struct {
	Admin   *testadmin.Fixture
	HomeDir string
}

// New creates a new Fixture for the given admin service test fixture and (optional) user access token.
func New(t *testing.T, adm *testadmin.Fixture, token string) *Fixture {
	homeDir := t.TempDir()
	dotRill := dotrill.New(homeDir)
	require.NoError(t, dotRill.SetDefaultAdminURL(adm.ExternalURL()))
	if token != "" {
		require.NoError(t, dotRill.SetAccessToken(token))
	}

	return &Fixture{
		Admin:   adm,
		HomeDir: homeDir,
	}
}

// NewWithUser is similar to New, but also creates a new user and authenticates the Fixture with a token for that user.
func NewWithUser(t *testing.T, adm *testadmin.Fixture) *Fixture {
	_, c := adm.NewUser(t)
	return New(t, adm, c.Token)
}

// Result represents the output of a CLI invocation.
type Result struct {
	ExitCode int
	Output   string
}

// Run executes the CLI with the given arguments and captures the output.
func (f *Fixture) Run(t *testing.T, args ...string) Result {
	return f.RunWithInput(t, "", args...)
}

// RunWithInput executes the CLI with the given input and arguments and captures the output.
func (f *Fixture) RunWithInput(t *testing.T, input string, args ...string) Result {
	// Buffer for capturing output
	var out bytes.Buffer

	// Create a new command helper configured for testing
	ch, err := cmdutil.NewHelper(version.Version{}, f.HomeDir)
	require.NoError(t, err)
	ch.Printer.OverrideDataOutput(&out)
	ch.Printer.OverrideHumanOutput(&out)

	// Configure the root command for testing
	root := cmd.RootCmd(ch)
	root.SetOut(&out)
	root.SetErr(&out)
	if input != "" {
		root.SetIn(bytes.NewBufferString(input + "\n"))
	}
	root.SetArgs(args)

	// Execute the command (mirrors the logic in cli/cmd.Run)
	err = root.ExecuteContext(t.Context())
	code := cmd.HandleExecuteError(ch, err)

	return Result{
		ExitCode: code,
		Output:   out.String(),
	}
}
