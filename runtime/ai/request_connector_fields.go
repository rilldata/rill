package ai

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

const RequestConnectorFieldsName = "request_connector_fields"

// connectorFieldNameRe matches typical top-level connector YAML keys (snake_case, ASCII).
var connectorFieldNameRe = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

type RequestConnectorFields struct {
	Runtime *runtime.Runtime
}

var _ Tool[*RequestConnectorFieldsArgs, *RequestConnectorFieldsResult] = (*RequestConnectorFields)(nil)

// RequestConnectorFieldsArgs is filled by the LLM. The runtime does not derive missing fields from JSON Schema;
// the client uses this payload to show connector forms (e.g. secrets to .env).
type RequestConnectorFieldsArgs struct {
	Driver        string         `json:"driver" jsonschema:"Connector driver name (e.g. clickhouse, s3, postgres)."`
	EnteredFields map[string]any `json:"entered_fields,omitempty" jsonschema:"Optional YAML property keys already provided by the user (e.g. username, password)."`
	MissingFields []string       `json:"missing_fields" jsonschema:"YAML property keys still needed from the user (e.g. username, password)."`
	Message       string         `json:"message,omitempty" jsonschema:"Optional short explanation for the end user."`
	ResourceName  string         `json:"resource_name,omitempty" jsonschema:"Optional connector resource name if known (filename stem without path)."`
}

// RequestConnectorFieldsResult echoes the handoff for the Rill UI and tests.
type RequestConnectorFieldsResult struct {
	Driver        string         `json:"driver"`
	EnteredFields map[string]any `json:"entered_fields,omitempty"`
	MissingFields []string       `json:"missing_fields"`
	Message       string         `json:"message,omitempty"`
	ResourceName  string         `json:"resource_name,omitempty"`
}

func (t *RequestConnectorFields) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        RequestConnectorFieldsName,
		Title:       "Request connector fields",
		Description: "UI handoff: request that the user supply specific connector YAML fields (often credentials). Use when the connector setup is incomplete or reconcile errors indicate missing auth or config. Pass driver and missing_fields you infer from instructions, the user message, and tool errors. The Rill UI may intercept this call to show forms and write secrets to .env.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(false),
			IdempotentHint:  true,
			OpenWorldHint:   boolPtr(false),
			ReadOnlyHint:    true,
		},
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Requesting connector fields...",
			"openai/toolInvocation/invoked":  "Connector fields requested",
		},
	}
}

func (t *RequestConnectorFields) CheckAccess(ctx context.Context) (bool, error) {
	return checkDeveloperAccess(ctx, t.Runtime, true)
}

func (t *RequestConnectorFields) Handler(ctx context.Context, args *RequestConnectorFieldsArgs) (*RequestConnectorFieldsResult, error) {
	driver := strings.TrimSpace(args.Driver)
	if driver == "" {
		return nil, fmt.Errorf("driver is required")
	}

	missing := make([]string, 0, len(args.MissingFields))
	seen := make(map[string]struct{})
	for _, raw := range args.MissingFields {
		f := strings.TrimSpace(raw)
		if f == "" {
			continue
		}
		if !connectorFieldNameRe.MatchString(f) {
			return nil, fmt.Errorf("invalid missing_fields entry %q: use snake_case keys like username or aws_access_key_id", raw)
		}
		if _, ok := seen[f]; ok {
			continue
		}
		seen[f] = struct{}{}
		missing = append(missing, f)
	}
	if len(missing) == 0 {
		return nil, fmt.Errorf("missing_fields must contain at least one non-empty field key")
	}

	return &RequestConnectorFieldsResult{
		Driver:        driver,
		EnteredFields: args.EnteredFields,
		MissingFields: missing,
		Message:       strings.TrimSpace(args.Message),
		ResourceName:  strings.TrimSpace(args.ResourceName),
	}, nil
}
