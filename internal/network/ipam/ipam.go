package ipam

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/shipengqi/container/pkg/log"
	"github.com/shipengqi/container/pkg/utils"
)

const DefaultAllocatorPath = "/var/run/q.container/network/ipam/subnet.json"

var defaultAllocator = &IPAM{
	SubnetAllocatorPath: DefaultAllocatorPath,
}

// IPAM IP Address Management
type IPAM struct {
	SubnetAllocatorPath string
	Subnets             *map[string]*BitMap
}

// Allocate an available ip address of a subnet
func (i *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	i.Subnets = &map[string]*BitMap{}
	err = i.load()
	if err != nil {
		log.Errorf("load allocation info: %v", err)
	}
	_, subnet, _ = net.ParseCIDR(subnet.String())
	one, size := subnet.Mask.Size()
	subnetKey := subnet.String()
	old, exist := (*i.Subnets)[subnetKey]
	if !exist || old.Capacity == 0 {
		(*i.Subnets)[subnetKey] = NewBitMap(1 << uint8(size-one))
	}
	subnetBitMap := (*i.Subnets)[subnetKey]
	// available ip index
	c := subnetBitMap.First()
	if c < 0 {
		return nil, errors.New("cannot find available ip address")
	}
	subnetBitMap.Set(uint(c))
	ip = subnet.IP
	for t := uint(4); t > 0; t -= 1 {
		[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
	}
	ip[3] += 1

	err = i.dump()
	if err != nil {
		log.Errorf("dump allocation info: %v", err)
	}
	return
}

func (i *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	i.Subnets = &map[string]*BitMap{}
	_, subnet, _ = net.ParseCIDR(subnet.String())
	err := i.load()
	if err != nil {
		log.Errorf("load allocation info: %v", err)
	}

	c := 0
	releaseIP := ipaddr.To4()
	releaseIP[3] -= 1

	for t := uint(4); t > 0; t -= 1 {
		c += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}

	subnetKey := subnet.String()
	subnetBitMap, exist := (*i.Subnets)[subnetKey]
	if !exist || subnetBitMap.Capacity == 0 {
		return nil
	}

	subnetBitMap.Reset(uint(c))

	err = i.dump()
	if err != nil {
		log.Errorf("dump allocation info: %v", err)
	}
	return nil
}

// load 加载网段地址分配信息
func (i *IPAM) load() error {
	if utils.IsNotExist(i.SubnetAllocatorPath) {
		return nil
	}

	subnetsBytes, err := ioutil.ReadFile(i.SubnetAllocatorPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(subnetsBytes, i.Subnets)
	if err != nil {
		return err
	}

	return nil
}

func (i *IPAM) dump() error {
	// 分隔目录和文件
	dir, _ := filepath.Split(i.SubnetAllocatorPath)
	// 不存在，则创建
	if utils.IsNotExist(dir) {
		err := os.MkdirAll(dir, 0644)
		if err != nil {
			return err
		}
	}

	// os.O_TRUNC 如果存在则清空
	f, err := os.OpenFile(i.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	subnetsBytes, err := json.Marshal(i.Subnets)
	if err != nil {
		return err
	}
	_, err = f.Write(subnetsBytes)
	if err != nil {
		return err
	}

	return nil
}
