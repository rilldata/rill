package service

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

func TestServiceWorkflow(t *testing.T) {
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
			Org:               "myorg",
		},
		Printer: p,
	}

	// Create Organization
	cmd := org.CreateCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--name", "myorg"})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)

	// Create service
	serviceName := "myservice"
	buf.Reset()
	p.SetHumanOutput(&buf)
	cmd = CreateCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{serviceName})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)

	expectedServiceMsg := []string{`Created service "myservice" in org "myorg".`}
	bufSlice := strings.Split(buf.String(), "\n")
	// Should we check for Access token as well?
	c.Assert(bufSlice, qt.HasLen, 2)
	c.Assert(bufSlice[0], qt.Contains, expectedServiceMsg[0])

	// Create one more service in same org
	cmd = CreateCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"myservice1"})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)

	// List service in org
	buf.Reset()
	format = printer.JSON
	p = printer.NewPrinter(&format)
	p.SetResourceOutput(&buf)

	cmd = ListCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)
	expectedServices := []string{"myservice", "myservice1"}
	serviceList := []Service{}
	err = json.Unmarshal([]byte(buf.String()), &serviceList)
	c.Assert(err, qt.IsNil)
	c.Assert(serviceList, qt.HasLen, 2)
	for _, service := range serviceList {
		c.Assert(expectedServices, qt.Contains, service.Name)
	}

	// Delete service
	buf.Reset()
	p.SetHumanOutput(&buf)
	cmd = DeleteCmd(helper)
	cmd.UsageString()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"myservice"})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)
	expectedMsg := fmt.Sprintf("Deleted service: %q", serviceName)
	c.Assert(buf.String(), qt.Equals, expectedMsg)

	// List service in org after delete
	buf.Reset()
	format = printer.JSON
	p = printer.NewPrinter(&format)
	p.SetResourceOutput(&buf)
	cmd = ListCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)
	serviceList = []Service{}
	err = json.Unmarshal([]byte(buf.String()), &serviceList)
	c.Assert(err, qt.IsNil)
	c.Assert(serviceList, qt.HasLen, 1)
	expectedServices = []string{"myservice1"}
	for _, service := range serviceList {
		c.Assert(expectedServices, qt.Contains, service.Name)
	}

	// Rename service
	buf.Reset()
	helper.Printer.SetHumanOutput(nil)
	helper.Printer.SetResourceOutput(&buf)
	cmd = RenameCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"myservice1", "--new-name", "myservice2"})
	err = cmd.Execute()
	c.Assert(err, qt.IsNil)
	expectedServices = []string{"myservice2"}
	serviceList = []Service{}
	err = json.Unmarshal([]byte(buf.String()), &serviceList)
	c.Assert(err, qt.IsNil)
	c.Assert(serviceList, qt.HasLen, 1)
	for _, service := range serviceList {
		c.Assert(expectedServices, qt.Contains, service.Name)
	}
}

type Service struct {
	Name    string `json:"name"`
	OrgName string `json:"org_name"`
}
