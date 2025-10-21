package user

import (
	"encoding/json"
	"fmt"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

func ShowCmd(ch *cmdutil.Helper) *cobra.Command {
	var email string

	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show detailed information about a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := ch.Client()
			if err != nil {
				return err
			}

			if email == "" {
				return fmt.Errorf("email is required")
			}

			member, err := client.GetOrganizationMemberUser(cmd.Context(), &adminv1.GetOrganizationMemberUserRequest{
				Org:   ch.Org,
				Email: email,
			})
			if err != nil {
				return err
			}

			ch.PrintfSuccess("User: %s (%s)\n", member.Member.UserName, member.Member.UserEmail)
			ch.PrintfSuccess("Role: %s\n", member.Member.RoleName)
			ch.PrintfSuccess("Projects count: %d\n", member.Member.ProjectsCount)
			ch.PrintfSuccess("Usergroups count: %d\n", member.Member.UsergroupsCount)

			if member.Member.Attributes != nil && len(member.Member.Attributes.Fields) > 0 {
				ch.PrintfSuccess("Attributes:\n")
				for key, value := range member.Member.Attributes.Fields {
					var valueStr string
					switch v := value.Kind.(type) {
					case *structpb.Value_StringValue:
						valueStr = v.StringValue
					case *structpb.Value_NumberValue:
						valueStr = fmt.Sprintf("%.0f", v.NumberValue)
					case *structpb.Value_BoolValue:
						valueStr = fmt.Sprintf("%t", v.BoolValue)
					default:
						// For complex types, marshal to JSON
						jsonBytes, _ := json.Marshal(value)
						valueStr = string(jsonBytes)
					}
					ch.PrintfSuccess("  %s: %s\n", key, valueStr)
				}
			} else {
				ch.PrintfSuccess("No custom attributes set\n")
			}

			return nil
		},
	}

	showCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	showCmd.Flags().StringVar(&email, "email", "", "Email of the user (required)")

	return showCmd
}
