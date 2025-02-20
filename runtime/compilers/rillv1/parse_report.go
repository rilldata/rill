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

// ReportYAML is the raw structure of a Report resource defined in YAML (does not include common fields)
type ReportYAML struct {
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
	Timeout string `yaml:"timeout"`
	Query   struct {
		Name     string         `yaml:"name"`
		Args     map[string]any `yaml:"args"`
		ArgsJSON string         `yaml:"args_json"`
	} `yaml:"query"`
	Export struct {
		Format string `yaml:"format"`
		Limit  uint   `yaml:"limit"`
	} `yaml:"export"`
	Email struct {
		Recipients []string `yaml:"recipients"`
	} `yaml:"email"`
	Notify struct {
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
}

// parseReport parses a report definition and adds the resulting resource to p.Resources.
func (p *Parser) parseReport(node *Node) error {
	// Parse YAML
	tmp := &ReportYAML{}
	err := p.decodeNodeYAML(node, true, tmp)
	if err != nil {
		return err
	}

	// Validate SQL or connector isn't set
	if node.SQL != "" {
		return fmt.Errorf("reports cannot have SQL")
	}
	if !node.ConnectorInferred && node.Connector != "" {
		return fmt.Errorf("reports cannot have a connector")
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

	// Parse export format
	exportFormat, err := parseExportFormat(tmp.Export.Format)
	if err != nil {
		return err
	}
	if exportFormat == runtimev1.ExportFormat_EXPORT_FORMAT_UNSPECIFIED {
		return fmt.Errorf(`missing required property "export.format"`)
	}

	if len(tmp.Email.Recipients) > 0 && len(tmp.Notify.Email.Recipients) > 0 {
		return errors.New(`cannot set both "email.recipients" and "notify.email.recipients"`)
	}

	isLegacySyntax := len(tmp.Email.Recipients) > 0

	// Validate recipients
	if isLegacySyntax {
		// Backward compatibility
		for _, email := range tmp.Email.Recipients {
			_, err := mail.ParseAddress(email)
			if err != nil {
				return fmt.Errorf("invalid recipient email address %q", email)
			}
		}
	} else {
		if len(tmp.Notify.Email.Recipients) == 0 && len(tmp.Notify.Slack.Channels) == 0 &&
			len(tmp.Notify.Slack.Users) == 0 && len(tmp.Notify.Slack.Webhooks) == 0 {
			return fmt.Errorf(`missing notification recipients`)
		}
		for _, email := range tmp.Notify.Email.Recipients {
			_, err := mail.ParseAddress(email)
			if err != nil {
				return fmt.Errorf("invalid recipient email address %q", email)
			}
		}
		for _, email := range tmp.Notify.Slack.Users {
			_, err := mail.ParseAddress(email)
			if err != nil {
				return fmt.Errorf("invalid recipient email address %q", email)
			}
		}
	}

	// Track report
	r, err := p.insertResource(ResourceKindReport, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	r.ReportSpec.DisplayName = tmp.DisplayName
	if r.ReportSpec.DisplayName == "" {
		r.ReportSpec.DisplayName = ToDisplayName(node.Name)
	}
	if schedule != nil {
		r.ReportSpec.RefreshSchedule = schedule
	}
	r.ReportSpec.WatermarkInherit = watermarkInherit
	r.ReportSpec.IntervalsIsoDuration = tmp.Intervals.Duration
	r.ReportSpec.IntervalsLimit = int32(tmp.Intervals.Limit)
	r.ReportSpec.IntervalsCheckUnclosed = tmp.Intervals.CheckUnclosed
	if timeout != 0 {
		r.ReportSpec.TimeoutSeconds = uint32(timeout.Seconds())
	}
	r.ReportSpec.QueryName = tmp.Query.Name
	r.ReportSpec.QueryArgsJson = tmp.Query.ArgsJSON
	r.ReportSpec.ExportLimit = uint64(tmp.Export.Limit)
	r.ReportSpec.ExportFormat = exportFormat

	if isLegacySyntax {
		// Backwards compatibility
		// Email settings
		notifier, err := structpb.NewStruct(map[string]any{
			"recipients": pbutil.ToSliceAny(tmp.Email.Recipients),
		})
		if err != nil {
			return fmt.Errorf("encountered invalid property type: %w", err)
		}
		r.ReportSpec.Notifiers = []*runtimev1.Notifier{
			{
				Connector:  "email",
				Properties: notifier,
			},
		}
	} else {
		// Email settings
		if len(tmp.Notify.Email.Recipients) > 0 {
			props, err := structpb.NewStruct(map[string]any{
				"recipients": pbutil.ToSliceAny(tmp.Notify.Email.Recipients),
			})
			if err != nil {
				return fmt.Errorf("encountered invalid property type: %w", err)
			}
			r.ReportSpec.Notifiers = append(r.ReportSpec.Notifiers, &runtimev1.Notifier{
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
			r.ReportSpec.Notifiers = append(r.ReportSpec.Notifiers, &runtimev1.Notifier{
				Connector:  "slack",
				Properties: props,
			})
		}
	}

	r.ReportSpec.Annotations = tmp.Annotations

	return nil
}

func parseExportFormat(s string) (runtimev1.ExportFormat, error) {
	switch strings.ToLower(s) {
	case "":
		return runtimev1.ExportFormat_EXPORT_FORMAT_UNSPECIFIED, nil
	case "csv":
		return runtimev1.ExportFormat_EXPORT_FORMAT_CSV, nil
	case "xlsx":
		return runtimev1.ExportFormat_EXPORT_FORMAT_XLSX, nil
	case "parquet":
		return runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET, nil
	default:
		if val, ok := runtimev1.ExportFormat_value[s]; ok {
			return runtimev1.ExportFormat(val), nil
		}
		return runtimev1.ExportFormat_EXPORT_FORMAT_UNSPECIFIED, fmt.Errorf("invalid export format %q", s)
	}
}
