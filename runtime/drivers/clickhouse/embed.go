package clickhouse

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"go.uber.org/zap"
)

const embedVersion = "25.5.1.2782"

var (
	embed             *embedClickHouse
	once              sync.Once
	errAlreadyRunning = fmt.Errorf("ClickHouse server is already running, please stop it before starting a new one")
)

type embedClickHouse struct {
	tcpPort int
	dataDir string
	tempDir string
	logger  *zap.Logger
	cmd     *exec.Cmd
	opts    *clickhouse.Options
	// number of calls to start. The server is stopped when the count reaches 0.
	refs int
	mu   sync.Mutex
}

func newEmbedClickHouse(tcpPort int, dataDir, tempDir string, logger *zap.Logger) (*embedClickHouse, error) {
	once.Do(func() {
		embed = &embedClickHouse{tcpPort: tcpPort, dataDir: dataDir, tempDir: tempDir, logger: logger}
	})
	if tcpPort != 0 && tcpPort != embed.tcpPort {
		return nil, fmt.Errorf("change of `embed_port` is not allowed while the application is running, please restart Rill")
	}
	return embed, nil
}

// start installs (depending on OS and platform) and starts ClickHouse server.
// The destination directory for the ClickHouse binary is .rill/clickhouse.
// The function returns the DSN for the ClickHouse server and close function.
//
// TODO: Since this can be a long-running process, we should accept a `ctx`,
// but the `drivers.Open` function currently doesn't propagate that.
func (e *embedClickHouse) start() (*clickhouse.Options, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.opts != nil {
		e.refs++
		return e.opts, nil
	}

	// Store the ClickHouse binary under .rill/clickhouse so that every project can use the same binary
	destDir, err := dotrill.New("").ResolveFilename("clickhouse", true)
	if err != nil {
		return nil, err
	}

	binPath, err := e.install(destDir, e.logger)
	if err != nil {
		return nil, err
	}

	if e.tcpPort == 0 {
		e.tcpPort, err = getFreePort()
		if err != nil {
			return nil, err
		}
	}

	configPath, err := e.prepareConfig()
	if err != nil {
		return nil, err
	}

	// Start the ClickHouse server
	stderr, err := e.startClickhouse(binPath, configPath)
	if err != nil {
		if !errors.Is(err, errAlreadyRunning) {
			return nil, err
		}
		e.logger.Warn("ClickHouse server is already running, attempting to kill the existing process")
		err = e.killExistingProcess()
		if err != nil {
			return nil, fmt.Errorf("failed to kill existing ClickHouse process: %w", err)
		}
		stderr, err = e.startClickhouse(binPath, configPath)
		if err != nil {
			return nil, err
		}
	}

	// If you're using cmd.StdoutPipe() or cmd.StderrPipe() and not reading from them fast enough,
	// the buffer can fill up, and the subprocess will block on writing output.
	// We read StderrPipe initially to check for clickhouse running status.
	// Once the process is closed the stderr pipe will be closed too, io.Copy will return EOF and the goroutine will exit.
	go func() {
		decoder := json.NewDecoder(stderr)

		for {
			var log clickhouseLog
			if err := decoder.Decode(&log); err != nil {
				if err == io.EOF {
					// EOF means the process has exited and the stderr pipe is closed.
					break
				}
				e.logger.Error("Failed to decode ClickHouse log", zap.Error(err))
				continue
			}

			switch log.Level {
			case "Fatal":
				e.logger.Error("ClickHouse embedded server: fatal log received, restart server", zap.String("logger_name", log.LoggerName), zap.String("message", log.Message))
			case "Critical", "Error":
				code := parseErrorCode(log.Message)
				if isUserError(code) {
					e.logger.Debug("ClickHouse embedded server", zap.String("logger_name", log.LoggerName), zap.String("message", log.Message))
				} else {
					e.logger.Error("ClickHouse embedded server", zap.String("logger_name", log.LoggerName), zap.String("message", log.Message))
				}
			case "Warning", "Notice":
				code := parseErrorCode(log.Message)
				if isUserError(code) {
					e.logger.Debug("ClickHouse embedded server", zap.String("logger_name", log.LoggerName), zap.String("message", log.Message))
				} else {
					e.logger.Warn("ClickHouse embedded server", zap.String("logger_name", log.LoggerName), zap.String("message", log.Message))
				}
			case "Information", "Debug", "Trace", "Test":
				// even the information logs are too verbose in clickhouse so we log them at debug level
				e.logger.Debug("ClickHouse embedded server", zap.String("logger_name", log.LoggerName), zap.String("message", log.Message))
			}
		}
		stderr.Close()
	}()

	addr := net.JoinHostPort("localhost", fmt.Sprintf("%d", e.tcpPort))
	e.logger.Info("Running an embedded ClickHouse server", zap.String("addr", addr))

	e.opts = &clickhouse.Options{
		Protocol: clickhouse.Native,
		Addr:     []string{addr},
	}
	return e.opts, nil
}

