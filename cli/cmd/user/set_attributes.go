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
	var email, attributes, attributesJSON string

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
	setAttributesCmd.Flags().StringVar(&attributes, "attributes", "", "Comma-separated attributes in key=value format (--attributes app=foo,dept=bar)")
	setAttributesCmd.Flags().StringVar(&attributesJSON, "json", "", "Attributes as JSON object (--json '{\"app\":\"foo\",\"dept\":\"bar\"}')")

	return setAttributesCmd
}

func parseAttributes(attributes, attributesJSON string) (map[string]interface{}, error) {
	attrs := make(map[string]interface{})

	switch {
	case attributesJSON != "" && attributes != "":
		return nil, fmt.Errorf("cannot use both --attributes and --json flags")
	case attributesJSON != "":
		if err := json.Unmarshal([]byte(attributesJSON), &attrs); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
	case attributes != "":
		for _, pair := range strings.Split(attributes, ",") {
			if pair = strings.TrimSpace(pair); pair == "" {
				continue
			}

			parts := strings.SplitN(pair, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid attribute format '%s'. Use key=value format", pair)
			}

			key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			if key == "" {
				return nil, fmt.Errorf("attribute key cannot be empty")
			}

			attrs[key] = parseValue(value)
		}
	default:
		return nil, fmt.Errorf("must provide either --attributes or --json flag")
	}

	return attrs, nil
}

func parseValue(value string) interface{} {
	switch {
	case value == "true":
		return true
	case value == "false":
		return false
	case !strings.Contains(value, "."):
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		}
	default:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return value
}
