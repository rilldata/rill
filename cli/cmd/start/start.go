package start

import (
	"context"
	"log"
	"net/http"

	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/rilldata/rill/cli/pkg/web"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// StartCmd represents the start command
func StartCmd() *cobra.Command {
	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "A brief description of rill start",
		Long:  `A longer description.`,
		Run: func(cmd *cobra.Command, args []string) {
			var logger *zap.Logger
			url := "http://localhost:8080"

			ctx := graceful.WithCancelOnTerminate(context.Background())
			uiHandler, err := web.StaticHandler()
			if err != nil {
				logger.Error("failed to set up ui handler: %w", zap.Error(err))
			}

			err = browser.Open(url)
			if err != nil {
				log.Fatalf("Couldn't open browser: %v", err)
			}

			server := &http.Server{Handler: uiHandler}
			err = graceful.ServeHTTP(ctx, server, 8080)
			if err != nil {
				logger.Error("server crashed", zap.Error(err))
			}
		},
	}

	return startCmd
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
