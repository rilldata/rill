package admin

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/ai"
	"github.com/rilldata/rill/admin/server"
	"github.com/rilldata/rill/admin/worker"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/debugserver"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"

	// Load database drivers for admin
	_ "github.com/rilldata/rill/admin/database/postgres"
)

// Config describes admin server config derived from environment variables.
// Env var keys must be prefixed with RILL_ADMIN_ and are converted from snake_case to CamelCase.
// For example RILL_ADMIN_HTTP_PORT is mapped to Config.HTTPPort.
type Config struct {
	DatabaseDriver           string                 `default:"postgres" split_words:"true"`
	DatabaseURL              string                 `split_words:"true"`
	RedisURL                 string                 `default:"" split_words:"true"`
	ProvisionerSetJSON       string                 `split_words:"true"`
	DefaultProvisioner       string                 `split_words:"true"`
	Jobs                     []string               `split_words:"true"`
	LogLevel                 zapcore.Level          `default:"info" split_words:"true"`
	MetricsExporter          observability.Exporter `default:"prometheus" split_words:"true"`
	TracesExporter           observability.Exporter `default:"" split_words:"true"`
	HTTPPort                 int                    `default:"8080" split_words:"true"`
	GRPCPort                 int                    `default:"9090" split_words:"true"`
	DebugPort                int                    `split_words:"true"`
	ExternalURL              string                 `default:"http://localhost:8080" split_words:"true"`
	ExternalGRPCURL          string                 `envconfig:"external_grpc_url"`
	FrontendURL              string                 `default:"http://localhost:3000" split_words:"true"`
	AllowedOrigins           []string               `default:"*" split_words:"true"`
	SessionKeyPairs          []string               `split_words:"true"`
	SigningJWKS              string                 `split_words:"true"`
	SigningKeyID             string                 `split_words:"true"`
	AuthDomain               string                 `split_words:"true"`
	AuthClientID             string                 `split_words:"true"`
	AuthClientSecret         string                 `split_words:"true"`
	GithubAppID              int64                  `split_words:"true"`
	GithubAppName            string                 `split_words:"true"`
	GithubAppPrivateKey      string                 `split_words:"true"`
	GithubAppWebhookSecret   string                 `split_words:"true"`
	GithubClientID           string                 `split_words:"true"`
	GithubClientSecret       string                 `split_words:"true"`
	EmailSMTPHost            string                 `split_words:"true"`
	EmailSMTPPort            int                    `split_words:"true"`
	EmailSMTPUsername        string                 `split_words:"true"`
	EmailSMTPPassword        string                 `split_words:"true"`
	EmailSenderEmail         string                 `split_words:"true"`
	EmailSenderName          string                 `split_words:"true"`
	EmailBCC                 string                 `split_words:"true"`
	OpenAIAPIKey             string                 `envconfig:"openai_api_key"`
	ActivitySinkType         string                 `default:"" split_words:"true"`
	ActivitySinkKafkaBrokers string                 `default:"" split_words:"true"`
	ActivityUISinkKafkaTopic string                 `default:"" split_words:"true"`
}

