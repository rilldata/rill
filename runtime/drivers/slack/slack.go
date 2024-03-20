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
		templates: template.Must(template.New("").ParseFS(templatesFS, "templates/slack/*.slack")),
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

func (h *handle) SendScheduledReport(s *drivers.ScheduledReport, r drivers.RecipientOpts) error {
	opts, ok := r.(*RecipientsOpts)
	if !ok {
		return fmt.Errorf("invalid recipient type: %T", r)
	}
	buf := new(bytes.Buffer)
	err := h.templates.Lookup("scheduled_report.slack").Execute(buf, s)
	if err != nil {
		return fmt.Errorf("slack template error: %w", err)
	}
	txt := buf.String()

	if err := h.sendTextToChannels(txt, opts.Channels); err != nil {
		return err
	}
	if err := h.sendTextToEmails(txt, opts.Emails); err != nil {
		return err
	}
	return h.sendTextViaWebhooks(txt, opts.Webhooks)
}

func (h *handle) SendAlertStatus(s *drivers.AlertStatus, r drivers.RecipientOpts) error {
	slackSpec, ok := r.(*RecipientsOpts)
	if !ok {
		return fmt.Errorf("invalid recipient type: %T", r)
	}
	switch s.Status {
	case runtimev1.AssertionStatus_ASSERTION_STATUS_PASS:
		return h.sendAlertStatus(slackSpec, &AlertStatusData{
			Title:               s.Title,
			ExecutionTimeString: s.ExecutionTime.Format(time.RFC1123),
			IsPass:              true,
			IsRecover:           s.IsRecover,
			OpenLink:            htemplate.URL(s.OpenLink),
			EditLink:            htemplate.URL(s.EditLink),
		})
	case runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL:
		return h.sendAlertFail(slackSpec, &AlertFailData{
			Title:               s.Title,
			ExecutionTimeString: s.ExecutionTime.Format(time.RFC1123),
			FailRow:             s.FailRow,
			OpenLink:            htemplate.URL(s.OpenLink),
			EditLink:            htemplate.URL(s.EditLink),
		})
	case runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR:
		return h.sendAlertStatus(slackSpec, &AlertStatusData{
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

func (h *handle) sendAlertStatus(opts *RecipientsOpts, data *AlertStatusData) error {
	subject := fmt.Sprintf("%s (%s)", data.Title, data.ExecutionTimeString)
	if data.IsRecover {
		subject = fmt.Sprintf("Recovered: %s", subject)
	}
	data.Subject = subject

	buf := new(bytes.Buffer)
	err := h.templates.Lookup("alert_status.slack").Execute(buf, data)
	if err != nil {
		return fmt.Errorf("slack template error: %w", err)
	}
	txt := buf.String()

	if err := h.sendTextToChannels(txt, opts.Channels); err != nil {
		return err
	}
	if err := h.sendTextToEmails(txt, opts.Emails); err != nil {
		return err
	}
	return h.sendTextViaWebhooks(txt, opts.Webhooks)
}

func (h *handle) sendAlertFail(opts *RecipientsOpts, data *AlertFailData) error {
	data.Subject = fmt.Sprintf("%s (%s)", data.Title, data.ExecutionTimeString)

	buf := new(bytes.Buffer)
	err := h.templates.Lookup("alert_fail.slack").Execute(buf, data)
	if err != nil {
		return fmt.Errorf("slack template error: %w", err)
	}
	txt := buf.String()

	if err := h.sendTextToChannels(txt, opts.Channels); err != nil {
		return err
	}
	if err := h.sendTextToEmails(txt, opts.Emails); err != nil {
		return err
	}
	return h.sendTextViaWebhooks(txt, opts.Webhooks)
}

func (h *handle) sendTextToChannels(txt string, channels []string) error {
	api := slack.New(h.config.BotToken)
	for _, channel := range channels {
		_, _, err := api.PostMessage(channel, slack.MsgOptionText(txt, false), slack.MsgOptionDisableLinkUnfurl())
		if err != nil {
			return fmt.Errorf("slack api error: %w", err)
		}
	}
	return nil
}

func (h *handle) sendTextToEmails(txt string, emails []string) error {
	api := slack.New(h.config.BotToken)
	for _, email := range emails {
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

func (h *handle) sendTextViaWebhooks(txt string, webhooks []string) error {
	for _, webhook := range webhooks {
		payload := slack.WebhookMessage{
			Text: txt,
		}
		err := slack.PostWebhook(webhook, &payload)
		if err != nil {
			return fmt.Errorf("slack webhook error: %w", err)
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
	Webhooks []string
}

type AlertStatusData struct {
	Subject             string
	Title               string
	ExecutionTimeString string // Will be inferred from ExecutionTime
	IsPass              bool
	IsRecover           bool
	IsError             bool
	ErrorMessage        string
	OpenLink            htemplate.URL
	EditLink            htemplate.URL
}

type AlertFailData struct {
	Subject             string
	Title               string
	ExecutionTimeString string // Will be inferred from ExecutionTime
	FailRow             map[string]any
	OpenLink            htemplate.URL
	EditLink            htemplate.URL
}
