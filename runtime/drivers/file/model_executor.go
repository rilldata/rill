package file

import (
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/driverutil"
)

type ModelOutputProperties struct {
	Path               string             `mapstructure:"path"`
	Format             drivers.FileFormat `mapstructure:"format"`
	Headers            []string           `mapstructure:"headers"`
	FileSizeLimitBytes int64              `mapstructure:"file_size_limit_bytes"`
	// Conditional formats for XLSX output.
	ConditionalFormats []driverutil.XLSXConditionalFormat `mapstructure:"conditional_formats"`
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
