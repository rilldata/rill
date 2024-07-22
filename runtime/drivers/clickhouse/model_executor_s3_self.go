package clickhouse

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
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
	// Build the model executor options with updated input properties
	opts := &drivers.ModelExecutorOptions{
		Env:              e.opts.Env,
		ModelName:        e.opts.ModelName,
		InputConnector:   e.opts.InputConnector,
		InputProperties:  propsMap,
		OutputConnector:  e.opts.OutputConnector,
		OutputProperties: e.opts.OutputProperties,
		Priority:         e.opts.Priority,
	}
	executor := &selfToSelfExecutor{c: e.c, opts: opts}
	return executor.Execute(ctx)
}

func (e *s3ToSelfExecutor) genSQL(path string) (string, error) {
	props := &s3ConfigProperties{}
	if err := mapstructure.Decode(e.s3.Config(), props); err != nil {
		return "", err
	}

	// SELECT * FROM S3(path, [id, secret], format)
	var sb strings.Builder
	sb.WriteString("SELECT * FROM s3(")
	sb.WriteString(fmt.Sprintf("'%s'", path))
	if props.AccessKeyID != "" {
		sb.WriteString(", ")
		sb.WriteString(fmt.Sprintf("'%s'", props.AccessKeyID))
		sb.WriteString(", ")
		sb.WriteString(fmt.Sprintf("'%s'", props.SecretAccessKey))
	}
	sb.WriteString(")")
	return sb.String(), nil
}

type s3ConfigProperties struct {
	AccessKeyID     string `mapstructure:"aws_access_key_id"`
	SecretAccessKey string `mapstructure:"aws_secret_access_key"`
	SessionToken    string `mapstructure:"aws_access_token"`
}
