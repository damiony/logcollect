package conf

import (
	"logagent/utils"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type config struct {
	Root        string
	BaseName    string
	FullName    string
	Endpoints   []string
	DialTimeOut int64
}

var Configs config

func Init(path, name, t string) {
	initConfigs(path, name, t)
	initEtcd()
}

func initConfigs(path, name, t string) {
	viper.AddConfigPath(path)
	viper.SetConfigName(name)
	viper.SetConfigType(t)

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	subv := viper.Sub("logagent.etcd")
	Configs.Endpoints = []string{}
	err = subv.Unmarshal(&Configs)
	if err != nil {
		logrus.Fatal("Viper unmarshal config error: ", err)
	}

	// 获取本机ip
	ip, err := utils.OutBoundIp()
	if err != nil {
		logrus.Fatal("get out bound ip error: ", err)
	}
	// 生成etcd的key: /root/ip/basename
	Configs.FullName = Configs.Root + "/" + ip + "/" + Configs.BaseName
}
