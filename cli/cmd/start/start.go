package start

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/web"
	"github.com/rilldata/rill/runtime/api"
	_ "github.com/rilldata/rill/runtime/connectors/gcs"
	_ "github.com/rilldata/rill/runtime/connectors/https"
	_ "github.com/rilldata/rill/runtime/connectors/s3"
	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/druid"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/postgres"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

var localInstanceID = "local"
var localRepoID = "local"

// StartCmd represents the start command
func StartCmd() *cobra.Command {
	var olapDriver string
	var olapDSN string
	var repoDSN string
	var httpPort int
	var grpcPort int
	var verbose bool

	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the Rill Developer application",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create a logger
			lvl := zapcore.ErrorLevel
			if verbose {
				lvl = zapcore.DebugLevel
			}
			logger, err := zap.NewDevelopment(zap.IncreaseLevel(lvl))
			if err != nil {
				return err
			}

			// Create an in-memory metastore
			metastore, err := drivers.Open("sqlite", "file:rill?mode=memory&cache=shared")
			if err != nil {
				return fmt.Errorf("error: could not connect to metadata db: %s", err)
			}
			err = metastore.Migrate(context.Background())
			if err != nil {
				return fmt.Errorf("error: metadata db migration: %s", err)
			}

			// Create the local runtime server
			opts := &server.ServerOptions{
				HTTPPort:             httpPort,
				GRPCPort:             grpcPort,
				ConnectionCacheSize:  100,
				CatalogCacheSize:     100,
				CatalogCacheDuration: 1 * time.Second,
			}
			server, err := server.NewServer(opts, metastore, logger)
			if err != nil {
				return fmt.Errorf("error: could not create server: %s", err)
			}

			// Create instance and repo configured for local use
			inst, err := server.CreateInstance(context.Background(), &api.CreateInstanceRequest{
				InstanceId:   localInstanceID,
				Driver:       olapDriver,
				Dsn:          olapDSN,
				Exposed:      true,
				EmbedCatalog: olapDriver == "duckdb",
			})
			if err != nil {
				return err
			}
			repo, err := server.CreateRepo(context.Background(), &api.CreateRepoRequest{
				RepoId: localRepoID,
				Driver: "file",
				Dsn:    repoDSN,
			})
			if err != nil {
				return err
			}
			logger.Sugar().Infof("Serving local instance '%s' and repo '%s'", inst.Instance.InstanceId, repo.Repo.RepoId)

			// Create config object to serve on /local/config
			localConfig := map[string]any{
				"instance_id": localInstanceID,
				"repo_id":     localRepoID,
				"grpc_port":   grpcPort,
			}

			// Prepare errgroup with graceful shutdown
			gctx := graceful.WithCancelOnTerminate(context.Background())
			group, ctx := errgroup.WithContext(gctx)

			// Create one HTTP server serving both the UI and the runtime's REST service
			uiHandler, err := web.StaticHandler()
			if err != nil {
				return err
			}
			runtimeHandler, err := server.HTTPHandler(ctx)
			if err != nil {
				return err
			}
			mux := http.NewServeMux()
			mux.Handle("/", uiHandler)
			mux.Handle("/v1/", runtimeHandler)
			mux.Handle("/local/config", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				data, err := json.Marshal(localConfig)
				if err != nil {
					w.WriteHeader(400)
					return
				}
				w.Header().Add("Content-Type", "application/json")
				w.Write(data)
			}))

			// Open the browser
			uiURL := fmt.Sprintf("http://localhost:%d", httpPort)
			err = browser.Open(uiURL)
			if err != nil {
				return fmt.Errorf("could not open browser: %v", err)
			}

			// Start the gRPC and combined UI/REST servers
			group.Go(func() error {
				return server.ServeGRPC(ctx)
			})
			group.Go(func() error {
				server := &http.Server{Handler: mux}
				logger.Info("serving HTTP", zap.Int("port", opts.HTTPPort))
				return graceful.ServeHTTP(ctx, server, opts.HTTPPort)
			})
			err = group.Wait()
			if err != nil {
				return fmt.Errorf("server crashed: %v", err)
			}

			logger.Info("server shutdown gracefully")
			return nil
		},
	}

	startCmd.Flags().StringVar(&olapDriver, "db-driver", "duckdb", "OLAP database driver")
	startCmd.Flags().StringVar(&olapDSN, "db", "stage.db", "OLAP database DSN")
	startCmd.Flags().StringVar(&repoDSN, "dir", ".", "Project directory")
	startCmd.Flags().IntVar(&httpPort, "port", 9009, "Port for the UI and runtime")
	startCmd.Flags().IntVar(&grpcPort, "port-grpc", 9010, "Port for the runtime's gRPC service")
	startCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")

	return startCmd
}
