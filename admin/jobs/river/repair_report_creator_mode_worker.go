package river

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type MigrateReportsCreatorModeArgs struct{}

func (MigrateReportsCreatorModeArgs) Kind() string { return "migrate_reports_creator_mode" }

type MigrateReportsCreatorModeWorker struct {
	river.WorkerDefaults[MigrateReportsCreatorModeArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// Work This worker changes the web_open_mode of reports created to "creator".
func (w *MigrateReportsCreatorModeWorker) Work(ctx context.Context, job *river.Job[MigrateReportsCreatorModeArgs]) error {
	w.logger.Info("Starting migrate_reports_creator_mode job")
	afterPath := ""
	limit := 100
	for {
		reports, err := w.admin.DB.ListVirtualReportFiles(ctx, afterPath, limit)
		if err != nil {
			return err
		}
		for _, report := range reports {
			y, err := unmarshalReportYAML(report.Data)
			if err != nil {
				w.logger.Error("failed to unmarshal report yaml", zap.String("report", report.Path), zap.Error(err))
				return err
			}
			if y.Annotations.WebOpenMode == WebOpenModeCreator {
				w.logger.Info("Skipping report already in creator mode", zap.String("report", report.Path))
				afterPath = report.Path
				continue
			}
			y.Annotations.WebOpenMode = WebOpenModeCreator
			newData, err := yaml.Marshal(y)
			if err != nil {
				w.logger.Error("failed to marshal report yaml", zap.String("report", report.Path), zap.Error(err))
				return err
			}
			if err := w.admin.DB.UpsertVirtualFile(ctx, &database.InsertVirtualFileOptions{
				ProjectID:   report.ProjectID,
				Environment: "prod",
				Path:        report.Path,
				Data:        newData,
			}); err != nil {
				w.logger.Error("failed to update report yaml", zap.String("report", report.Path), zap.Error(err))
				return err
			}
			w.logger.Info("Updated report", zap.String("report", report.Path), zap.String("web_open_mode", string(y.Annotations.WebOpenMode)))
			afterPath = report.Path
		}
		if len(reports) < limit {
			break
		}
	}
	w.logger.Info("Completed migrate_reports_creator_mode job")
	return nil
}

func unmarshalReportYAML(data []byte) (*reportYAML, error) {
	var res reportYAML
	if err := yaml.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	if res.Type != "report" {
		return nil, fmt.Errorf("not a report")
	}
	return &res, nil
}

type reportYAML struct {
	Type        string `yaml:"type"`
	DisplayName string `yaml:"display_name"`
	Title       string `yaml:"title,omitempty"` // Deprecated: replaced by display_name, but kept for backwards compatibility
	Refresh     struct {
		Cron     string `yaml:"cron"`
		TimeZone string `yaml:"time_zone"`
	} `yaml:"refresh"`
	Watermark string `yaml:"watermark"`
	Intervals struct {
		Duration string `yaml:"duration"`
	} `yaml:"intervals"`
	Query struct {
		Name     string         `yaml:"name"`
		Args     map[string]any `yaml:"args,omitempty"`
		ArgsJSON string         `yaml:"args_json,omitempty"`
	} `yaml:"query"`
	Export struct {
		Format        string `yaml:"format"`
		IncludeHeader bool   `yaml:"include_header"`
		Limit         uint   `yaml:"limit"`
	} `yaml:"export"`
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
	Annotations reportAnnotations `yaml:"annotations,omitempty"`
}

type reportAnnotations struct {
	AdminOwnerUserID string      `yaml:"admin_owner_user_id"`
	AdminManaged     bool        `yaml:"admin_managed"`
	AdminNonce       string      `yaml:"admin_nonce"` // To ensure spec version gets updated on writes, to enable polling in TriggerReconcileAndAwaitReport
	WebOpenPath      string      `yaml:"web_open_path"`
	WebOpenState     string      `yaml:"web_open_state"`
	WebOpenMode      WebOpenMode `yaml:"web_open_mode,omitempty"`
	Explore          string      `yaml:"explore,omitempty"`
	Canvas           string      `yaml:"canvas,omitempty"`
}

type WebOpenMode string

const (
	WebOpenModeRecipient WebOpenMode = "recipient"
	WebOpenModeCreator   WebOpenMode = "creator"
	WebOpenModeNone      WebOpenMode = "none"
	WebOpenModeFiltered  WebOpenMode = "filtered"
)
