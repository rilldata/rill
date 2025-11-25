package project

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	runtimeclient "github.com/rilldata/rill/runtime/client"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func LogsCmd(ch *cmdutil.Helper) *cobra.Command {
	var name, path string
	var follow bool
	var tail int
	var level string

	logsCmd := &cobra.Command{
		Use:   "logs [<project-name>]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Show project logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if len(args) > 0 {
				name = args[0]
			}

			if !cmd.Flags().Changed("project") && len(args) == 0 && ch.Interactive {
				name, err = ch.InferProjectName(cmd.Context(), ch.Org, path)
				if err != nil {
					return fmt.Errorf("unable to infer project name (use `--project` to explicitly specify the name): %w", err)
				}
			}

			proj, err := client.GetProject(cmd.Context(), &adminv1.GetProjectRequest{
				Org:     ch.Org,
				Project: name,
			})
			if err != nil {
				return err
			}

			depl := proj.ProdDeployment
			if depl == nil {
				return fmt.Errorf("project %q is not currently deployed", name)
			}

			if depl.Status != adminv1.DeploymentStatus_DEPLOYMENT_STATUS_RUNNING {
				ch.PrintfWarn("Deployment status not RUNNING: %s\n", depl.Status.String())
				return nil
			}

			rt, err := runtimeclient.New(depl.RuntimeHost, proj.Jwt)
			if err != nil {
				return fmt.Errorf("failed to connect to runtime: %w", err)
			}

			lvl := toRuntimeLogLevel(level)
			if lvl == runtimev1.LogLevel_LOG_LEVEL_UNSPECIFIED {
				return fmt.Errorf("invalid log level: %s", level)
			}

			if follow {
				ctx, cancel := context.WithCancelCause(cmd.Context())
				defer cancel(nil)

				go func() {
					logs, err := rt.WatchLogs(ctx, &runtimev1.WatchLogsRequest{InstanceId: depl.RuntimeInstanceId, Replay: true, ReplayLimit: int32(tail), Level: lvl})
					if err != nil {
						cancel(fmt.Errorf("failed to watch logs: %w", err))
						return
					}

					for {
						res, err := logs.Recv()
						if err != nil {
							cancel(fmt.Errorf("failed to receive logs: %w", err))
							return
						}

						printLog(res.Log)
					}
				}()

				// keep on receiving logs util context is cancelled
				<-ctx.Done()
				err := context.Cause(ctx)
				if errors.Is(err, context.Canceled) {
					// Since user cancellation is expected for --follow, don't return an error if the cause was a context cancellation.
					return nil
				}
				return context.Cause(ctx)
			}

			res, err := rt.GetLogs(cmd.Context(),
				&runtimev1.GetLogsRequest{InstanceId: depl.RuntimeInstanceId, Ascending: true, Limit: int32(tail), Level: lvl},
				grpc.MaxCallRecvMsgSize(17*1024*1024)) // setting grpc received message size to 16MB(default max logbuffer size)+1MB(buffer).
			if err != nil {
				return fmt.Errorf("failed to get logs: %w", err)
			}

			for _, log := range res.Logs {
				printLog(log)
			}
			return nil
		},
	}

	logsCmd.Flags().SortFlags = false
	logsCmd.Flags().StringVar(&name, "project", "", "Project Name")
	logsCmd.Flags().StringVar(&path, "path", ".", "Project directory")
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow logs")
	logsCmd.Flags().IntVarP(&tail, "tail", "t", -1, "Number of lines to show from the end of the logs, use -1 for all logs")
	logsCmd.Flags().StringVar(&level, "level", "INFO", "Minimum log level to show (DEBUG, INFO, WARN, ERROR, FATAL)")

	return logsCmd
}

func printLog(log *runtimev1.Log) {
	fmt.Printf("%s\t%s\t%s\t%s\n", printTime(log.Time), printLogLevel(log.Level), log.Message, log.JsonPayload)
}

func printTime(t *timestamppb.Timestamp) string {
	return t.AsTime().Format("2006-01-02T15:04:05.000000")
}

func printLogLevel(logLevel runtimev1.LogLevel) string {
	switch logLevel {
	case runtimev1.LogLevel_LOG_LEVEL_DEBUG:
		return "DEBUG"
	case runtimev1.LogLevel_LOG_LEVEL_INFO:
		return "INFO"
	case runtimev1.LogLevel_LOG_LEVEL_WARN:
		return "WARN"
	case runtimev1.LogLevel_LOG_LEVEL_ERROR:
		return "ERROR"
	case runtimev1.LogLevel_LOG_LEVEL_FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

func toRuntimeLogLevel(lvl string) runtimev1.LogLevel {
	switch strings.ToUpper(lvl) {
	case "DEBUG":
		return runtimev1.LogLevel_LOG_LEVEL_DEBUG
	case "INFO":
		return runtimev1.LogLevel_LOG_LEVEL_INFO
	case "WARN":
		return runtimev1.LogLevel_LOG_LEVEL_WARN
	case "ERROR":
		return runtimev1.LogLevel_LOG_LEVEL_ERROR
	case "FATAL":
		return runtimev1.LogLevel_LOG_LEVEL_FATAL
	default:
		return runtimev1.LogLevel_LOG_LEVEL_UNSPECIFIED
	}
}
