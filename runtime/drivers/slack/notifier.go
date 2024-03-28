package slack

import (
	"bytes"
	"fmt"
	htemplate "html/template"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/slack-go/slack"
)

func (h *handle) SendScheduledReport(s *drivers.ScheduledReport, spec drivers.NotifierSpec) error {
	slackSpec, err := h.validateSlackSpec(spec)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	err = h.templates.Lookup("scheduled_report.slack").Execute(buf, s)
	if err != nil {
		return fmt.Errorf("slack template error: %w", err)
	}
	txt := buf.String()

	if err := h.sendTextToChannels(txt, slackSpec.Channels); err != nil {
		return err
	}
	if err := h.sendTextToEmails(txt, slackSpec.Emails); err != nil {
		return err
	}
	return h.sendTextViaWebhooks(txt, slackSpec.Webhooks)
}

func (h *handle) SendAlertStatus(s *drivers.AlertStatus, spec drivers.NotifierSpec) error {
	slackSpec, err := h.validateSlackSpec(spec)
	if err != nil {
		return err
	}
	switch s.Status {
	case runtimev1.AssertionStatus_ASSERTION_STATUS_PASS:
		return h.sendAlertStatus(&AlertStatusData{
			Title:               s.Title,
			ExecutionTimeString: s.ExecutionTime.Format(time.RFC1123),
			IsPass:              true,
			IsRecover:           s.IsRecover,
			OpenLink:            htemplate.URL(s.OpenLink),
			EditLink:            htemplate.URL(s.EditLink),
		}, slackSpec)
	case runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL:
		return h.sendAlertFail(&AlertFailData{
			Title:               s.Title,
			ExecutionTimeString: s.ExecutionTime.Format(time.RFC1123),
			FailRow:             s.FailRow,
			OpenLink:            htemplate.URL(s.OpenLink),
			EditLink:            htemplate.URL(s.EditLink),
		}, slackSpec)
	case runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR:
		return h.sendAlertStatus(&AlertStatusData{
			Title:               s.Title,
			ExecutionTimeString: s.ExecutionTime.Format(time.RFC1123),
			IsError:             true,
			ErrorMessage:        s.ExecutionError,
			OpenLink:            htemplate.URL(s.EditLink),
			EditLink:            htemplate.URL(s.EditLink),
		}, slackSpec)
	default:
		return fmt.Errorf("unknown assertion status: %v", s.Status)
	}
}

func (h *handle) validateSlackSpec(spec drivers.NotifierSpec) (*runtimev1.SlackNotifierSpec, error) {
	notifierSpec, ok := spec.(*runtimev1.NotifierSpec_Slack)
	if !ok {
		return nil, fmt.Errorf("invalid notifier spec type: %T", spec)
	}
	slackSpec := notifierSpec.Slack
	if len(slackSpec.Channels) == 0 && len(slackSpec.Emails) == 0 && len(slackSpec.Webhooks) == 0 {
		return nil, fmt.Errorf("no slack recipients specified")
	}
	return slackSpec, nil
}

func (h *handle) sendAlertStatus(data *AlertStatusData, spec *runtimev1.SlackNotifierSpec) error {
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

	if err := h.sendTextToChannels(txt, spec.Channels); err != nil {
		return err
	}
	if err := h.sendTextToEmails(txt, spec.Emails); err != nil {
		return err
	}
	return h.sendTextViaWebhooks(txt, spec.Webhooks)
}

func (h *handle) sendAlertFail(data *AlertFailData, spec *runtimev1.SlackNotifierSpec) error {
	data.Subject = fmt.Sprintf("%s (%s)", data.Title, data.ExecutionTimeString)

	buf := new(bytes.Buffer)
	err := h.templates.Lookup("alert_fail.slack").Execute(buf, data)
	if err != nil {
		return fmt.Errorf("slack template error: %w", err)
	}
	txt := buf.String()

	if err := h.sendTextToChannels(txt, spec.Channels); err != nil {
		return err
	}
	if err := h.sendTextToEmails(txt, spec.Emails); err != nil {
		return err
	}
	return h.sendTextViaWebhooks(txt, spec.Webhooks)
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
