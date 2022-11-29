package start

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/mattn/go-colorable"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/web"
	"github.com/rilldata/rill/runtime"
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

var localInstanceID = "default"

// StartCmd represents the start command
func StartCmd() *cobra.Command {
	var olapDriver string
	var olapDSN string
	var repoDSN string
	var httpPort int
	var grpcPort int
	var verbose bool
	var noUI bool
	var noOpen bool

	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the Rill Developer application",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create base logger
			config := zap.NewDevelopmentEncoderConfig()
			config.EncodeLevel = zapcore.CapitalColorLevelEncoder
			l := zap.New(zapcore.NewCore(
				zapcore.NewConsoleEncoder(config),
				zapcore.AddSync(colorable.NewColorableStdout()),
				zapcore.DebugLevel,
			))

			// Create derived loggers
			cliLevel := zap.InfoLevel
			serverLevel := zap.ErrorLevel
			if verbose {
				cliLevel = zap.DebugLevel
				serverLevel = zap.DebugLevel
			}
			logger := l.WithOptions(zap.IncreaseLevel(cliLevel))
			serverLogger := l.WithOptions(zap.IncreaseLevel(serverLevel))

			// Create local runtime
			rtOpts := &runtime.Options{
				ConnectionCacheSize: 100,
				MetastoreDriver:     "sqlite",
				MetastoreDSN:        "file:rill?mode=memory&cache=shared",
			}
			rt, err := runtime.New(rtOpts, logger)
			if err != nil {
				return err
			}

			// Create instance and repo configured for local use
			inst := &drivers.Instance{
				ID:           localInstanceID,
				OLAPDriver:   olapDriver,
				OLAPDSN:      olapDSN,
				RepoDriver:   "file",
				RepoDSN:      repoDSN,
				EmbedCatalog: olapDriver == "duckdb",
			}
			err = rt.CreateInstance(context.Background(), inst)
			if err != nil {
				return err
			}

			// Get full path to repo for logging
			repoAbs, err := filepath.Abs(repoDSN)
			if err != nil {
				return err
			}

			// Trigger reconciliation
			logger.Sugar().Infof("Hydrating project at '%s'", repoAbs)
			res, err := rt.Reconcile(context.Background(), inst.ID, nil, false, false)
			if err != nil {
				return err
			}
			for _, merr := range res.Errors {
				logger.Sugar().Errorf("%s: %s", merr.FilePath, merr.Message)
			}
			for _, path := range res.AffectedPaths {
				logger.Sugar().Infof("Reconciled: %s", path)
			}
			logger.Sugar().Infof("Hydration completed!")

			// Create the local runtime server
			srvOpts := &server.Options{
				HTTPPort: httpPort,
				GRPCPort: grpcPort,
			}
			server, err := server.NewServer(srvOpts, rt, serverLogger)
			if err != nil {
				return err
			}

			// Create config object to serve on /local/config
			localConfig := map[string]any{
				"instance_id": localInstanceID,
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
				return fmt.Errorf("could not create runtime http handler:%v", err)
			}

			// Add UI, runtime and local/config handlers on HTTP server
			localConfigHandler := localConfigHandler(localConfig)
			mux := http.NewServeMux()
			if !noUI {
				mux.Handle("/", uiHandler)
			}
			mux.Handle("/v1/", runtimeHandler)
			mux.Handle("/local/config", localConfigHandler)

			// Open the browser
			if !noUI && !noOpen {
				uiURL := fmt.Sprintf("http://localhost:%d", httpPort)
				err = browser.Open(uiURL)
				if err != nil {
					logger.Sugar().Warnf("could not open browser, error: %v, copy and paste this URL into your browser: %s", err, uiURL)
				}
			}

			// Start the gRPC and combined UI/REST servers
			group.Go(func() error {
				logger.Sugar().Debugf("Serving runtime gRPC on http://localhost:%d", srvOpts.GRPCPort)
				return server.ServeGRPC(ctx)
			})
			group.Go(func() error {
				server := &http.Server{Handler: mux}
				logger.Sugar().Infof("Serving Rill on: http://localhost:%d", srvOpts.HTTPPort)
				return graceful.ServeHTTP(ctx, server, srvOpts.HTTPPort)
			})

			err = group.Wait()
			if err != nil {
				return fmt.Errorf("rill crashed: %v", err)
			}

			logger.Info("Rill shutdown gracefully")
			return nil
		},
	}

	startCmd.Flags().StringVar(&olapDriver, "db-driver", "duckdb", "OLAP database driver")
	startCmd.Flags().StringVar(&olapDSN, "db", "stage.db", "OLAP database DSN")
	startCmd.Flags().StringVar(&repoDSN, "dir", ".", "Project directory")
	startCmd.Flags().IntVar(&httpPort, "port", 9009, "Port for the UI and runtime")
	startCmd.Flags().IntVar(&grpcPort, "port-grpc", 9010, "Port for the runtime's gRPC service")
	startCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")
	startCmd.Flags().BoolVar(&noUI, "no-ui", false, "Serve only the runtime")
	startCmd.Flags().BoolVar(&noOpen, "no-open", false, "Disable opening the browser window")

	return startCmd
}

func localConfigHandler(localConfig map[string]any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(localConfig)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.Write(data)
	})
}
