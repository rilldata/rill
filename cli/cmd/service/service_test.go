package service

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
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/rilldata/rill/cli/pkg/mock"
	"github.com/rilldata/rill/cli/pkg/printer"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func TestServiceWorkflow(t *testing.T) {
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
	require.NoError(t, err)

	// Create service
	serviceName := "myservice"
	buf.Reset()
	p.SetHumanOutput(&buf)
	cmd = CreateCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{serviceName})
	err = cmd.Execute()
	require.NoError(t, err)

	expectedServiceMsg := []string{`Created service "myservice" in org "myorg".`}
	bufSlice := strings.Split(buf.String(), "\n")
	// Should we check for Access token as well?
	require.Equal(t, len(bufSlice), 3)
	require.Contains(t, bufSlice[0], expectedServiceMsg[0])

	// Create one more service in same org
	cmd = CreateCmd(helper)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"myservice1"})
	err = cmd.Execute()
	require.NoError(t, err)

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
	require.NoError(t, err)
	expectedServices := []string{"myservice", "myservice1"}
	serviceList := []Service{}
	err = json.Unmarshal([]byte(buf.String()), &serviceList)
	require.NoError(t, err)
	require.Equal(t, len(serviceList), 2)
	for _, service := range serviceList {
		require.Contains(t, expectedServices, service.Name)
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
	require.NoError(t, err)
	expectedMsg := fmt.Sprintf("Deleted service: %q\n", serviceName)
	require.Equal(t, buf.String(), expectedMsg)

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
	require.NoError(t, err)
	serviceList = []Service{}
	err = json.Unmarshal([]byte(buf.String()), &serviceList)
	require.NoError(t, err)
	require.Equal(t, len(serviceList), 1)
	expectedServices = []string{"myservice1"}
	for _, service := range serviceList {
		require.Contains(t, expectedServices, service.Name)
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
	require.NoError(t, err)
	expectedServices = []string{"myservice2"}
	serviceList = []Service{}
	err = json.Unmarshal([]byte(buf.String()), &serviceList)
	require.NoError(t, err)
	require.Equal(t, len(serviceList), 1)
	for _, service := range serviceList {
		require.Contains(t, expectedServices, service.Name)
	}
}

type Service struct {
	Name    string `json:"name"`
	OrgName string `json:"org_name"`
}
