package sudo

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

func lookupCmd(cfg *config.Config) *cobra.Command {
	lookupCmd := &cobra.Command{
		Use:       "lookup [user|org|project] <id>",
		ValidArgs: []string{"user", "org", "project"},
		Args:      cobra.MatchAll(cobra.ExactArgs(2), validateFirstArg),
		Short:     "Lookup resource by id",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			var err error

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			switch args[0] {
			case "user":
				err = getUser(ctx, client, args[1])
			case "org":
				err = getOrganization(ctx, client, args[1])
			case "project":
				err = getProject(ctx, client, args[1])
			}
			return err
		},
	}

	return lookupCmd
}

func validateFirstArg(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("first argument is required")
	}

	firstArg := args[0]
	if !slices.Contains[string](cmd.ValidArgs, firstArg) {
		return fmt.Errorf("invalid first argument: %s", firstArg)
	}

	return nil
}

func getOrganization(ctx context.Context, c *client.Client, orgID string) error {
	res, err := c.SudoGetResource(ctx, &adminv1.SudoGetResourceRequest{
		Id: &adminv1.SudoGetResourceRequest_OrgId{OrgId: orgID},
	})
	if err != nil {
		return err
	}

	cmdutil.PrintlnSuccess("Organization info\n")
	fmt.Printf("  Org: %s\n", res.GetOrg())

	return nil
}

func getUser(ctx context.Context, c *client.Client, userID string) error {
	res, err := c.SudoGetResource(ctx, &adminv1.SudoGetResourceRequest{
		Id: &adminv1.SudoGetResourceRequest_UserId{UserId: userID},
	})
	if err != nil {
		return err
	}

	cmdutil.PrintlnSuccess("User info\n")
	fmt.Printf("  User: %s\n", res.GetUser())

	return nil
}

func getProject(ctx context.Context, c *client.Client, projectID string) error {
	res, err := c.SudoGetResource(ctx, &adminv1.SudoGetResourceRequest{
		Id: &adminv1.SudoGetResourceRequest_ProjectId{ProjectId: projectID},
	})
	if err != nil {
		return err
	}

	cmdutil.PrintlnSuccess("Project info\n")
	fmt.Printf("  Project: %s\n", res.GetProject())

	return nil
}
