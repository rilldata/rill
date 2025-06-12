package project

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

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
				OrganizationName: ch.Org,
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
				ch.PrintfWarn("Deployment status not OK: %s\n", depl.Status.String())
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
			level = printLogLevel(lvl)

			if follow {
				ctx, cancel := context.WithCancelCause(cmd.Context())
				defer cancel(nil)

				// Build the SSE URL for the logs endpoint using net/url
				baseURL := fmt.Sprintf("%s/v1/instances/%s/logs/watch", depl.RuntimeHost, depl.RuntimeInstanceId)
				u, err := url.Parse(baseURL)
				if err != nil {
					return fmt.Errorf("failed to parse logs base URL: %w", err)
				}
				q := u.Query()
				q.Set("stream", "logs")
				q.Set("replay", "true")
				q.Set("replay_limit", fmt.Sprintf("%d", tail))
				q.Set("level", level)
				u.RawQuery = q.Encode()
				logsURL := u.String()

				req, err := http.NewRequestWithContext(ctx, http.MethodGet, logsURL, http.NoBody)
				if err != nil {
					return fmt.Errorf("failed to create logs SSE request: %w", err)
				}
				// Add auth header if needed
				if proj.Jwt != "" {
					req.Header.Set("Authorization", "Bearer "+proj.Jwt)
				}

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return fmt.Errorf("failed to connect to logs SSE endpoint: %w", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					return fmt.Errorf("logs SSE endpoint returned status: %s", resp.Status)
				}

				dec := newSSEDecoder(resp.Body)
				for {
					select {
					case <-ctx.Done():
						if errors.Is(context.Cause(ctx), context.Canceled) {
							return nil
						}
						return context.Cause(ctx)
					default:
						event, err := dec.Decode()
						if err != nil {
							if errors.Is(err, io.EOF) {
								return nil
							}
							if errors.Is(err, context.Canceled) || errors.Is(context.Cause(ctx), context.Canceled) {
								return nil
							}
							if errors.Is(err, io.ErrClosedPipe) {
								return nil
							}
							if strings.Contains(err.Error(), "use of closed network connection") || strings.Contains(err.Error(), "connection reset by peer") {
								return nil
							}
							return fmt.Errorf("failed to decode SSE log event: %w", err)
						}
						if event != nil && event.Data != nil {
							var resp runtimev1.WatchLogsResponse
							err := json.Unmarshal(event.Data, &resp)
							if err == nil && resp.Log != nil {
								printLog(resp.Log)
								continue
							}
							fmt.Printf("%s\n", event.Data)
						}
					}
				}
			}

			res, err := rt.GetLogs(cmd.Context(), &runtimev1.GetLogsRequest{InstanceId: depl.RuntimeInstanceId, Ascending: true, Limit: int32(tail), Level: lvl})
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

// Replace sseEvent, sseDecoder, and newSSEDecoder with a line-oriented SSE parser

type sseEvent struct {
	Data json.RawMessage
}

type sseDecoder struct {
	r io.Reader
}

func newSSEDecoder(r io.Reader) *sseDecoder {
	return &sseDecoder{r: r}
}

// Decode reads the next SSE event, extracting the JSON payload from 'data: ' lines.
func (d *sseDecoder) Decode() (*sseEvent, error) {
	var dataLines []string
	buf := make([]byte, 4096)
	var lineBuf string
	for {
		n, err := d.r.Read(buf)
		if n > 0 {
			lineBuf += string(buf[:n])
			for {
				idx := strings.Index(lineBuf, "\n")
				if idx == -1 {
					break
				}
				line := lineBuf[:idx]
				lineBuf = lineBuf[idx+1:]
				line = strings.TrimRight(line, "\r")
				if strings.HasPrefix(line, "data: ") {
					dataLines = append(dataLines, line[len("data: "):])
				} else if line == "" && len(dataLines) > 0 {
					// End of event
					joined := strings.Join(dataLines, "\n")
					return &sseEvent{Data: json.RawMessage(joined)}, nil
				}
			}
		}
		if err != nil {
			if err == io.EOF && len(dataLines) > 0 {
				joined := strings.Join(dataLines, "\n")
				return &sseEvent{Data: json.RawMessage(joined)}, nil
			}
			return nil, err
		}
	}
}
