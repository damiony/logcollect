package main

import (
	"logagent/collects"
	"logagent/conf"
	"logagent/utils"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func signalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	s := <-c
	switch s {
	case syscall.SIGINT:
		logrus.Debug("SIGINT......")
		collects.CloseManagers()
		logrus.Debug("EXIT.")
	case syscall.SIGTERM:
		logrus.Debug("SIGTERM...")
		collects.CloseManagers()
		logrus.Debug("EXIT.")
	}
	os.Exit(1)
}

func main() {
	defer collects.CloseManagers()
	defer func() {
		err := recover()
		if err != nil {
			logrus.Error("Catch panic error: ", err)
		}
	}()
	logrus.SetLevel(logrus.DebugLevel)

	// 监听信号
	go signalHandler()

	// 初始化配置信息
	dir := utils.Getpwd()
	conf.Init(dir+"docs", "configs", "yml")

	// 进行收集
	collects.UpdateManagers()

	// 监听etcd变化
	conf.WatchEtcd(collects.UpdateManagers)
}
