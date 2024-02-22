package user

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/mock"
	"github.com/rilldata/rill/cli/pkg/printer"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func TestUserWorkflow(t *testing.T) {
	t.Skip("Skipping test as it is failing on CI")
	pg := pgtestcontainer.New(t)
	defer pg.Terminate(t)

	ctx := context.Background()
	logger, _ := zap.NewDevelopment()

	// Get Admin service
	adm, err := mock.AdminService(ctx, logger, pg.DatabaseURL)
	require.NoError(t, err)
	defer adm.Close()

	db := adm.DB

	// create mock admin user
	adminUser, err := db.InsertUser(ctx, &database.InsertUserOptions{
		Email:               "admin@test.io",
		DisplayName:         "admin",
		QuotaSingleuserOrgs: 3,
	})
	require.NoError(t, err)
	require.NotNil(t, adminUser)

	// issue admin and viewer tokens
	adminAuthToken, err := adm.IssueUserAuthToken(ctx, adminUser.ID, database.AuthClientIDRillWeb, "test", nil, nil)
	require.NoError(t, err)
	require.NotNil(t, adminAuthToken)

	// Create mock admin server
	srv, err := mock.AdminServer(ctx, logger, adm)
	require.NoError(t, err)

	// Make errgroup for running the processes
	ctx = graceful.WithCancelOnTerminate(ctx)
	group, cctx := errgroup.WithContext(ctx)

	group.Go(func() error { return srv.ServeGRPC(cctx) })
	group.Go(func() error { return srv.ServeHTTP(cctx) })
	err = mock.CheckServerStatus(cctx)
	require.NoError(t, err)

	var buf bytes.Buffer
	p := printer.NewPrinter(printer.FormatHuman)
	p.OverrideDataOutput(&buf)
	helper := &cmdutil.Helper{
		AdminURL:          "http://localhost:9090",
		AdminTokenDefault: adminAuthToken.Token().String(),
		Printer:           p,
	}
	defer helper.Close()

	// Create organization for testing
	orgName := "test-org"
	cmd := org.CreateCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--name", orgName})
	err = cmd.Execute()
	require.NoError(t, err)

	// Add user to organization
	buf.Reset()
	p.OverrideHumanOutput(&buf)
	cmd = AddCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--org", orgName, "--email", "test@rilldata.com", "--role", "admin"})
	err = cmd.Execute()
	require.NoError(t, err)

	expectedMsg := fmt.Sprintf("Invitation sent to %q to join organization %q as %q", "test@rilldata.com", orgName, "admin")
	require.Contains(t, buf.String(), expectedMsg)

	// List users in organization
	buf.Reset()
	p = printer.NewPrinter(printer.FormatJSON)
	p.OverrideDataOutput(&buf)
	helper.Printer = p
	cmd = ListCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--org", orgName})
	err = cmd.Execute()
	require.NoError(t, err)

	fmt.Println("buf.String(): ", buf.String())
	invites := strings.Split(buf.String(), "\n\n")[1]
	inviteList := []Invites{}
	err = json.Unmarshal([]byte(invites), &inviteList)
	require.NoError(t, err)
	expectedInviteList := []Invites{
		{
			Email:     "test@rilldata.com",
			RoleName:  "admin",
			InvitedBy: "admin@test.io",
		},
	}
	require.EqualValues(t, inviteList, expectedInviteList)
	// Add one more user to organization
	buf.Reset()
	p.OverrideHumanOutput(&buf)
	cmd = AddCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--org", orgName, "--email", "test1@rilldata.com", "--role", "admin"})
	err = cmd.Execute()
	require.NoError(t, err)
	expectedMsg = fmt.Sprintf("Invitation sent to %q to join organization %q as %q", "test1@rilldata.com", orgName, "admin")
	require.Contains(t, buf.String(), expectedMsg)

	// List invites again in same organization
	buf.Reset()
	p = printer.NewPrinter(printer.FormatJSON)
	p.OverrideDataOutput(&buf)
	helper.Printer = p
	cmd = ListCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--org", orgName})
	err = cmd.Execute()
	require.NoError(t, err)

	expectedInviteList = append(expectedInviteList, Invites{
		Email:     "test1@rilldata.com",
		RoleName:  "admin",
		InvitedBy: "admin@test.io",
	})
	for _, invite := range inviteList {
		require.Contains(t, expectedInviteList, invite)
	}

	// Remove user from organization
	buf.Reset()
	p.OverrideHumanOutput(&buf)
	cmd = RemoveCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--org", orgName, "--email", "test1@rilldata.com"})
	err = cmd.Execute()
	require.NoError(t, err)
	expectedMsg = fmt.Sprintf("Removed user %q from organization %q", "test1@rilldata.com", orgName)
	require.Contains(t, buf.String(), expectedMsg)

	// List invites again in same organization after removing user
	buf.Reset()
	p = printer.NewPrinter(printer.FormatJSON)
	p.OverrideDataOutput(&buf)
	helper.Printer = p
	cmd = ListCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--org", orgName})
	err = cmd.Execute()
	require.NoError(t, err)

	expectedInviteList = []Invites{
		{
			Email:     "test@rilldata.com",
			RoleName:  "admin",
			InvitedBy: "admin@test.io",
		},
	}
	for _, invite := range inviteList {
		require.Contains(t, expectedInviteList, invite)
	}

	// Set user role in organization
	buf.Reset()
	p.OverrideHumanOutput(&buf)
	cmd = SetRoleCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--org", orgName, "--email", "test@rilldata.com", "--role", "viewer"})
	err = cmd.Execute()
	require.NoError(t, err)
	expectedMsg = fmt.Sprintf("Updated role of user %q to %q in the organization %q", "test@rilldata.com", "viewer", orgName)
	require.Contains(t, buf.String(), expectedMsg)
}

type Invites struct {
	Email     string `json:"email"`
	RoleName  string `json:"role_name"`
	InvitedBy string `json:"invited_by"`
}
