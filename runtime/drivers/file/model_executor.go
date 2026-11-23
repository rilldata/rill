package file

import (
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/driverutil"
)

type ModelOutputProperties struct {
	Path               string             `mapstructure:"path"`
	Format             drivers.FileFormat `mapstructure:"format"`
	FileSizeLimitBytes int64              `mapstructure:"file_size_limit_bytes"`
	// The header lines, and for XLSX output the conditional formats, number formats and time columns.
	driverutil.FileOptions `mapstructure:",squash"`
}

func (p *ModelOutputProperties) Validate() error {
	if p.Path == "" {
		return fmt.Errorf("missing property 'path'")
	}
	if p.Format == "" {
		return fmt.Errorf("missing property 'format'")
	} else if !p.Format.Valid() {
		return fmt.Errorf("invalid property 'format': %q", p.Format)
	}
	return nil
}

type ModelResultProperties struct {
	Path   string             `mapstructure:"path"`
	Format drivers.FileFormat `mapstructure:"format"`
}
