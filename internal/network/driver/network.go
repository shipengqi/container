package driver

import (
	"net"
	"os"
	"path/filepath"

	jsoniter "github.com/json-iterator/go"
	"github.com/vishvananda/netlink"

	"github.com/shipengqi/container/pkg/log"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Network configuration
type Network struct {
	Name    string     `json:"name"`
	Driver  string     `json:"driver"`
	IpRange *net.IPNet `json:"iprange"`
}

// Endpoint Contains connection information, IP address, Veth device, port mapping, connected containers and networks, etc.
type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"dev"`
	IPAddress   net.IP           `json:"ip"`
	MacAddress  net.HardwareAddr `json:"mac"`
	Network     *Network         `json:"network"`
	PortMapping []string         `json:"portmap"`
}

func (n *Network) Dump(dumpPath string) error {
	// 不存在则创建
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(dumpPath, 0644)
		} else {
			return err
		}
	}

	nwPath := filepath.Join(dumpPath, n.Name)
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer nwFile.Close()

	nwJson, err := json.Marshal(n)
	if err != nil {
		return err
	}

	_, err = nwFile.Write(nwJson)
	if err != nil {
		return err
	}
	return nil
}

func (n *Network) Remove(dumpPath string) error {
	if _, err := os.Stat(filepath.Join(dumpPath, n.Name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else {
		return os.Remove(filepath.Join(dumpPath, n.Name))
	}
}

func (n *Network) Load(dumpPath string) error {
	f, err := os.Open(dumpPath)
	defer f.Close()
	if err != nil {
		return err
	}
	nwJson := make([]byte, 2000)
	c, err := f.Read(nwJson)
	if err != nil {
		return err
	}

	err = json.Unmarshal(nwJson[:c], n)
	if err != nil {
		log.Errorf("Error load network info", err)
		return err
	}
	return nil
}