// StartCmd starts an admin server. It only allows configuration using environment variables.
func StartCmd(ch *cmdutil.Helper) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start [jobs|server|worker]",
		Short: "Start admin service",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Load .env (note: fails silently if .env has errors)
			_ = godotenv.Load()

			// Init config
			var conf Config
			err := envconfig.Process("rill_admin", &conf)
			if err != nil {
				fmt.Printf("failed to load config: %s\n", err.Error())
				os.Exit(1)
			}

			// Init logger
			cfg := zap.NewProductionConfig()
			cfg.Level.SetLevel(conf.LogLevel)
			logger, err := cfg.Build()
			if err != nil {
				fmt.Printf("error: failed to create logger: %s\n", err.Error())
				os.Exit(1)
			}

			// Let ExternalGRPCURL default to ExternalURL, unless ExternalURL is itself the default.
			// NOTE: This is temporary until we migrate to a server that can host HTTP and gRPC on the same port.
			if conf.ExternalGRPCURL == "" {
				if conf.ExternalURL == "http://localhost:8080" {
					conf.ExternalGRPCURL = "http://localhost:9090"
				} else {
					conf.ExternalGRPCURL = conf.ExternalURL
				}
			}

			// Validate frontend and external URLs
			_, err = url.Parse(conf.FrontendURL)
			if err != nil {
				fmt.Printf("error: invalid frontend URL: %s\n", err.Error())
				os.Exit(1)
			}
			_, err = url.Parse(conf.ExternalURL)
			if err != nil {
				fmt.Printf("error: invalid external URL: %s\n", err.Error())
				os.Exit(1)
			}
			_, err = url.Parse(conf.ExternalGRPCURL)
			if err != nil {
				fmt.Printf("error: invalid external grpc URL: %s\n", err.Error())
				os.Exit(1)
			}

			// Init observability
			shutdown, err := observability.Start(cmd.Context(), logger, &observability.Options{
				MetricsExporter: conf.MetricsExporter,
				TracesExporter:  conf.TracesExporter,
				ServiceName:     "admin-server",
				ServiceVersion:  ch.Version.String(),
			})
			if err != nil {
				logger.Fatal("error starting observability", zap.Error(err))
			}
			defer func() {
				// Allow 10 seconds to gracefully shutdown observability
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				err := shutdown(ctx)
				if err != nil {
					logger.Error("observability shutdown failed", zap.Error(err))
				}
			}()

			// Init activity client
			var activityClient *activity.Client
			switch conf.ActivitySinkType {
			case "", "noop":
				activityClient = activity.NewNoopClient()
			case "kafka":
				// NOTE: ActivityUISinkKafkaTopic specifically denotes a topic for UI events.
				// This is acceptable since the UI is presently the only source that records events on the admin server's telemetry.
				// However, if other events are emitted from the admin server in the future, we should refactor to emit all events of any kind to a single topic.
				// (And handle multiplexing of different event types downstream.)
				sink, err := activity.NewKafkaSink(conf.ActivitySinkKafkaBrokers, conf.ActivityUISinkKafkaTopic, logger)
				if err != nil {
					logger.Fatal("error creating kafka sink", zap.Error(err))
				}
				activityClient = activity.NewClient(sink, logger)
			default:
				logger.Fatal("unknown activity sink type", zap.String("type", conf.ActivitySinkType))
			}
			defer activityClient.Close(context.Background())

			// Add service info to activity client
			activityClient = activityClient.WithServiceName("admin-server")
			if ch.Version.Number != "" || ch.Version.Commit != "" {
				activityClient = activityClient.WithServiceVersion(ch.Version.Number, ch.Version.Commit)
			}
			if ch.Version.IsDev() {
				activityClient = activityClient.WithIsDev()
			}

			// Init runtime JWT issuer
			issuer, err := auth.NewIssuer(conf.ExternalURL, conf.SigningKeyID, []byte(conf.SigningJWKS))
			if err != nil {
				logger.Fatal("error creating runtime jwt issuer", zap.Error(err))
			}

			// Init email client
			var sender email.Sender
			if conf.EmailSMTPHost != "" {
				sender, err = email.NewSMTPSender(&email.SMTPOptions{
					SMTPHost:     conf.EmailSMTPHost,
					SMTPPort:     conf.EmailSMTPPort,
					SMTPUsername: conf.EmailSMTPUsername,
					SMTPPassword: conf.EmailSMTPPassword,
					FromEmail:    conf.EmailSenderEmail,
					FromName:     conf.EmailSenderName,
					BCC:          conf.EmailBCC,
				})
			} else {
				sender, err = email.NewConsoleSender(logger, conf.EmailSenderEmail, conf.EmailSenderName)
			}
			if err != nil {
				logger.Fatal("error creating email sender", zap.Error(err))
			}
			emailClient := email.New(sender)

			// Init github client
			gh, err := admin.NewGithub(conf.GithubAppID, conf.GithubAppPrivateKey)
			if err != nil {
				logger.Fatal("error creating github client", zap.Error(err))
			}

			// Init AI client
			var aiClient ai.Client
			if conf.OpenAIAPIKey != "" {
				aiClient, err = ai.NewOpenAI(conf.OpenAIAPIKey)
				if err != nil {
					logger.Fatal("error creating OpenAI client", zap.Error(err))
				}
			} else {
				aiClient = ai.NewNoop()
			}

			// Init admin service
			admOpts := &admin.Options{
				DatabaseDriver:     conf.DatabaseDriver,
				DatabaseDSN:        conf.DatabaseURL,
				ProvisionerSetJSON: conf.ProvisionerSetJSON,
				DefaultProvisioner: conf.DefaultProvisioner,
				ExternalURL:        conf.ExternalGRPCURL, // NOTE: using gRPC url
				VersionNumber:      ch.Version.Number,
			}
			adm, err := admin.New(cmd.Context(), admOpts, logger, issuer, emailClient, gh, aiClient)
			if err != nil {
				logger.Fatal("error creating service", zap.Error(err))
			}
			defer adm.Close()

			// Parse session keys as hex strings
			keyPairs := make([][]byte, len(conf.SessionKeyPairs))
			for idx, keyHex := range conf.SessionKeyPairs {
				key, err := hex.DecodeString(keyHex)
				if err != nil {
					logger.Fatal("failed to parse session key from hex string to bytes")
				}
				keyPairs[idx] = key
			}

			// Make errgroup for running the processes
			ctx := graceful.WithCancelOnTerminate(context.Background())
			group, cctx := errgroup.WithContext(ctx)

			// Determine services to run. If no service name was provided, run them all.
			// We just have three currently, so keeping this basic.
			runServer := len(args) == 0 || args[0] == "server"
			runWorker := len(args) == 0 || args[0] == "worker"
			runJobs := len(args) == 0 || args[0] == "jobs"

			// Init and run server
			if runServer {
				var limiter ratelimit.Limiter
				if conf.RedisURL == "" {
					limiter = ratelimit.NewNoop()
				} else {
					opts, err := redis.ParseURL(conf.RedisURL)
					if err != nil {
						logger.Fatal("failed to parse redis url", zap.Error(err))
					}
					limiter = ratelimit.NewRedis(redis.NewClient(opts))
				}
				srv, err := server.New(logger, adm, issuer, limiter, activityClient, &server.Options{
					HTTPPort:               conf.HTTPPort,
					GRPCPort:               conf.GRPCPort,
					ExternalURL:            conf.ExternalURL,
					FrontendURL:            conf.FrontendURL,
					AllowedOrigins:         conf.AllowedOrigins,
					SessionKeyPairs:        keyPairs,
					ServePrometheus:        conf.MetricsExporter == observability.PrometheusExporter,
					AuthDomain:             conf.AuthDomain,
					AuthClientID:           conf.AuthClientID,
					AuthClientSecret:       conf.AuthClientSecret,
					GithubAppName:          conf.GithubAppName,
					GithubAppWebhookSecret: conf.GithubAppWebhookSecret,
					GithubClientID:         conf.GithubClientID,
					GithubClientSecret:     conf.GithubClientSecret,
				})
				if err != nil {
					logger.Fatal("error creating server", zap.Error(err))
				}
				group.Go(func() error { return srv.ServeGRPC(cctx) })
				group.Go(func() error { return srv.ServeHTTP(cctx) })
				if conf.DebugPort != 0 {
					group.Go(func() error { return debugserver.ServeHTTP(cctx, conf.DebugPort) })
				}
			}

			// Init and run worker
			if runWorker || runJobs {
				wkr := worker.New(logger, adm)
				if runWorker {
					group.Go(func() error { return wkr.Run(cctx) })
					if !runServer {
						// If we're not running the server, lets start a http server with /ping endpoint for health checks
						group.Go(func() error { return worker.StartPingServer(cctx, conf.HTTPPort) })
					}
				}
				if runJobs {
					for _, job := range conf.Jobs {
						job := job
						group.Go(func() error { return wkr.RunJob(cctx, job) })
					}
				}
			}

			// Run tasks
			err = group.Wait()
			if err != nil {
				logger.Error("crashed", zap.Error(err))
				return
			}

			logger.Info("shutdown gracefully")
		},
	}
	return startCmd
}
