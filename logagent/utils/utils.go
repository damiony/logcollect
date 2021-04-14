package utils

import (
	"errors"
	"net"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

func NetworkIps() (map[string]string, error) {
	ips := make(map[string]string)
	// 返回全部网卡信息
	interfaces, err := net.Interfaces()
	if err != nil {
		logrus.Error("Get net interfaces error:", err)
		return ips, err
	}

	for _, i := range interfaces {
		// address是网卡上的全部ip列表
		address, err := i.Addrs()
		if err != nil {
			logrus.Error("Get address error:", err)
			continue
		}

		for _, v := range address {
			if ipnet, ok := v.(*net.IPNet); ok && !ipnet.IP.IsLoopback() { // 去除回环地址
				if ipnet.IP.To4() != nil { // 获取ipv4
					ips[i.Name] += ipnet.IP.String() + " "
				}
			}
		}
	}

	// 去除首尾空格
	for k := range ips {
		strings.Trim(ips[k], " ")
	}
	return ips, nil
}

func OutBoundIp() (ip string, err error) {
	// ip地址可以随意填写 只要符合ip规则即可
	conn, err := net.Dial("udp", "1.1.1.1:3000")
	if err != nil {
		return "", err
	}

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = strings.Split(localAddr.String(), ":")[0]

	if ip == "" {
		err = errors.New("Ip is invalid")
	}

	return
}

func Getpwd() string {
	str, err := os.Getwd()
	if err != nil {
		return "./"
	}
	return str + "/"
}
