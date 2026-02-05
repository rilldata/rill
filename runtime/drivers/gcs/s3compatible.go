package gcs

import (
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers/s3"
)

// s3CompatibleConn is implemented to provide operations for GCS driver that is configured using S3 compatible credentials only.
type s3CompatibleConn struct {
	*s3.Connection
	config *ConfigProperties
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
