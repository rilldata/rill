package slack

import (
	"bytes"
	"fmt"
	htemplate "html/template"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/admin/pkg/urlutil"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/slack-go/slack"
)

type notifier struct {
	h     *handle
	props map[string]any
}

func (n *notifier) SendScheduledReport(s *drivers.ScheduledReport) error {
	d := ReportStatusData{
		DisplayName:      s.DisplayName,
		ReportTimeString: s.ReportTime.Format(time.RFC1123),
		DownloadFormat:   s.DownloadFormat,
		OpenLink:         s.OpenLink,
		DownloadLink:     s.DownloadLink,
	}

	buf := new(bytes.Buffer)
	err := n.h.templates.Lookup("scheduled_report.slack").Execute(buf, d)
	if err != nil {
		return fmt.Errorf("slack template error: %w", err)
	}
	txt := buf.String()

	if err := n.sendTextToChannels(txt); err != nil {
		return err
	}
	if err := n.sendTextViaWebhooks(txt); err != nil {
		return err
	}

	d.UnsubscribeLink = s.UnsubscribeLink
	return n.sendReportToUsers(d)
}

func (n *notifier) sendReportToUsers(d ReportStatusData) error {
	props, err := n.parseProps()
	if err != nil {
		return err
	}

	if len(props.Users) == 0 {
		return nil
	}

	if n.h.api == nil {
		return fmt.Errorf("slack api is not configured, consider setting a bot token")
	}

	unsubLink := d.UnsubscribeLink

	for _, email := range props.Users {
		d.UnsubscribeLink = urlutil.MustWithQuery(unsubLink, map[string]string{"slack_user": email})

		buf := new(bytes.Buffer)
		err := n.h.templates.Lookup("scheduled_report.slack").Execute(buf, d)
		if err != nil {
			return fmt.Errorf("slack template error: %w", err)
		}
		txt := buf.String()

		user, err := n.h.api.GetUserByEmail(email)
		if err != nil {
			return fmt.Errorf("slack api error: %w", err)
		}
		_, _, err = n.h.api.PostMessage(user.ID, slack.MsgOptionText(txt, false), slack.MsgOptionDisableLinkUnfurl())
		if err != nil {
			return fmt.Errorf("slack api error: %w", err)
		}
	}
	return nil
}

func (n *notifier) SendAlertStatus(s *drivers.AlertStatus) error {
	switch s.Status {
	case runtimev1.AssertionStatus_ASSERTION_STATUS_PASS:
		return n.sendAlertStatus(&AlertStatusData{
			DisplayName:         s.DisplayName,
			ExecutionTimeString: s.ExecutionTime.Format(time.RFC1123),
			IsPass:              true,
			IsRecover:           s.IsRecover,
			OpenLink:            htemplate.URL(s.OpenLink),
			EditLink:            htemplate.URL(s.EditLink),
		})
	case runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL:
		return n.sendAlertFail(&AlertFailData{
			DisplayName:         s.DisplayName,
			ExecutionTimeString: s.ExecutionTime.Format(time.RFC1123),
			FailRow:             s.FailRow,
			OpenLink:            htemplate.URL(s.OpenLink),
			EditLink:            htemplate.URL(s.EditLink),
		})
	case runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR:
		return n.sendAlertStatus(&AlertStatusData{
			DisplayName:         s.DisplayName,
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
	subject := fmt.Sprintf("%s (%s)", data.DisplayName, data.ExecutionTimeString)
	if data.IsRecover {
		subject = fmt.Sprintf("Recovered: %s", subject)
	}
	data.Subject = subject

	buf := new(bytes.Buffer)
	err := n.h.templates.Lookup("alert_status.slack").Execute(buf, data)
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
	data.Subject = fmt.Sprintf("%s (%s)", data.DisplayName, data.ExecutionTimeString)

	buf := new(bytes.Buffer)
	err := n.h.templates.Lookup("alert_fail.slack").Execute(buf, data)
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
	props, err := n.parseProps()
	if err != nil {
		return err
	}

	if len(props.Channels) == 0 {
		return nil
	}

	if n.h.api == nil {
		return fmt.Errorf("slack api is not configured, consider setting a bot token")
	}

	for _, channel := range props.Channels {
		_, _, err := n.h.api.PostMessage(channel, slack.MsgOptionText(txt, false), slack.MsgOptionDisableLinkUnfurl())
		if err != nil {
			return fmt.Errorf("slack api error: %w", err)
		}
	}
	return nil
}

func (n *notifier) sendTextToUsers(txt string) error {
	props, err := n.parseProps()
	if err != nil {
		return err
	}

	if len(props.Users) == 0 {
		return nil
	}

	if n.h.api == nil {
		return fmt.Errorf("slack api is not configured, consider setting a bot token")
	}

	for _, email := range props.Users {
		user, err := n.h.api.GetUserByEmail(email)
		if err != nil {
			return fmt.Errorf("slack api error: %w", err)
		}
		_, _, err = n.h.api.PostMessage(user.ID, slack.MsgOptionText(txt, false), slack.MsgOptionDisableLinkUnfurl())
		if err != nil {
			return fmt.Errorf("slack api error: %w", err)
		}
	}
	return nil
}

func (n *notifier) sendTextViaWebhooks(txt string) error {
	props, err := n.parseProps()
	if err != nil {
		return err
	}
	for _, webhook := range props.Webhooks {
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

type notifierProperties struct {
	Users    []string `mapstructure:"users"`
	Channels []string `mapstructure:"channels"`
	Webhooks []string `mapstructure:"webhooks"`
}

func (n *notifier) parseProps() (*notifierProperties, error) {
	props := &notifierProperties{}
	err := mapstructure.WeakDecode(n.props, props)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Slack properties: %w", err)
	}
	return props, nil
}

type ReportStatusData struct {
	DisplayName      string
	ReportTimeString string
	DownloadFormat   string
	OpenLink         string
	DownloadLink     string
	UnsubscribeLink  string
}

type AlertStatusData struct {
	Subject             string
	DisplayName         string
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
	DisplayName         string
	ExecutionTimeString string // Will be inferred from ExecutionTime
	FailRow             map[string]any
	OpenLink            htemplate.URL
	EditLink            htemplate.URL
}
