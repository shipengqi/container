package driver

// Interface network driver interface
type Interface interface {
	// Name return name of driver
	Name() string
	Create(subnet string, name string) (*Network, error)
	Delete(network string) error
	// Connect container Endpoint to Network
	Connect(network string, endpoint *Endpoint) error
	// Disconnect remove container Endpoint in Network
	Disconnect(network string, endpoint *Endpoint) error
}
