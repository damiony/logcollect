package main

import (
	"logtransfer/conf"
	"logtransfer/services"
	"logtransfer/utils"
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
		services.CloseManagers()
		logrus.Debug("EXIT.")
	case syscall.SIGTERM:
		logrus.Debug("SIGTERM......")
		services.CloseManagers()
		logrus.Debug("EXIT.")
	}
	os.Exit(1)
}

func main() {
	defer services.CloseManagers()
	defer func() {
		err := recover()
		if err != nil {
			logrus.Error("Catch panic error: ", err)
		}
	}()
	logrus.SetLevel(logrus.DebugLevel)
	// 监听信号
	go signalHandler()

	// 读取配置
	dir := utils.GetPwd()
	conf.Init(dir+"docs", "configs", "yml")

	// 初始化服务
	services.InitManagers()

	defer services.CloseManagers()
	// 监听配置变化
	conf.WatchEtcd(services.UpdateManagers)
}
