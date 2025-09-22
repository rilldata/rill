package admin

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/billing/payment"
	"github.com/rilldata/rill/admin/jobs/river"
	"github.com/rilldata/rill/admin/server"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ai"
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
	"google.golang.org/api/option"

	// Register database and provisioner implementations
	_ "github.com/rilldata/rill/admin/database/postgres"
	_ "github.com/rilldata/rill/admin/provisioner/clickhousestatic"
	_ "github.com/rilldata/rill/admin/provisioner/kubernetes"
	_ "github.com/rilldata/rill/admin/provisioner/static"
)

// Config describes admin server config derived from environment variables.
// Env var keys must be prefixed with RILL_ADMIN_ and are converted from snake_case to CamelCase.
// For example RILL_ADMIN_HTTP_PORT is mapped to Config.HTTPPort.
type Config struct {
	DatabaseDriver string `default:"postgres" split_words:"true"`
	DatabaseURL    string `split_words:"true"`
	// json encoded array of database.EncryptionKey
	DatabaseEncryptionKeyring string                 `split_words:"true"`
	RiverDatabaseURL          string                 `split_words:"true"`
	RedisURL                  string                 `default:"" split_words:"true"`
	ProvisionerSetJSON        string                 `split_words:"true"`
	ProvisionerMaxConcurrency int                    `default:"30" split_words:"true"`
	DefaultProvisioner        string                 `split_words:"true"`
	Jobs                      []string               `split_words:"true"`
	LogLevel                  zapcore.Level          `default:"info" split_words:"true"`
	MetricsExporter           observability.Exporter `default:"prometheus" split_words:"true"`
	TracesExporter            observability.Exporter `default:"" split_words:"true"`
	HTTPPort                  int                    `default:"8080" split_words:"true"`
	GRPCPort                  int                    `default:"8080" split_words:"true"`
	DebugPort                 int                    `split_words:"true"`
	ExternalURL               string                 `default:"http://localhost:8080" split_words:"true"`
	ExternalGRPCURL           string                 `envconfig:"external_grpc_url"`
	FrontendURL               string                 `default:"http://localhost:3000" split_words:"true"`
	AllowedOrigins            []string               `default:"*" split_words:"true"`
	SessionKeyPairs           []string               `split_words:"true"`
	SigningJWKS               string                 `split_words:"true"`
	SigningKeyID              string                 `split_words:"true"`
	AuthDomain                string                 `split_words:"true"`
	AuthClientID              string                 `split_words:"true"`
	AuthClientSecret          string                 `split_words:"true"`
	GithubAppID               int64                  `split_words:"true"`
	GithubAppName             string                 `split_words:"true"`
	GithubAppPrivateKey       string                 `split_words:"true"`
	GithubAppWebhookSecret    string                 `split_words:"true"`
	GithubClientID            string                 `split_words:"true"`
	GithubClientSecret        string                 `split_words:"true"`
	GithubManagedAccount      string                 `split_words:"true"`
	AssetsBucket              string                 `split_words:"true"`
	// AssetsBucketGoogleCredentialsJSON is only required to be set for local development.
	// For production use cases the service account will be directly attached to pods which is the recommended way of setting credentials.
	AssetsBucketGoogleCredentialsJSON string `split_words:"true"`
	EmailSMTPHost                     string `split_words:"true"`
	EmailSMTPPort                     int    `split_words:"true"`
	EmailSMTPUsername                 string `split_words:"true"`
	EmailSMTPPassword                 string `split_words:"true"`
	EmailSenderEmail                  string `split_words:"true"`
	EmailSenderName                   string `split_words:"true"`
	EmailBCC                          string `split_words:"true"`
	OpenAIAPIKey                      string `envconfig:"openai_api_key"`
	ActivitySinkType                  string `default:"" split_words:"true"`
	ActivitySinkKafkaBrokers          string `default:"" split_words:"true"`
	ActivityUISinkKafkaTopic          string `default:"" split_words:"true"`
	MetricsProject                    string `default:"" split_words:"true"`
	AutoscalerCron                    string `default:"CRON_TZ=America/Los_Angeles 0 0 * * 1" split_words:"true"`
	ScaleDownConstraint               int    `default:"0" split_words:"true"`
	OrbAPIKey                         string `split_words:"true"`
	OrbWebhookSecret                  string `split_words:"true"`
	OrbIntegratedTaxProvider          string `default:"avalara" split_words:"true"`
	StripeAPIKey                      string `split_words:"true"`
	StripeWebhookSecret               string `split_words:"true"`
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
			cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
			logger, err := cfg.Build()
			if err != nil {
				fmt.Printf("error: failed to create logger: %s\n", err.Error())
				os.Exit(1)
			}

			// Let ExternalGRPCURL default to ExternalURL, unless ExternalURL is itself the default.
			if conf.ExternalGRPCURL == "" {
				conf.ExternalGRPCURL = conf.ExternalURL
			}

			// Validate frontend and external URLs
			_, err = url.Parse(conf.FrontendURL)
			if err != nil {
				logger.Fatal("invalid frontend URL", zap.Error(err))
			}
			_, err = url.Parse(conf.ExternalURL)
			if err != nil {
				logger.Fatal("invalid external URL", zap.Error(err))
			}
			_, err = url.Parse(conf.ExternalGRPCURL)
			if err != nil {
				logger.Fatal("invalid external grpc URL", zap.Error(err))
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
			gh, err := admin.NewGithub(cmd.Context(), conf.GithubAppID, conf.GithubAppPrivateKey, conf.GithubManagedAccount, logger)
			if err != nil {
				logger.Fatal("error creating github client", zap.Error(err))
			}

			// Init AI client
			var aiClient ai.Client
			if conf.OpenAIAPIKey != "" {
				aiClient, err = ai.NewOpenAI(conf.OpenAIAPIKey, nil)
				if err != nil {
					logger.Fatal("error creating OpenAI client", zap.Error(err))
				}
			} else {
				aiClient = ai.NewNoop()
			}

			// Init AssetsBucket handle
			var clientOpts []option.ClientOption
			if conf.AssetsBucketGoogleCredentialsJSON != "" {
				clientOpts = append(clientOpts, option.WithCredentialsJSON([]byte(conf.AssetsBucketGoogleCredentialsJSON)))
			}
			storageClient, err := storage.NewClient(cmd.Context(), clientOpts...)
			if err != nil {
				logger.Fatal("failed to create assets bucket handle", zap.Error(err))
			}
			assetsBucket := storageClient.Bucket(conf.AssetsBucket)

			// Parse metrics project name
			var metricsProjectOrg, metricsProjectName string
			if conf.MetricsProject != "" {
				parts := strings.Split(conf.MetricsProject, "/")
				if len(parts) != 2 {
					logger.Fatal("invalid metrics project slug", zap.String("name", conf.MetricsProject))
				}
				metricsProjectOrg = parts[0]
				metricsProjectName = parts[1]
			}

			var biller billing.Biller
			if conf.OrbAPIKey != "" {
				biller = billing.NewOrb(logger, conf.OrbAPIKey, conf.OrbWebhookSecret, strings.ToLower(conf.OrbIntegratedTaxProvider))
			} else {
				biller = billing.NewNoop()
			}

			var p payment.Provider
			if conf.StripeAPIKey != "" {
				p = payment.NewStripe(logger, conf.StripeAPIKey, conf.StripeWebhookSecret)
			} else {
				p = payment.NewNoop()
			}

			// Init admin service
			admOpts := &admin.Options{
				DatabaseDriver:            conf.DatabaseDriver,
				DatabaseDSN:               conf.DatabaseURL,
				DatabaseEncryptionKeyring: conf.DatabaseEncryptionKeyring,
				ExternalURL:               conf.ExternalGRPCURL, // NOTE: using gRPC url
				FrontendURL:               conf.FrontendURL,
				ProvisionerSetJSON:        conf.ProvisionerSetJSON,
				ProvisionerMaxConcurrency: conf.ProvisionerMaxConcurrency,
				DefaultProvisioner:        conf.DefaultProvisioner,
				Version:                   ch.Version,
				MetricsProjectOrg:         metricsProjectOrg,
				MetricsProjectName:        metricsProjectName,
				AutoscalerCron:            conf.AutoscalerCron,
				ScaleDownConstraint:       conf.ScaleDownConstraint,
			}
			adm, err := admin.New(cmd.Context(), admOpts, logger, issuer, emailClient, gh, aiClient, assetsBucket, biller, p)
			if err != nil {
				logger.Fatal("error creating service", zap.Error(err))
			}
			defer adm.Close()

			// Init river jobs client
			jobs, err := river.New(cmd.Context(), conf.RiverDatabaseURL, adm)
			if err != nil {
				logger.Fatal("error creating river jobs client", zap.Error(err))
			}
			defer jobs.Close(cmd.Context())

			// Set initialized jobs client on admin so jobs can be triggered from admin
			adm.Jobs = jobs

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
					GithubManagedAccount:   conf.GithubManagedAccount,
					AssetsBucket:           conf.AssetsBucket,
				})
				if err != nil {
					logger.Fatal("error creating server", zap.Error(err))
				}
				group.Go(func() error { return srv.ServeHTTP(cctx) })
				if conf.DebugPort != 0 {
					group.Go(func() error { return debugserver.ServeHTTP(cctx, conf.DebugPort) })
				}
			}

			// Init and run worker
			if runWorker || runJobs {
				if runWorker {
					group.Go(func() error { return jobs.Work(cctx) })
					if !runServer {
						// If we're not running the server, lets start a http server with /ping endpoint for health checks
						mux := http.NewServeMux()
						mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(http.StatusOK)
							_, err := w.Write([]byte("pong"))
							if err != nil {
								panic(err)
							}
						})
						group.Go(func() error {
							return graceful.ServeHTTP(cctx, mux, graceful.ServeOptions{Port: conf.HTTPPort})
						})
					}
				}

				if runJobs {
					for _, job := range conf.Jobs {
						job := job
						group.Go(func() error {
							_, err := jobs.EnqueueByKind(cmd.Context(), job)
							if err != nil {
								return err
							}
							return nil
						})
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
