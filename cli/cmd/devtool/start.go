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
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/hashicorp/go-version"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const (
	minGoVersion   = "1.22"
	minNodeVersion = "18"
	stateDirCloud  = "dev-cloud-state"
	stateDirLocal  = "dev-project"
	rillGithubURL  = "https://github.com/rilldata/rill"
)

var (
	logErr  = color.New(color.FgHiRed)
	logWarn = color.New(color.FgHiYellow)
	logInfo = color.New(color.FgHiGreen)
)

func StartCmd(ch *cmdutil.Helper) *cobra.Command {
	var verbose, reset, refreshDotenv bool
	services := &servicesCfg{}

	cmd := &cobra.Command{
		Use:   "start [cloud|local]",
		Short: "Start a local development environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			var preset string
			if len(args) > 0 {
				preset = args[0]
			} else {
				preset = cmdutil.SelectPrompt("Select preset", []string{"cloud", "local"}, "cloud")
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
		checkRillRepo(ctx),
	)
	if err != nil {
		return err
	}

	switch preset {
	case "cloud":
		err = cloud{}.start(ctx, ch, verbose, reset, refreshDotenv, services)
	case "local":
		err = local{}.start(ctx, verbose, reset, services)
	default:
		err = fmt.Errorf("Unknown preset %q", preset)
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
	min := version.Must(version.NewVersion(minGoVersion))
	if v.LessThan(min) {
		return fmt.Errorf("Go version %s or higher is required", minGoVersion)
	}
	return nil
}

func checkNodeVersion(ctx context.Context) error {
	out, err := newCmd(ctx, "node", "--version").Output()
	if err != nil {
		return fmt.Errorf("error executing the 'node --version' command: %w", err)
	}

	v := version.Must(version.NewVersion(strings.TrimSpace(string(out))))
	min := version.Must(version.NewVersion(minNodeVersion))
	if v.LessThan(min) {
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

func checkRillRepo(ctx context.Context) error {
	_, err := os.Stat(".git")
	if err != nil {
		return fmt.Errorf("you must run `rill devtool` from the root of the rill repository")
	}

	_, githubURL, err := gitutil.ExtractGitRemote("", "")
	if err != nil {
		return fmt.Errorf("error extracting git remote: %w", err)
	}

	if githubURL != rillGithubURL {
		return fmt.Errorf("you must run `rill devtool` from the rill repository (expected remote %q, got %q)", rillGithubURL, githubURL)
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

func (s cloud) start(ctx context.Context, ch *cmdutil.Helper, verbose, reset, refreshDotenv bool, services *servicesCfg) error {
	if reset {
		err := s.resetState(ctx)
		if err != nil {
			return fmt.Errorf("reset cloud deps: %w", err)
		}
		logInfo.Printf("Reset cloud dependencies\n")
	}

	err := os.MkdirAll(stateDirCloud, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create state dir %q: %w", stateDirCloud, err)
	}

	if refreshDotenv {
		err := downloadDotenv(ctx, "cloud")
		if err != nil {
			return fmt.Errorf("failed to refresh .env: %w", err)
		}
		logInfo.Printf("Refreshed .env\n")
	}

	// Validate the .env file is well-formed.
	err = checkDotenv()
	if err != nil {
		return err
	}
	_, err = godotenv.Read()
	if err != nil {
		return fmt.Errorf("error parsing .env: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	if services.deps {
		g.Go(func() error { return s.runDeps(ctx, verbose) })
	}

	depsReadyCh := make(chan struct{})
	g.Go(func() error {
		if services.deps {
			err := s.awaitPostgres(ctx)
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
			return s.runAdmin(ctx, verbose)
		})
	}

	if services.runtime {
		g.Go(func() error {
			if err := awaitClose(ctx, depsReadyCh); err != nil {
				return err
			}
			return s.runRuntime(ctx, verbose)
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

	_ = os.RemoveAll(stateDirCloud)
	return newCmd(ctx, "docker", "compose", "-f", "cli/cmd/devtool/data/cloud-deps.docker-compose.yml", "down", "--volumes").Run()
}

func (s cloud) runDeps(ctx context.Context, verbose bool) error {
	logInfo.Printf("Starting dependencies\n")
	defer logInfo.Printf("Stopped dependencies\n")

	cmd := newCmd(ctx, "docker", "compose", "-f", "cli/cmd/devtool/data/cloud-deps.docker-compose.yml", "up", "--no-recreate")
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
	}
	return cmd.Run()
}

func (s cloud) awaitPostgres(ctx context.Context) error {
	logInfo.Printf("Waiting for Postgres\n")

	dbURL := lookupDotenv("RILL_ADMIN_DATABASE_URL")
	for {
		conn, err := pgx.Connect(ctx, dbURL)
		if err == nil {
			conn.Close(ctx)
			logInfo.Printf("Postgres ready\n")
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
			logInfo.Printf("Redis ready\n")
			return nil
		}

		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s cloud) runAdmin(ctx context.Context, verbose bool) (err error) {
	logInfo.Printf("Starting admin\n")
	defer logInfo.Printf("Stopped admin\n")

	cmd := newCmd(ctx, "go", "run", "cli/main.go", "admin", "start")
	if verbose {
		cmd.Env = append(os.Environ(), "RILL_ADMIN_LOG_LEVEL=debug")
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
				logInfo.Printf("Admin ready\n")
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

func (s cloud) runRuntime(ctx context.Context, verbose bool) (err error) {
	logInfo.Printf("Starting runtime\n")
	defer logInfo.Printf("Stopped runtime\n")

	cmd := newCmd(ctx, "go", "run", "cli/main.go", "runtime", "start")
	if verbose {
		cmd.Env = append(os.Environ(), "RILL_RUNTIME_LOG_LEVEL=debug")
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
				logInfo.Printf("Runtime ready\n")
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
			logErr.Printf("Failed running `npm install -w web-admin`: %v", err)
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

	args := []string{"run", "cli/main.go", "start", stateDirLocal, "--no-ui", "--debug"}
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

	cmd := newCmd(ctx, "npm", "run", "dev", "-w", "web-local")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	return cmd.Run()
}

func (s local) awaitUI(ctx context.Context) error {
	uiURL := "http://localhost:3000"
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
