package start

import (
	"bufio"
	"context"
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

	"github.com/rilldata/rill/cli/pkg/dotrill"
)

// startClickHouse installs (depending on OS and platform) and starts ClickHouse server.
// The destination directory for the ClickHouse binary is .rill/clickhouse.
// The function returns the DSN for the ClickHouse server.
func startClickHouse(ctx context.Context, projectPath string) (string, error) {
	// Store the ClickHouse binary under .rill/clickhouse so that every project can use the same binary
	destDir, err := dotrill.ResolveFilename("clickhouse", true)
	if err != nil {
		return "", err
	}

	binPath, err := installClickHouse(destDir)
	if err != nil {
		return "", err
	}

	projectAbsPath, err := filepath.Abs(projectPath)
	if err != nil {
		return "", err
	}

	configPath, err := createClickHouseConfig(projectAbsPath)
	if err != nil {
		return "", err
	}

	// Start ClickHouse server as a subprocess
	go func() {
		cmd := newCmd(ctx, binPath, "server", fmt.Sprintf("--config-file=%s", configPath))
		err = cmd.Run()
		if err != nil {
			fmt.Println("Error running clickhouse server", err)
		}
	}()

	// Wait for ClickHouse to be up and running
	address := net.JoinHostPort("localhost", "9000")
	err = tcpCheck(address, 30*time.Second)
	if err != nil {
		return "", err
	}

	return "clickhouse://" + address, nil
}

func installClickHouse(destDir string) (string, error) {
	destPath := filepath.Join(destDir, "clickhouse")

	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		err = os.MkdirAll(destDir, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("error creating ClickHouse directory: %w", err)
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

	url := fmt.Sprintf("https://builds.clickhouse.com/master/%s/clickhouse", dir)
	fmt.Printf("Will download %s into %s\n", url, destPath)
	if err := downloadFile(destPath, url); err != nil {
		return "", fmt.Errorf("error downloading ClickHouse: %w", err)
	}

	err := os.Chmod(destPath, 0x755)
	if err != nil {
		return "", fmt.Errorf("error setting executable permission: %w", err)
	}

	fmt.Println("Successfully downloaded the ClickHouse binary")

	return destPath, nil
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

func createClickHouseConfig(projectPath string) (string, error) {
	fc, err := os.CreateTemp("", "clickhouse-config*.xml")
	if err != nil {
		return "", err
	}
	defer fc.Close()

	content := clickHouseConfigContent(projectPath)

	if _, err := fc.Write(content); err != nil {
		return "", err
	}
	return fc.Name(), nil
}

// clickHouseConfigContent returns the content of the ClickHouse config file.
// ClickHouse requires a config file with a minimal set of properties to be passed:
// https://github.com/ClickHouse/ClickHouse/blob/master/programs/server/embedded.xml
// Full list of properties can be found here:
// https://github.com/ClickHouse/ClickHouse/blob/master/programs/server/config.xml
func clickHouseConfigContent(projectPath string) []byte {
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

// newCmd initializes an exec.Cmd that sends SIGINT instead of SIGKILL when the ctx is canceled.
func newCmd(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Cancel = func() error {
		return cmd.Process.Signal(os.Interrupt)
	}
	return cmd
}

func tcpCheck(address string, timeout time.Duration) error {
	start := time.Now()
	for {
		conn, err := net.DialTimeout("tcp", address, 1*time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
		if time.Since(start) > timeout {
			return fmt.Errorf("timed out waiting for ClickHouse to be ready")
		}
		time.Sleep(1 * time.Second)
	}
}
