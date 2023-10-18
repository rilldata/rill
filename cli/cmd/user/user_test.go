package user

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/rilldata/rill/cli/cmd/org"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/mock"
	"github.com/rilldata/rill/cli/pkg/printer"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func TestUserWorkflow(t *testing.T) {
	c := qt.New(t)
	pg := pgtestcontainer.New(t)
	defer pg.Terminate(t)

	ctx := context.Background()
	logger, _ := zap.NewDevelopment()

	// Get Admin service
	adm, err := mock.AdminService(ctx, logger, pg.DatabaseURL)
	defer adm.Close()
	c.Assert(err, qt.IsNil)
	db := adm.DB

	// create mock admin user
	adminUser, err := db.InsertUser(ctx, &database.InsertUserOptions{
		Email:               "admin@test.io",
		DisplayName:         "admin",
		QuotaSingleuserOrgs: 3,
	})
	c.Assert(err, qt.IsNil)

	// issue admin and viewer tokens
	adminAuthToken, err := adm.IssueUserAuthToken(ctx, adminUser.ID, database.AuthClientIDRillWeb, "test", nil, nil)
	c.Assert(err, qt.IsNil)
	c.Assert(adminAuthToken, qt.Not(qt.IsNil))

	// Create mock admin server
	srv, err := mock.AdminServer(ctx, logger, adm)
	c.Assert(err, qt.IsNil)

	// Make errgroup for running the processes
	ctx = graceful.WithCancelOnTerminate(ctx)
	group, cctx := errgroup.WithContext(ctx)

	group.Go(func() error { return srv.ServeGRPC(cctx) })
	group.Go(func() error { return srv.ServeHTTP(cctx) })
	// time.Sleep(15 * time.Second)
	err = mock.CheckServerStatus(srv)
	c.Assert(err, qt.IsNil)

	var buf bytes.Buffer
	format := printer.Human
	p := printer.NewPrinter(&format)
	p.SetResourceOutput(&buf)
	helper := &cmdutil.Helper{
		Config: &config.Config{
			AdminURL:          "http://localhost:9090",
			AdminTokenDefault: adminAuthToken.Token().String(),
		},
		Printer: p,
	}

	// Create organization for testing
	orgName := "test-org"
	cmd := org.CreateCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--name", orgName})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)

	// Add user to organization
	buf.Reset()
	p.SetHumanOutput(&buf)
	cmd = AddCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--org", orgName, "--email", "test@rilldata.com", "--role", "admin"})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)

	expectedMsg := fmt.Sprintf("Invitation sent to %q to join organization %q as %q", "test@rilldata.com", orgName, "admin")
	c.Assert(buf.String(), qt.Contains, expectedMsg)

	// List users in organization
	buf.Reset()
	format = printer.JSON
	p = printer.NewPrinter(&format)
	p.SetResourceOutput(&buf)
	helper.Printer = p
	cmd = ListCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--org", orgName})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)

	fmt.Println("buf.String(): ", buf.String())
	invites := strings.Split(buf.String(), "\n\n")[1]
	inviteList := []Invites{}
	err = json.Unmarshal([]byte(invites), &inviteList)
	c.Assert(err, qt.IsNil)
	expectedInviteList := []Invites{
		{
			Email:     "test@rilldata.com",
			RoleName:  "admin",
			InvitedBy: "admin@test.io",
		},
	}
	c.Assert(inviteList, qt.DeepEquals, expectedInviteList)

	// Add one more user to organization
	buf.Reset()
	p.SetHumanOutput(&buf)
	cmd = AddCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--org", orgName, "--email", "test1@rilldata.com", "--role", "admin"})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)
	expectedMsg = fmt.Sprintf("Invitation sent to %q to join organization %q as %q", "test1@rilldata.com", orgName, "admin")
	c.Assert(buf.String(), qt.Contains, expectedMsg)

	// List invites again in same organization
	buf.Reset()
	format = printer.JSON
	p = printer.NewPrinter(&format)
	p.SetResourceOutput(&buf)
	helper.Printer = p
	cmd = ListCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--org", orgName})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)

	expectedInviteList = append(expectedInviteList, Invites{
		Email:     "test1@rilldata.com",
		RoleName:  "admin",
		InvitedBy: "admin@test.io",
	})
	for _, invite := range inviteList {
		c.Assert(expectedInviteList, qt.Contains, invite)
	}

	// Remove user from organization
	buf.Reset()
	p.SetHumanOutput(&buf)
	cmd = RemoveCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--org", orgName, "--email", "test1@rilldata.com"})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)
	expectedMsg = fmt.Sprintf("Removed user %q from organization %q", "test1@rilldata.com", orgName)
	c.Assert(buf.String(), qt.Contains, expectedMsg)

	// List invites again in same organization after removing user
	buf.Reset()
	format = printer.JSON
	p = printer.NewPrinter(&format)
	p.SetResourceOutput(&buf)
	helper.Printer = p
	cmd = ListCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--org", orgName})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)

	expectedInviteList = []Invites{
		{
			Email:     "test@rilldata.com",
			RoleName:  "admin",
			InvitedBy: "admin@test.io",
		},
	}
	for _, invite := range inviteList {
		c.Assert(expectedInviteList, qt.Contains, invite)
	}

	// Set user role in organization
	buf.Reset()
	p.SetHumanOutput(&buf)
	cmd = SetRoleCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--org", orgName, "--email", "test@rilldata.com", "--role", "viewer"})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)
	expectedMsg = fmt.Sprintf("Updated role of user %q to %q in the organization %q", "test@rilldata.com", "viewer", orgName)
	c.Assert(buf.String(), qt.Contains, expectedMsg)
}

type Invites struct {
	Email     string `json:"email"`
	RoleName  string `json:"role_name"`
	InvitedBy string `json:"invited_by"`
}
