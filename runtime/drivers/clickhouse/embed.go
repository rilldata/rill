package clickhouse

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"go.uber.org/zap"
)

const embedVersion = "24.7.4.51"

var (
	embed *embedClickHouse
	once  sync.Once
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
	if tcpPort != embed.tcpPort {
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
	destDir, err := dotrill.ResolveFilename("clickhouse", true)
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

	e.cmd = exec.Command(binPath, "server", "--config-file", configPath)

	ready := make(chan error, 1)
	go func() {
		err := e.startAndWaitUntilReady()
		ready <- err
		if err != nil && e.cmd != nil && e.cmd.Process != nil {
			_ = e.cmd.Process.Kill()
			return
		}

		if err := e.cmd.Wait(); err != nil {
			e.logger.Error("clickhouse server exited with an error", zap.Error(err))
		}
	}()

	if err := <-ready; err != nil {
		return nil, err
	}

	addr := net.JoinHostPort("localhost", fmt.Sprintf("%d", e.tcpPort))
	e.logger.Info("Running an embedded ClickHouse server", zap.String("addr", addr))

	e.opts = &clickhouse.Options{
		Protocol: clickhouse.Native,
		Addr:     []string{addr},
	}
	return e.opts, nil
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

	err := e.cmd.Process.Signal(os.Interrupt)
	if err != nil {
		return err
	}
	e.opts = nil
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
			fileName = "clickhouse-common-static-24.7.4.51-amd64.tgz"
		case "arm64":
			fileName = "clickhouse-common-static-24.7.4.51-arm64.tgz"
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
        <level>information</level>
        <console>true</console>
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
            <named_collection_control>1</named_collection_control>
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

func (e *embedClickHouse) startAndWaitUntilReady() error {
	stderr, err := e.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

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
			if scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "<Error>") {
					e.logger.Error(line)
				} else if strings.Contains(line, "Application: Ready for connections") {
					return nil
				}
			} else {
				if err := scanner.Err(); err != nil {
					return fmt.Errorf("error reading clickhouse logs: %w", err)
				}
				return fmt.Errorf("clickhouse is not ready")
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
