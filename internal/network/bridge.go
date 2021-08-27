package network

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"

	"github.com/shipengqi/container/pkg/log"
)

type BridgeNetworkDriver struct {
	name string
}

func (b *BridgeNetworkDriver) Name() string {
	return "bridge"
}

// Connect 连接一个网络和网络端点
func (b *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	// 获取网络名
	bridgeName := network.Name
	// 获取到 Linux Bridge 接口的对象和接口属性
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	// 创建 Veth 接口的配置
	la := netlink.NewLinkAttrs()
	// 由于 Linux 接口名的限制，名字取 endpoint ID 的前 5 位
	la.Name = endpoint.ID[:5]
	// 通过设置 Veth 接口的 master 属性，设置这个 Veth 的一端挂载到网络对应的 Linux Bridge 上
	la.MasterIndex = br.Attrs().Index

	// 创建 Veth 对象，通过 PeerName 配置 Veth 另外一端的接口名
	// 配置 Veth 另外一端的名字 cif-{endpoint ID 的前 5 位}
	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + endpoint.ID[:5],
	}

	// LinkAdd 创建这个 Veth 接口
	// 因为上面指定了 link 的 Master Index 是网络对应的 Linux Bridge
	// 所以 Veth 的一端就己经挂载到了网络对应的 Linux Bridge 上
	if err = netlink.LinkAdd(&endpoint.Device); err != nil {
		return errors.Errorf("Add Endpoint Device: %v", err)
	}

	// 启动 Veth 设备，相当于 `ip link set xxx up`
	if err = netlink.LinkSetUp(&endpoint.Device); err != nil {
		return errors.Errorf("Add Endpoint Device: %v", err)
	}
	return nil
}

func (b *BridgeNetworkDriver) Disconnect(network Network, endpoint *Endpoint) error {
	return nil
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
	log.Debugf("subnet: %s", subnet)
	ip, ipRange, err := net.ParseCIDR(subnet)
	ipRange.IP = ip
	if err != nil {
		return nil, err
	}

	n := &Network{
		Name:    name,
		Driver: b.Name(),
		IpRange: ipRange,
	}

	// 配置 linux bridge
	err = b.initBridge(n)
	if err != nil {
		return nil, errors.Wrap(err, "init bridge")
	}
	return n, nil
}

func (b *BridgeNetworkDriver) Delete(network Network) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	return netlink.LinkDel(br)
}

// initBridge 初始化网桥
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

	// 网络设备只有设置成 UP 状态后才能处理和转发请求
	if err := setInterfaceUP(bridgeName); err != nil {
		return errors.Errorf("Error set bridge up: %s, Error: %v", bridgeName, err)
	}

	// Setup iptables
	if err := setupIPTables(bridgeName, n.IpRange); err != nil {
		return errors.Errorf("Error setting iptables for %s: %v", bridgeName, err)
	}

	return nil
}

func createBridgeInterface(bridgeName string) error {
	// 检查是否存在同名网桥设备
	_, err := net.InterfaceByName(bridgeName)
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}

	// 初始化一个 netlink 的 Link 对象， Link 的名字即 Bridge 虚拟设备的名字
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName

	// 创建 netlink 的 Bridge 对象
	br := &netlink.Bridge{LinkAttrs: la}
	// Linkadd 方法，创建 Bridge 虚拟网络设备 相当于 `ip link add xxxx`
	if err := netlink.LinkAdd(br); err != nil {
		return errors.Errorf("Bridge creation failed for bridge %s: %v", bridgeName, err)
	}
	return nil
}

// setInterfaceIP 设置一个网络接口的 IP 地址
func setInterfaceIP(name string, rawIP string) error {
	retries := 2
	var iface netlink.Link
	var err error
	// LinkByName 方法找到需要设置的网络接口
	for i := 0; i < retries; i++ {
		iface, err = netlink.LinkByName(name)
		if err == nil {
			break
		}
		log.Debugf("retrieving new bridge netlink link [ %s ]... retrying", name)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return errors.Errorf("Abandoning retrieving the new bridge link from netlink, "+
			"Run [ ip link ] to troubleshoot the error: %v", err)
	}

	// netlink.ParseIPNet 是对 net.ParseCIDR 的一个封装，可以将 net.ParseCIDR 的返回值中的 IP 和 net 整合
	// ipNet 既包含了网段的信息， 192.168.0.0/24，也包含了原始的 ip 192.168.0.1
	ipNet, err := netlink.ParseIPNet(rawIP)
	if err != nil {
		return err
	}

	// netlink.AddrAdd 给网络接口配置地址 相当于 `ip addr add xxx` 的命令
	// 同时如果配置了地址所在网段的信息，例如 192.168.0.0/24
	// 还会配置路由表 192.168.0.0/24 转发到这个 bridge 的网络接口上，例如在宿主机上将 192.168.0.0/24 的网段请求路由到 brO 的网桥
	// route add -net 192.168.0.0/24 dev br0
	addr := &netlink.Addr{
		IPNet: ipNet,
		Label: "",
		Flags: 0,
		Scope: 0,
		Peer:  nil,
	}
	return netlink.AddrAdd(iface, addr)
}

func setInterfaceUP(interfaceName string) error {
	iface, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return errors.Errorf("retrieving a link named [ %s ]: %v", iface.Attrs().Name, err)
	}

	// 等价于 `ip link set xxx up` 命令
	if err := netlink.LinkSetUp(iface); err != nil {
		return errors.Errorf("enabling interface for %s: %v", interfaceName, err)
	}
	return nil
}

// setupIPTables 创建 SNAT 规则， 只要是从这个网桥上出来的包，都会对其
// 做源 IP 的转换，保证了容器经过宿主机访问到宿主机外部网络请求的包转换成机器的 IP，
// 从而能正确的送达和接收。
func setupIPTables(bridgeName string, subnet *net.IPNet) error {
	// Go 语言没有直接操控 iptables 操作的库
	// 例如 对 namespace 中发出的包添加网络地址转换。在 namespace 中请求宿主机外部地址时，将 namespace 中的源地址转换成宿主机的地址
	// 作为源地址，就可以在 namespace 中访问宿主机外的网络了。
	// iptables -t nat -A POSTROUTING -s 192.168.0.0/24 -o ethO -j MASQUERADE
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	output, err := cmd.Output()
	if err != nil {
		log.Errorf("iptables Output, %v", output)
	}
	return err
}
