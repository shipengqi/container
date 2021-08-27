package network

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/shipengqi/container/internal/container"
	"github.com/shipengqi/container/pkg/log"
	"github.com/vishvananda/netns"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/vishvananda/netlink"
)

const DefaultNetworkPath = "/var/run/q.container/network/network/"

var (
	drivers = map[string]Driver{}
	networks = map[string]*Network{}
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

func (nw *Network) dump(dumpPath string) error {
	// 不存在则创建
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(dumpPath, 0644)
		} else {
			return err
		}
	}

	nwPath := filepath.Join(dumpPath, nw.Name)
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC | os.O_WRONLY | os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer nwFile.Close()

	nwJson, err := json.Marshal(nw)
	if err != nil {
		return err
	}

	_, err = nwFile.Write(nwJson)
	if err != nil {
		return err
	}
	return nil
}

func (nw *Network) remove(dumpPath string) error {
	if _, err := os.Stat(filepath.Join(dumpPath, nw.Name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else {
		return os.Remove(filepath.Join(dumpPath, nw.Name))
	}
}

func (nw *Network) load(dumpPath string) error {
	f, err := os.Open(dumpPath)
	defer f.Close()
	if err != nil {
		return err
	}
	nwJson := make([]byte, 2000)
	n, err := f.Read(nwJson)
	if err != nil {
		return err
	}

	err = json.Unmarshal(nwJson[:n], nw)
	if err != nil {
		log.Errorf("Error load network info", err)
		return err
	}
	return nil
}

func Init() error {
	// 加载网络驱动
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver

	// 判断网络的配置目录是否存在，不存在则创建
	if _, err := os.Stat(DefaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(DefaultNetworkPath, 0644)
		} else {
			return err
		}
	}

	// 检查网络配置目录中的所有文件
	// filepath.Walk(path, func(string, os.FileInfo, error)) 函数会遍历 path 目录
	// 并执行回调函数去处理每一个文件
	_ = filepath.Walk(DefaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		// 如果是目录则跳过
		if strings.HasSuffix(nwPath, "/") {
			return nil
		}

		// 加载文件名作为网络名
		_, nwName := path.Split(nwPath)
		nw := &Network{
			Name: nwName,
		}

		// 加载网络的配置信息
		log.Debugf("load network: %s", nwPath)
		if err := nw.load(nwPath); err != nil {
			log.Errorf("error load network: %s", err)
		}

		log.Debugs(nw)
		networks[nwName] = nw
		return nil
	})

	return nil
}

func ListNetwork() {
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

func DeleteNetwork(networkName string) error {
	nw, ok := networks[networkName]
	if !ok {
		return errors.Errorf("No Such Network: %s", networkName)
	}

	// 调用 IPAM 的实例 defaultAllocator 释放网络网关的 IP
	if err := defaultAllocator.Release(nw.IpRange, &nw.IpRange.IP); err != nil {
		return errors.Errorf("Error Remove Network gateway ip: %s", err)
	}

	// 调用网络驱动删除网络创建的设备与配置
	if err := drivers[nw.Driver].Delete(*nw); err != nil {
		return errors.Errorf("Error Remove Network DriverError: %s", err)
	}

	return nw.remove(DefaultNetworkPath)
}

func CreateNetwork(driver, name, subnet string) error {
	// ParseCIDR 将网段的字符串转为 net.IPNet 对象
	// ParseCIDR parses s as a CIDR notation IP address and prefix length,
	// like "192.0.2.0/24" or "2001:db8::/32", as defined in
	// RFC 4632 and RFC 4291.
	//
	// It returns the IP address and the network implied by the IP and
	// prefix length.
	// For example, ParseCIDR("192.0.2.1/24") returns the IP address
	// 192.0.2.1 and the network 192.0.2.0/24.
	_, cidr, _ := net.ParseCIDR(subnet)
	// 通过 IPAM 分配的网关 IP，获取到网段中的第一个 IP 作为网关 IP
	gatewayIp, err := defaultAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = gatewayIp

	// 调用指定的网络驱动创建网络，这里的 drivers 字典是各个网络驱动的实例字典，通过调用网络驱动
	// 的 Create 方法创建网络
	nw, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}

	return nw.dump(DefaultNetworkPath)
}


func Connect(networkName string, cinfo *container.Information) error {
	// 读取网络配置信息
	network, ok := networks[networkName]
	if !ok {
		return errors.Errorf("No Such Network: %s", networkName)
	}

	log.Debugf("ip: %s", network.IpRange.IP)
	log.Debugf("mask: %s", network.IpRange.Mask)
	// 分配容器IP地址
	ip, err := defaultAllocator.Allocate(network.IpRange)
	if err != nil {
		return errors.Wrap(err, "allocate ip")
	}

	// 创建网络端点
	ep := &Endpoint{
		ID: fmt.Sprintf("%s-%s", cinfo.Id, networkName),
		IPAddress: ip,
		Network: network,
		PortMapping: cinfo.PortMapping,
	}

	// 调用网络驱动挂载和配置网络端点
	if err = drivers[network.Driver].Connect(network, ep); err != nil {
		return errors.Wrap(err, "driver connect")
	}

	// 到容器的 namespace 配置容器网络设备 IP 地址
	if err = configEndpointIpAddressAndRoute(ep, cinfo); err != nil {
		return errors.Wrap(err, "config endpoint")
	}

	// 配置端口l映射信息 例如 run -p 80:8080
	return configPortMapping(ep, cinfo)
}

