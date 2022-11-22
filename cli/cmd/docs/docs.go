package docs

import (
	"github.com/mattn/go-colorable"
	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var docsUrl = "https://docs.rilldata.com"

// docsCmd represents the docs command
func DocsCmd() *cobra.Command {
	var docsCmd = &cobra.Command{
		Use:   "docs",
		Short: "Show rill docs",
		Long:  `A longer description`,
		Run: func(cmd *cobra.Command, args []string) {
			// Create base logger
			config := zap.NewDevelopmentEncoderConfig()
			config.EncodeLevel = zapcore.CapitalColorLevelEncoder
			logger := zap.New(zapcore.NewCore(
				zapcore.NewConsoleEncoder(config),
				zapcore.AddSync(colorable.NewColorableStdout()),
				zapcore.DebugLevel,
			))

			err := browser.Open(docsUrl)
			if err != nil {
				logger.Sugar().Warnf("could not open browser error: %v, copy and paste this URL into your browser: %s", err, docsUrl)
			}
		},
	}
	return docsCmd
}
