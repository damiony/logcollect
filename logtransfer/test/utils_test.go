package test

import (
	"logtransfer/utils"
	"testing"
)

func TestOutBoundIp(t *testing.T) {
	ip, err := utils.OutBoundIp()
	if err != nil {
		t.Error(err)
	} else {
		t.Log(ip)
	}
}
