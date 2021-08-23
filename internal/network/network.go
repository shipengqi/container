package network

import (
	"net"

	"github.com/vishvananda/netlink"
)

// Network 网络是容器的一个集合，在这个网络上的容器可以通过这个网络互相通信，就像挂载到同一个 Linux Bridge 设备上的网络设备一样，可以直接
// 通过 Bridge 设备实现网络互连。连接到同一个网络中的容器也可以通过这个网络和网络中别的容器互连。
// Network 包括这个网络相关的配置，比如网络的容器地址段、网络操作所调用的网络驱动等信息。
type Network struct {
	Name    string // network name
	Driver  string // network driver name
	IpRange *net.IPNet
}

// Endpoint 网络端点是用于连接容器与网络的，保证容器内部与网络的通信。像 Veth 设备，一端挂载到容器内部，另一端挂载到 Bridge 上，
// 就能保证容器和网络的通信。
// Endpoint 包括连接到网络的一些信息，比如 IP、Veth 设备、端口映射、连接的容器和网络等信息。
type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"dev"`
	IPAddress   net.IP           `json:"ip"`
	MacAddress  net.HardwareAddr `json:"mac"`
	Network     *Network
	PortMapping []string
}

// Driver 网络驱动是一个网络功能中的组件， 不同的驱动对网络的创建、连接、
// 销毁的策略不同，通过在创建网络时指定不同的网络驱动来定义使用哪个驱动做网络的配置。
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
