package start

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/mattn/go-colorable"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/web"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
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
			logger := l.WithOptions(zap.IncreaseLevel(cliLevel)).Sugar()
			serverLogger := l.WithOptions(zap.IncreaseLevel(serverLevel))

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
				HTTPPort:            httpPort,
				GRPCPort:            grpcPort,
				ConnectionCacheSize: 100,
			}
			srv, err := server.NewServer(opts, metastore, serverLogger)
			if err != nil {
				return err
			}

			// create dir if it doesn't exist
			err = os.MkdirAll(repoDSN, os.ModePerm)
			if err != nil {
				return err
			}
			// Create instance and repo configured for local use
			_, err = srv.CreateInstance(context.Background(), &runtimev1.CreateInstanceRequest{
				InstanceId:   localInstanceID,
				OlapDriver:   olapDriver,
				OlapDsn:      olapDSN,
				RepoDriver:   "file",
				RepoDsn:      repoDSN,
				EmbedCatalog: olapDriver == "duckdb",
			})
			if err != nil {
				return err
			}

			// Get full path to repo for logging
			repoAbs, err := filepath.Abs(repoDSN)
			if err != nil {
				return err
			}

			// Trigger reconciliation
			logger.Infof("Hydrating project at '%s'", repoAbs)
			res, err := srv.Reconcile(context.Background(), &runtimev1.ReconcileRequest{
				InstanceId: localInstanceID,
			})
			if err != nil {
				return err
			}
			for _, merr := range res.Errors {
				logger.Errorf("%s: %s", merr.FilePath, merr.Message)
			}
			for _, p := range res.AffectedPaths {
				logger.Infof("Reconciled: %s", p)
			}
			logger.Infof("Hydration completed!")

			installId, err := initGlobalConfig()
			if err != nil {
				return err
			}
			_, isDev := os.LookupEnv("RILL_IS_DEV")
			// Create config object to serve on /local/config
			localConfig := map[string]any{
				"instance_id": localInstanceID,
				"grpc_port":   grpcPort,
				"install_id":  installId,
				// TODO: do we need to only get the last folder?
				"project_id": fmt.Sprintf("%x", md5.Sum([]byte(olapDSN))),
				"is_dev":     isDev,
			}

			// Prepare errgroup with graceful shutdown
			gctx := graceful.WithCancelOnTerminate(context.Background())
			group, ctx := errgroup.WithContext(gctx)

			// Create one HTTP server serving both the UI and the runtime's REST service
			uiHandler, err := web.StaticHandler()
			if err != nil {
				return err
			}
			runtimeHandler, err := srv.HTTPHandler(ctx)
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
					logger.Warnf("could not open browser, error: %v, copy and paste this URL into your browser: %s", err, uiURL)
				}
			}

			// Start the gRPC and combined UI/REST servers
			group.Go(func() error {
				logger.Debugf("Serving runtime gRPC on http://localhost:%d", opts.GRPCPort)
				return srv.ServeGRPC(ctx)
			})
			group.Go(func() error {
				srv := &http.Server{Handler: mux}
				logger.Infof("Serving Rill on: http://localhost:%d", opts.HTTPPort)
				return graceful.ServeHTTP(ctx, srv, opts.HTTPPort)
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

func initGlobalConfig() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	confFolder := path.Join(home, ".rill")
	_, err = os.Stat(confFolder)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// create folder if not exists
			err := os.MkdirAll(confFolder, os.ModePerm)
			if err != nil {
				return "", err
			}
		} else {
			// unknown error
			return "", err
		}
	}

	globalConf := map[string]any{}
	var installId string

	confFile := path.Join(confFolder, "local.json")
	_, err = os.Stat(confFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			// return if unknown error
			return "", err
		}
	} else {
		// read file if exists
		conf, err := os.ReadFile(confFile)
		if err != nil {
			return "", err
		}
		err = json.Unmarshal(conf, &globalConf)
		if err != nil {
			return "", err
		}
	}

	// installId was used in nodejs.
	// keeping it as is to retain the same ID for existing users
	installIdAny, ok := globalConf["installId"]
	if !ok {
		// create install id if not exists
		installId = uuid.New().String()
		globalConf["installId"] = installId
		globalConfJson, err := json.Marshal(&globalConf)
		if err != nil {
			return "", err
		}
		err = os.WriteFile(confFile, globalConfJson, 0644)
		if err != nil {
			return "", err
		}
	} else {
		installId = installIdAny.(string)
	}

	return installId, nil
}