func (e *embedClickHouse) startClickhouse(binPath, configPath string) (io.ReadCloser, error) {
	e.cmd = exec.Command(binPath, "server", "--config-file", configPath)
	e.cmd.Stdout = io.Discard

	stderr, err := e.cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	ready := make(chan error, 1)
	go func() {
		err := e.startAndWaitUntilReady(stderr)
		ready <- err
		if err != nil && e.cmd != nil && e.cmd.Process != nil {
			_ = e.cmd.Process.Kill()
			return
		}
	}()

	if err := <-ready; err != nil {
		return nil, err
	}
	return stderr, nil
}

func (e *embedClickHouse) stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.opts == nil || e.refs < 0 {
		// should never happen
		return nil
	}
	e.refs--
	if e.refs > 0 {
		return nil
	}

	addr := net.JoinHostPort("localhost", fmt.Sprintf("%d", e.tcpPort))
	e.logger.Info("Stopping embedded ClickHouse server", zap.String("addr", addr))
	if e.cmd == nil || e.cmd.Process == nil {
		return nil
	}

	err := e.cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		return err
	}
	_ = e.cmd.Wait()
	e.opts = nil
	e.cmd = nil
	return nil
}

func (e *embedClickHouse) install(destDir string, logger *zap.Logger) (string, error) {
	release := "v" + embedVersion + "-stable"
	destPath := filepath.Join(destDir, embedVersion, "clickhouse")

	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		err = os.MkdirAll(destDir, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("error creating clickhouse directory: %w", err)
		}
	}

	if _, err := os.Stat(destPath); err == nil {
		// ClickHouse binary already exists
		// TODO: Check compatibility in case the binary is outdated.
		return destPath, nil
	}

	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// The following OS and platform matching mostly repeats the logic in the ClickHouse installation script
	// https://github.com/ClickHouse/ClickHouse/blob/master/docs/_includes/install/universal.sh
	switch goos {
	case "darwin":
		fileName := ""
		switch goarch {
		case "amd64":
			fileName = "clickhouse-macos"
		case "arm64":
			fileName = "clickhouse-macos-aarch64"
		default:
			return "", fmt.Errorf("unsupported architecture %q for embedded Clickhouse", goarch)
		}
		url := "https://github.com/ClickHouse/ClickHouse/releases/download/" + release + "/" + fileName
		logger.Info("Downloading ClickHouse binary", zap.String("url", url), zap.String("dst", destPath))
		if err := downloadFile(destPath, url); err != nil {
			return "", fmt.Errorf("error downloading ClickHouse: %w", err)
		}
	case "linux":
		fileName := ""
		switch goarch {
		case "amd64":
			fileName = fmt.Sprintf("clickhouse-common-static-%s-amd64.tgz", embedVersion)
		case "arm64":
			fileName = fmt.Sprintf("clickhouse-common-static-%s-arm64.tgz", embedVersion)
		default:
			return "", fmt.Errorf("unsupported architecture %q for embedded Clickhouse", goarch)
		}
		url := "https://github.com/ClickHouse/ClickHouse/releases/download/" + release + "/" + fileName
		destTgzPath := filepath.Join(destDir, release, fileName)
		logger.Info("Downloading ClickHouse binary", zap.String("url", url), zap.String("dst", destTgzPath))
		if err := downloadFile(destTgzPath, url); err != nil {
			return "", fmt.Errorf("error downloading ClickHouse: %w", err)
		}
		fileToExtract := filepath.Join("clickhouse-common-static-"+embedVersion, "usr", "bin", "clickhouse")
		if err := extractFileFromTgz(destTgzPath, fileToExtract, destPath); err != nil {
			return "", fmt.Errorf("error extracting ClickHouse: %w", err)
		}
	default:
		return "", fmt.Errorf("unsupported OS %q for embedded Clickhouse", goos)
	}

	err := os.Chmod(destPath, 0x755)
	if err != nil {
		return "", fmt.Errorf("error setting executable permission: %w", err)
	}

	logger.Info("Successfully downloaded the ClickHouse binary")

	return destPath, nil
}

