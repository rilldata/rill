package main

import (
	"context"
	"flag"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"

	// Load postgres driver
	_ "github.com/jackc/pgx/v4/stdlib"
)

var (
	red    = color.New(color.FgHiRed)
	green  = color.New(color.FgHiGreen)
	yellow = color.New(color.FgHiYellow)
)

var (
	reset        = false
	startAdmin   = true
	startRuntime = true
	startUI      = true
)

func init() {
	flag.BoolVar(&reset, "reset", false, "reset db")
	flag.BoolVar(&startAdmin, "admin", true, "start admin server")
	flag.BoolVar(&startRuntime, "runtime", true, "start runtime server")
	flag.BoolVar(&startUI, "ui", true, "start ui")
}

func main() {
	flag.Parse()
	// check go version
	versionString := runtime.Version()
	major, minor := parseVersion(versionString)

	if major > 1 || (major == 1 && minor >= 20) {
		panic("require go version greater than 1.20")
	}

	// check node version
	nodeVersion, err := exec.Command("node", "--version").Output()
	if err != nil {
		panic("Error executing the 'node --version' command:" + err.Error())
	}

	versionString = strings.TrimPrefix(string(nodeVersion), "v")
	major, minor = parseVersion(versionString)
	if major == 1 && minor < 18.0 {
		panic("require Node.js version greater than 18")
	}

	// check docker version
	cmd := exec.Command("docker-compose", "--version")
	err = cmd.Run()

	if err != nil {
		panic("require docker-compose")
	}

	// observability - otel,zipkin,prometheus
	cmd = exec.Command("docker-compose", "-f", "scripts/observability/docker-compose.yaml", "up", "--no-recreate")
	err = cmd.Start()
	if err != nil {
		panic("could not start observability services")
	}

	if reset {
		cmd = exec.Command("docker-compose", "-f", "admin/docker-compose.yml", "down", "--volumes")
		err = cmd.Run()
		if err != nil {
			panic("could not stop db")
		}
		yellow.Println("DELETED EXISTING POSTGRES")
	}

	// postgres
	cmd = exec.Command("docker-compose", "-f", "admin/docker-compose.yml", "up", "--no-recreate")
	yellow.Println("STARTING POSTGRES")
	err = cmd.Start()
	if err != nil {
		panic("could not start db")
	}
	waitDBUP()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		yellow.Println("stopping services")
		time.Sleep(2 * time.Second)
		os.Exit(1)
	}()

	wg := sync.WaitGroup{}
	if startAdmin {
		// admin
		wg.Add(1)
		cmd = exec.Command("go", "run", "cli/main.go", "admin", "start")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
		yellow.Println("STARTING ADMIN")
		err = cmd.Start()
		if err != nil {
			panic("unable to start admin server")
		}
		go func(cmd *exec.Cmd) {
			_ = cmd.Wait()
			// nolint:all //required
			red.Println("ADMIN SERVER STOPPED")
			wg.Done()
		}(cmd)
		waitAdmin()
	}

	if startRuntime {
		// runtime
		wg.Add(1)
		cmd = exec.Command("go", "run", "cli/main.go", "runtime", "start")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
		yellow.Println("STARTING RUNTIME")
		err = cmd.Start()
		if err != nil {
			panic("unable to start runtime server")
		}
		go func(cmd *exec.Cmd) {
			_ = cmd.Wait()
			// nolint:all //required
			red.Println("RUNTIME SERVER STOPPED")
			wg.Done()
		}(cmd)
		waitRuntime()
	}

	if startUI {
		// UI
		wg.Add(1)
		yellow.Println("INSTALLING UI DEPENDENCIES")
		cmd = exec.Command("npm", "install", "-w", "web-admin")
		_ = cmd.Run()

		yellow.Println("STARTING UI")
		cmd = exec.Command("npm", "run", "dev", "-w", "web-admin")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
		err = cmd.Start()
		if err != nil {
			panic("unable to start UI")
		}
		go func(cmd *exec.Cmd) {
			_ = cmd.Wait()
			// nolint:all //required
			red.Println("UI STOPPED")
			wg.Done()
		}(cmd)
		// todo :: add health check ui
	}
	wg.Wait()
}

func parseVersion(versionStr string) (int, int) {
	versionComponents := strings.Split(versionStr, ".")
	major, _ := strconv.Atoi(versionComponents[0])
	minor, _ := strconv.Atoi(versionComponents[1])
	return major, minor
}

func waitDBUP() {
	_ = godotenv.Load()
	connectionString := os.Getenv("RILL_ADMIN_DATABASE_URL")
	for i := 0; i < 10; i++ {
		conn, err := pgx.Connect(context.Background(), connectionString)
		if err == nil {
			conn.Close(context.Background())
			green.Println("POSTGRES STARTED")
			return
		}
		time.Sleep(2 * time.Second)
	}
	red.Println("Could not start postgres server")
}

func waitAdmin() {
	for i := 0; i < 10; i++ {
		cmd := exec.Command("go", "run", "cli/main.go", "admin", "ping", "--url", "http://localhost:9090")
		out, _ := cmd.Output()
		if strings.Contains(string(out), "Pong") {
			green.Println("ADMIN STARTED")
			return
		}
		time.Sleep(2 * time.Second)
	}
	red.Println("Could not start admin server")
}

func waitRuntime() {
	for i := 0; i < 10; i++ {
		cmd := exec.Command("go", "run", "cli/main.go", "runtime", "ping", "--url", "http://localhost:9091")
		out, _ := cmd.Output()
		if strings.Contains(string(out), "Pong") {
			green.Println("RUNTIME STARTED")
			return
		}
		time.Sleep(2 * time.Second)
	}
	red.Println("Could not start runtime server")
}
