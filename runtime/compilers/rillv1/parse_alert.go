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
	Watermark  string           `yaml:"watermark"` // options: "trigger_time", "inherit"
	Intervals  struct {
		Duration      string `yaml:"duration"`
		Limit         uint   `yaml:"limit"`
		CheckUnclosed bool   `yaml:"check_unclosed"`
	} `yaml:"intervals"`
	Timeout string `yaml:"timeout"`
	Query   struct {
		Name     string         `yaml:"name"`
		Args     map[string]any `yaml:"args"`
		ArgsJSON string         `yaml:"args_json"`
		For      struct {
			UserID     string         `yaml:"user_id"`
			UserEmail  string         `yaml:"user_email"`
			Attributes map[string]any `yaml:"attributes"`
		} `yaml:"for"`
	} `yaml:"query"`
	Email struct {
		Recipients    []string `yaml:"recipients"`
		OnPass        *bool    `yaml:"on_pass"`
		OnFail        *bool    `yaml:"on_fail"`
		OnError       *bool    `yaml:"on_error"`
		Renotify      *bool    `yaml:"renotify"`
		RenotifyAfter string   `yaml:"renotify_after"`
	} `yaml:"email"`
	Annotations map[string]string `yaml:"annotations"`
}

// parseAlert parses an alert definition and adds the resulting resource to p.Resources.
func (p *Parser) parseAlert(ctx context.Context, node *Node) error {
	// Parse YAML
	tmp := &AlertYAML{}
	if node.YAMLRaw != "" {
		// Can't use node.YAML because we want to set KnownFields for alerts
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

	// Parse watermark
	watermarkInherit := false
	if tmp.Watermark != "" {
		switch strings.ToLower(tmp.Watermark) {
		case "inherit":
			watermarkInherit = true
		case "trigger_time":
			// Do nothing
		default:
			return fmt.Errorf(`invalid value %q for property "watermark"`, tmp.Watermark)
		}
	}

	// Validate the interval duration as a standard ISO8601 duration (without Rill extensions)
	if tmp.Intervals.Duration != "" {
		err := validateISO8601(tmp.Intervals.Duration, true)
		if err != nil {
			return fmt.Errorf(`invalid value %q for property "intervals.duration"`, tmp.Intervals.Duration)
		}
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
	for _, email := range tmp.Email.Recipients {
		_, err := mail.ParseAddress(email)
		if err != nil {
			return fmt.Errorf("invalid recipient email address %q", email)
		}
	}

	// Validate email.renotify_after
	var emailRenotifyAfter time.Duration
	if tmp.Email.RenotifyAfter != "" {
		emailRenotifyAfter, err = parseDuration(tmp.Email.RenotifyAfter)
		if err != nil {
			return fmt.Errorf(`invalid value for property "email.renotify_after": %w`, err)
		}
	}

	// Track alert
	r, err := p.insertResource(ResourceKindAlert, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	r.AlertSpec.Title = tmp.Title
	if schedule != nil {
		r.AlertSpec.RefreshSchedule = schedule
	}
	r.AlertSpec.WatermarkInherit = watermarkInherit
	r.AlertSpec.IntervalsIsoDuration = tmp.Intervals.Duration
	r.AlertSpec.IntervalsLimit = int32(tmp.Intervals.Limit)
	r.AlertSpec.IntervalsCheckUnclosed = tmp.Intervals.CheckUnclosed
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
	r.AlertSpec.EmailRenotify = false

	// Override email notification defaults
	if tmp.Email.OnPass != nil {
		r.AlertSpec.EmailOnPass = *tmp.Email.OnPass
	}
	if tmp.Email.OnFail != nil {
		r.AlertSpec.EmailOnFail = *tmp.Email.OnFail
	}
	if tmp.Email.OnError != nil {
		r.AlertSpec.EmailOnError = *tmp.Email.OnError
	}
	if tmp.Email.Renotify != nil {
		r.AlertSpec.EmailRenotify = *tmp.Email.Renotify
		r.AlertSpec.EmailRenotifyAfterSeconds = uint32(emailRenotifyAfter.Seconds())
	}

	r.AlertSpec.Annotations = tmp.Annotations

	return nil
}
