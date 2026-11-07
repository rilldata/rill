package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

type mockSender struct {
	fromEmail   string
	fromName    string
	toEmail     string
	toName      string
	subject     string
	body        string
	attachments []Attachment
}

func (m *mockSender) Send(toEmail, toName, subject, body string, attachments []Attachment) error {
	m.toEmail = toEmail
	m.toName = toName
	m.subject = subject
	m.body = body
	m.attachments = attachments
	return nil
}

func TestCopyrightYear(t *testing.T) {
	mock := &mockSender{}
	client := New(mock)

	opts := &CallToAction{
		ToEmail:    uuid.New().String(),
		ToName:     uuid.New().String(),
		Subject:    uuid.New().String(),
		ButtonText: uuid.New().String(),
		ButtonLink: uuid.New().String(),
		ShowFooter: true,
	}
	err := client.SendCallToAction(opts)
	require.NoError(t, err)

	require.Equal(t, opts.ToEmail, mock.toEmail)
	require.Equal(t, opts.ToName, mock.toName)
	require.Equal(t, opts.Subject, mock.subject)
	require.Contains(t, mock.body, opts.ButtonText)
	require.Contains(t, mock.body, opts.ButtonLink)

	year := time.Now().Year()
	require.Contains(t, mock.body, fmt.Sprintf("© %d Rill Data, Inc", year))
}

func TestOrganizationInvite(t *testing.T) {
	mock := &mockSender{}
	client := New(mock)

	opts := &OrganizationInvite{
		ToEmail:       uuid.New().String(),
		ToName:        uuid.New().String(),
		AcceptURL:     "https://api.example.com",
		OrgName:       uuid.New().String(),
		RoleName:      uuid.New().String(),
		InvitedByName: uuid.New().String(),
	}
	err := client.SendOrganizationInvite(opts)
	require.NoError(t, err)

	require.Equal(t, opts.ToEmail, mock.toEmail)
	require.Equal(t, opts.ToName, mock.toName)
	require.NotEmpty(t, mock.subject)
	require.Contains(t, mock.body, opts.OrgName)
	require.Contains(t, mock.body, opts.RoleName)
	require.Contains(t, mock.body, opts.InvitedByName)
}

func TestAlertFail(t *testing.T) {
	mock := &mockSender{}
	client := New(mock)

	opts := &AlertStatus{
		ToEmail:       uuid.New().String(),
		ToName:        uuid.New().String(),
		DisplayName:   "Foobar",
		ExecutionTime: time.Date(2024, 01, 27, 0, 0, 0, 0, time.UTC),
		Status:        runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL,
		FailRow:       map[string]any{"hello": "world", "pi": 3.14},
		OpenLink:      "https://example.com",
		EditLink:      "https://example.com",
	}
	err := client.SendAlertStatus(opts)
	require.NoError(t, err)

	require.Equal(t, opts.ToEmail, mock.toEmail)
	require.Equal(t, opts.ToName, mock.toName)
	require.NotEmpty(t, mock.subject)
	require.Contains(t, mock.body, opts.DisplayName)
	require.Contains(t, mock.body, opts.ExecutionTime.Format(time.RFC1123))
	for k, v := range opts.FailRow {
		require.Contains(t, mock.body, k)
		require.Contains(t, mock.body, fmt.Sprintf("%v", v))
	}
	require.NotContains(t, mock.body, "are attached")
	require.NotContains(t, mock.body, "too large to attach")
	require.Empty(t, mock.attachments)

	// The export is attached
	opts.Attachment = &Attachment{Filename: "foobar.csv", ContentType: "text/csv", Data: []byte("a,b\n1,2\n")}
	err = client.SendAlertStatus(opts)
	require.NoError(t, err)
	require.Contains(t, mock.body, "The rows that matched your alert criteria are attached.")
	require.Equal(t, []Attachment{*opts.Attachment}, mock.attachments)

	// The export was too large to attach
	opts.Attachment = nil
	opts.AttachmentTooLarge = true
	err = client.SendAlertStatus(opts)
	require.NoError(t, err)
	require.Contains(t, mock.body, "too large to attach to this email")
	require.Empty(t, mock.attachments)
}

func TestAlertRecover(t *testing.T) {
	mock := &mockSender{}
	client := New(mock)

	opts := &AlertStatus{
		ToEmail:       uuid.New().String(),
		ToName:        uuid.New().String(),
		DisplayName:   "Foobar",
		ExecutionTime: time.Date(2024, 01, 27, 0, 0, 0, 0, time.UTC),
		Status:        runtimev1.AssertionStatus_ASSERTION_STATUS_PASS,
		IsRecover:     true,
		OpenLink:      "https://example.com",
		EditLink:      "https://example.com",
	}
	err := client.SendAlertStatus(opts)
	require.NoError(t, err)

	require.Equal(t, opts.ToEmail, mock.toEmail)
	require.Equal(t, opts.ToName, mock.toName)
	require.NotEmpty(t, mock.subject)
	require.Contains(t, mock.body, opts.DisplayName)
	require.Contains(t, mock.body, opts.ExecutionTime.Format(time.RFC1123))
	require.Contains(t, mock.body, "recovered")
}

