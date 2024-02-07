package provisioner

import (
	"context"
	"testing"

	"github.com/c2h5oh/datasize"
	"github.com/google/uuid"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/pgtestcontainer"
	"github.com/stretchr/testify/require"

	_ "github.com/rilldata/rill/admin/database/postgres"
)

func Test_staticProvisioner_Provision(t *testing.T) {
	pg := pgtestcontainer.New(t)
	defer pg.Terminate(t)

	db, err := database.Open("postgres", pg.DatabaseURL)
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
		opts    *ProvisionOptions
		want    *Allocation
		wantErr bool
	}{
		{
			name:    "all applicable ",
			spec:    spec,
			opts:    &ProvisionOptions{OLAPDriver: "duckdb", Slots: 1},
			want:    &Allocation{CPU: 1, MemoryGB: 2, StorageBytes: int64(40) * int64(datasize.GB)},
			wantErr: false,
		},
		{
			name:    "one applicable ",
			spec:    spec,
			opts:    &ProvisionOptions{OLAPDriver: "duckdb", Slots: 4},
			want:    &Allocation{CPU: 4, MemoryGB: 8, StorageBytes: int64(160) * int64(datasize.GB), Host: "host_1"},
			wantErr: false,
		},
		{
			name:    "none applicable ",
			spec:    spec,
			opts:    &ProvisionOptions{OLAPDriver: "duckdb", Slots: 5},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &StaticProvisioner{
				Spec: tt.spec,
				db:   db,
			}
			got, err := p.Provision(ctx, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("staticProvisioner.Provision() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareAllocation(got, tt.want) {
				t.Errorf("staticProvisioner.Provision() = %v, want %v", got, tt.want)
			}
		})
	}
}

func compareAllocation(got, want *Allocation) bool {
	if (got != nil) != (want != nil) {
		return false
	}

	if got == nil {
		return true
	}

	if want.Host != "" && want.Host != got.Host {
		return false
	}

	return got.CPU == want.CPU && got.MemoryGB == want.MemoryGB && got.StorageBytes == want.StorageBytes
}
