package utils

import (
	"net"
	"os"
	"strings"
)

func OutBoundIp() (string, error) {
	conn, err := net.Dial("udp", "1.1.1.1:3000")
	if err != nil {
		return "", err
	}

	addr := conn.LocalAddr().String()
	ip := strings.Split(addr, ":")[0]
	return ip, nil
}

func GetPwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return "./"
	}
	return dir + "/"
}
