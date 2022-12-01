package start

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/config"
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
	runtimeserver "github.com/rilldata/rill/runtime/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

const localInstanceID = "default"
const defaultOLAPDriver = "duckdb"
const defaultOLAPDSN = "stage.db"

// StartCmd represents the start command
func StartCmd(ver string) *cobra.Command {
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
			conf := zap.NewDevelopmentEncoderConfig()
			conf.EncodeLevel = zapcore.CapitalColorLevelEncoder
			l := zap.New(zapcore.NewCore(
				zapcore.NewConsoleEncoder(conf),
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

			// Create project dir if it doesn't exist
			err = os.MkdirAll(repoDSN, os.ModePerm)
			if err != nil {
				return err
			}

			// If no OLAP is specifically set, initialize it in the repo dir, not the working directory
			if olapDriver == defaultOLAPDriver && olapDSN == defaultOLAPDSN {
				olapDSN = path.Join(repoDSN, olapDSN)
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
			res, err := rt.Reconcile(context.Background(), inst.ID, nil, nil, false, false)
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

			// Build local info for frontend
			installID, err := config.InstallID()
			if err != nil {
				return err
			}
			inf := &localInfo{
				InstanceID:  localInstanceID,
				GRPCPort:    grpcPort,
				InstallID:   installID,
				ProjectPath: repoAbs,
				IsDev:       ver == "",
			}

			// Create the local server
			srv := &localServer{
				runtime:   rt,
				logger:    serverLogger,
				httpPort:  httpPort,
				grpcPort:  grpcPort,
				disableUI: noUI,
				info:      inf,
			}

			// Open the browser when health check succeeds
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				// Basic health check
				uri := fmt.Sprintf("http://localhost:%d", httpPort)
				client := http.Client{Timeout: time.Second}
				for {
					// Check for cancellation
					if ctx.Err() != nil {
						return
					}

					// Check if server is up
					resp, err := client.Get(uri + "/local/health")
					if err == nil {
						defer resp.Body.Close()
						if resp.StatusCode < http.StatusInternalServerError {
							break
						}
					}

					// Wait a bit and retry
					time.Sleep(10 * time.Millisecond)
				}

				// Health check succeeded
				logger.Sugar().Infof("Serving Rill on: %s", uri)
				if !noUI && !noOpen {
					err = browser.Open(uri)
					if err != nil {
						logger.Debug("could not open browser", zap.Error(err))
					}
				}
			}()

			// Serve the local server
			err = srv.ListenAndServe()
			if err != nil {
				return fmt.Errorf("server crashed: %w", err)
			}
			logger.Info("Rill shutdown gracefully")
			return nil
		},
	}

	startCmd.Flags().StringVar(&olapDriver, "db-driver", defaultOLAPDriver, "OLAP database driver")
	startCmd.Flags().StringVar(&olapDSN, "db", defaultOLAPDSN, "OLAP database DSN")
	startCmd.Flags().StringVar(&repoDSN, "dir", ".", "Project directory")
	startCmd.Flags().IntVar(&httpPort, "port", 9009, "Port for the UI and runtime")
	startCmd.Flags().IntVar(&grpcPort, "port-grpc", 9010, "Port for the runtime's gRPC service")
	startCmd.Flags().BoolVar(&verbose, "verbose", false, "Sets the log level to debug")
	startCmd.Flags().BoolVar(&noUI, "no-ui", false, "Serve only the runtime")
	startCmd.Flags().BoolVar(&noOpen, "no-open", false, "Disable opening the browser window")

	return startCmd
}

type localInfo struct {
	InstanceID  string `json:"instance_id"`
	GRPCPort    int    `json:"grpc_port"`
	InstallID   string `json:"install_id"`
	ProjectPath string `json:"project_path"`
	IsDev       bool   `json:"is_dev"`
}

type localServer struct {
	runtime   *runtime.Runtime
	logger    *zap.Logger
	httpPort  int
	grpcPort  int
	disableUI bool
	info      *localInfo
}

func (s *localServer) ListenAndServe() error {
	// Prepare errgroup and context with graceful shutdown
	gctx := graceful.WithCancelOnTerminate(context.Background())
	group, ctx := errgroup.WithContext(gctx)

	// Create a runtime server
	opts := &runtimeserver.Options{
		HTTPPort: s.httpPort,
		GRPCPort: s.grpcPort,
	}
	runtimeServer, err := runtimeserver.NewServer(opts, s.runtime, s.logger)
	if err != nil {
		return err
	}
	runtimeHandler, err := runtimeServer.HTTPHandler(ctx) // NOTE: This context is used for grpc-gateway connection
	if err != nil {
		return err
	}

	// Create a single HTTP handler for both the local UI, local backend endpoints, and local runtime
	mux := http.NewServeMux()
	if !s.disableUI {
		mux.Handle("/", web.StaticHandler())
	}
	mux.Handle("/v1/", runtimeHandler)
	mux.Handle("/local/config", http.HandlerFunc(s.infoHandler))
	mux.Handle("/local/track", http.HandlerFunc(s.trackingHandler))
	mux.Handle("/local/health", http.HandlerFunc(s.healthHandler))

	// Start the gRPC server
	group.Go(func() error {
		return runtimeServer.ServeGRPC(ctx)
	})

	// Start the local HTTP server
	group.Go(func() error {
		server := &http.Server{Handler: cors(mux)}
		return graceful.ServeHTTP(ctx, server, s.httpPort)
	})

	return group.Wait()
}

// infoHandler servers the local info struct
func (s *localServer) infoHandler(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(s.info)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

// trackingHandler proxies events to intake.rilldata.io
func (s *localServer) trackingHandler(w http.ResponseWriter, r *http.Request) {
	// Proxy request to rill intake
	proxyReq, err := http.NewRequest(r.Method, "https://intake.rilldata.io/events/data-modeler-metrics", r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	// Copy the auth header
	proxyReq.Header = http.Header{
		"Authorization": r.Header["Authorization"],
	}

	// Send proxied request
	resp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
}

// healthHandler is a basic health check
func (s *localServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

// Fully open CORS policy. This is very much local-only.
// TODO: Adapt before recommending hosting Rill using the local server.
func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				w.Header().Set("Access-Control-Allow-Headers", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, PUT, PATCH, DELETE")
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}
