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
	"github.com/rilldata/rill/runtime/pkg/duration"
	"gopkg.in/yaml.v3"
)

// ReportYAML is the raw structure of a Report resource defined in YAML (does not include common fields)
type ReportYAML struct {
	commonYAML `yaml:",inline"` // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	Title      string           `yaml:"title"`
	Refresh    *ScheduleYAML    `yaml:"refresh"`
	Timeout    string           `yaml:"timeout"`
	Query      struct {
		Name      string         `yaml:"name"`
		TimeRange string         `yaml:"time_range"`
		Args      map[string]any `yaml:"args"`
		ArgsJSON  string         `yaml:"args_json"`
	} `yaml:"query"`
	Export struct {
		Format string `yaml:"format"`
		Limit  uint   `yaml:"limit"`
	} `yaml:"export"`
	Email struct {
		Recipients []string `yaml:"recipients"`
		Template   struct {
			OpenURL   string `yaml:"open_url"`
			EditURL   string `yaml:"edit_url"`
			ExportURL string `yaml:"export_url"`
		} `yaml:"template"`
	} `yaml:"email"`
	Security struct {
		OwnerUserID string `yaml:"owner_user_id"`
		ManagedByUI bool   `yaml:"managed_by_ui"`
	} `yaml:"security"`
}

// parseReport parses a report definition and adds the resulting resource to p.Resources.
func (p *Parser) parseReport(ctx context.Context, node *Node) error {
	// Parse YAML
	tmp := &ReportYAML{}
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

	// Query time range
	if tmp.Query.TimeRange != "" {
		_, err := duration.ParseISO8601(tmp.Query.TimeRange)
		if err != nil {
			return fmt.Errorf(`invalid "query.time_range": %w`, err)
		}
	}

	// Parse export format
	exportFormat, err := parseExportFormat(tmp.Export.Format)
	if err != nil {
		return err
	}
	if exportFormat == runtimev1.ExportFormat_EXPORT_FORMAT_UNSPECIFIED {
		return fmt.Errorf(`missing required property "export.format"`)
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
		r.SourceSpec.TimeoutSeconds = uint32(timeout.Seconds())
	}
	r.ReportSpec.QueryName = tmp.Query.Name
	r.ReportSpec.QueryArgsJson = tmp.Query.ArgsJSON
	r.ReportSpec.QueryTimeRange = tmp.Query.TimeRange
	r.ReportSpec.ExportLimit = uint64(tmp.Export.Limit)
	r.ReportSpec.ExportFormat = exportFormat
	r.ReportSpec.EmailRecipients = tmp.Email.Recipients
	r.ReportSpec.EmailOpenUrl = tmp.Email.Template.OpenURL
	r.ReportSpec.EmailEditUrl = tmp.Email.Template.EditURL
	r.ReportSpec.EmailExportUrl = tmp.Email.Template.ExportURL
	if tmp.Security.OwnerUserID != "" || tmp.Security.ManagedByUI {
		r.ReportSpec.Security = &runtimev1.ReportSpec_Security{}
		r.ReportSpec.Security.OwnerUserId = tmp.Security.OwnerUserID
		r.ReportSpec.Security.ManagedByUi = tmp.Security.ManagedByUI
	}

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
