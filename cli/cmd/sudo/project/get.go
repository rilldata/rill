package project

import (
	"fmt"
	"strings"
	"time"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func GetCmd(ch *cmdutil.Helper) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get <org> <project>",
		Args:  cobra.ExactArgs(2),
		Short: "Get project details",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client, err := ch.Client()
			if err != nil {
				return err
			}
			res, err := client.GetProject(ctx, &adminv1.GetProjectRequest{
				OrganizationName: args[0],
				Name:             args[1],
			})
			if err != nil {
				return err
			}

			annotations := make([]string, 0, len(res.Project.Annotations))
			for k, v := range res.Project.Annotations {
				annotations = append(annotations, fmt.Sprintf("%s=%s", k, v))
			}

			project := res.Project
			fmt.Printf("Name: %s (ID: %s)\n", project.Name, project.Id)
			fmt.Printf("Org: %s (ID: %s)\n", project.OrgName, project.OrgId)
			fmt.Printf("Created on: %s\n", project.CreatedOn.AsTime().Format(time.RFC3339Nano))
			fmt.Printf("Public: %t\n", project.Public)
			fmt.Printf("Created by user ID: %s\n", project.CreatedByUserId)
			fmt.Printf("Provisioner: %s\n", project.Provisioner)
			fmt.Printf("Github URL: %s\n", project.GithubUrl)
			fmt.Printf("Subpath: %s\n", project.Subpath)
			fmt.Printf("Prod version: %s\n", project.ProdVersion)
			fmt.Printf("Prod branch: %s\n", project.ProdBranch)
			fmt.Printf("Prod OLAP driver: %s\n", project.ProdOlapDriver)
			fmt.Printf("Prod OLAP DSN: %s\n", project.ProdOlapDsn)
			fmt.Printf("Prod slots: %d\n", project.ProdSlots)
			fmt.Printf("Prod deployment ID: %s\n", project.ProdDeploymentId)
			fmt.Printf("Prod hibernation TTL: %s\n", time.Duration(project.ProdTtlSeconds)*time.Second)
			fmt.Printf("Annotations: %s\n", strings.Join(annotations, "; "))

			return nil
		},
	}

	return getCmd
}
