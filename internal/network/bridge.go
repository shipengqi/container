package network

import (
	"net"

	"github.com/pkg/errors"
)

type BridgeNetworkDriver struct {

}

func (b *BridgeNetworkDriver) Create(subnet, name string) (*Network, error) {
	// 通过 net.ParseCIDR 方法，获取网段的字符串中网关 IP 地址和网络 IP 段
	// ParseCIDR parses s as a CIDR notation IP address and prefix length,
	// like "192.0.2.0/24" or "2001:db8::/32", as defined in
	// RFC 4632 and RFC 4291.
	//
	// It returns the IP address and the network implied by the IP and
	// prefix length.
	// For example, ParseCIDR("192.0.2.1/24") returns the IP address
	// 192.0.2.1 and the network 192.0.2.0/24.
	ip, ipRange, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, err
	}

	n := &Network{
		Name:    name,
		IpRange: ipRange,
	}

	// 配置 linux bridge
	err = b.initBridge(n)
	if err != nil {
		return nil, errors.Wrap(err, "init bridge")
	}
	return n, nil
}

// initBridge 初始化网桥
// 1. 创建 bridge 虚拟设备
// 2. 设置 bridge 设备地址和路由
// 3. 启动 bridge 设备
// 4. 设置 iptables SNAT 规则
func (b *BridgeNetworkDriver) initBridge(n *Network) error {
	// 1. 创建 bridge 虚拟设备
	bridgeName := n.Name
	if err := createBridgeInterface(bridgeName); err != nil {
		return errors.Wrap(err, "create bridge")
	}
	// 2. 设置 bridge 设备地址和路由
	gatewayIP := *n.IpRange
	gatewayIP.IP = n.IpRange.IP
	if err := setInterfaceIP(bridgeName, gatewayIP.String()); err != nil {
		return errors.Wrap(err, "create bridge")
	}
	return nil
}
