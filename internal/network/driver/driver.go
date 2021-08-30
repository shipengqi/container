package driver

// Driver network driver
type Driver interface {
	// Name return name of driver
	Name() string
	Create(subnet string, name string) (*Network, error)
	Delete(network Network) error
	// Connect container Endpoint to Network
	Connect(network *Network, endpoint *Endpoint) error
	// Disconnect remove container Endpoint in Network
	Disconnect(network Network, endpoint *Endpoint) error
}
