package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"golang.org/x/sync/errgroup"

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

	cctx := graceful.WithCancelOnTerminate(context.Background())
	group, ctx := errgroup.WithContext(cctx)

	// check node version
	nodeVersion, err := exec.CommandContext(ctx, "node", "--version").Output()
	if err != nil {
		panic("Error executing the 'node --version' command:" + err.Error())
	}

	versionString = strings.TrimPrefix(string(nodeVersion), "v")
	major, _ = parseVersion(versionString)
	if major < 18 {
		panic("require Node.js version greater than 18")
	}

	// check docker version
	cmd := exec.CommandContext(ctx, "docker-compose", "--version")
	err = cmd.Run()

	if err != nil {
		panic("require docker-compose")
	}

	// observability - otel,zipkin,prometheus
	cmd = exec.CommandContext(ctx, "docker-compose", "-f", "scripts/observability/docker-compose.yaml", "up", "--no-recreate")
	err = cmd.Start()
	if err != nil {
		panic("could not start observability services")
	}

	if reset {
		cmd = exec.CommandContext(ctx, "docker-compose", "-f", "admin/docker-compose.yml", "down", "--volumes")
		err = cmd.Run()
		if err != nil {
			panic("could not stop db")
		}
		yellow.Println("DELETED EXISTING POSTGRES")
	}

	// postgres
	cmd = exec.CommandContext(ctx, "docker-compose", "-f", "admin/docker-compose.yml", "up", "--no-recreate")
	yellow.Println("STARTING POSTGRES")
	err = cmd.Start()
	if err != nil {
		panic("could not start db")
	}
	waitDBUP(ctx)
	waitRedis(ctx)

	if startAdmin {
		// admin
		admin := exec.CommandContext(ctx, "go", "run", "cli/main.go", "admin", "start")
		admin.Stdout = os.Stdout
		admin.Stderr = os.Stdout
		yellow.Println("STARTING ADMIN")
		err = admin.Start()
		if err != nil {
			panic("unable to start admin server")
		}
		group.Go(func() error {
			err := admin.Wait()
			if err != nil {
				yellow.Println(err)
			}
			// nolint:all //required
			red.Println("ADMIN SERVER STOPPED")
			return err
		})
		waitAdmin(ctx)
	}

	if startRuntime {
		// runtime
		rt := exec.CommandContext(ctx, "go", "run", "cli/main.go", "runtime", "start")
		rt.Stdout = os.Stdout
		rt.Stderr = os.Stdout
		yellow.Println("STARTING RUNTIME")
		err = rt.Start()
		if err != nil {
			panic("unable to start runtime server")
		}
		group.Go(func() error {
			err := rt.Wait()
			if err != nil {
				yellow.Println(err)
			}
			// nolint:all //required
			red.Println("RUNTIME SERVER STOPPED")
			return err
		})
		waitRuntime(ctx)
	}

	if startUI {
		// UI
		yellow.Println("INSTALLING UI DEPENDENCIES")
		ui := exec.CommandContext(ctx, "npm", "install", "-w", "web-admin")
		err = ui.Run()
		if err != nil {
			panic(err)
		}

		yellow.Println("STARTING UI")
		ui = exec.CommandContext(ctx, "npm", "run", "dev", "-w", "web-admin")
		ui.Stdout = os.Stdout
		ui.Stderr = os.Stdout
		err = ui.Start()
		if err != nil {
			panic("unable to start UI")
		}
		group.Go(func() error {
			err := ui.Wait()
			if err != nil {
				yellow.Println(err)
			}
			// nolint:all //required
			red.Println("UI STOPPED")
			return err
		})
		// todo :: add health check ui
	}
	_ = group.Wait()
}

func parseVersion(versionStr string) (int, int) {
	versionComponents := strings.Split(versionStr, ".")
	major, _ := strconv.Atoi(versionComponents[0])
	minor, _ := strconv.Atoi(versionComponents[1])
	return major, minor
}

func waitDBUP(ctx context.Context) {
	_ = godotenv.Load()
	connectionString := os.Getenv("RILL_ADMIN_DATABASE_URL")
	for i := 0; i < 10; i++ {
		conn, err := pgx.Connect(ctx, connectionString)
		if err == nil {
			conn.Close(ctx)
			green.Println("POSTGRES STARTED")
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}
		time.Sleep(2 * time.Second)
	}
	red.Println("Could not start postgres server")
}

func waitAdmin(ctx context.Context) {
	for i := 0; i < 10; i++ {
		cmd := exec.CommandContext(ctx, "go", "run", "cli/main.go", "admin", "ping", "--url", "http://localhost:9090")
		out, err := cmd.Output()
		if errors.Is(err, context.Canceled) {
			return
		}
		if strings.Contains(string(out), "Pong") {
			green.Println("ADMIN STARTED")
			return
		}
		time.Sleep(2 * time.Second)
	}
	red.Println("Could not start admin server")
}

func waitRedis(ctx context.Context) {
	connectionString := os.Getenv("RILL_ADMIN_REDIS_URL")
	if connectionString == "" {
		return
	}
	opts, err := redis.ParseURL(connectionString)
	if err != nil {
		panic("failed to parse redis url " + err.Error())
	}
	c := redis.NewClient(opts)
	defer c.Close()
	res, err := c.Echo(ctx, "hello").Result()
	if err != nil || res != "hello" {
		panic("redis not started")
	}
}

func waitRuntime(ctx context.Context) {
	for i := 0; i < 10; i++ {
		cmd := exec.CommandContext(ctx, "go", "run", "cli/main.go", "runtime", "ping", "--url", "http://localhost:9091")
		out, err := cmd.Output()
		if strings.Contains(string(out), "Pong") {
			green.Println("RUNTIME STARTED")
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}
		time.Sleep(2 * time.Second)
	}
	red.Println("Could not start runtime server")
}
