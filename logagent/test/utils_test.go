package test

import (
	"logagent/utils"
	"testing"
)

func TestNetworkIps(t *testing.T) {
	ips, err := utils.NetworkIps()
	if err != nil {
		t.Fatal(err)
	}
	for name, ip := range ips {
		t.Logf("name: %v\tip: %v", name, ip)
	}
}

func TestOutBoundIp(t *testing.T) {
	ip, err := utils.OutBoundIp()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ip)
}
