package devtool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/hashicorp/go-version"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const (
	composeFile    = "cli/cmd/devtool/data/cloud-deps.docker-compose.yml"
	minGoVersion   = "1.24"
	minNodeVersion = "18"
	stateDirLocal  = "dev-project"
	rillGitRemote  = "https://github.com/rilldata/rill.git"
)

var (
	// Console log colors
	logErr  = color.New(color.FgHiRed)
	logWarn = color.New(color.FgHiYellow)
	logInfo = color.New(color.FgHiGreen)

	// Preset options
	presets = []string{
		// Full cloud setup
		"cloud",
		// Minimal cloud setup (no Clickhouse, no telemetry)
		"minimal",
		// Rill Developer setup (equivalent to `rill start`)
		"local",
		// Cloud setup for e2e tests
		"e2e",
		// TODO: What is this?
		"other",
	}
)

func StartCmd(ch *cmdutil.Helper) *cobra.Command {
	var verbose, reset, refreshDotenv bool
	services := &servicesCfg{}

	cmd := &cobra.Command{
		Use:   "start [cloud|minimal|local|e2e]",
		Short: "Start a local development environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			var preset string
			if len(args) > 0 {
				preset = args[0]
			} else {
				res, err := cmdutil.SelectPrompt("Select preset", presets, "cloud")
				if err != nil {
					return err
				}
				preset = res
			}

			err := services.parse()
			if err != nil {
				return fmt.Errorf("failed to parse services: %w", err)
			}

			return start(ch, preset, verbose, reset, refreshDotenv, services)
		},
	}

	cmd.Flags().BoolVar(&verbose, "verbose", false, "Set log level to debug")
	cmd.Flags().BoolVar(&reset, "reset", false, "Reset local development state")
	cmd.Flags().BoolVar(&refreshDotenv, "refresh-dotenv", true, "Refresh .env file from shared storage")
	services.addFlags(cmd)

	return cmd
}

func start(ch *cmdutil.Helper, preset string, verbose, reset, refreshDotenv bool, services *servicesCfg) error {
	ctx := graceful.WithCancelOnTerminate(context.Background())

	err := errors.Join(
		checkGoVersion(),
		checkNodeVersion(ctx),
		checkDocker(ctx),
		checkRillRepo(),
	)
	if err != nil {
		return err
	}

	switch preset {
	case "cloud", "minimal", "e2e", "other":
		err = cloud{}.start(ctx, ch, verbose, reset, refreshDotenv, preset, services)
	case "local":
		err = local{}.start(ctx, verbose, reset, services)
	default:
		err = fmt.Errorf("unknown preset %q", preset)
	}
	// If ctx.Err() != nil, we don't return the err because any graceful shutdown will cause sub-commands to return non-zero exit code errors.
	// In these cases, ignoring the error doesn't really matter since "real" errors are probably also logged to stdout anyway.
	if err != nil && (ctx.Err() == nil || verbose) {
		return err
	}
	return nil
}

func checkGoVersion() error {
	v := version.Must(version.NewVersion(strings.TrimPrefix(runtime.Version(), "go")))
	minVersion := version.Must(version.NewVersion(minGoVersion))
	if v.LessThan(minVersion) {
		return fmt.Errorf("go version %s or higher is required", minGoVersion)
	}
	return nil
}

func checkNodeVersion(ctx context.Context) error {
	out, err := newCmd(ctx, "node", "--version").Output()
	if err != nil {
		return fmt.Errorf("error executing the 'node --version' command: %w", err)
	}

	v := version.Must(version.NewVersion(strings.TrimSpace(string(out))))
	minVersion := version.Must(version.NewVersion(minNodeVersion))
	if v.LessThan(minVersion) {
		return fmt.Errorf("node.js version %s or higher is required", minNodeVersion)
	}

	return nil
}

func checkDocker(ctx context.Context) error {
	out, err := newCmd(ctx, "docker", "info", "--format", "json").Output()
	if err != nil {
		return fmt.Errorf("error executing the 'docker info' command: %w", err)
	}

	info := make(map[string]any)
	err = json.Unmarshal(out, &info)
	if err != nil {
		return fmt.Errorf("error parsing the output of 'docker info': %w", err)
	}

	if sv, ok := info["ServerVersion"].(string); !ok || sv == "" {
		return errors.New("error extracting the Docker server version (is Docker running?)")
	}

	if se, ok := info["ServerErrors"].([]string); ok && len(se) > 0 {
		return fmt.Errorf("docker not available: found errors: %v", se)
	}

	return nil
}