func (e *embedClickHouse) prepareConfig() (string, error) {
	err := os.MkdirAll(e.dataDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	fc, err := os.Create(filepath.Join(e.dataDir, "config.xml"))
	if err != nil {
		return "", err
	}
	defer fc.Close()

	content, err := e.getConfigContent()
	if err != nil {
		return "", err
	}

	if _, err := fc.Write(content); err != nil {
		return "", err
	}
	return fc.Name(), nil
}

// getConfigContent returns the content of the ClickHouse config file.
// ClickHouse requires a config file with a minimal set of properties to be passed:
// https://github.com/ClickHouse/ClickHouse/blob/master/programs/server/embedded.xml
// Full list of properties can be found here:
// https://github.com/ClickHouse/ClickHouse/blob/master/programs/server/config.xml
func (e *embedClickHouse) getConfigContent() ([]byte, error) {
	dataDirAbsPath, err := filepath.Abs(e.dataDir)
	if err != nil {
		return nil, err
	}

	tempDirAbsPath, err := filepath.Abs(e.tempDir)
	if err != nil {
		return nil, err
	}

	config := []byte(fmt.Sprintf(`<clickhouse>
    <logger>
        <level>debug</level>
        <console>true</console>
		<formatting>
			<type>json</type>
			<names>
				<level>level</level>
				<logger_name>logger_name</logger_name>
				<message>message</message>
			</names>
		</formatting>
    </logger>

    <tcp_port>%d</tcp_port>
    <http_port>0</http_port>
    <mysql_port>0</mysql_port>

    <path>%s</path>
    <tmp_path>%s</tmp_path>

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
        </default>
    </users>

    <profiles>
        <default/>
    </profiles>

    <quotas>
        <default />
    </quotas>
</clickhouse>`, e.tcpPort, dataDirAbsPath, tempDirAbsPath))
	return config, nil
}

func (e *embedClickHouse) startAndWaitUntilReady(stderr io.Reader) error {
	if err := e.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start clickhouse: %w", err)
	}

	scanner := bufio.NewScanner(stderr)

	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			return fmt.Errorf("clickhouse is not ready: timeout")
		default:
			if !scanner.Scan() {
				if scanner.Err() != nil {
					return fmt.Errorf("error reading clickhouse logs: %w", scanner.Err())
				}
				return fmt.Errorf("clickhouse is not ready")
			}
			line := scanner.Text()
			var logLine clickhouseLog
			if err := json.Unmarshal([]byte(line), &logLine); err != nil {
				// Till the clickhouse configs are parsed the logs may not be in JSON format.
				continue
			}
			if logLine.LoggerName == "Application" && strings.Contains(logLine.Message, "Ready for connections") {
				return nil
			}
			if logLine.Level == "Error" {
				if strings.Contains(logLine.Message, "Another server instance in same directory is already running") {
					return errAlreadyRunning
				}
				e.logger.Error("ClickHouse error", zap.String("message", logLine.Message), zap.String("logger_name", logLine.LoggerName))
			}
		}
	}
}

