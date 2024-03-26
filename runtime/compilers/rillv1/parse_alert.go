package rillv1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/types/known/structpb"
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
	Notify struct {
		OnRecover     *bool  `yaml:"on_recover"`
		OnFail        *bool  `yaml:"on_fail"`
		OnError       *bool  `yaml:"on_error"`
		Renotify      *bool  `yaml:"renotify"`
		RenotifyAfter string `yaml:"renotify_after"`
		Email         struct {
			Recipients []string `yaml:"recipients"`
		} `yaml:"email"`
		Slack struct {
			Channels []string `yaml:"channels"`
			Emails   []string `yaml:"emails"`
			Webhooks []string `yaml:"webhooks"`
		} `yaml:"slack"`
	} `yaml:"notify"`
	Annotations map[string]string `yaml:"annotations"`
	// Backwards compatibility
	Email struct {
		Recipients    []string `yaml:"recipients"`
		OnRecover     *bool    `yaml:"on_recover"`
		OnFail        *bool    `yaml:"on_fail"`
		OnError       *bool    `yaml:"on_error"`
		Renotify      *bool    `yaml:"renotify"`
		RenotifyAfter string   `yaml:"renotify_after"`
	} `yaml:"email"`
}

// parseAlert parses an alert definition and adds the resulting resource to p.Resources.
func (p *Parser) parseAlert(node *Node) error {
	// Parse YAML
	tmp := &AlertYAML{}
	err := p.decodeNodeYAML(node, true, tmp)
	if err != nil {
		return err
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

	// Validate the interval duration as a standard ISO8601 duration (without Rill extensions) with only one component
	if tmp.Intervals.Duration != "" {
		err := validateISO8601(tmp.Intervals.Duration, true, true)
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

	if len(tmp.Email.Recipients) > 0 && len(tmp.Notify.Email.Recipients) > 0 {
		return errors.New(`cannot set both "email.recipients" and "notify.email.recipients"`)
	}

	isLegacySyntax := len(tmp.Email.Recipients) > 0

	// Validate the input
	var renotifyAfter time.Duration
	if isLegacySyntax {
		// Backwards compatibility
		// Validate email recipients
		for _, email := range tmp.Email.Recipients {
			_, err := mail.ParseAddress(email)
			if err != nil {
				return fmt.Errorf("invalid recipient email address %q", email)
			}
		}
		// Validate email.renotify_after
		if tmp.Email.RenotifyAfter != "" {
			renotifyAfter, err = parseDuration(tmp.Email.RenotifyAfter)
			if err != nil {
				return fmt.Errorf(`invalid value for property "email.renotify_after": %w`, err)
			}
		}
	} else {
		// Validate email recipients
		for _, email := range tmp.Notify.Email.Recipients {
			_, err := mail.ParseAddress(email)
			if err != nil {
				return fmt.Errorf("invalid recipient email address %q", email)
			}
		}
		// Validate notify.renotify_after
		if tmp.Notify.RenotifyAfter != "" {
			renotifyAfter, err = parseDuration(tmp.Notify.RenotifyAfter)
			if err != nil {
				return fmt.Errorf(`invalid value for property "notify.renotify_after": %w`, err)
			}
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

	// Notification default settings
	r.AlertSpec.NotifySpec = &runtimev1.AlertNotifySpec{}
	r.AlertSpec.NotifySpec.NotifyOnRecover = false
	r.AlertSpec.NotifySpec.NotifyOnFail = true
	r.AlertSpec.NotifySpec.NotifyOnError = false
	r.AlertSpec.NotifySpec.Renotify = false

	if isLegacySyntax {
		// Backwards compatibility
		// Override email notification defaults
		if tmp.Email.OnRecover != nil {
			r.AlertSpec.NotifySpec.NotifyOnRecover = *tmp.Email.OnRecover
		}
		if tmp.Email.OnFail != nil {
			r.AlertSpec.NotifySpec.NotifyOnFail = *tmp.Email.OnFail
		}
		if tmp.Email.OnError != nil {
			r.AlertSpec.NotifySpec.NotifyOnError = *tmp.Email.OnError
		}
		if tmp.Email.Renotify != nil {
			r.AlertSpec.NotifySpec.Renotify = *tmp.Email.Renotify
			r.AlertSpec.NotifySpec.RenotifyAfterSeconds = uint32(renotifyAfter.Seconds())
		}
		// Email settings
		r.AlertSpec.NotifySpec.Notifiers = []*runtimev1.NotifierSpec{
			{
				Connector: "email",
				Spec: &runtimev1.NotifierSpec_Email{
					Email: &runtimev1.EmailNotifierSpec{
						Recipients: tmp.Email.Recipients,
					},
				},
			},
		}
	} else {
		// Override notification defaults
		if tmp.Notify.OnRecover != nil {
			r.AlertSpec.NotifySpec.NotifyOnRecover = *tmp.Notify.OnRecover
		}
		if tmp.Notify.OnFail != nil {
			r.AlertSpec.NotifySpec.NotifyOnFail = *tmp.Notify.OnFail
		}
		if tmp.Notify.OnError != nil {
			r.AlertSpec.NotifySpec.NotifyOnError = *tmp.Notify.OnError
		}
		if tmp.Notify.Renotify != nil {
			r.AlertSpec.NotifySpec.Renotify = *tmp.Notify.Renotify
			r.AlertSpec.NotifySpec.RenotifyAfterSeconds = uint32(renotifyAfter.Seconds())
		}
		// Email settings
		if len(tmp.Notify.Email.Recipients) > 0 {
			r.AlertSpec.NotifySpec.Notifiers = append(r.AlertSpec.NotifySpec.Notifiers, &runtimev1.NotifierSpec{
				Connector: "email",
				Spec: &runtimev1.NotifierSpec_Email{
					Email: &runtimev1.EmailNotifierSpec{
						Recipients: tmp.Notify.Email.Recipients,
					},
				},
			})
		}
		// Slack settings
		if len(tmp.Notify.Slack.Channels) > 0 || len(tmp.Notify.Slack.Emails) > 0 || len(tmp.Notify.Slack.Webhooks) > 0 {
			r.AlertSpec.NotifySpec.Notifiers = append(r.AlertSpec.NotifySpec.Notifiers, &runtimev1.NotifierSpec{
				Connector: "slack",
				Spec: &runtimev1.NotifierSpec_Slack{
					Slack: &runtimev1.SlackNotifierSpec{
						Emails:   tmp.Notify.Slack.Emails,
						Channels: tmp.Notify.Slack.Channels,
						Webhooks: tmp.Notify.Slack.Webhooks,
					},
				},
			})
		}
	}

	r.AlertSpec.Annotations = tmp.Annotations

	return nil
}
