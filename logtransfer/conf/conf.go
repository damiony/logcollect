package conf

import (
	"logtransfer/utils"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type config struct {
	Root        string   // etcd的根路径
	BaseName    string   // etcd配置的文件名
	FullName    string   // etcd配置的全路径
	Endpoints   []string // etcd的ip
	DialTimeout int64    // etcd连接超时时间
}

var Configs config

func Init(path string, name string, filetype string) {
	InitConfig(path, name, filetype)
	initEtcdConfig()
}

func InitConfig(path string, name string, filetype string) {
	// 读取配置
	viper.AddConfigPath(path)
	viper.SetConfigName(name)
	viper.SetConfigType(filetype)
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatalf("Read config %s error: %s", path+name, err.Error())
	}

	// 提取子树
	subv := viper.Sub("logtransfer.etcd")
	Configs.Endpoints = []string{}
	err = subv.Unmarshal(&Configs)
	if err != nil {
		logrus.Fatal("viper unmarshal config error: ", err)
	}

	// 获取ip
	ip, err := utils.OutBoundIp()
	if err != nil {
		logrus.Fatal("get ip error: ", err)
	}
	// 生成etcd的key: /root/ip/basename
	Configs.FullName = Configs.Root + "/" + ip + "/" + Configs.BaseName
}
