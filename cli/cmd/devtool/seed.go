package devtool

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/rilldata/rill/cli/cmd/project"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func SeedCmd(ch *cmdutil.Helper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seed {cloud|e2e}",
		Short: "Authenticate and deploy a seed project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			preset := args[0]
			if preset != "cloud" && preset != "e2e" {
				return fmt.Errorf("seeding not available for preset %q", preset)
			}

			// clone examples to a temp dir and deploy
			temp, err := os.MkdirTemp("", "rill-seed-*")
			if err != nil {
				return err
			}
			defer os.RemoveAll(temp)
			_, err = git.PlainClone(temp, false, &git.CloneOptions{
				URL: "https://github.com/rilldata/rill-examples.git",
			})
			if err != nil {
				return err
			}

			// create org if not exists
			if ch.Org == "" {
				client, err := ch.Client()
				if err != nil {
					return err
				}
				_, err = client.CreateOrganization(cmd.Context(), &adminv1.CreateOrganizationRequest{
					Name: "rilldata",
				})
				if err != nil {
					return err
				}
			}
			ch.Interactive = false // disable interactive prompts for seeding
			return project.ConnectGithubFlow(cmd.Context(), ch, &project.DeployOpts{
				GitPath:     temp,
				SubPath:     "rill-openrtb-prog-ads",
				Name:        "rill-openrtb-prog-ads",
				RemoteName:  "origin",
				ProdVersion: "latest",
				Slots:       2,
				Github:      true,
			})
		},
	}

	return cmd
}
