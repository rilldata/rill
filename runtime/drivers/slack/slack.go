package slack

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	htemplate "html/template"
	"text/template"
	"time"

	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

//go:embed templates/slack/*
var templatesFS embed.FS

var spec = drivers.Spec{
	DisplayName: "Slack",
	Description: "Slack Notifier",
	ConfigProperties: []drivers.PropertySchema{
		{
			Key:    "bot_token",
			Secret: true,
		},
	},
}

func init() {
	drivers.Register("slack", driver{})
	drivers.RegisterAsConnector("slack", driver{})
}

type driver struct{}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) Open(config map[string]any, shared bool, client *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if shared {
		return nil, fmt.Errorf("slack driver can't be shared")
	}
	conf := &configProperties{}
	err := mapstructure.Decode(config, conf)
	if err != nil {
		return nil, err
	}

	conn := &handle{
		config:    conf,
		logger:    logger,
		templates: template.Must(template.New("").ParseFS(templatesFS, "templates/slack/*.md")),
	}
	return conn, nil
}

func (d driver) Drop(config map[string]any, logger *zap.Logger) error {
	return nil
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, props map[string]any, logger *zap.Logger) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

type handle struct {
	config    *configProperties
	logger    *zap.Logger
	templates *template.Template
}

var _ drivers.Handle = &handle{}

func (h *handle) Driver() string {
	return "slack"
}

func (h *handle) Config() map[string]any {
	return map[string]any{}
}

func (h *handle) Migrate(ctx context.Context) error {
	return nil
}

func (h *handle) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

func (h *handle) Close() error {
	return nil
}

func (h *handle) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

func (h *handle) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

func (h *handle) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

func (h *handle) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

func (h *handle) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

func (h *handle) AsSQLStore() (drivers.SQLStore, bool) {
	return nil, false
}

func (h *handle) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

func (h *handle) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

func (h *handle) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

func (h *handle) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	return nil, false
}

func (h *handle) AsNotifier() (drivers.Notifier, bool) {
	return h, true
}

func (h *handle) SendAlertStatus(s *drivers.AlertStatus, r drivers.RecipientOpts) error {
	slackSpec, ok := r.(*RecipientsOpts)
	if !ok {
		return fmt.Errorf("invalid recipient s type: %T", r)
	}
	switch s.Status {
	case runtimev1.AssertionStatus_ASSERTION_STATUS_PASS:
		return h.sendAlertStatus(slackSpec, &drivers.AlertStatusData{
			Title:               s.Title,
			ExecutionTimeString: s.ExecutionTime.Format(time.RFC1123),
			IsPass:              true,
			IsRecover:           s.IsRecover,
			OpenLink:            htemplate.URL(s.OpenLink),
			EditLink:            htemplate.URL(s.EditLink),
		})
	case runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL:
		return h.sendAlertFail(slackSpec, &drivers.AlertFailData{
			Title:               s.Title,
			ExecutionTimeString: s.ExecutionTime.Format(time.RFC1123),
			FailRow:             s.FailRow,
			OpenLink:            htemplate.URL(s.OpenLink),
			EditLink:            htemplate.URL(s.EditLink),
		})
	case runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR:
		return h.sendAlertStatus(slackSpec, &drivers.AlertStatusData{
			Title:               s.Title,
			ExecutionTimeString: s.ExecutionTime.Format(time.RFC1123),
			IsError:             true,
			ErrorMessage:        s.ExecutionError,
			OpenLink:            htemplate.URL(s.EditLink),
			EditLink:            htemplate.URL(s.EditLink),
		})
	default:
		return fmt.Errorf("unknown assertion status: %v", s.Status)
	}
}

func (h *handle) sendAlertStatus(spec *RecipientsOpts, data *drivers.AlertStatusData) error {
	subject := fmt.Sprintf("%s (%s)", data.Title, data.ExecutionTimeString)
	if data.IsRecover {
		subject = fmt.Sprintf("Recovered: %s", subject)
	}
	data.Subject = subject

	buf := new(bytes.Buffer)
	err := h.templates.Lookup("alert_status.md").Execute(buf, data)
	if err != nil {
		return fmt.Errorf("slack template error: %w", err)
	}
	txt := buf.String()
	api := slack.New(h.config.BotToken)
	for _, channel := range spec.Channels {
		_, _, err := api.PostMessage(channel, slack.MsgOptionText(txt, false))
		if err != nil {
			return fmt.Errorf("slack api error: %w", err)
		}
	}
	for _, email := range spec.Emails {
		user, err := api.GetUserByEmail(email)
		if err != nil {
			return fmt.Errorf("slack api error: %w", err)
		}
		_, _, err = api.PostMessage(user.ID, slack.MsgOptionText(txt, false))
		if err != nil {
			return fmt.Errorf("slack api error: %w", err)
		}
	}

	return nil
}

func (h *handle) sendAlertFail(spec *RecipientsOpts, data *drivers.AlertFailData) error {
	data.Subject = fmt.Sprintf("%s (%s)", data.Title, data.ExecutionTimeString)

	buf := new(bytes.Buffer)
	err := h.templates.Lookup("alert_fail.md").Execute(buf, data)
	if err != nil {
		return fmt.Errorf("slack template error: %w", err)
	}
	txt := buf.String()

	api := slack.New(h.config.BotToken)
	for _, channel := range spec.Channels {
		_, _, err := api.PostMessage(channel, slack.MsgOptionText(txt, false))
		if err != nil {
			return fmt.Errorf("slack api error: %w", err)
		}
	}
	for _, email := range spec.Emails {
		user, err := api.GetUserByEmail(email)
		if err != nil {
			return fmt.Errorf("slack api error: %w", err)
		}
		_, _, err = api.PostMessage(user.ID, slack.MsgOptionText(txt, false), slack.MsgOptionDisableLinkUnfurl())
		if err != nil {
			return fmt.Errorf("slack api error: %w", err)
		}
	}
	return nil
}

type configProperties struct {
	BotToken string `mapstructure:"bot_token"`
}

type RecipientsOpts struct {
	Channels []string
	Emails   []string
}
