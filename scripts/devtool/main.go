package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
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

	c, cancel := context.WithCancel(context.Background())
	defer cancel()
	group, ctx := errgroup.WithContext(graceful.WithCancelOnTerminate(c))

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

	if reset {
		cmd = exec.CommandContext(ctx, "docker-compose", "-f", "admin/docker-compose.yml", "down", "--volumes")
		err = cmd.Run()
		if err != nil {
			panic("could not stop db")
		}
		yellow.Println("DELETED EXISTING VOLUMES")
	}

	// postgres,redis,observability services
	pgCmd := exec.CommandContext(ctx, "docker-compose", "-f", "admin/docker-compose.yml", "up", "--no-recreate")
	yellow.Println("STARTING DOCKER SERVICES")
	err = pgCmd.Start()
	if err != nil {
		panic("could not start db")
	}
	group.Go(func() error {
		err := pgCmd.Wait()
		red.Println("STOPPING DOCKER SERVICES")
		cancel()
		return err
	})
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
			red.Println("ADMIN SERVER STOPPED")
			cancel()
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
			red.Println("RUNTIME SERVER STOPPED")
			cancel()
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
			red.Println("UI STOPPED")
			cancel()
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
		resp, err := http.Get("http://localhost:8080/v1/ping")
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			time.Sleep(2 * time.Second)
			continue
		}
		statusCode := resp.StatusCode
		resp.Body.Close()
		if statusCode == http.StatusOK {
			green.Println("ADMIN STARTED")
			return
		}
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
	for i := 0; i < 10; i++ {
		c := redis.NewClient(opts)
		res, err := c.Echo(ctx, "hello").Result()
		c.Close()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			time.Sleep(2 * time.Second)
			continue
		}
		if res == "hello" {
			green.Println("REDIS STARTED")
			return
		}
	}
	red.Println("redis not started")
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
