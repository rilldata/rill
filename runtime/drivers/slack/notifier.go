package slack

import (
	"bytes"
	"embed"
	"fmt"
	htemplate "html/template"
	"text/template"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"github.com/slack-go/slack"
)

const (
	UsersField    = "users"
	ChannelsField = "channels"
	WebhooksField = "webhooks"
)

//go:embed templates/slack/*
var templatesFS embed.FS

type notifier struct {
	token     string
	users     []string
	channels  []string
	webhooks  []string
	templates *template.Template
}

func newNotifier(token string, props map[string]any) *notifier {
	users := pbutil.ToSliceString(props[UsersField].([]any))
	channels := pbutil.ToSliceString(props[ChannelsField].([]any))
	webhooks := pbutil.ToSliceString(props[WebhooksField].([]any))
	return &notifier{
		token:     token,
		users:     users,
		channels:  channels,
		webhooks:  webhooks,
		templates: template.Must(template.New("").ParseFS(templatesFS, "templates/slack/*.slack")),
	}
}

func (n *notifier) SendScheduledReport(s *drivers.ScheduledReport) error {
	buf := new(bytes.Buffer)
	err := n.templates.Lookup("scheduled_report.slack").Execute(buf, s)
	if err != nil {
		return fmt.Errorf("slack template error: %w", err)
	}
	txt := buf.String()

	if err := n.sendTextToChannels(txt); err != nil {
		return err
	}
	if err := n.sendTextToUsers(txt); err != nil {
		return err
	}
	return n.sendTextViaWebhooks(txt)
}

func (n *notifier) SendAlertStatus(s *drivers.AlertStatus) error {
	switch s.Status {
	case runtimev1.AssertionStatus_ASSERTION_STATUS_PASS:
		return n.sendAlertStatus(&AlertStatusData{
			Title:               s.Title,
			ExecutionTimeString: s.ExecutionTime.Format(time.RFC1123),
			IsPass:              true,
			IsRecover:           s.IsRecover,
			OpenLink:            htemplate.URL(s.OpenLink),
			EditLink:            htemplate.URL(s.EditLink),
		})
	case runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL:
		return n.sendAlertFail(&AlertFailData{
			Title:               s.Title,
			ExecutionTimeString: s.ExecutionTime.Format(time.RFC1123),
			FailRow:             s.FailRow,
			OpenLink:            htemplate.URL(s.OpenLink),
			EditLink:            htemplate.URL(s.EditLink),
		})
	case runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR:
		return n.sendAlertStatus(&AlertStatusData{
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

func (n *notifier) sendAlertStatus(data *AlertStatusData) error {
	subject := fmt.Sprintf("%s (%s)", data.Title, data.ExecutionTimeString)
	if data.IsRecover {
		subject = fmt.Sprintf("Recovered: %s", subject)
	}
	data.Subject = subject

	buf := new(bytes.Buffer)
	err := n.templates.Lookup("alert_status.slack").Execute(buf, data)
	if err != nil {
		return fmt.Errorf("slack template error: %w", err)
	}
	txt := buf.String()

	if err := n.sendTextToChannels(txt); err != nil {
		return err
	}
	if err := n.sendTextToUsers(txt); err != nil {
		return err
	}
	return n.sendTextViaWebhooks(txt)
}

func (n *notifier) sendAlertFail(data *AlertFailData) error {
	data.Subject = fmt.Sprintf("%s (%s)", data.Title, data.ExecutionTimeString)

	buf := new(bytes.Buffer)
	err := n.templates.Lookup("alert_fail.slack").Execute(buf, data)
	if err != nil {
		return fmt.Errorf("slack template error: %w", err)
	}
	txt := buf.String()

	if err := n.sendTextToChannels(txt); err != nil {
		return err
	}
	if err := n.sendTextToUsers(txt); err != nil {
		return err
	}
	return n.sendTextViaWebhooks(txt)
}

func (n *notifier) sendTextToChannels(txt string) error {
	if len(n.channels) == 0 {
		return nil
	}

	api, err := n.api()
	if err != nil {
		return err
	}
	for _, channel := range n.channels {
		_, _, err = api.PostMessage(channel, slack.MsgOptionText(txt, false), slack.MsgOptionDisableLinkUnfurl())
		if err != nil {
			return fmt.Errorf("slack api error: %w", err)
		}
	}
	return nil
}

func (n *notifier) sendTextToUsers(txt string) error {
	if len(n.users) == 0 {
		return nil
	}

	api, err := n.api()
	if err != nil {
		return err
	}
	for _, email := range n.users {
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

func (n *notifier) sendTextViaWebhooks(txt string) error {
	for _, webhook := range n.webhooks {
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

func (n *notifier) api() (*slack.Client, error) {
	if n.token == "" {
		return nil, fmt.Errorf("slack bot token is required")
	}
	return slack.New(n.token), nil
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