// configEndpointIpAddressAndRoute 配置容器网络端点的地址和路由
// BridgeNetworkDriver.Connect 将容器的网络的一个端点挂载到了 Bridge 网络的Linux Bridge 上
// configEndpointIpAddressAndRoute 是在容器的 Net Namespace 中，将网络端点的 Veth 设备的另外一端移到这个
// Net Namespace 中并配置
func configEndpointIpAddressAndRoute(ep *Endpoint, cinfo *container.Information) error {
	// 通过网络端点中 Veth 的另一端
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return errors.Errorf("fail config endpoint: %v", err)
	}

	// 将容器的网络端点加入到容器的网络空间中
	// 并使这个函数下面的操作都在这个网络空间中进行
	// 执行完函数后，恢复为默认的网络空间
	defer enterContainerNetns(&peerLink, cinfo)()

	// 获取到容器的 IP 地址及网段， 用于配置容器内部 Veth 端点配置
	log.Debugf("ep ip: %s",  ep.IPAddress)
	log.Debugf("ep ip range: %s",  *ep.Network.IpRange)
	interfaceIP := *ep.Network.IpRange
	interfaceIP.IP = ep.IPAddress

	// 设置容器内 Veth 端点的 IP
	if err = setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return errors.Errorf("set interface ip: %v,%s", ep.Network, err)
	}

	// 启动容器内的 Veth 端点
	if err = setInterfaceUP(ep.Device.PeerName); err != nil {
		return errors.Wrap(err, "set interface up")
	}

	// Net Namespace 中默认本地地址 127.0.0.1 的 loopback 网卡是关闭状态的
	// 启动容器内 lo 网卡
	if err = setInterfaceUP("lo"); err != nil {
		return errors.Wrap(err, "set lo interface up")
	}

	// 设置容器内的外部请求都通过容器内的 Veth 端点访问
	// 0.0.0.0/0 表示所有的 IP 地址段
	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")

	// 构建要添加的路由数据，包括网络设备、网关 IP 及目的网段
	// 相当于 ip netns exec ns1 route add 0.0.0.0/0 dev veth1
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw: ep.Network.IpRange.IP,
		Dst: cidr,
	}

	// RouteAdd 添加路由到容器的网络空间
	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return errors.Wrap(err, "route add")
	}

	return nil
}

// enterContainerNetns 将容器的网络端点加入到容器的网络空间中
// 并锁定当前程序所执行的线程，使当前线程进入到容器的网络空间
// 返回值是一个函数指针，执行这个返回函数才会退出容器的网络空间，回归到宿主机的网络空间
func enterContainerNetns(enLink *netlink.Link, cinfo *container.Information) func() {
	// 找到容器的 Net Namespace
	// /proc/<pid>/ns/net 打开这个文件的文件描述符就可以来操作 Net Namespace
	// 而 cinfo 中的 PID，即容器在宿主机上映射的进程 ID
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cinfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		log.Errorf("error get container net namespace, %v", err)
	}

	nsFD := f.Fd()
	// 锁定当前程序所执行的线程，如果不锁定操作系统线程的话
	// Go 语言的 goroutine 可能会被调度到别的线程上去
	// 就不能保证一直在所需要的网络空间中了
	runtime.LockOSThread()

	// 修改网络端点 Veth 的另外一端，将其移动到容器的 Net Namespace 中
	if err = netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		log.Errorf("error set link netns , %v", err)
	}

	// 获取当前的网络 namespace
	// 以便后面从容器的 Net Namespace 中退出，回到原本网络的 Net Namespace 中
	origns, err := netns.Get()
	if err != nil {
		log.Errorf("error get current netns, %v", err)
	}

	// 将当前进程加入容器的 Net Namespace
	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		log.Errorf("error set netns, %v", err)
	}
	// 返回之前 Net Namespace 的函数
	// 在容器的网络空间中，执行完容器配置之后调用此函数就可以将程序恢复到原生的N et Namespace
	return func () {
		// 恢复到上面获取到的之前的 Net Namespace
		_ = netns.Set(origns)
		// 关闭 Namespace 文件
		origns.Close()
		// 取消对当附程序的线程锁定
		runtime.UnlockOSThread()
		// 关闭 Namespace 文件
		f.Close()
	}
}

func configPortMapping(ep *Endpoint, cinfo *container.Information) error {
	for _, pm := range ep.PortMapping {
		portMapping :=strings.Split(pm, ":")
		if len(portMapping) != 2 {
			log.Errorf("port mapping format error, %v", pm)
			continue
		}
		// 在 iptables 的 PREROUTING 中添加 DNAT 规则
		// 将宿主机的端口请求转发到容器的地址和端口上
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		output, err := cmd.Output()
		if err != nil {
			log.Errorf("iptables Output, %v", output)
			continue
		}
	}
	return nil
}
