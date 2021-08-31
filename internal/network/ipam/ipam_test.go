package ipam

import (
	"net"
	"testing"
)

var defaultAllocator = New("/var/run/q.container/network/ipam/subnet.json")

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