func downloadFile(path, url string) error {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}

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

// extractFileFromTgz extracts a file from a .tgz archive.
// The function searches for the file with the given name in the archive and extracts it to the destination path.
func extractFileFromTgz(tgzPath, fileName, destPath string) error {
	file, err := os.Open(tgzPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()

	tarReader := tar.NewReader(gz)

	// Iterate through the files in the tar archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		if header.Typeflag == tar.TypeReg && header.Name == fileName {
			if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
				return err
			}

			destFile, err := os.Create(destPath)
			if err != nil {
				return err
			}

			if _, err := io.Copy(destFile, tarReader); err != nil { //nolint:gosec // Source is trusted, no risk of G110: Potential DoS vulnerability
				destFile.Close()
				return err
			}

			if err := destFile.Close(); err != nil {
				return err
			}
			break
		}
	}

	return nil
}

func getFreePort() (int, error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

var errorCodeRegexp = regexp.MustCompile(`Code:\s*(\d+)`)

func parseErrorCode(msg string) int {
	match := errorCodeRegexp.FindStringSubmatch(msg)
	if len(match) < 2 {
		return -1
	}
	code, err := strconv.Atoi(match[1])
	if err != nil {
		return -1
	}
	return code
}

// isUserError checks if the error code is a user error.
// Clickhouse returns lots of exceptions that are not server errors or an issue with the server but rather a user error.
// Exceptions are usually accompanied by an error code, which is a number that can be used to identify the type of error.
// This function checks for specific error codes that are considered user errors.
// As of writing this is not an exhaustive list so if you find an error that is not to be logged add the error code here.
func isUserError(code int) bool {
	switch code {
	case 16: // no such column in table
		return true
	case 20: // number of columns does not match
		return true
	case 34, 35: // too many/too few arguments for function
		return true
	case 50: // unknown type
		return true
	case 38, 42, 43, 44, 46, 47, 51, 52, 53, 57, 60, 62, 63, 81, 82, 179: // different query syntax error
		return true
	case 181, 182, 183, 184, 215: // aggregate syntax error
		return true
	case 210: // network error but also thrown on query cancellation, connection failures etc which are usually auto recovered
		return true
	case 394: // query was cancelled
		return true

	default:
		return false
	}
}

func (e *embedClickHouse) killExistingProcess() error {
	file, err := os.Open(filepath.Join(e.dataDir, "status"))
	if err != nil {
		return err
	}
	defer file.Close()

	var pid int
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "PID:") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			pid, err = strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return fmt.Errorf("failed to parse PID from status file: %w", err)
			}
			break
		}
	}

	if pid <= 0 {
		return fmt.Errorf("no valid PID found in status file")
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process with PID %d: %w", pid, err)
	}
	err = p.Signal(syscall.SIGTERM)
	if err != nil {
		return fmt.Errorf("failed to kill process with PID %d: %w", pid, err)
	}

	// wait for the process to exit
	// unfortunately no way to wait for the process to exit gracefully since we don't have the original handle that started this process
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout while waiting for process with PID %d to exit: %w", pid, ctx.Err())
		case <-time.After(500 * time.Millisecond):
			// check if the process is still running
			err = p.Signal(syscall.Signal(0))
			if err != nil {
				if errors.Is(err, os.ErrProcessDone) {
					// process has exited
					return nil
				}
				// some other error occurred, return it
				return fmt.Errorf("failed to check if process with PID %d is running: %w", pid, err)
			}
			// process is still running, continue waiting
		}
	}
}

type clickhouseLog struct {
	LoggerName string `json:"logger_name"`
	Message    string `json:"message"`
	Level      string `json:"level"`
}