func checkRillRepo() error {
	_, err := os.Stat(".git")
	if err != nil {
		return fmt.Errorf("you must run `rill devtool` from the root of the rill repository")
	}

	remote, err := gitutil.ExtractGitRemote("", "", false)
	if err != nil {
		return fmt.Errorf("error extracting git remote: %w", err)
	}
	githubRemote, _ := remote.Github()

	if githubRemote != rillGitRemote {
		return fmt.Errorf("you must run `rill devtool` from the rill repository (expected remote %q, got %q)", rillGitRemote, githubRemote)
	}

	return nil
}

type servicesCfg struct {
	admin   bool
	deps    bool
	runtime bool
	ui      bool
	only    []string
	except  []string
}

func (s *servicesCfg) addFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&s.only, "only", []string{}, "Only start the listed services (options: admin, deps, runtime, ui)")
	cmd.Flags().StringSliceVar(&s.except, "except", []string{}, "Start all except the listed services (options: admin, deps, runtime, ui)")
}

func (s *servicesCfg) parse() error {
	if len(s.only) > 0 && len(s.except) > 0 {
		return errors.New("cannot use both --only and --except")
	}

	vals := s.except
	def := true
	if len(s.only) > 0 {
		vals = s.only
		def = false
	}

	s.admin = def
	s.deps = def
	s.runtime = def
	s.ui = def

	for _, v := range vals {
		switch v {
		case "admin":
			s.admin = !def
		case "deps":
			s.deps = !def
		case "runtime":
			s.runtime = !def
		case "ui":
			s.ui = !def
		default:
			return fmt.Errorf("invalid service %q", v)
		}
	}

	return nil
}

type cloud struct{}

func (s cloud) start(ctx context.Context, ch *cmdutil.Helper, verbose, reset, refreshDotenv bool, preset string, services *servicesCfg) error {
	if refreshDotenv {
		err := downloadDotenv(ctx, preset)
		if err != nil {
			return fmt.Errorf("failed to refresh .env: %w", err)
		}
		logInfo.Printf("Refreshed .env\n")
	}

	if preset == "other" {
		preset = "e2e"
	}

	// Validate the .env file is well-formed.
	err := checkDotenv()
	if err != nil {
		return err
	}
	_, err = godotenv.Read()
	if err != nil {
		return fmt.Errorf("error parsing .env: %w", err)
	}

	if reset {
		err := s.resetState(ctx)
		if err != nil {
			return fmt.Errorf("reset cloud deps: %w", err)
		}
		logInfo.Printf("Reset cloud dependencies\n")
	}

	g, ctx := errgroup.WithContext(ctx)

	err = os.MkdirAll(stateDirectory(), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create state dir %q: %w", stateDirectory(), err)
	}
	logInfo.Printf("State directory is %q\n", stateDirectory())

	if services.deps {
		g.Go(func() error { return s.runDeps(ctx, verbose, preset) })
	}

	depsReadyCh := make(chan struct{})
	g.Go(func() error {
		if services.deps {
			err := s.awaitPostgres(ctx, preset)
			if err != nil {
				return err
			}
			err = s.awaitRedis(ctx)
			if err != nil {
				return err
			}
		}
		close(depsReadyCh)
		return nil
	})

	if services.admin {
		g.Go(func() error {
			if err := awaitClose(ctx, depsReadyCh); err != nil {
				return err
			}
			return s.runAdmin(ctx, verbose, preset)
		})
	}

	if services.runtime {
		g.Go(func() error {
			if err := awaitClose(ctx, depsReadyCh); err != nil {
				return err
			}
			return s.runRuntime(ctx, verbose, preset)
		})
	}

	backendReadyCh := make(chan struct{})
	g.Go(func() error {
		if err := awaitClose(ctx, depsReadyCh); err != nil {
			return err
		}
		if services.admin {
			err := s.awaitAdmin(ctx)
			if err != nil {
				return err
			}
		}
		if services.runtime {
			err := s.awaitRuntime(ctx)
			if err != nil {
				return err
			}
		}
		close(backendReadyCh)
		return nil
	})

	g.Go(func() error {
		if !services.admin {
			return nil
		}

		if err := awaitClose(ctx, backendReadyCh); err != nil {
			return err
		}

		// NOTE: Will revert back to previous env on ctx.Done()
		switchEnvToDevTemporarily(ctx, ch)

		return nil
	})

	if services.ui {
		npmReadyCh := make(chan struct{})
		g.Go(func() error {
			err := s.runUIInstall(ctx)
			if err != nil {
				return err
			}
			close(npmReadyCh)
			return nil
		})

		g.Go(func() error {
			if err := awaitClose(ctx, backendReadyCh, npmReadyCh); err != nil {
				return err
			}
			return s.runUI(ctx)
		})
	}

	uiReadyCh := make(chan struct{})
	g.Go(func() error {
		if services.ui {
			err := s.awaitUI(ctx)
			if err != nil {
				return err
			}
		}
		close(uiReadyCh)
		return nil
	})

	g.Go(func() error {
		if err := awaitClose(ctx, backendReadyCh, uiReadyCh); err != nil {
			return err
		}
		logInfo.Printf("All services ready\n")
		return nil
	})

	return g.Wait()
}

