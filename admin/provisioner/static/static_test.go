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
			wantCfg: &provisioner.RuntimeConfig{Host: "host_1", CPU: 1, MemoryGB: 4, StorageBytes: int64(40) * int64(datasize.GB)},
			wantErr: false,
		},
		{
			name:    "some applicable",
			spec:    spec,
			args:    &provisioner.RuntimeArgs{Slots: 6},
			wantCfg: &provisioner.RuntimeConfig{Host: "host_2", CPU: 6, MemoryGB: 24, StorageBytes: int64(240) * int64(datasize.GB)},
			wantErr: false,
		},
		{
			name:    "none applicable",
			spec:    spec,
			args:    &provisioner.RuntimeArgs{Slots: 8},
			wantCfg: nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specJSON, err := json.Marshal(tt.spec)
			require.NoError(t, err)

			p, err := NewStatic(specJSON, db, zap.NewNop())
			p.(*StaticProvisioner).nextIdx.Store(0) // Make host assignment deterministic
			require.NoError(t, err)

			in := &provisioner.Resource{
				ID:   uuid.NewString(),
				Type: provisioner.ResourceTypeRuntime,
			}
			opts := &provisioner.ResourceOptions{
				Args: tt.args.AsMap(),
			}
			out, err := p.Provision(ctx, in, opts)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, in.ID, out.ID)
			require.Equal(t, in.Type, out.Type)
			require.Equal(t, tt.wantCfg.AsMap(), out.Config)
		})
	}
}