func TestAlertError(t *testing.T) {
	mock := &mockSender{}
	client := New(mock)

	opts := &AlertStatus{
		ToEmail:        uuid.New().String(),
		ToName:         uuid.New().String(),
		DisplayName:    "Foobar",
		ExecutionTime:  time.Date(2024, 01, 27, 0, 0, 0, 0, time.UTC),
		Status:         runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR,
		ExecutionError: "hello error",
		OpenLink:       "https://example.com",
		EditLink:       "https://example.com",
	}
	err := client.SendAlertStatus(opts)
	require.NoError(t, err)

	require.Equal(t, opts.ToEmail, mock.toEmail)
	require.Equal(t, opts.ToName, mock.toName)
	require.NotEmpty(t, mock.subject)
	require.Contains(t, mock.body, opts.DisplayName)
	require.Contains(t, mock.body, opts.ExecutionTime.Format(time.RFC1123))
	require.Contains(t, mock.body, "hello error")
}

func TestScheduledReport(t *testing.T) {
	mock := &mockSender{}
	client := New(mock)

	opts := &ScheduledReport{
		ToEmail:        uuid.New().String(),
		ToName:         uuid.New().String(),
		DisplayName:    "Foobar",
		ReportTime:     time.Date(2024, 01, 27, 0, 0, 0, 0, time.UTC),
		DownloadFormat: "CSV",
		OpenLink:       "https://example.com/open",
		DownloadLink:   "https://example.com/download",
	}
	err := client.SendScheduledReport(opts)
	require.NoError(t, err)
	require.Equal(t, opts.ToEmail, mock.toEmail)
	require.Contains(t, mock.body, opts.DisplayName)
	require.Contains(t, mock.body, opts.DownloadLink)
	require.NotContains(t, mock.body, "file is attached")
	require.NotContains(t, mock.body, "too large to attach")
	require.Empty(t, mock.attachments)

	// The export is attached
	opts.Attachment = &Attachment{Filename: "foobar.csv", ContentType: "text/csv", Data: []byte("a,b\n1,2\n")}
	err = client.SendScheduledReport(opts)
	require.NoError(t, err)
	require.Contains(t, mock.body, "The CSV file is attached.")
	require.Contains(t, mock.body, opts.DownloadLink)
	require.Equal(t, []Attachment{*opts.Attachment}, mock.attachments)

	// The export was too large to attach
	opts.Attachment = nil
	opts.AttachmentTooLarge = true
	err = client.SendScheduledReport(opts)
	require.NoError(t, err)
	require.Contains(t, mock.body, "The CSV file is too large to attach to this email")
	require.Contains(t, mock.body, opts.DownloadLink)
	require.Empty(t, mock.attachments)
}

func TestComposeMessage(t *testing.T) {
	from := mail.Address{Name: "Rill", Address: "noreply@example.com"}
	to := mail.Address{Name: "Jane", Address: "jane@example.com"}
	// A long line with non-ASCII characters, which must be encoded to keep lines within the SMTP line length limit
	body := "<html><body>" + strings.Repeat("Rapport généré ", 100) + "</body></html>"

	// Without attachments, the message has a single HTML part
	msg, err := composeMessage(from, to, "Foobar", body, nil)
	require.NoError(t, err)
	parsed, err := mail.ReadMessage(bytes.NewReader(msg))
	require.NoError(t, err)
	require.Equal(t, "text/html; charset=utf-8", parsed.Header.Get("Content-Type"))
	data, err := io.ReadAll(parsed.Body)
	require.NoError(t, err)
	require.Equal(t, body+"\r\n", string(data))

	// With an attachment, the message is multipart with the HTML body as the first part
	attachment := Attachment{Filename: "rapport_généré.csv", ContentType: "text/csv", Data: make([]byte, 1000)}
	for i := range attachment.Data {
		attachment.Data[i] = byte(i)
	}
	msg, err = composeMessage(from, to, "Foobar", body, []Attachment{attachment})
	require.NoError(t, err)
	for _, line := range strings.Split(string(msg), "\r\n") {
		require.LessOrEqual(t, len(line), 998)
	}

	parsed, err = mail.ReadMessage(bytes.NewReader(msg))
	require.NoError(t, err)
	require.Equal(t, "Foobar", parsed.Header.Get("Subject"))
	mediaType, params, err := mime.ParseMediaType(parsed.Header.Get("Content-Type"))
	require.NoError(t, err)
	require.Equal(t, "multipart/mixed", mediaType)
	r := multipart.NewReader(parsed.Body, params["boundary"])

	part, err := r.NextPart()
	require.NoError(t, err)
	require.Equal(t, "text/html; charset=utf-8", part.Header.Get("Content-Type"))
	data, err = io.ReadAll(part) // The reader transparently decodes quoted-printable parts
	require.NoError(t, err)
	require.Equal(t, body, string(data))

	part, err = r.NextPart()
	require.NoError(t, err)
	require.Equal(t, attachment.Filename, part.FileName())
	mediaType, _, err = mime.ParseMediaType(part.Header.Get("Content-Type"))
	require.NoError(t, err)
	require.Equal(t, attachment.ContentType, mediaType)
	require.Equal(t, "base64", part.Header.Get("Content-Transfer-Encoding"))
	data, err = io.ReadAll(base64.NewDecoder(base64.StdEncoding, part))
	require.NoError(t, err)
	require.Equal(t, attachment.Data, data)

	_, err = r.NextPart()
	require.ErrorIs(t, err, io.EOF)
}
