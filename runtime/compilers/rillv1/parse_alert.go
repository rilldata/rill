package rillv1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
)

// AlertYAML is the raw structure of an Alert resource defined in YAML (does not include common fields)
type AlertYAML struct {
	commonYAML `yaml:",inline"` // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	Title      string           `yaml:"title"`
	Refresh    *ScheduleYAML    `yaml:"refresh"`
	Timeout    string           `yaml:"timeout"`
	Query      struct {
		Name     string         `yaml:"name"`
		Args     map[string]any `yaml:"args"`
		ArgsJSON string         `yaml:"args_json"`
		For      struct {
			UserID     string         `yaml:"user_id"`
			UserEmail  string         `yaml:"user_email"`
			Attributes map[string]any `yaml:"attributes"`
		} `yaml:"for"`
	} `yaml:"query"`
	Export struct {
		Format string `yaml:"format"`
		Limit  uint   `yaml:"limit"`
	} `yaml:"export"`
	Email struct {
		Recipients    []string `yaml:"recipients"`
		OnPass        *bool    `yaml:"on_pass"`
		OnFail        *bool    `yaml:"on_fail"`
		OnError       *bool    `yaml:"on_error"`
		SkipUnchanged *bool    `yaml:"skip_unchanged"`
	} `yaml:"email"`
	Annotations map[string]string `yaml:"annotations"`
}

// parseAlert parses an alert definition and adds the resulting resource to p.Resources.
func (p *Parser) parseAlert(ctx context.Context, node *Node) error {
	// Parse YAML
	tmp := &AlertYAML{}
	if node.YAMLRaw != "" {
		// Can't use node.YAML because we want to set KnownFields for reports
		dec := yaml.NewDecoder(strings.NewReader(node.YAMLRaw))
		dec.KnownFields(true)
		if err := dec.Decode(tmp); err != nil {
			return pathError{path: node.YAMLPath, err: newYAMLError(err)}
		}
	}

	// Validate SQL or connector isn't set
	if node.SQL != "" {
		return fmt.Errorf("alerts cannot have SQL")
	}
	if !node.ConnectorInferred && node.Connector != "" {
		return fmt.Errorf("alerts cannot have a connector")
	}

	// Parse refresh schedule
	schedule, err := parseScheduleYAML(tmp.Refresh)
	if err != nil {
		return err
	}

	// Parse timeout
	var timeout time.Duration
	if tmp.Timeout != "" {
		timeout, err = parseDuration(tmp.Timeout)
		if err != nil {
			return err
		}
	}

	// Query name
	if tmp.Query.Name == "" {
		return fmt.Errorf(`invalid value %q for property "query.name"`, tmp.Query.Name)
	}

	// Query args
	if tmp.Query.ArgsJSON != "" {
		// Validate JSON
		if !json.Valid([]byte(tmp.Query.ArgsJSON)) {
			return errors.New(`failed to parse "query.args_json" as JSON`)
		}
	} else {
		// Fall back to query.args if query.args_json is not set
		data, err := json.Marshal(tmp.Query.Args)
		if err != nil {
			return fmt.Errorf(`failed to serialize "query.args" to JSON: %w`, err)
		}
		tmp.Query.ArgsJSON = string(data)
	}
	if tmp.Query.ArgsJSON == "" {
		return errors.New(`missing query args (must set either "query.args" or "query.args_json")`)
	}

	// Query for: validate only one of user_id, user_email, or attributes is set
	n := 0
	var queryForUserID, queryForUserEmail string
	var queryForAttributes *structpb.Struct
	if tmp.Query.For.UserID != "" {
		n++
		queryForUserID = tmp.Query.For.UserID
	}
	if tmp.Query.For.UserEmail != "" {
		n++
		_, err := mail.ParseAddress(tmp.Query.For.UserEmail)
		if err != nil {
			return fmt.Errorf(`invalid value %q for property "query.for.user_email"`, tmp.Query.For.UserEmail)
		}
		queryForUserEmail = tmp.Query.For.UserEmail
	}
	if len(tmp.Query.For.Attributes) > 0 {
		n++
		queryForAttributes, err = structpb.NewStruct(tmp.Query.For.Attributes)
		if err != nil {
			return fmt.Errorf(`failed to serialize property "query.for.attributes": %w`, err)
		}
	}
	if n > 1 {
		return fmt.Errorf(`only one of "query.for.user_id", "query.for.user_email", or "query.for.attributes" may be set`)
	}

	// Validate recipients
	if len(tmp.Email.Recipients) == 0 {
		return fmt.Errorf(`missing required property "recipients"`)
	}
	for _, email := range tmp.Email.Recipients {
		_, err := mail.ParseAddress(email)
		if err != nil {
			return fmt.Errorf("invalid recipient email address %q", email)
		}
	}

	// Track report
	r, err := p.insertResource(ResourceKindAlert, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	r.AlertSpec.Title = tmp.Title
	if schedule != nil {
		r.AlertSpec.RefreshSchedule = schedule
	}
	if timeout != 0 {
		r.AlertSpec.TimeoutSeconds = uint32(timeout.Seconds())
	}
	r.AlertSpec.QueryName = tmp.Query.Name
	r.AlertSpec.QueryArgsJson = tmp.Query.ArgsJSON

	// Note: have already validated that at most one of the cases match
	if queryForUserID != "" {
		r.AlertSpec.QueryFor = &runtimev1.AlertSpec_QueryForUserId{QueryForUserId: queryForUserID}
	} else if queryForUserEmail != "" {
		r.AlertSpec.QueryFor = &runtimev1.AlertSpec_QueryForUserEmail{QueryForUserEmail: queryForUserEmail}
	} else if queryForAttributes != nil {
		r.AlertSpec.QueryFor = &runtimev1.AlertSpec_QueryForAttributes{QueryForAttributes: queryForAttributes}
	}

	r.AlertSpec.EmailRecipients = tmp.Email.Recipients

	// Email notification default settings
	r.AlertSpec.EmailOnPass = false
	r.AlertSpec.EmailOnFail = true
	r.AlertSpec.EmailOnError = false
	r.AlertSpec.EmailSkipUnchanged = true
	if tmp.Email.OnPass != nil {
		r.AlertSpec.EmailOnPass = *tmp.Email.OnPass
	}
	if tmp.Email.OnFail != nil {
		r.AlertSpec.EmailOnFail = *tmp.Email.OnFail
	}
	if tmp.Email.OnError != nil {
		r.AlertSpec.EmailOnError = *tmp.Email.OnError
	}
	if tmp.Email.SkipUnchanged != nil {
		r.AlertSpec.EmailSkipUnchanged = *tmp.Email.SkipUnchanged
	}

	r.AlertSpec.Annotations = tmp.Annotations

	return nil
}
