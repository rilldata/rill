package provisioner

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

// ResourceType enumerates the provisionable resource types.
type ResourceType string

const (
	ResourceTypeRuntime ResourceType = "runtime"
)

// RuntimeArgs describe the expected arguments for provisioning a runtime resource.
type RuntimeArgs struct {
	Slots   int    `mapstructure:"slots"`
	Version string `mapstructure:"version"`
}

func NewRuntimeArgs(args map[string]any) (*RuntimeArgs, error) {
	res := &RuntimeArgs{}
	err := mapstructure.Decode(args, res)
	if err != nil {
		return nil, fmt.Errorf("failed to parse runtime args: %w", err)
	}
	return res, nil
}

func (r *RuntimeArgs) AsMap() map[string]any {
	res := make(map[string]any)
	err := mapstructure.Decode(r, &res)
	if err != nil {
		panic(err)
	}
	return res
}

// RuntimeConfig describes the expected config for a provisioned runtime resource.
type RuntimeConfig struct {
	Host         string `mapstructure:"host"`
	Audience     string `mapstructure:"audience"`
	CPU          int    `mapstructure:"cpu"`
	MemoryGB     int    `mapstructure:"memory_gb"`
	StorageBytes int64  `mapstructure:"storage_bytes"`
}

func NewRuntimeConfig(cfg map[string]any) (*RuntimeConfig, error) {
	res := &RuntimeConfig{}
	err := mapstructure.Decode(cfg, res)
	if err != nil {
		return nil, fmt.Errorf("failed to parse runtime config: %w", err)
	}
	return res, nil
}

func (r *RuntimeConfig) AsMap() map[string]any {
	res := make(map[string]any)
	err := mapstructure.Decode(r, &res)
	if err != nil {
		panic(err)
	}
	return res
}
