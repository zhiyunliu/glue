package balancer

type node struct {
	addr        string
	serviceName string
}

func (n node) Address() string {
	return n.addr
}

// ServiceName is service name
func (n node) ServiceName() string {
	return n.serviceName
}

// InitialWeight is the initial value of scheduling weight
// if not set return nil
func (n node) InitialWeight() *int64 {
	return nil
}

// Version is service node version
func (n node) Version() string {
	return "0.0.0.0"
}

// Metadata is the kv pair metadata associated with the service instance.
// version,namespace,region,protocol etc..
func (n node) Metadata() map[string]string {
	return map[string]string{}
}
