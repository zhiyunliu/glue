package registry

import (
	"context"
)

type RegistrarOptions interface {
	GetGroup() string
	GetCluster() string
}

// Registrar is service registrar.
type Registrar interface {
	Name() string

	ServerConfigs() string
	// Register the registration.
	Register(ctx context.Context, service *ServiceInstance) error
	// Deregister the registration.
	Deregister(ctx context.Context, service *ServiceInstance) error

	// GetService return the service instances in memory according to the service name.
	GetService(ctx context.Context, serviceName string, opts ...GetServiceOption) ([]*ServiceInstance, error)
	// Watch creates a watcher according to the service name.
	Watch(ctx context.Context, serviceName string) (Watcher, error)

	// GetAllServicesInfo return all services in memory.
	GetAllServicesInfo(ctx context.Context) (ServiceList, error)

	// GetOptions return the options of the registrar.
	GetOptions() RegistrarOptions

	// GetImpl return the implementation of the registrar.
	GetImpl() any
}

// Watcher is service watcher.
type Watcher interface {
	// Next returns services in the following two cases:
	// 1.the first time to watch and the service instance list is not empty.
	// 2.any service instance changes found.
	// if the above two conditions are not met, it will block until context deadline exceeded or canceled
	Next() ([]*ServiceInstance, error)
	// Stop close the watcher.
	Stop() error
}

// ServiceInstance is an instance of a service in a discovery system.
type ServiceInstance struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Version  string            `json:"version"`
	Metadata map[string]string `json:"metadata"`
	Weight   int64             `json:"weight"`
	//http://localhost:8000
	Endpoints []ServerItem `json:"endpoints"`
	Healthy   bool         `json:"healthy"`
}

type ServerItem struct {
	ServiceName string
	EndpointURL string //scheme://host:port/path
}

type ServiceList struct {
	Count    int64    `json:"count"`
	NameList []string `json:"name_list"`
}

type GetServiceOptions struct {
	HealthyOnly bool
}

type GetServiceOption func(*GetServiceOptions)

func WithHealthyOnly(healthyOnly bool) GetServiceOption {
	return func(o *GetServiceOptions) {
		o.HealthyOnly = healthyOnly
	}
}
