package clickhouse

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/s3"
)

type s3ToSelfExecutor struct {
	s3   drivers.Handle
	c    *connection
	opts *drivers.ModelExecutorOptions
}

type inputProps struct {
	Path string `mapstructure:"path"`
}

func (p *inputProps) Validate() error {
	if p.Path == "" {
		return fmt.Errorf("path is mandatory for s3 input connector")
	}
	return nil
}

func (e *s3ToSelfExecutor) Execute(ctx context.Context) (*drivers.ModelResult, error) {
	iProps := &inputProps{}
	if err := mapstructure.WeakDecode(e.opts.InputProperties, iProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if err := iProps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input properties: %w", err)
	}

	sql, err := e.genSQL(iProps.Path)
	if err != nil {
		return nil, err
	}
	props := &ModelInputProperties{SQL: sql}
	propsMap := make(map[string]any)
	if err := mapstructure.Decode(props, &propsMap); err != nil {
		return nil, err
	}
	// May be its not safe to update input props ?
	e.opts.InputProperties = propsMap

	executor := &selfToSelfExecutor{c: e.c, opts: e.opts}
	return executor.Execute(ctx)
}

func (e *s3ToSelfExecutor) genSQL(path string) (string, error) {
	props := &s3.ConfigProperties{}
	if err := mapstructure.Decode(e.s3.Config(), props); err != nil {
		return "", err
	}

	// SELECT * FROM S3(path, [id, secret], format)
	var sb strings.Builder
	sb.WriteString("SELECT * FROM s3(")
	sb.WriteString(fmt.Sprintf("%q", path))
	if props.AccessKeyID != "" {
		sb.WriteString(", ")
		sb.WriteString(props.AccessKeyID)
		sb.WriteString(", ")
		sb.WriteString(props.SecretAccessKey)
	}
	sb.WriteString(")")
	return sb.String(), nil
}
