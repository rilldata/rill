package s3

import (
	"context"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

type warehouseToSelfExecutor struct {
	w    drivers.Warehouse
	c    *Connection
	opts *drivers.ModelExecutorOptions
}

var _ drivers.ModelExecutor = &warehouseToSelfExecutor{}

func (e *warehouseToSelfExecutor) Execute(ctx context.Context) (*drivers.ModelResult, error) {
	props := &ModelOutputProperties{}
	if err := mapstructure.Decode(e.opts.OutputProperties, props); err != nil {
		return nil, err
	}
	export, err := e.w.Export(ctx, e.opts.InputProperties, e.c, props.Path)
	if err != nil {
		return nil, err
	}
	outputGlob, err := export.Glob()
	if err != nil {
		return nil, err
	}
	resProps := &ModelResultProperties{Path: outputGlob}
	res := make(map[string]any)
	err = mapstructure.Decode(resProps, &res)
	if err != nil {
		return nil, err
	}

	return &drivers.ModelResult{
		Connector:  e.opts.OutputConnector,
		Properties: res,
	}, nil
}
