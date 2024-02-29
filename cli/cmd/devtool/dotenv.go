package devtool

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func DotenvCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dotenv",
		Short: "Utilities for managing .env files",
	}

	cmd.AddCommand(DotenvRefreshCmd(ch))
	cmd.AddCommand(DotenvUploadCmd(ch))

	return cmd
}

func DotenvRefreshCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refresh {cloud}",
		Short: "Refresh .env file from shared storage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			preset := args[0]
			if preset != "cloud" {
				return fmt.Errorf(".env not used for preset %q", preset)
			}

			err := checkRillRepo(cmd.Context())
			if err != nil {
				return err
			}

			err = downloadDotenv(cmd.Context(), preset)
			if err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

func DotenvUploadCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload {cloud}",
		Short: "Distribute your current .env file to the team",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			preset := args[0]
			if preset != "cloud" {
				return fmt.Errorf(".env not used for preset %q", preset)
			}

			err := checkRillRepo(cmd.Context())
			if err != nil {
				return err
			}

			err = checkDotenv()
			if err != nil {
				return err
			}

			ch.PrintfWarn("This will overwrite the .env file in shared storage with the contents of your local .env file.\n")
			ch.PrintfWarn("The updated .env will automatically be used by other users of the devtool.\n")
			if !cmdutil.ConfirmPrompt("Do you want to continue?", "", false) {
				return nil
			}

			err = uploadDotenv(cmd.Context(), preset)
			if err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

var dotenvURLs = map[string]string{
	"cloud": "gs://rill-devtool/dotenv/cloud-dev.env",
}

func checkDotenv() error {
	_, err := os.Stat(".env")
	if err != nil {
		return fmt.Errorf(".env file not found at the root of the rill repository")
	}
	return nil
}

func downloadDotenv(ctx context.Context, preset string) error {
	err := exec.CommandContext(ctx, "gcloud", "storage", "cp", dotenvURLs[preset], ".env").Run()
	if err != nil {
		return fmt.Errorf("error syncing '.env' file from GCS (you must be a Rill team member and have authenticated `gcloud`): %w", err)
	}
	return nil
}

func uploadDotenv(ctx context.Context, preset string) error {
	return exec.CommandContext(ctx, "gcloud", "storage", "cp", ".env", dotenvURLs[preset]).Run()
}
