package bridge

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"

	"github.com/shipengqi/container/internal/network/driver"
	"github.com/shipengqi/container/pkg/log"
)

// bridged network bridge driver
type bridged struct {
	name string
}

// New create a bridge driver
func New() driver.Interface {
	return &bridged{"bridge"}
}

func (b *bridged) Name() string {
	return b.name
}

// Connect a network and a endpoint
func (b *bridged) Connect(network string, endpoint *driver.Endpoint) error {
	// search Linux Bridge Interface
	br, err := netlink.LinkByName(network)
	if err != nil {
		return err
	}

	// create LinkAttrs
	la := netlink.NewLinkAttrs()
	// Linux Interface name limitation
	la.Name = endpoint.ID[:5]
	// set index of a bridge, indicates that one end of this Veth is mounted to the Linux Bridge
	la.MasterIndex = br.Attrs().Index

	// create a Veth object
	// PeerName the other end of the Veth
	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + la.Name,
	}

	// LinkAdd create Veth Interface
	if err = netlink.LinkAdd(&endpoint.Device); err != nil {
		return errors.Errorf("Add Endpoint Device: %v", err)
	}

	// UP Veth Interface => `ip link set xxx up`
	if err = netlink.LinkSetUp(&endpoint.Device); err != nil {
		return errors.Errorf("Add Endpoint Device: %v", err)
	}
	return nil
}

func (b *bridged) Disconnect(network string, endpoint *driver.Endpoint) error {
	return nil
}

func (b *bridged) Create(subnet, name string) (*driver.Network, error) {
	log.Debugf("subnet: %s", subnet)
	ip, ipRange, err := net.ParseCIDR(subnet)
	ipRange.IP = ip
	if err != nil {
		return nil, err
	}

	n := &driver.Network{
		Name:    name,
		Driver:  b.name,
		IpRange: ipRange,
	}

	err = b.initBridge(n)
	if err != nil {
		return nil, errors.Wrap(err, "init bridge")
	}
	return n, nil
}

func (b *bridged) Delete(network string) error {
	br, err := netlink.LinkByName(network)
	if err != nil {
		return err
	}
	// `ip link del $link`
	return netlink.LinkDel(br)
}

// initBridge Initializing a bridge interface
func (b *bridged) initBridge(n *driver.Network) error {
	bridgeName := n.Name
	if err := CreateBridgeInterface(bridgeName); err != nil {
		return errors.Wrap(err, "create bridge")
	}

	gatewayIP := *n.IpRange
	gatewayIP.IP = n.IpRange.IP
	if err := SetInterfaceIP(bridgeName, gatewayIP.String()); err != nil {
		return errors.Wrap(err, "set interface ip")
	}

	if err := SetInterfaceUP(bridgeName); err != nil {
		return errors.Errorf("set bridge up: %s: %v", bridgeName, err)
	}

	if err := setupIPTables(bridgeName, n.IpRange); err != nil {
		return errors.Errorf("set iptables for %s: %v", bridgeName, err)
	}

	return nil
}

// CreateBridgeInterface create a Network Bridge Interface
func CreateBridgeInterface(bridgeName string) error {
	_, err := net.InterfaceByName(bridgeName)
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}

	la := netlink.NewLinkAttrs()
	la.Name = bridgeName

	br := &netlink.Bridge{LinkAttrs: la}
	// ip link add xxxx
	if err := netlink.LinkAdd(br); err != nil {
		return errors.Errorf("Bridge creation failed for bridge %s: %v", bridgeName, err)
	}
	return nil
}

// SetInterfaceIP Set ip address of a Network Interface
func SetInterfaceIP(name string, rawIP string) error {
	retries := 2
	var iface netlink.Link
	var err error
	// try to find the Network Interface
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

	ipNet, err := netlink.ParseIPNet(rawIP)
	if err != nil {
		return err
	}

	// route add -net 192.168.0.0/24 dev br0
	addr := &netlink.Addr{
		IPNet: ipNet,
		Label: "",
		Flags: 0,
		Scope: 0,
		Peer:  nil,
	}
	// ip addr add xxx
	return netlink.AddrAdd(iface, addr)
}

// SetInterfaceUP set Interface up
func SetInterfaceUP(interfaceName string) error {
	iface, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return errors.Errorf("retrieving a link named [ %s ]: %v", iface.Attrs().Name, err)
	}

	// ip link set xxx up
	if err := netlink.LinkSetUp(iface); err != nil {
		return errors.Errorf("enabling interface for %s: %v", interfaceName, err)
	}
	return nil
}

func setupIPTables(bridgeName string, subnet *net.IPNet) error {
	// iptables -t nat -A POSTROUTING -s 192.168.0.0/24 -o testbridge -j MASQUERADE
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	output, err := cmd.Output()
	if err != nil {
		log.Errorf("iptables Output: %v", output)
	}
	return err
}
