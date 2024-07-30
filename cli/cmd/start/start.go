package start

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/rilldata/rill/cli/pkg/gitutil"
	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/rilldata/rill/runtime/compilers/rillv1beta"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// maxProjectFiles is the maximum number of files that can be in a project directory.
// It corresponds to the file watcher limit in runtime/drivers/file/repo.go.
const maxProjectFiles = 1000

// StartCmd represents the start command
func StartCmd(ch *cmdutil.Helper) *cobra.Command {
	var olapDriver string
	var olapDSN string
	var httpPort int
	var grpcPort int
	var verbose bool
	var debug bool
	var readonly bool
	var reset bool
	var noUI bool
	var noOpen bool
	var logFormat string
	var env []string
	var vars []string
	var tlsCertPath string
	var tlsKeyPath string

	startCmd := &cobra.Command{
		Use:   "start [<path>]",
		Short: "Build project and start web app",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var projectPath string
			if len(args) > 0 {
				projectPath = args[0]
				if strings.HasSuffix(projectPath, ".git") {
					repoName, err := gitutil.CloneRepo(projectPath)
					if err != nil {
						return fmt.Errorf("clone repo error: %w", err)
					}

					projectPath = repoName
				}
			} else if !rillv1beta.HasRillProject("") {
				if !ch.Interactive {
					return fmt.Errorf("required arg <path> missing")
				}

				currentDir, err := filepath.Abs("")
				if err != nil {
					return err
				}

				projectPath = currentDir
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return err
				}

				displayPath := currentDir
				defval := true
				if strings.HasPrefix(currentDir, homeDir) {
					displayPath = strings.Replace(currentDir, homeDir, "~", 1)
					if currentDir == homeDir {
						defval = false
						displayPath = "~/"
					}
				}

				msg := fmt.Sprintf("Rill will create project files in %q. Do you want to continue?", displayPath)
				confirm, err := cmdutil.ConfirmPrompt(msg, "", defval)
				if err != nil {
					return err
				}
				if !confirm {
					ch.PrintfWarn("Aborted\n")
					return nil
				}
			}

			// Check that projectPath doesn't have an excessive number of files
			n, err := countFilesInDirectory(projectPath)
			if err != nil {
				return err
			}
			if n > maxProjectFiles {
				ch.PrintfError("The project directory exceeds the limit of %d files (found %d files). Please open Rill against a directory with fewer files.\n", maxProjectFiles, n)
				return nil
			}

			parsedLogFormat, ok := local.ParseLogFormat(logFormat)
			if !ok {
				return fmt.Errorf("invalid log format %q", logFormat)
			}

			// Backwards compatibility for --env (see comment on the flag definition for details)
			environment := "dev"
			for _, v := range env {
				if strings.Contains(v, "=") {
					vars = append(vars, v)
				} else {
					environment = v
				}
			}

			// Parser variables from "a=b" format to map
			varsMap, err := parseVariables(vars)
			if err != nil {
				return err
			}

			// If keypath or certpath provided, but not the other, display error
			// If keypath and certpath provided, check if the file exists
			if (tlsCertPath != "" && tlsKeyPath == "") || (tlsCertPath == "" && tlsKeyPath != "") {
				return fmt.Errorf("both --tls-cert and --tls-key must be provided")
			} else if tlsCertPath != "" && tlsKeyPath != "" {
				// Check to ensure the paths are valid
				if _, err := os.Stat(tlsCertPath); os.IsNotExist(err) {
					return fmt.Errorf("certificate not found: %s", tlsCertPath)
				}
				if _, err := os.Stat(tlsKeyPath); os.IsNotExist(err) {
					return fmt.Errorf("key not found: %s", tlsKeyPath)
				}
			}

			scheme := "http"
			if tlsCertPath != "" && tlsKeyPath != "" {
				scheme = "https"
			}
			localURL := fmt.Sprintf("%s://localhost:%d", scheme, httpPort)

			// if olapDriver is clickhouse and no olapDSN is specified, install and use clickhouse
			if olapDriver == "clickhouse" && olapDSN == local.DefaultOLAPDSN {
				chBinPath, err := installClickHouse()
				if err != nil {
					return err
				}

				chConfig, err := os.CreateTemp("", "clickhouse-config*.xml")
				if err != nil {
					return err
				}

				config := clickHouseConfigContent(projectPath)

				if _, err := chConfig.Write(config); err != nil {
					return err
				}

				go func() {
					// Ensure the config file is closed and deleted at the end
					defer os.Remove(chConfig.Name())
					defer chConfig.Close()

					name := chConfig.Name()
					cmd := newCmd(cmd.Context(), chBinPath, "server", fmt.Sprintf("--config-file=%s", name))
					err = cmd.Run()
					if err != nil {
						fmt.Println("Error running clickhouse server", err)
					}
				}()

				olapDSN = "clickhouse://localhost:9000"
			}

			app, err := local.NewApp(cmd.Context(), &local.AppOptions{
				Version:     ch.Version,
				Verbose:     verbose,
				Debug:       debug,
				Reset:       reset,
				Environment: environment,
				OlapDriver:  olapDriver,
				OlapDSN:     olapDSN,
				ProjectPath: projectPath,
				LogFormat:   parsedLogFormat,
				Variables:   varsMap,
				Activity:    ch.Telemetry(cmd.Context()),
				AdminURL:    ch.AdminURL,
				AdminToken:  ch.AdminToken(),
				CMDHelper:   ch,
				LocalURL:    localURL,
			})
			if err != nil {
				return err
			}
			defer app.Close()

			userID, _ := ch.CurrentUserID(cmd.Context())

			err = app.Serve(httpPort, grpcPort, !noUI, !noOpen, readonly, userID, tlsCertPath, tlsKeyPath)
			if err != nil {
				return fmt.Errorf("serve: %w", err)
			}

			return nil
		},
	}

	startCmd.Flags().SortFlags = false
	startCmd.Flags().BoolVar(&noOpen, "no-open", false, "Do not open browser")
	startCmd.Flags().IntVar(&httpPort, "port", 9009, "Port for HTTP")
	startCmd.Flags().IntVar(&grpcPort, "port-grpc", 49009, "Port for gRPC (internal)")
	startCmd.Flags().BoolVar(&readonly, "readonly", false, "Show only dashboards in UI")
	startCmd.Flags().BoolVar(&noUI, "no-ui", false, "Serve only the backend")
	startCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")
	startCmd.Flags().BoolVar(&debug, "debug", false, "Collect additional debug info")
	startCmd.Flags().BoolVar(&reset, "reset", false, "Clear and re-ingest source data")
	startCmd.Flags().StringVar(&logFormat, "log-format", "console", "Log format (options: \"console\", \"json\")")
	startCmd.Flags().StringVar(&tlsCertPath, "tls-cert", "", "Path to TLS certificate")
	startCmd.Flags().StringVar(&tlsKeyPath, "tls-key", "", "Path to TLS key file")

	// --env was previously used for variables, but is now used to set the environment name. We maintain backwards compatibility by keeping --env as a slice var, and setting any value containing an equals sign as a variable.
	startCmd.Flags().StringSliceVarP(&env, "env", "e", []string{}, `Environment name (default "dev")`)
	startCmd.Flags().StringSliceVarP(&vars, "var", "v", []string{}, "Set project variables")

	// We have deprecated the ability configure the OLAP database via the CLI. This should now be done via rill.yaml.
	// Keeping these for backwards compatibility for a while.
	startCmd.Flags().StringVar(&olapDSN, "db", local.DefaultOLAPDSN, "Database DSN")
	startCmd.Flags().StringVar(&olapDriver, "db-driver", local.DefaultOLAPDriver, "Database driver")
	if err := startCmd.Flags().MarkHidden("db"); err != nil {
		panic(err)
	}
	if err := startCmd.Flags().MarkHidden("db-driver"); err != nil {
		panic(err)
	}

	return startCmd
}

