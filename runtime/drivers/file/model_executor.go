package file

import "fmt"

type ModelOutputProperties struct {
	Path   string `mapstructure:"path"`
	Format string `mapstructure:"format"`
}

func (p *ModelOutputProperties) Validate() error {
	if p.Path == "" {
		return fmt.Errorf("missing property 'path'")
	}
	if p.Format == "" {
		return fmt.Errorf("missing property 'format'")
	}
	return nil
}

type ModelResultProperties struct {
	Path   string `mapstructure:"path"`
	Format string `mapstructure:"format"`
}