func (s cloud) resetState(ctx context.Context) (err error) {
	logInfo.Printf("Resetting state\n")
	defer func() {
		if err == nil {
			logInfo.Printf("Reset state\n")
		} else {
			logErr.Printf("Failed to reset state: %v", err)
		}
	}()

	_ = os.RemoveAll(stateDirectory())

	// tear down all containers regardless of profile
	return newCmd(ctx, "docker", "compose", "--env-file", ".env", "-f", composeFile, "down", "--volumes").Run()
}

func (s cloud) runDeps(ctx context.Context, verbose bool, preset string) error {
	composeFile := "cli/cmd/devtool/data/cloud-deps.docker-compose.yml"
	profile := "full"
	if preset == "minimal" {
		profile = "minimal"
	} else if preset == "e2e" {
		profile = "e2e"
	}

	args := []string{"docker", "compose", "--env-file", ".env", "-f", composeFile, "--profile", profile, "up"}

	logInfo.Printf("Starting dependencies: %s\n", strings.Join(args, " "))
	defer logInfo.Printf("Stopped dependencies\n")

	err := prepareStripeConfig()
	if err != nil {
		return fmt.Errorf("failed to prepare stripe config: %w", err)
	}

	cmd := newCmd(ctx, args[0], args[1:]...)
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
	}
	return cmd.Run()
}