// a smaller subset of relevant parts of rill.yaml
type rillYAML struct {
	IgnorePaths []string `yaml:"ignore_paths"`
}

func countFilesInDirectory(projectPath string) (int, error) {
	var fileCount int

	if projectPath == "" {
		projectPath = "."
	}

	var ignorePaths []string
	// Read rill.yaml and get `ignore_paths`
	rawYaml, err := os.ReadFile(filepath.Join(projectPath, "/rill.yaml"))
	if err == nil {
		yml := &rillYAML{}
		err = yaml.Unmarshal(rawYaml, yml)
		if err == nil {
			ignorePaths = yml.IgnorePaths
		}
	}

	err = filepath.WalkDir(projectPath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		path = strings.TrimPrefix(path, projectPath)

		if drivers.IsIgnored(path, ignorePaths) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.IsDir() {
			fileCount++
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	return fileCount, nil
}

func parseVariables(vals []string) (map[string]string, error) {
	res := make(map[string]string)
	for _, v := range vals {
		v, err := godotenv.Unmarshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to parse variable %q: %w", v, err)
		}
		for k, v := range v {
			res[k] = v
		}
	}
	return res, nil
}

func installClickHouse() (string, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	dir := ""

	switch goos {
	case "linux":
		switch goarch {
		case "amd64":
			cpuInfo, err := getCPUFeatures()
			if err != nil {
				return "", fmt.Errorf("error reading CPU info: %w", err)
			}
			if strings.Contains(cpuInfo, "sse4_2") {
				dir = "amd64"
			} else {
				dir = "amd64compat"
			}
		case "arm64":
			cpuInfo, err := getCPUFeatures()
			if err != nil {
				return "", fmt.Errorf("error reading CPU info: %w", err)
			}
			if strings.Contains(cpuInfo, "asimd") && strings.Contains(cpuInfo, "sha1") &&
				strings.Contains(cpuInfo, "aes") && strings.Contains(cpuInfo, "atomics") &&
				strings.Contains(cpuInfo, "lrcpc") {
				dir = "aarch64"
			} else {
				dir = "aarch64v80compat"
			}
		}
	case "darwin":
		switch goarch {
		case "amd64":
			dir = "macos"
		case "arm64":
			dir = "macos-aarch64"
		}
	}

	if dir == "" {
		return "", fmt.Errorf("operating system '%s' / architecture '%s' is unsupported.\n", goos, goarch)
	}

	clickhouseDownloadFilenamePrefix := "clickhouse"
	// Store the ClickHouse binary under .rill/clickhouse so that every project can use the same binary
	clickhouseDestDir, err := dotrill.ResolveFilename("clickhouse", true)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(clickhouseDestDir); os.IsNotExist(err) {
		err = os.MkdirAll(clickhouseDestDir, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("error creating ClickHouse directory: %w", err)
		}
	}
	clickhouseDestPath := filepath.Join(clickhouseDestDir, clickhouseDownloadFilenamePrefix)

	if _, err := os.Stat(clickhouseDestPath); err == nil {
		fmt.Printf("ClickHouse binary %s already exists\n", clickhouseDestPath)
		return clickhouseDestPath, nil
	}

	URL := fmt.Sprintf("https://builds.clickhouse.com/master/%s/clickhouse", dir)
	fmt.Printf("Will download %s into %s\n", URL, clickhouseDestPath)
	if err := downloadFile(clickhouseDestPath, URL); err != nil {
		return "", fmt.Errorf("error downloading ClickHouse: %w", err)
	}

	err = os.Chmod(clickhouseDestPath, 0o755)
	if err != nil {
		return "", fmt.Errorf("error setting executable permission: %w", err)
	}

	fmt.Println("Successfully downloaded the ClickHouse binary")

	return clickhouseDestPath, nil
}

func getCPUFeatures() (string, error) {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return "", err
	}
	defer file.Close()

	var result strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		result.WriteString(scanner.Text() + "\n")
	}
	return result.String(), scanner.Err()
}

