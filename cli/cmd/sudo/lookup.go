package sudo

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func lookupCmd(cfg *config.Config) *cobra.Command {
	lookupCmd := &cobra.Command{
		Use:   "lookup {user|org|project|deployment|instance} <id>",
		Short: "Lookup resource by ID",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := cmdutil.Client(cfg)
			if err != nil {
				return err
			}
			defer client.Close()

			switch args[0] {
			case "user":
				return getUser(ctx, client, args[1])
			case "org":
				return getOrganization(ctx, client, args[1])
			case "project":
				return getProject(ctx, client, args[1])
			case "deployment":
				// TODO
				return nil
			case "instance":
				// TODO
				return nil
			default:
				return fmt.Errorf("invalid resource type %q", args[0])
			}
		},
	}

	return lookupCmd
}

func getOrganization(ctx context.Context, c *client.Client, orgID string) error {
	res, err := c.SudoGetResource(ctx, &adminv1.SudoGetResourceRequest{
		Id: &adminv1.SudoGetResourceRequest_OrgId{OrgId: orgID},
	})
	if err != nil {
		return err
	}

	org := res.GetOrg()
	fmt.Printf("Name: %s\n", org.Name)
	fmt.Printf("Created on: %s\n", org.CreatedOn.AsTime().Format(time.RFC3339Nano))

	return nil
}

func getUser(ctx context.Context, c *client.Client, userID string) error {
	res, err := c.SudoGetResource(ctx, &adminv1.SudoGetResourceRequest{
		Id: &adminv1.SudoGetResourceRequest_UserId{UserId: userID},
	})
	if err != nil {
		return err
	}

	user := res.GetUser()
	fmt.Printf("Email: %s\n", user.Email)
	fmt.Printf("Name: %s\n", user.DisplayName)
	fmt.Printf("Created on: %s\n", user.CreatedOn.AsTime().Format(time.RFC3339Nano))

	return nil
}

func getProject(ctx context.Context, c *client.Client, projectID string) error {
	res, err := c.SudoGetResource(ctx, &adminv1.SudoGetResourceRequest{
		Id: &adminv1.SudoGetResourceRequest_ProjectId{ProjectId: projectID},
	})
	if err != nil {
		return err
	}

	project := res.GetProject()
	fmt.Printf("Name: %s\n", project.Name)
	fmt.Printf("Org: %s (ID: %s)\n", project.OrgName, project.OrgId)
	fmt.Printf("Created on: %s\n", project.CreatedOn)
	fmt.Printf("Public: %s", project.Public)
	fmt.Printf("Region: %s", project.Region)
	fmt.Printf("Github URL: %s", project.GithubUrl)
	fmt.Printf("Subpath: %s", project.Subpath)
	fmt.Printf("Prod branch: %s", project.ProdBranch)
	fmt.Printf("Prod OLAP driver: %s", project.ProdOlapDriver)
	fmt.Printf("Prod OLAP DSN: %s", project.ProdOlapDsn)
	fmt.Printf("Prod slots: %s", project.ProdSlots)
	fmt.Printf("Prod deployment ID: %s", project.ProdDeploymentId)

	return nil
}
