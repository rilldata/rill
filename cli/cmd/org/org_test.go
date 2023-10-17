package org

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-github/v50/github"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/dotrill"
	"github.com/rilldata/rill/cli/pkg/mock"
	"github.com/rilldata/rill/cli/pkg/printer"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func TestOrganizationWorkflow(t *testing.T) {
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
	time.Sleep(15 * time.Second)

	var buf bytes.Buffer
	format := printer.JSON
	p := printer.NewPrinter(&format)
	p.SetResourceOutput(&buf)
	helper := &cmdutil.Helper{
		Config: &config.Config{
			AdminURL:          "http://localhost:9090",
			AdminTokenDefault: adminAuthToken.Token().String(),
		},
		Printer: p,
	}

	// Create organization with name
	cmd := CreateCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--name", "myorg"})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)

	orgList := []Org{}
	err = json.Unmarshal([]byte(buf.String()), &orgList)
	c.Assert(err, qt.IsNil)
	c.Assert(len(orgList), qt.Equals, 1)
	c.Assert(orgList[0].Name, qt.Equals, "myorg")

	// Create new organization with name
	buf.Reset()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--name", "test"})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)
	err = json.Unmarshal([]byte(buf.String()), &orgList)
	c.Assert(err, qt.IsNil)
	c.Assert(len(orgList), qt.Equals, 1)
	c.Assert(orgList[0].Name, qt.Equals, "test")

	// List organizations
	buf.Reset()
	cmd = ListCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)

	err = json.Unmarshal([]byte(buf.String()), &orgList)
	c.Assert(err, qt.IsNil)
	c.Assert(len(orgList), qt.Equals, 2)

	// 1 more way to check org list
	// eq := !reflect.DeepEqual(expectedOrgs, orgList)
	// c.Assert(eq, qt.Equals, false)
	expectedOrgs := []string{"myorg", "test"}
	for _, org := range orgList {
		c.Assert(expectedOrgs, qt.Contains, org.Name)
	}

	// Delete organization
	buf.Reset()
	cmd = DeleteCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--org", "myorg", "--force"})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)

	// List organizations
	buf.Reset()
	cmd = ListCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)

	orgList = []Org{}
	err = json.Unmarshal([]byte(buf.String()), &orgList)
	c.Assert(err, qt.IsNil)
	c.Assert(len(orgList), qt.Equals, 1)
	expectedOrgs = []string{"test"}
	for _, org := range orgList {
		c.Assert(expectedOrgs, qt.Contains, org.Name)
	}

	// rename organization
	buf.Reset()
	cmd = RenameCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--org", "test", "--new-name", "new-test", "--force"})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)

	err = json.Unmarshal([]byte(buf.String()), &orgList)
	c.Assert(err, qt.IsNil)
	c.Assert(len(orgList), qt.Equals, 1)
	c.Assert(orgList[0].Name, qt.Equals, "new-test")

	// Switch organization
	buf.Reset()
	helper.Printer.SetHumanOutput(&buf)
	cmd = SwitchCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"new-test"})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)

	expectedMsg := fmt.Sprintf("Set default organization to %q.\n", "new-test")
	c.Assert(buf.String(), qt.Contains, expectedMsg)
	org, err := dotrill.GetDefaultOrg()
	c.Assert(err, qt.IsNil)
	c.Assert(org, qt.Equals, "new-test")

}

type Org struct {
	Name string `json:"Name"`
}

// mockGithub provides a mock implementation of admin.Github.
type mockGithub struct{}

func (m *mockGithub) AppClient() *github.Client {
	return nil
}

func (m *mockGithub) InstallationClient(installationID int64) (*github.Client, error) {
	return nil, nil
}

func (m *mockGithub) InstallationToken(ctx context.Context, installationID int64) (string, error) {
	return "", nil
}
