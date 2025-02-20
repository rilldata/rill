package rillv1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers/slack"
	"github.com/rilldata/rill/runtime/pkg/duration"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"google.golang.org/protobuf/types/known/structpb"
)

// AlertYAML is the raw structure of an Alert resource defined in YAML (does not include common fields)
type AlertYAML struct {
	commonYAML  `yaml:",inline"` // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	DisplayName string           `yaml:"display_name"`
	Title       string           `yaml:"title"` // Deprecated: use display_name
	Refresh     *ScheduleYAML    `yaml:"refresh"`
	Watermark   string           `yaml:"watermark"` // options: "trigger_time", "inherit"
	Intervals   struct {
		Duration      string `yaml:"duration"`
		Limit         uint   `yaml:"limit"`
		CheckUnclosed bool   `yaml:"check_unclosed"`
	} `yaml:"intervals"`
	Timeout string    `yaml:"timeout"`
	Data    *DataYAML `yaml:"data"`
	For     struct {
		UserID     string         `yaml:"user_id"`
		UserEmail  string         `yaml:"user_email"`
		Attributes map[string]any `yaml:"attributes"`
	} `yaml:"for"`
	Query struct { // legacy query
		Name     string         `yaml:"name"`
		Args     map[string]any `yaml:"args"`
		ArgsJSON string         `yaml:"args_json"`
		For      struct {
			UserID     string         `yaml:"user_id"`
			UserEmail  string         `yaml:"user_email"`
			Attributes map[string]any `yaml:"attributes"`
		} `yaml:"for"`
	} `yaml:"query"`
	OnRecover     *bool  `yaml:"on_recover"`
	OnFail        *bool  `yaml:"on_fail"`
	OnError       *bool  `yaml:"on_error"`
	Renotify      *bool  `yaml:"renotify"`
	RenotifyAfter string `yaml:"renotify_after"`
	Notify        struct {
		Email struct {
			Recipients []string `yaml:"recipients"`
		} `yaml:"email"`
		Slack struct {
			Users    []string `yaml:"users"`
			Channels []string `yaml:"channels"`
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

	// Display name backwards compatibility
	if tmp.Title != "" && tmp.DisplayName == "" {
		tmp.DisplayName = tmp.Title
	}

	// Parse refresh schedule
	schedule, err := p.parseScheduleYAML(tmp.Refresh)
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
		err := duration.ValidateISO8601(tmp.Intervals.Duration, true, true)
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

	// Data and query
	var resolver string
	var resolverProps *structpb.Struct
	var queryForUserID, queryForUserEmail string
	var queryForAttributes *structpb.Struct
	isLegacyQuery := tmp.Data == nil

	if !isLegacyQuery {
		var refs []ResourceName
		resolver, resolverProps, refs, err = p.parseDataYAML(tmp.Data, node.Connector)
		if err != nil {
			return fmt.Errorf(`failed to parse "data": %w`, err)
		}
		node.Refs = append(node.Refs, refs...)

		// Query for: validate only one of user_id, user_email, or attributes is set
		n := 0
		if tmp.For.UserID != "" {
			n++
			queryForUserID = tmp.For.UserID
		}
		if tmp.For.UserEmail != "" {
			n++
			_, err := mail.ParseAddress(tmp.For.UserEmail)
			if err != nil {
				return fmt.Errorf(`invalid value %q for property "for.user_email"`, tmp.For.UserEmail)
			}
			queryForUserEmail = tmp.For.UserEmail
		}
		if len(tmp.For.Attributes) > 0 {
			n++
			queryForAttributes, err = structpb.NewStruct(tmp.For.Attributes)
			if err != nil {
				return fmt.Errorf(`failed to serialize property "for.attributes": %w`, err)
			}
		}
		if n > 1 {
			return fmt.Errorf(`only one of "for.user_id", "for.user_email", or "for.attributes" may be set`)
		}
	} else {
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

		resolver = "legacy_metrics"
		props := map[string]any{
			"query_name":      tmp.Query.Name,
			"query_args_json": tmp.Query.ArgsJSON,
		}
		resolverProps, err = structpb.NewStruct(props)
		if err != nil {
			return fmt.Errorf("encountered invalid property type: %w", err)
		}
	}

	if len(tmp.Email.Recipients) > 0 && len(tmp.Notify.Email.Recipients) > 0 {
		return errors.New(`cannot set both "email.recipients" and "notify.email.recipients"`)
	}

	isLegacyNotify := len(tmp.Email.Recipients) > 0

	// Validate the input
	var renotifyAfter time.Duration
	if isLegacyNotify {
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
		// Validate renotify_after
		if tmp.RenotifyAfter != "" {
			renotifyAfter, err = parseDuration(tmp.RenotifyAfter)
			if err != nil {
				return fmt.Errorf(`invalid value for property "renotify_after": %w`, err)
			}
		}
	}

	// Track alert
	r, err := p.insertResource(ResourceKindAlert, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	r.AlertSpec.DisplayName = tmp.DisplayName
	if r.AlertSpec.DisplayName == "" {
		r.AlertSpec.DisplayName = ToDisplayName(node.Name)
	}
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

	r.AlertSpec.Resolver = resolver
	r.AlertSpec.ResolverProperties = resolverProps

	// Note: have already validated that at most one of the cases match
	if queryForUserID != "" {
		r.AlertSpec.QueryFor = &runtimev1.AlertSpec_QueryForUserId{QueryForUserId: queryForUserID}
	} else if queryForUserEmail != "" {
		r.AlertSpec.QueryFor = &runtimev1.AlertSpec_QueryForUserEmail{QueryForUserEmail: queryForUserEmail}
	} else if queryForAttributes != nil {
		r.AlertSpec.QueryFor = &runtimev1.AlertSpec_QueryForAttributes{QueryForAttributes: queryForAttributes}
	}

	// Notification default settings
	r.AlertSpec.NotifyOnRecover = false
	r.AlertSpec.NotifyOnFail = true
	r.AlertSpec.NotifyOnError = false
	r.AlertSpec.Renotify = false

	if isLegacyNotify {
		// Backwards compatibility
		// Override email notification defaults
		if tmp.Email.OnRecover != nil {
			r.AlertSpec.NotifyOnRecover = *tmp.Email.OnRecover
		}
		if tmp.Email.OnFail != nil {
			r.AlertSpec.NotifyOnFail = *tmp.Email.OnFail
		}
		if tmp.Email.OnError != nil {
			r.AlertSpec.NotifyOnError = *tmp.Email.OnError
		}
		if tmp.Email.Renotify != nil {
			r.AlertSpec.Renotify = *tmp.Email.Renotify
			r.AlertSpec.RenotifyAfterSeconds = uint32(renotifyAfter.Seconds())
		}
		// Email settings
		notifier, err := structpb.NewStruct(map[string]any{
			"recipients": pbutil.ToSliceAny(tmp.Email.Recipients),
		})
		if err != nil {
			return fmt.Errorf("encountered invalid property type: %w", err)
		}
		r.AlertSpec.Notifiers = []*runtimev1.Notifier{
			{
				Connector:  "email",
				Properties: notifier,
			},
		}
	} else {
		// Override notification defaults
		if tmp.OnRecover != nil {
			r.AlertSpec.NotifyOnRecover = *tmp.OnRecover
		}
		if tmp.OnFail != nil {
			r.AlertSpec.NotifyOnFail = *tmp.OnFail
		}
		if tmp.OnError != nil {
			r.AlertSpec.NotifyOnError = *tmp.OnError
		}
		if tmp.Renotify != nil {
			r.AlertSpec.Renotify = *tmp.Renotify
			r.AlertSpec.RenotifyAfterSeconds = uint32(renotifyAfter.Seconds())
		}
		// Email settings
		if len(tmp.Notify.Email.Recipients) > 0 {
			props, err := structpb.NewStruct(map[string]any{
				"recipients": pbutil.ToSliceAny(tmp.Notify.Email.Recipients),
			})
			if err != nil {
				return fmt.Errorf("encountered invalid property type: %w", err)
			}
			r.AlertSpec.Notifiers = append(r.AlertSpec.Notifiers, &runtimev1.Notifier{
				Connector:  "email",
				Properties: props,
			})
		}
		// Slack settings
		if len(tmp.Notify.Slack.Channels) > 0 || len(tmp.Notify.Slack.Users) > 0 || len(tmp.Notify.Slack.Webhooks) > 0 {
			props, err := structpb.NewStruct(slack.EncodeProps(tmp.Notify.Slack.Users, tmp.Notify.Slack.Channels, tmp.Notify.Slack.Webhooks))
			if err != nil {
				return err
			}
			r.AlertSpec.Notifiers = append(r.AlertSpec.Notifiers, &runtimev1.Notifier{
				Connector:  "slack",
				Properties: props,
			})
		}
	}

	r.AlertSpec.Annotations = tmp.Annotations

	return nil
}
