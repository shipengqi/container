package network

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"

	"github.com/shipengqi/container/internal/container"
	"github.com/shipengqi/container/internal/network/driver"
	"github.com/shipengqi/container/internal/network/driver/bridge"
	"github.com/shipengqi/container/internal/network/ipam"
	"github.com/shipengqi/container/pkg/log"
)

const (
	DefaultNetworkPath   = "/var/run/q.container/network/network/"
	DefaultAllocatorPath = "/var/run/q.container/network/ipam/subnet.json"
)

var (
	drivers          = map[string]driver.Interface{}
	networks         = map[string]*driver.Network{}
	defaultAllocator = ipam.New(DefaultAllocatorPath)
)


func Init() error {
	var bridgeDriver = bridge.New()
	drivers[bridgeDriver.Name()] = bridgeDriver

	if _, err := os.Stat(DefaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(DefaultNetworkPath, 0644)
		} else {
			return err
		}
	}

	_ = filepath.Walk(DefaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		if strings.HasSuffix(nwPath, "/") {
			return nil
		}
		_, nwName := path.Split(nwPath)
		nw := &driver.Network{
			Name: nwName,
		}
		log.Debugf("load network: %s", nwPath)
		if err := nw.Load(nwPath); err != nil {
			log.Errorf("error load network: %s", err)
		}
		log.Debugs(nw)
		networks[nwName] = nw
		return nil
	})

	return nil
}

func List() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, _ = fmt.Fprint(w, "NAME\tIpRange\tDriver\n")
	for _, nw := range networks {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n",
			nw.Name,
			nw.IpRange.String(),
			nw.Driver,
		)
	}
	if err := w.Flush(); err != nil {
		log.Errorf("Flush error %v", err)
		return
	}
}

func Delete(name string) error {
	nw, ok := networks[name]
	if !ok {
		return errors.Errorf("No Such Network: %s", name)
	}

	if err := defaultAllocator.Release(nw.IpRange, &nw.IpRange.IP); err != nil {
		return errors.Errorf("Error Remove Network gateway ip: %s", err)
	}

	if err := drivers[nw.Driver].Delete(nw.Name); err != nil {
		return errors.Errorf("Error Remove Network DriverError: %s", err)
	}

	return nw.Remove(DefaultNetworkPath)
}

func Create(driver, name, subnet string) error {
	_, cidr, _ := net.ParseCIDR(subnet)
	gatewayIp, err := defaultAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = gatewayIp

	nw, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}

	return nw.Dump(DefaultNetworkPath)
}

func Connect(name string, cinfo *container.Information) error {
	network, ok := networks[name]
	if !ok {
		return errors.Errorf("No Such Network: %s", name)
	}

	log.Debugf("ip: %s", network.IpRange.IP)
	log.Debugf("mask: %s", network.IpRange.Mask)
	ip, err := defaultAllocator.Allocate(network.IpRange)
	if err != nil {
		return errors.Wrap(err, "allocate ip")
	}
	log.Debugf("alloc ip: %s", ip)
	ep := &driver.Endpoint{
		ID:          fmt.Sprintf("%s-%s", cinfo.Id, name),
		IPAddress:   ip,
		Network:     network,
		PortMapping: cinfo.PortMapping,
	}

	if err = drivers[network.Driver].Connect(network.Name, ep); err != nil {
		return errors.Wrap(err, "driver connect")
	}

	if err = configEndpointIpAddressAndRoute(ep, cinfo); err != nil {
		return errors.Wrap(err, "config endpoint")
	}

	return configPortMapping(ep, cinfo)
}

func configEndpointIpAddressAndRoute(ep *driver.Endpoint, cinfo *container.Information) error {
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return errors.Errorf("fail config endpoint: %v", err)
	}
	defer enterContainerNetns(&peerLink, cinfo)()

	log.Debugf("ep ip: %s", ep.IPAddress)
	log.Debugf("ep ip range: %s", *ep.Network.IpRange)
	interfaceIP := *ep.Network.IpRange
	interfaceIP.IP = ep.IPAddress

	// ip netns exec ns1 ifconfig veth1 172.18.0.2/24 up
	// set ip address of the end of Veth in container
	if err = bridge.SetInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return errors.Errorf("set interface ip: %s: %v", ep.Network, err)
	}

	// start up one end of Veth in container
	if err = bridge.SetInterfaceUP(ep.Device.PeerName); err != nil {
		return errors.Wrap(err, "set interface up")
	}

	// lo Interface is off by default in Net Namespace
	if err = bridge.SetInterfaceUP("lo"); err != nil {
		return errors.Wrap(err, "set lo interface up")
	}

	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")

	// ip netns exec ns1 route add 0.0.0.0/0 dev veth1
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IpRange.IP,
		Dst:       cidr,
	}

	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return errors.Wrap(err, "route add")
	}

	return nil
}

func enterContainerNetns(enLink *netlink.Link, cinfo *container.Information) func() {
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cinfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		log.Errorf("error get container net namespace: %v", err)
	}

	nsFD := f.Fd()
	runtime.LockOSThread()

	// ip link set $link netns $ns
	// ip link set veth0 netns ns1
	if err = netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		log.Errorf("error set link netns : %v", err)
	}

	origins, err := netns.Get()
	if err != nil {
		log.Errorf("get current netns: %v", err)
	}

	// sets namespace using syscall
	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		log.Errorf("set netns: %v", err)
	}
	return func() {
		_ = netns.Set(origins)
		origins.Close()
		runtime.UnlockOSThread()
		f.Close()
	}
}

func configPortMapping(ep *driver.Endpoint, cinfo *container.Information) error {
	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			log.Errorf("port mapping format: %v", pm)
			continue
		}
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])
		log.Debugf("poat mapping: %s", iptablesCmd)
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		output, err := cmd.Output()
		if err != nil {
			log.Errorf("iptables Output: %v", output)
			continue
		}
	}
	return nil
}
