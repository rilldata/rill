package file

import (
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
)

type ModelOutputProperties struct {
	Path   string             `mapstructure:"path"`
	Format drivers.FileFormat `mapstructure:"format"`
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