func (s cloud) awaitPostgres(ctx context.Context, preset string) error {
	logInfo.Printf("Waiting for Postgres (%s)\n", preset)

	dbURL := lookupDotenv("RILL_ADMIN_DATABASE_URL")
	for {
		conn, err := pgx.Connect(ctx, dbURL)
		if err == nil {
			conn.Close(ctx)
			logInfo.Printf("Postgres ready at %s\n", dbURL)
			return nil
		}

		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s cloud) awaitRedis(ctx context.Context) error {
	dbURL := lookupDotenv("RILL_ADMIN_REDIS_URL")
	if dbURL == "" {
		return nil
	}

	logInfo.Printf("Waiting for Redis\n")

	opts, err := redis.ParseURL(dbURL)
	if err != nil {
		return fmt.Errorf("failed to parse redis url: %w", err)
	}
	for {
		c := redis.NewClient(opts)
		res, err := c.Echo(ctx, "hello").Result()
		c.Close()
		if err == nil && res == "hello" {
			logInfo.Printf("Redis ready at %s\n", dbURL)
			return nil
		}

		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s cloud) runAdmin(ctx context.Context, verbose bool, preset string) (err error) {
	logInfo.Printf("Starting admin\n")
	defer logInfo.Printf("Stopped admin\n")

	cmd := newCmd(ctx, "go", "run", "cli/main.go", "admin", "start")
	cmd.Env = os.Environ()
	if preset == "minimal" {
		cmd.Env = append(
			cmd.Env,
			// This differs from the usual dev provisioner set in not having a Clickhouse provisioner.
			`RILL_ADMIN_PROVISIONER_SET_JSON={"static":{"type":"static","spec":{"runtimes":[{"host":"http://localhost:8081","slots":50,"data_dir":"dev-cloud-state","audience_url":"http://localhost:8081"}]}}}`,
			// Disable traces
			"RILL_ADMIN_TRACES_EXPORTER="+string(observability.NoopExporter),
			// Change metrics to Prometheus, which unlike Otel doesn't require an external collector.
			"RILL_ADMIN_METRICS_EXPORTER="+string(observability.PrometheusExporter),
		)
	}
	if verbose {
		cmd.Env = append(cmd.Env, "RILL_ADMIN_LOG_LEVEL=debug")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	return cmd.Run()
}

func (s cloud) awaitAdmin(ctx context.Context) error {
	pingURL := lookupDotenv("RILL_ADMIN_EXTERNAL_URL")
	pingURL, err := url.JoinPath(pingURL, "/v1/ping")
	if err != nil {
		return fmt.Errorf("failed to parse admin url: %w", err)
	}

	for {
		resp, err := http.Get(pingURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				logInfo.Printf("Admin ready at %s\n", pingURL)
				return nil
			}
		}

		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s cloud) runRuntime(ctx context.Context, verbose bool, preset string) (err error) {
	logInfo.Printf("Starting runtime\n")
	defer logInfo.Printf("Stopped runtime\n")

	cmd := newCmd(ctx, "go", "run", "cli/main.go", "runtime", "start")
	cmd.Env = os.Environ()
	if preset == "minimal" {
		cmd.Env = append(
			cmd.Env,
			// Disable traces
			"RILL_RUNTIME_TRACES_EXPORTER="+string(observability.NoopExporter),
			// Change metrics to Prometheus, which unlike Otel doesn't require an external collector.
			"RILL_RUNTIME_METRICS_EXPORTER="+string(observability.PrometheusExporter),
		)
	}
	if verbose {
		cmd.Env = append(cmd.Env, "RILL_RUNTIME_LOG_LEVEL=debug")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	return cmd.Run()
}

func (s cloud) awaitRuntime(ctx context.Context) error {
	pingURL := lookupDotenv("RILL_RUNTIME_AUTH_AUDIENCE_URL") // TODO: This is a proxy for the runtime's external URL. Should be less implicit.
	pingURL, err := url.JoinPath(pingURL, "/v1/ping")
	if err != nil {
		return fmt.Errorf("failed to parse admin url: %w", err)
	}

	for {
		resp, err := http.Get(pingURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				logInfo.Printf("Runtime ready at %s\n", pingURL)
				return nil
			}
		}

		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s cloud) runUIInstall(ctx context.Context) (err error) {
	logInfo.Printf("Running `npm install -w web-admin`\n")
	defer func() {
		if err == nil {
			logInfo.Printf("Finished `npm install -w web-admin`\n")
		} else {
			logErr.Printf("Failed running `npm install -w web-admin`: %v\n", err)
		}
	}()

	return newCmd(ctx, "npm", "install", "-w", "web-admin").Run()
}

func (s cloud) runUI(ctx context.Context) (err error) {
	logInfo.Printf("Starting UI\n")
	defer logInfo.Printf("Stopped UI\n")

	cmd := newCmd(ctx, "npm", "run", "dev", "-w", "web-admin")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	return cmd.Run()
}

func (s cloud) awaitUI(ctx context.Context) error {
	uiURL := lookupDotenv("RILL_ADMIN_FRONTEND_URL") // TODO: This is a proxy for the frontend's external URL. Should be less implicit.

	for {
		resp, err := http.Get(uiURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				logInfo.Printf("UI ready at %s\n", uiURL)
				return nil
			}
		}

		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type local struct{}

func (s local) start(ctx context.Context, verbose, reset bool, services *servicesCfg) error {
	g, ctx := errgroup.WithContext(ctx)

	if services.runtime {
		g.Go(func() error { return s.runRuntime(ctx, verbose, reset) })
	}

	runtimeReadyCh := make(chan struct{})
	g.Go(func() error {
		if services.runtime {
			err := s.awaitRuntime(ctx)
			if err != nil {
				return err
			}
		}
		close(runtimeReadyCh)
		return nil
	})

	if services.ui {
		npmReadyCh := make(chan struct{})
		g.Go(func() error {
			err := s.runUIInstall(ctx)
			if err != nil {
				return err
			}
			close(npmReadyCh)
			return nil
		})

		g.Go(func() error {
			if err := awaitClose(ctx, runtimeReadyCh, npmReadyCh); err != nil {
				return err
			}
			return s.runUI(ctx)
		})
	}

	uiReadyCh := make(chan struct{})
	g.Go(func() error {
		if services.ui {
			err := s.awaitUI(ctx)
			if err != nil {
				return err
			}
		}
		close(uiReadyCh)
		return nil
	})

	g.Go(func() error {
		// Wait for runtime, then UI
		if err := awaitClose(ctx, runtimeReadyCh, uiReadyCh); err != nil {
			return err
		}
		logInfo.Printf("All services ready\n")
		return nil
	})

	return g.Wait()
}

func (s local) runRuntime(ctx context.Context, verbose, reset bool) error {
	logInfo.Printf("Starting runtime\n")
	defer func() { logInfo.Printf("Stopped runtime\n") }()

	args := []string{"run", "cli/main.go", "start", stateDirLocal, "--no-ui", "--debug", "--allowed-origins", "http://localhost:3001"}
	if verbose {
		args = append(args, "--verbose")
	}
	if reset {
		args = append(args, "--reset")
	}

	cmd := newCmd(ctx, "go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	return cmd.Run()
}

func (s local) awaitRuntime(ctx context.Context) error {
	pingURL := "http://localhost:9009/v1/ping"
	for {
		resp, err := http.Get(pingURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				logInfo.Printf("Backend ready\n")
				return nil
			}
		}

		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s local) runUIInstall(ctx context.Context) (err error) {
	logInfo.Printf("Running `npm install -w web-local`\n")
	defer func() {
		if err == nil {
			logInfo.Printf("Finished `npm install -w web-local`\n")
		} else {
			logErr.Printf("Failed running `npm install -w web-local`: %v", err)
		}
	}()

	return newCmd(ctx, "npm", "install", "-w", "web-local").Run()
}

func (s local) runUI(ctx context.Context) (err error) {
	logInfo.Printf("Starting UI\n")
	defer logInfo.Printf("Stopped UI\n")

	cmd := newCmd(ctx, "npm", "run", "dev", "-w", "web-local", "--", "--port", "3001")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	return cmd.Run()
}

func (s local) awaitUI(ctx context.Context) error {
	uiURL := "http://localhost:3001"
	for {
		resp, err := http.Get(uiURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				logInfo.Printf("UI ready\n")
				return nil
			}
		}

		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func prepareStripeConfig() error {
	templateFile := "cli/cmd/devtool/data/stripe-config.template"
	outputFile := filepath.Join(stateDirectory(), "stripe-config.toml")

	apiKey := lookupDotenv("RILL_DEVTOOL_STRIPE_CLI_API_KEY")
	if apiKey == "" {
		logWarn.Printf("No Stripe API key found in .env, Stripe webhook events will not be processed\n")
	}

	// Parse the template
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create the output file
	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	// Execute the template, writing to the output file
	err = tmpl.Execute(out, map[string]string{"APIKey": apiKey})
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// awaitClose waits for all of the given channels to close.
// It returns an error if the context is canceled.
func awaitClose(ctx context.Context, chs ...<-chan struct{}) error {
	for _, ch := range chs {
		select {
		case <-ch:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// newCmd initializes an exec.Cmd that sends SIGINT instead of SIGKILL when the ctx is canceled.
func newCmd(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Cancel = func() error {
		return cmd.Process.Signal(os.Interrupt)
	}
	return cmd
}

// lookupDotenv returns a value from the .env file.
// NOTE: Not using godotenv.Load() since the OpenTelemetry-related env vars clash with `docker compose`.
func lookupDotenv(key string) string {
	env, err := godotenv.Read()
	if err != nil {
		return ""
	}
	return env[key]
}

// stateDirectory returns the directory where the devtool's state is stored.
// Deleting this directory will reset the state of the local development environment.
func stateDirectory() string {
	dir := lookupDotenv("RILL_DEVTOOL_STATE_DIRECTORY")
	if dir == "" {
		dir = "dev-cloud-state"
	}
	return dir
}
