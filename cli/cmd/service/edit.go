package service

import (
	"encoding/json"
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

func EditCmd(ch *cmdutil.Helper) *cobra.Command {
	var newName string
	var attributes string

	editCmd := &cobra.Command{
		Use:   "edit <service-name>",
		Args:  cobra.ExactArgs(1),
		Short: "edit service properties",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			req := &adminv1.UpdateServiceRequest{
				Name: args[0],
				Org:  ch.Org,
			}

			if newName != "" {
				req.NewName = &newName
			}

			if cmd.Flags().Changed("attributes") {
				if attributes == "" {
					attributes = "{}" // Default to empty JSON object if not provided
				}
				var attrs map[string]any
				if err = json.Unmarshal([]byte(attributes), &attrs); err != nil {
					return fmt.Errorf("failed to parse --attributes as JSON: %w", err)
				}
				req.Attributes, err = structpb.NewStruct(attrs)
				if err != nil {
					return fmt.Errorf("failed to convert attributes to struct: %w", err)
				}
			}

			res, err := client.UpdateService(cmd.Context(), req)
			if err != nil {
				return err
			}

			ch.PrintfSuccess("Updated service\n")
			ch.PrintServices([]*adminv1.Service{res.Service})

			return nil
		},
	}
	editCmd.Flags().SortFlags = false
	editCmd.Flags().StringVar(&newName, "new-name", "", "New service name")
	editCmd.Flags().StringVar(&attributes, "attributes", "", "JSON object of key-value pairs for service attributes")

	return editCmd
}