func downloadFile(path, url string) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	return err
}

// newCmd initializes an exec.Cmd that sends SIGINT instead of SIGKILL when the ctx is canceled.
func newCmd(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Cancel = func() error {
		return cmd.Process.Signal(os.Interrupt)
	}
	return cmd
}

func clickHouseConfigContent(projectPath string) []byte {
	// https://github.com/ClickHouse/ClickHouse/blob/master/programs/server/embedded.xml
	config := []byte(fmt.Sprintf(`<clickhouse>
    <logger>
        <level>trace</level>
        <console>true</console>
    </logger>

    <http_port>8123</http_port>
    <tcp_port>9000</tcp_port>
    <mysql_port>9004</mysql_port>

    <path>%s/tmp/</path>
    <tmp_path>%s/tmp/clickhouse/tmp/</tmp_path>
    <user_files_path>%s/data/</user_files_path>

    <mlock_executable>true</mlock_executable>

    <users>
        <default>
            <password></password>

            <networks>
                <ip>::/0</ip>
            </networks>

            <profile>default</profile>
            <quota>default</quota>

            <access_management>1</access_management>
            <named_collection_control>1</named_collection_control>
        </default>
    </users>

    <profiles>
        <default/>
    </profiles>

    <quotas>
        <default />
    </quotas>
</clickhouse>`, projectPath, projectPath, projectPath))
	return config
}
