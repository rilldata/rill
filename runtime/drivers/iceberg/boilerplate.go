package iceberg

import (
	"context"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return false, nil
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, nil
}

func (c Connection) Migrate(ctx context.Context) error {
	return nil
}

func (c Connection) MigrationStatus(ctx context.Context) (current int, desired int, err error) {
	return 0, 0, nil
}

func (c Connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

func (c Connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

func (c Connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

func (c Connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

func (c Connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

func (c Connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

func (c Connection) AsSQLStore() (drivers.SQLStore, bool) {
	return nil, false
}

func (c Connection) AsTransporter(from drivers.Handle, to drivers.Handle) (drivers.Transporter, bool) {
	return nil, false
}
