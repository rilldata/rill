package email

import (
	"testing"

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
	client := New(mock, "https://example.com")

	opts := &OrganizationInvite{
		ToEmail:       uuid.New().String(),
		ToName:        uuid.New().String(),
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
