package user

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
)

func SetAttributesCmd(ch *cmdutil.Helper) *cobra.Command {
	var email, attributesJSON string
	var attributes map[string]string
	var force bool

	setAttributesCmd := &cobra.Command{
		Use:   "set-attributes",
		Short: "Set custom attributes for a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			if email == "" {
				return fmt.Errorf("email is required")
			}

			attrs, err := parseAttributes(attributes, attributesJSON)
			if err != nil {
				return err
			}

			client, err := ch.Client()
			if err != nil {
				return err
			}

			// Check for existing attributes unless force flag is set
			if !force {
				existingUser, err := client.GetOrganizationMemberUser(cmd.Context(), &adminv1.GetOrganizationMemberUserRequest{
					Org:   ch.Org,
					Email: email,
				})
				if err != nil {
					return fmt.Errorf("failed to get existing user: %w", err)
				}

				if existingUser.Member.Attributes != nil && len(existingUser.Member.Attributes.Fields) > 0 {
					ch.PrintfWarn("User already has attributes. This will overwrite existing attributes.\n")
					ch.PrintfWarn("Current attributes:\n")
					for key, value := range existingUser.Member.Attributes.AsMap() {
						ch.Printf("  %s: %v\n", key, value)
					}
					ch.Printf("\nUse --force to proceed without this warning.\n")

					if !ch.Interactive {
						return fmt.Errorf("user already has attributes, use --force to overwrite")
					}

					confirmed, err := cmdutil.ConfirmPrompt("Do you want to overwrite the existing attributes?", "", false)
					if err != nil {
						return fmt.Errorf("failed to prompt for confirmation: %w", err)
					}
					if !confirmed {
						ch.Printf("Cancelled.\n")
						return nil
					}
				}
			}

			attributesStruct, err := structpb.NewStruct(attrs)
			if err != nil {
				return fmt.Errorf("failed to parse attributes: %w", err)
			}

			_, err = client.UpdateOrganizationMemberUserAttributes(cmd.Context(), &adminv1.UpdateOrganizationMemberUserAttributesRequest{
				Org:        ch.Org,
				Email:      email,
				Attributes: attributesStruct,
			})
			if err != nil {
				return fmt.Errorf("failed to update user attributes: %w", err)
			}

			ch.PrintfSuccess("Successfully updated attributes for user %s\n", email)
			return nil
		},
	}

	setAttributesCmd.Flags().StringVar(&ch.Org, "org", ch.Org, "Organization")
	setAttributesCmd.Flags().StringVar(&email, "email", "", "Email of the user (required)")
	setAttributesCmd.Flags().StringToStringVar(&attributes, "attribute", nil, "Attributes in key=value format (--attribute app=foo --attribute dept=bar)")
	setAttributesCmd.Flags().StringVar(&attributesJSON, "json", "", "Attributes as JSON object (--json '{\"app\":\"foo\",\"dept\":\"bar\"}')")
	setAttributesCmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt when overwriting existing attributes")

	return setAttributesCmd
}

func parseAttributes(attributes map[string]string, attributesJSON string) (map[string]interface{}, error) {
	attrs := make(map[string]interface{})

	switch {
	case attributesJSON != "" && len(attributes) > 0:
		return nil, fmt.Errorf("cannot use both --attributes and --json flags")
	case attributesJSON != "":
		if err := json.Unmarshal([]byte(attributesJSON), &attrs); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
	case len(attributes) > 0:
		for key, value := range attributes {
			attrs[key] = parseValue(value)
		}
	default:
		return nil, fmt.Errorf("must provide either --attributes or --json flag")
	}

	return attrs, nil
}

func parseValue(value string) interface{} {
	switch strings.ToLower(value) {
	case "true":
		return true
	case "false":
		return false
	}

	if i, err := strconv.ParseInt(value, 10, 64); err == nil {
		return i
	}

	if strings.Contains(value, ".") {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return value
}
