package email

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type mockSender struct {
	fromEmail string
	fromName  string
	toEmail   string
	toName    string
	subject   string
	body      string
}

func (m *mockSender) Send(toEmail, toName, subject, body string) error {
	m.toEmail = toEmail
	m.toName = toName
	m.subject = subject
	m.body = body
	return nil
}

func TestOrganizationInvite(t *testing.T) {
	mock := &mockSender{}
	client := New(mock)

	opts := &OrganizationInvite{
		ToEmail:       uuid.New().String(),
		ToName:        uuid.New().String(),
		AdminURL:      "https://api.example.com",
		FrontendURL:   "https://ui.example.com",
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

func TestAlert(t *testing.T) {
	mock := &mockSender{}
	client := New(mock)

	opts := &Alert{
		ToEmail:       uuid.New().String(),
		ToName:        uuid.New().String(),
		Title:         "Foobar",
		ExecutionTime: time.Date(2024, 01, 27, 0, 0, 0, 0, time.UTC),
		FailRow:       map[string]any{"hello": "world", "pi": 3.14},
		OpenLink:      "https://example.com",
		EditLink:      "https://example.com",
	}
	err := client.SendAlert(opts)
	require.NoError(t, err)

	require.Equal(t, opts.ToEmail, mock.toEmail)
	require.Equal(t, opts.ToName, mock.toName)
	require.NotEmpty(t, mock.subject)
	require.Contains(t, mock.body, opts.Title)
	require.Contains(t, mock.body, opts.ExecutionTime.Format(time.RFC1123))
	for k, v := range opts.FailRow {
		require.Contains(t, mock.body, k)
		require.Contains(t, mock.body, fmt.Sprintf("%v", v))
	}
}
