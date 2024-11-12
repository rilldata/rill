package static

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/c2h5oh/datasize"
	"github.com/google/uuid"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/rilldata/rill/admin/provisioner"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/rilldata/rill/admin/database/postgres"
)

func TestProvision(t *testing.T) {
	pg := pgtestcontainer.New(t)
	defer pg.Terminate(t)

	db, err := database.Open("postgres", pg.DatabaseURL, "")
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	ctx := context.Background()
	require.NoError(t, db.Migrate(ctx))

	org, err := db.InsertOrganization(ctx, &database.InsertOrganizationOptions{
		Name: "test",
	})
	require.NoError(t, err)

	p1, err := db.InsertProject(ctx, &database.InsertProjectOptions{
		OrganizationID: org.ID,
		Name:           "p-q",
		ProdBranch:     "main",
		ProdSlots:      1,
	})
	require.NoError(t, err)

	// insert data
	_, err = db.InsertDeployment(ctx, &database.InsertDeploymentOptions{
		ProjectID:         p1.ID,
		Provisioner:       "static",
		ProvisionID:       uuid.NewString(),
		Slots:             2,
		Branch:            "main",
		RuntimeHost:       "host_1",
		RuntimeInstanceID: uuid.NewString(),
	})

	require.NoError(t, err)
	_, err = db.InsertDeployment(ctx, &database.InsertDeploymentOptions{
		ProjectID:         p1.ID,
		Provisioner:       "static",
		ProvisionID:       uuid.NewString(),
		Slots:             5,
		Branch:            "main",
		RuntimeHost:       "host_2",
		RuntimeInstanceID: uuid.NewString(),
	})
	require.NoError(t, err)

	_, err = db.InsertDeployment(ctx, &database.InsertDeploymentOptions{
		ProjectID:         p1.ID,
		Provisioner:       "static",
		ProvisionID:       uuid.NewString(),
		Slots:             4,
		Branch:            "main",
		RuntimeHost:       "host_3",
		RuntimeInstanceID: uuid.NewString(),
	})
	require.NoError(t, err)

	// spec
	spec := &StaticSpec{
		Runtimes: []*StaticRuntimeSpec{
			{Host: "host_1", Slots: 6},
			{Host: "host_2", Slots: 6},
			{Host: "host_3", Slots: 6},
		},
	}

	tests := []struct {
		name    string
		spec    *StaticSpec
		args    *provisioner.RuntimeArgs
		wantCfg *provisioner.RuntimeConfig
		wantErr bool
	}{
		{
			name:    "all applicable",
			spec:    spec,
			args:    &provisioner.RuntimeArgs{Slots: 1},
			wantCfg: &provisioner.RuntimeConfig{CPU: 1, MemoryGB: 4, StorageBytes: int64(40) * int64(datasize.GB)},
			wantErr: false,
		},
		{
			name:    "one applicable",
			spec:    spec,
			args:    &provisioner.RuntimeArgs{Slots: 4},
			wantCfg: &provisioner.RuntimeConfig{CPU: 4, MemoryGB: 16, StorageBytes: int64(160) * int64(datasize.GB), Host: "host_1"},
			wantErr: false,
		},
		{
			name:    "none applicable",
			spec:    spec,
			args:    &provisioner.RuntimeArgs{Slots: 5},
			wantCfg: nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specJSON, err := json.Marshal(tt.spec)
			require.NoError(t, err)

			p, err := NewStatic(specJSON, db, zap.NewNop())
			require.NoError(t, err)

			opts := &provisioner.ProvisionOptions{
				ID:   uuid.NewString(),
				Type: provisioner.ResourceTypeRuntime,
				Args: tt.args.AsMap(),
			}
			res, err := p.Provision(ctx, opts)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			// Since host assignment is random, if the host is not set in the expected config, we ignore it.
			if tt.wantCfg.Host == "" {
				require.NotEmpty(t, res.Config["host"])
				res.Config["host"] = ""
			}

			require.Equal(t, opts.ID, res.ID)
			require.Equal(t, opts.Type, res.Type)
			require.Equal(t, tt.wantCfg.AsMap(), res.Config)
		})
	}
}
