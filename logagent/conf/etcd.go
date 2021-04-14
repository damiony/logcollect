package conf

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
)

type EtcdInfo struct {
	Name    string   `json:"name"`
	MqHosts []string `json:"mqhosts"`
	Path    string   `json:"path"`
}

type Option func()

var etcdClient *clientv3.Client
var EtcdInfos []EtcdInfo

func initEtcd() {
	var err error
	// 初始化etcd客户端
	etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   Configs.Endpoints,                                // etcd集群地址
		DialTimeout: time.Duration(Configs.DialTimeOut) * time.Second, // 连接超时时间
	})
	if err != nil {
		logrus.Fatal("Create etcd client error: ", err)
	}

	// 检查etcd的健康状况
	if !checkEtcdHealth() {
		logrus.Fatal("Connect etcd fail.")
	}

	// 获取配置信息
	resp, err := etcdClient.Get(context.TODO(), Configs.FullName)
	if err != nil {
		logrus.Error("Etcd get info error: ", err)
	}

	// 如果配置不存在，初始化etcd的信息
	// 如果配置存在，初始化消息队列的信息
	kvs := resp.Kvs
	if len(kvs) == 0 {
		etcdClient.Put(context.TODO(), Configs.FullName, "[]")
	} else {
		err = json.Unmarshal(kvs[0].Value, &EtcdInfos)
		if err != nil {
			logrus.Fatal("unmarshal json error: ", err)
		}
		logrus.Debugf("Unmarshal etcd json suc: %v", EtcdInfos)
	}
}

// checkEtcdHealth 检查etcd的健康状况
func checkEtcdHealth() bool {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Duration(Configs.DialTimeOut)*time.Second)
	defer cancel()
	for _, ip := range Configs.Endpoints {
		_, err := etcdClient.Status(timeoutCtx, ip)
		if err != nil {
			return false
		}
	}
	return true
}

// 重新读取etcd的配置信息
func reloadEtcd() error {
	resp, err := etcdClient.Get(context.TODO(), Configs.FullName)
	if err != nil {
		logrus.Error("Get etcd key: %s error: %v", Configs.FullName, err)
		return err
	}

	kvs := resp.Kvs
	if len(kvs) == 0 {
		return nil
	}

	newEtcdInfos := []EtcdInfo{}
	err = json.Unmarshal(kvs[0].Value, &newEtcdInfos)
	if err != nil {
		logrus.Error("Unmarshal etcd json error: ", err)
		return err
	}

	EtcdInfos = newEtcdInfos
	logrus.Debugf("Reload etcd json suc: %v", EtcdInfos)
	return nil
}

// 监听etcd的key 并且执行回调函数
func WatchEtcd(option Option) {
	watcher := clientv3.NewWatcher(etcdClient)
	wCh := watcher.Watch(context.TODO(), Configs.FullName)
	for wresp := range wCh {
		if len(wresp.Events) == 0 {
			continue
		}
		if err := reloadEtcd(); err == nil {
			option()
		}
	}
}
