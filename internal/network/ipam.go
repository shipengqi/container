package network

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/shipengqi/container/pkg/log"
	"github.com/shipengqi/container/pkg/utils"
)

const DefaultAllocatorPath = "/var/run/q.container/network/ipam/subnet.json"

// IPAM 保存 ip 地址的分配信息
type IPAM struct {
	// 文件存放路径
	SubnetAllocatorPath string
	// 网段与位图算法的数组 map，key 是网段，value 是分配的位图数组
	// 使用 string 中的一个字符表示一个状态位
	// todo 采用一位表示一个是否分配的状态位
	Subnets *map[string]string
}

var defaultAllocator = &IPAM{
	SubnetAllocatorPath: DefaultAllocatorPath,
}

// Allocate 在网段中分配一个可用的 ip 地址
func (i *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	// 餐放在网段中地址分配信息的数组
	i.Subnets = &map[string]string{}
	err = i.load()
	if err != nil {
		log.Errorf("load allocation info: %v", err)
	}

	// net.IPNet.Mask.Size 返回网段的子网掩码的总长度和网段前面的固定位的长度
	// 比如 127.0.0.0/8 网段的子网掩码是 255.0.0.0
	// 那么 subnet.Mask.Size 返回值就是前面 255 所对应的位数和总位数，也就是 8 和 32
	_, subnet, _ = net.ParseCIDR(subnet.String())
	one, size := subnet.Mask.Size()
	// 如果之前没有分配过这个网段，则初始化网段配置
	if _, exist := (*i.Subnets)[subnet.String()]; !exist {
		// 用 0 填满这个网段的配置
		// 1 << uint8(size - one) 表示这个网段中有多少个可用地址
		// size - one 是子网掩码后面的网络位数，2^(size - one) 表示网段中的可用 IP 数
		// 而 2^(size - one) 等价于 1 << uint8(size - one)
		(*i.Subnets)[subnet.String()] = strings.Repeat("0", 1 << uint8(size - one))
	}

	// 遍历网段的位图数组
	for c := range (*i.Subnets)[subnet.String()] {
		// 找到数组中为 0 的项和数组序号，即可以分配的 IP
		if (*i.Subnets)[subnet.String()][c] == '0' {
			// 设置这个为 0 的序号值为 1， 即分配这个 IP
			ipalloc := []byte((*i.Subnets)[subnet.String()])
			// Go 的字符串，创建之后就不能修改，所以通过转换成 byte 数组，修改后再转换成字符串赋值
			ipalloc[c] = '1'
			(*i.Subnets)[subnet.String()] = string(ipalloc)
			// 这里的 IP 为初始 ip，比如对于网段 192.168.0.0/16，这里就是 192.168.0.0
			ip = subnet.IP
			// 通过网段的 IP 与上面的偏移相加计算出分配的 IP 地址，由于 IP 地址是 uint 的一个数组，
			// 需要通过数组中的每一项加所需要的值，比如网段是 172.16.0.0/12 数组序号是 65555，
			// 那么在 [172,16,0,0] 上依次加 [unit8(65555 >> 24), unit8(65555 >> 16), unit8(65555 >> 8), unit8(65555 >> 0)]
			// 即 [0,1,0,19]，那么获得的 ip 就是 172.17.0.19
			for t := uint(4); t > 0; t-=1 {
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}
			// 由于此处 IP 是从 1 开始分配的，所以最后再加 1 ，最终得到分配的 IP 是172.17.0.20
			ip[3]+=1
			break
		}
	}

	err = i.dump()
	if err != nil {
		log.Errorf("dump allocation info: %v", err)
	}
	return
}

func (i *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	i.Subnets = &map[string]string{}
	_, subnet, _ = net.ParseCIDR(subnet.String())
	err := i.load()
	if err != nil {
		log.Errorf("load allocation info: %v", err)
	}

	// 计算 ip 地址在位图数组中的索引
	c := 0
	// 将 ip 地址转换成 4 个字节的表示方式
	releaseIP := ipaddr.To4()
	// 由于 IP 是从 1 开始分配的，所以转换成成索引要减 1
	// 第 4 个字节数减 1
	releaseIP[3]-=1

	for t := uint(4); t > 0; t-=1 {
		// 与分配 IP 相反，释放 IP 获得索引的方式是 IP 地址的每一位相减之后分别左移将对应的数值加到索引上。
		c += int(releaseIP[t-1] - subnet.IP[t-1]) << ((4-t) * 8)
	}
	// 将分配的位图数组中索引位置的值置为 0
	ipalloc := []byte((*i.Subnets)[subnet.String()])
	ipalloc[c] = '0'
	(*i.Subnets)[subnet.String()] = string(ipalloc)

	err = i.dump()
	if err != nil {
		log.Errorf("dump allocation info: %v", err)
	}
	return nil
}

// load 加载网段地址分配信息
func (i *IPAM) load() error {
	// 不存在，说明未分配
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

func (i *IPAM) dump() error{
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
	f, err := os.OpenFile(i.SubnetAllocatorPath, os.O_TRUNC | os.O_WRONLY | os.O_CREATE, 0644)
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

// https://github.com/tabalt/ipquery/
// https://github.com/thinkeridea/go-extend/tree/master/exnet
// https://juejin.cn/post/6844903882041065486
// IPv4 地址有 4 个字节，样式：b4 b3 b2 b1
// 每个字节表示的范围：
// byte4: 4294967296(1<<32)
// byte3: 16777216(1<<24)
// byte2: 65536(1<<16)
// byte1: 256(1<<8)
// 通用公式：b4<<24 | b3<<16 | b2<<8 | b1
// 例如，222.173.108.86
// 转换方法：222<<24 | 173<<16 | 108<<8 | 86 = 3735907414
// 再例如，1.0.1.1
// 转换方法：1<<24 | 0<<16 | 1<<8 | 1 = 16777473

// // IPString2Long 把ip字符串转为数值
// func IPString2Long(ip string) (uint, error) {
//	b := net.ParseIP(ip).To4()
//	if b == nil {
//		return 0, errors.New("invalid ipv4 format")
//	}
//
//	return uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24, nil
// }
//
// // Long2IPString 把数值转为ip字符串
// func Long2IPString(i uint) (string, error) {
//	if i > math.MaxUint32 {
//		return "", errors.New("beyond the scope of ipv4")
//	}
//
//	ip := make(net.IP, net.IPv4len)
//	ip[0] = byte(i >> 24)
//	ip[1] = byte(i >> 16)
//	ip[2] = byte(i >> 8)
//	ip[3] = byte(i)
//
//	return ip.String(), nil
// }
