package network

import (
	"net"
	"testing"
)

func TestIPAM_Allocate(t *testing.T) {
	_, ipnet, _ := net.ParseCIDR("192.168.0.1/24")
	ip, _ := defaultAllocator.Allocate(ipnet)
	t.Logf("alloc ip: %v", ip)
}

func TestIPAM_Release(t *testing.T) {
	ip, ipnet, _ := net.ParseCIDR("192.168.0.1/24")
	err := defaultAllocator.Release(ipnet, &ip)
	if err != nil {
		t.Errorf("release ip: %v", err)
	}
}

// [root@shcCDFrh75vm7 network]# go test -v -run TestIPAM_Allocate
// === RUN   TestIPAM_Allocate
//    ipam_test.go:11: alloc ip: 192.168.0.1
// --- PASS: TestIPAM_Allocate (0.00s)
// PASS
// ok      github.com/shipengqi/container/internal/network 0.010s
// [root@shcCDFrh75vm7 network] cat /var/run/q.container/network/ipam/subnet.json
// {"192.168.0.0/24":"10000000000000000000000000000000000000000000000000000000000000000000000
// 0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
// 0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"}
// // 第 1 位已经被置为 1。
// [root@shcCDFrh75vm7 network]# go test -v -run TestIPAM_Release
// === RUN   TestIPAM_Release
// --- PASS: TestIPAM_Release (0.00s)
// PASS
// ok      github.com/shipengqi/container/internal/network 0.009s
// [root@shcCDFrh75vm7 network]# cat /var/run/q.container/network/ipam/subnet.json
// {"192.168.0.0/24":"00000000000000000000000000000000000000000000000000000000000000000000000
// 0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
// 0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"}
// 第 1 位已经被重新置为 0。
