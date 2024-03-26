package rillv1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// ReportYAML is the raw structure of a Report resource defined in YAML (does not include common fields)
type ReportYAML struct {
	commonYAML `yaml:",inline"` // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	Title      string           `yaml:"title"`
	Refresh    *ScheduleYAML    `yaml:"refresh"`
	Timeout    string           `yaml:"timeout"`
	Query      struct {
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
			Channels []string `yaml:"channels"`
			Emails   []string `yaml:"emails"`
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
			len(tmp.Notify.Slack.Emails) == 0 && len(tmp.Notify.Slack.Webhooks) == 0 {
			return fmt.Errorf(`missing notification recipients`)
		}
		for _, email := range tmp.Notify.Email.Recipients {
			_, err := mail.ParseAddress(email)
			if err != nil {
				return fmt.Errorf("invalid recipient email address %q", email)
			}
		}
		for _, email := range tmp.Notify.Slack.Emails {
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

	r.ReportSpec.Title = tmp.Title
	if schedule != nil {
		r.ReportSpec.RefreshSchedule = schedule
	}
	if timeout != 0 {
		r.ReportSpec.TimeoutSeconds = uint32(timeout.Seconds())
	}
	r.ReportSpec.QueryName = tmp.Query.Name
	r.ReportSpec.QueryArgsJson = tmp.Query.ArgsJSON
	r.ReportSpec.ExportLimit = uint64(tmp.Export.Limit)
	r.ReportSpec.ExportFormat = exportFormat

	r.ReportSpec.NotifySpec = &runtimev1.ReportNotifySpec{}

	if isLegacySyntax {
		// Backwards compatibility
		// Email settings
		r.ReportSpec.NotifySpec.Notifiers = []*runtimev1.NotifierSpec{
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
		// Email settings
		if len(tmp.Notify.Email.Recipients) > 0 {
			r.ReportSpec.NotifySpec.Notifiers = append(r.ReportSpec.NotifySpec.Notifiers, &runtimev1.NotifierSpec{
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
			r.ReportSpec.NotifySpec.Notifiers = append(r.ReportSpec.NotifySpec.Notifiers, &runtimev1.NotifierSpec{
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
