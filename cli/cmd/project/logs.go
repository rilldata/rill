package project

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	runtimeclient "github.com/rilldata/rill/runtime/client"
	"github.com/spf13/cobra"
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
			cfg := ch.Config
			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			if len(args) > 0 {
				name = args[0]
			}

			if !cmd.Flags().Changed("project") && len(args) == 0 && cfg.Interactive {
				name, err = inferProjectName(cmd.Context(), client, cfg.Org, path)
				if err != nil {
					return err
				}
			}

			proj, err := client.GetProject(context.Background(), &adminv1.GetProjectRequest{
				OrganizationName: cfg.Org,
				Name:             name,
			})
			if err != nil {
				return err
			}

			depl := proj.ProdDeployment
			if depl == nil {
				return fmt.Errorf("project %q is not currently deployed", name)
			}

			if depl.Status != adminv1.DeploymentStatus_DEPLOYMENT_STATUS_OK {
				ch.Printer.PrintlnWarn(fmt.Sprintf("Deployment status not OK: %s", depl.Status.String()))
				return nil
			}

			rt, err := runtimeclient.New(depl.RuntimeHost, proj.Jwt)
			if err != nil {
				return fmt.Errorf("failed to connect to runtime: %w", err)
			}

			if follow {
				logClient, err := rt.WatchLogs(context.Background(), &runtimev1.WatchLogsRequest{InstanceId: depl.RuntimeInstanceId, Replay: true, ReplayLimit: int32(tail), Level: level})
				if err != nil {
					return fmt.Errorf("failed to watch logs: %w", err)
				}

				ctx, cancel := context.WithCancel(cmd.Context())

				go func() {
					for {
						res, err := logClient.Recv()
						if err != nil {
							fmt.Println("failed to receive logs: %w", err)
							cancel()
							break
						}

						printLog(res.Log)
					}
				}()

				// keep on receiving logs util context is cancelled
				<-ctx.Done()
				return nil
			}

			res, err := rt.GetLogs(context.Background(), &runtimev1.GetLogsRequest{InstanceId: depl.RuntimeInstanceId, Ascending: true, Limit: int32(tail), Level: level})
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
