package deploy

import (
	"fmt"

	"github.com/rilldata/rill/cli/cmd/project"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/local"
	"github.com/spf13/cobra"
)

// DeployCmd is the guided tour for deploying rill projects to rill cloud.
func DeployCmd(ch *cmdutil.Helper) *cobra.Command {
	var httpPort, grpcPort int
	var allowedOrigins []string

	deployCmd := &cobra.Command{
		Use:   "deploy [<path>]",
		Short: "Deploy project to Rill Cloud",
		RunE: func(cmd *cobra.Command, args []string) error {
			var projectPath string
			if len(args) > 0 {
				projectPath = args[0]
			}

			_, _, err := project.ValidateLocalProject(cmd.Context(), ch, projectPath, "")
			if err != nil {
				return err
			}

			// Parse log format
			parsedLogFormat, ok := local.ParseLogFormat("console")
			if !ok {
				return fmt.Errorf("invalid log format 'console'")
			}

			localURL := fmt.Sprintf("http://localhost:%d", httpPort)

			allowedOrigins = append(allowedOrigins, localURL)

			ctx := cmd.Context()

			app, err := local.NewApp(ctx, &local.AppOptions{
				Ch:             ch,
				Environment:    "dev",
				OlapDriver:     local.DefaultOLAPDriver,
				OlapDSN:        local.DefaultOLAPDSN,
				ProjectPath:    projectPath,
				LogFormat:      parsedLogFormat,
				LocalURL:       localURL,
				AllowedOrigins: allowedOrigins,
			})
			if err != nil {
				return err
			}
			defer app.Close()

			userID, _ := ch.CurrentUserID(ctx)

			// open the `/deploy` page once the app is up
			go func() {
				app.PollServer(ctx, httpPort, false, false)
				uri := fmt.Sprintf("http://localhost:%d/deploy", httpPort)

				ch.PrintfBold("\nOpen this URL in your browser (if not automatically opened) to deploy the project: %s\n\n", uri)

				err := browser.Open(uri)
				if err != nil {
					app.Logger.Debugf("could not open browser: %v", err)
				}
			}()

			err = app.Serve(httpPort, grpcPort, true, false, true, userID, "", "")
			if err != nil {
				return fmt.Errorf("serve: %w", err)
			}

			return nil
		},
	}

	deployCmd.Flags().IntVar(&httpPort, "port", 9009, "Port for HTTP")
	deployCmd.Flags().IntVar(&grpcPort, "port-grpc", 49009, "Port for gRPC (internal)")
	deployCmd.Flags().StringSliceVarP(&allowedOrigins, "allowed-origins", "", []string{}, "Override allowed origins for CORS")

	return deployCmd
}
