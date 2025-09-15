package provisioner

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

// ResourceType enumerates the provisionable resource types.
type ResourceType string

const (
	ResourceTypeRuntime    ResourceType = "runtime"
	ResourceTypeClickHouse ResourceType = "clickhouse"
)

func (r ResourceType) Valid() bool {
	switch r {
	case ResourceTypeRuntime, ResourceTypeClickHouse:
		return true
	}
	return false
}

// RuntimeArgs describe the expected arguments for provisioning a runtime resource.
type RuntimeArgs struct {
	Slots       int    `mapstructure:"slots"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

func NewRuntimeArgs(args map[string]any) (*RuntimeArgs, error) {
	res := &RuntimeArgs{}
	err := mapstructure.Decode(args, res)
	if err != nil {
		return nil, fmt.Errorf("failed to parse runtime args: %w", err)
	}
	if res.Slots < 1 {
		return nil, fmt.Errorf("runtime slots must be greater than 0 (received args: %v)", args)
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

// ClickhouseConfig describes the expected config for a provisioned Clickhouse resource.
type ClickhouseConfig struct {
	DSN      string `mapstructure:"dsn"`
	WriteDSN string `mapstructure:"write_dsn,omitempty"`
	Cluster  string `mapstructure:"cluster,omitempty"`
}

func NewClickhouseConfig(cfg map[string]any) (*ClickhouseConfig, error) {
	res := &ClickhouseConfig{}
	err := mapstructure.Decode(cfg, res)
	if err != nil {
		return nil, fmt.Errorf("failed to parse clickhouse config: %w", err)
	}
	return res, nil
}

func (c *ClickhouseConfig) AsMap() map[string]any {
	res := make(map[string]any)
	err := mapstructure.Decode(c, &res)
	if err != nil {
		panic(err)
	}
	return res
}
