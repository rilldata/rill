package clickhouse

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"go.uber.org/zap"
	"golang.org/x/sys/cpu"
)

type embedClickHouse struct {
	tcpPort int
	dataDir string
	tempDir string
	logger  *zap.Logger
	cmd     *exec.Cmd
}

func newEmbedClickHouse(tcpPort int, dataDir, tempDir string, logger *zap.Logger) *embedClickHouse {
	return &embedClickHouse{tcpPort: tcpPort, dataDir: dataDir, tempDir: tempDir, logger: logger}
}

// start installs (depending on OS and platform) and starts ClickHouse server.
// The destination directory for the ClickHouse binary is .rill/clickhouse.
// The function returns the DSN for the ClickHouse server and close function.
func (e *embedClickHouse) start() (*clickhouse.Options, error) {
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
	e.logger.Info("ClickHouse server: " + "clickhouse://" + addr)

	return &clickhouse.Options{
		Protocol: clickhouse.Native,
		Addr:     []string{addr},
	}, nil
}

func (e *embedClickHouse) stop() error {
	e.logger.Info("Stopping ClickHouse server: " + fmt.Sprintf("localhost:%d", e.tcpPort))
	if e.cmd == nil {
		return nil
	}

	if e.cmd.Process == nil {
		return nil
	}

	err := e.cmd.Process.Signal(os.Interrupt)
	if err != nil {
		return err
	}

	return nil
}

func (e *embedClickHouse) install(destDir string, logger *zap.Logger) (string, error) {
	destPath := filepath.Join(destDir, "clickhouse")

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
	dir := ""

	// The following OS and platform matching mostly repeats the logic in the ClickHouse installation script
	// https://github.com/ClickHouse/ClickHouse/blob/master/docs/_includes/install/universal.sh
	switch goos {
	case "linux":
		switch goarch {
		case "amd64":
			if cpu.X86.HasSSE42 {
				dir = "amd64"
			} else {
				dir = "amd64compat"
			}

		case "arm64":
			if cpu.ARM64.HasASIMD && cpu.ARM64.HasSHA1 && cpu.ARM64.HasAES &&
				cpu.ARM64.HasATOMICS && cpu.ARM64.HasLRCPC {
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

	url := fmt.Sprintf("https://builds.clickhouse.com/master/%s/clickhouse", dir)
	logger.Info(fmt.Sprintf("Will download %s into %s\n", url, destPath))
	if err := downloadFile(destPath, url); err != nil {
		return "", fmt.Errorf("error downloading ClickHouse: %w", err)
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

	return e.waitUntilReady(30*time.Second, stderr)
}

func (e *embedClickHouse) waitUntilReady(timeout time.Duration, reader io.ReadCloser) error {
	scanner := bufio.NewScanner(reader)
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			return fmt.Errorf("clickhouse is not ready: timeout after %v", timeout)
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

func getFreePort() (int, error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}
