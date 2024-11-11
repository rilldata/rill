package provisioner

// ServiceType enumerates the provisionable service types.
type ServiceType string

const (
	ServiceTypeRuntime ServiceType = "runtime"
)

// RuntimeArgs describe the expected arguments for provisioning a runtime service.
type RuntimeArgs struct {
	Slots int
}

// RuntimeConfig describes the expected config for a provisioned runtime service.
type RuntimeConfig struct {
	Host         string
	Audience     string
	CPU          int
	MemoryGB     int
	StorageBytes int64
}
