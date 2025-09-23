package gcs

import (
	"context"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/s3"
)

// s3CompatibleConn is a wrapper over s3.Connection.
// It is implemented to provide operations for GCS driver that is configured using S3 compatible credentials only.
type s3CompatibleConn struct {
	config *ConfigProperties
	s3Conn *s3.Connection
}

var _ drivers.ObjectStore = (*s3CompatibleConn)(nil)
var _ drivers.Handle = (*s3CompatibleConn)(nil)
var _ drivers.ModelManager = (*s3CompatibleConn)(nil)

// Ping implements drivers.Handle.
func (s *s3CompatibleConn) Ping(ctx context.Context) error {
	return s.s3Conn.Ping(ctx)
}

// AsAI implements drivers.Handle.
func (s *s3CompatibleConn) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

// AsAdmin implements drivers.Handle.
func (s *s3CompatibleConn) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Handle.
func (s *s3CompatibleConn) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsFileStore implements drivers.Handle.
func (s *s3CompatibleConn) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsInformationSchema implements drivers.Handle.
func (s *s3CompatibleConn) AsInformationSchema() (drivers.InformationSchema, bool) {
	return nil, false
}

// AsModelExecutor implements drivers.Handle.
func (s *s3CompatibleConn) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

// AsModelManager implements drivers.Handle.
func (s *s3CompatibleConn) AsModelManager(instanceID string) (drivers.ModelManager, bool) {
	return nil, false
}

// AsNotifier implements drivers.Handle.
func (s *s3CompatibleConn) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotImplemented
}

// AsOLAP implements drivers.Handle.
func (s *s3CompatibleConn) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

// AsObjectStore implements drivers.Handle.
func (s *s3CompatibleConn) AsObjectStore() (drivers.ObjectStore, bool) {
	return s, true
}

// AsRegistry implements drivers.Handle.
func (s *s3CompatibleConn) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Handle.
func (s *s3CompatibleConn) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsWarehouse implements drivers.Handle.
func (s *s3CompatibleConn) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

// Close implements drivers.Handle.
func (s *s3CompatibleConn) Close() error {
	return s.s3Conn.Close()
}

// Config implements drivers.Handle.
func (s *s3CompatibleConn) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(s.config, &m)
	return m
}

// Driver implements drivers.Handle.
func (s *s3CompatibleConn) Driver() string {
	return "gcs"
}

// Migrate implements drivers.Handle.
func (s *s3CompatibleConn) Migrate(ctx context.Context) error {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (s *s3CompatibleConn) MigrationStatus(ctx context.Context) (current int, desired int, err error) {
	return 0, 0, nil
}

// ObjectStore functions

// DownloadFiles implements drivers.ObjectStore.
func (s *s3CompatibleConn) DownloadFiles(ctx context.Context, path string) (drivers.FileIterator, error) {
	return s.s3Conn.DownloadFiles(ctx, rewriteToS3Path(path))
}

// ListObjects implements drivers.ObjectStore.
func (s *s3CompatibleConn) ListObjects(ctx context.Context, path string) ([]drivers.ObjectStoreEntry, error) {
	return s.s3Conn.ListObjects(ctx, rewriteToS3Path(path))
}

func rewriteToS3Path(s string) string {
	if after, ok := strings.CutPrefix(s, "gs://"); ok {
		return "s3://" + after
	}
	if after, ok := strings.CutPrefix(s, "gcs://"); ok {
		return "s3://" + after
	}
	return s
}

// ModelManager functions

// Delete implements drivers.ModelManager.
func (s *s3CompatibleConn) Delete(ctx context.Context, res *drivers.ModelResult) error {
	return s.s3Conn.Delete(ctx, res)
}

// Exists implements drivers.ModelManager.
func (s *s3CompatibleConn) Exists(ctx context.Context, res *drivers.ModelResult) (bool, error) {
	return s.s3Conn.Exists(ctx, res)
}

// MergePartitionResults implements drivers.ModelManager.
func (s *s3CompatibleConn) MergePartitionResults(a *drivers.ModelResult, b *drivers.ModelResult) (*drivers.ModelResult, error) {
	return s.s3Conn.MergePartitionResults(a, b)
}

// Rename implements drivers.ModelManager.
func (s *s3CompatibleConn) Rename(ctx context.Context, res *drivers.ModelResult, newName string, env *drivers.ModelEnv) (*drivers.ModelResult, error) {
	return s.s3Conn.Rename(ctx, res, newName, env)
}
