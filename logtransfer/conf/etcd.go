package conf

import (
	"context"
	"encoding/json"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"
)

type EtcdInfo struct {
	Title   string
	MqHosts []string
	DbHosts []string
}

type option func()

var etcdClient *clientv3.Client
var EtcdInfos []EtcdInfo

func initEtcdConfig() {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   Configs.Endpoints,
		DialTimeout: time.Duration(Configs.DialTimeout) * time.Second,
	})
	if err != nil {
		logrus.Fatal("new etcd client error: ", err)
	}
	etcdClient = client

	// 检查健康状况
	if !checkEtcdHealth() {
		logrus.Fatal("Connect etcd fail, clusters: %v", Configs.Endpoints)
	}

	resp, err := client.Get(context.TODO(), Configs.FullName)
	if err != nil {
		client.Close()
		logrus.Fatalf("get etcd key %s error: %s", Configs.FullName, err.Error())
	}

	kvs := resp.Kvs
	if len(kvs) == 0 {
		_, err = client.Put(context.TODO(), Configs.FullName, "[]")
		if err != nil {
			logrus.Fatalf("Put etcd key %s error: %s", Configs.FullName, err.Error())
		}
	} else {
		err = json.Unmarshal(kvs[0].Value, &EtcdInfos)
		if err != nil {
			logrus.Fatal("unmarshal json error: ", err)
		}
		logrus.Debugf("Unmarshal etcd json suc: %v", EtcdInfos)
	}
}

// 重新读取etcd的配置信息
func reloadEtcdConfigs() error {
	// EtcdInfos = []EtcdInfo{}
	resp, err := etcdClient.Get(context.TODO(), Configs.FullName)
	if err != nil {
		logrus.Error("Get key: %s error: %v", Configs.FullName, err)
		return err
	}

	kvs := resp.Kvs
	if len(kvs) == 0 {
		return nil
	}

	newEtcdInfos := []EtcdInfo{}
	err = json.Unmarshal(kvs[0].Value, &newEtcdInfos)
	if err != nil {
		logrus.Error("unmarshal json error: ", err)
		return err
	}

	EtcdInfos = newEtcdInfos
	logrus.Debugf("Reload etcd json suc: %v", EtcdInfos)
	return nil
}

// checkEtcdHealth 检查etcd的健康状况
func checkEtcdHealth() bool {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Duration(Configs.DialTimeout)*time.Second)
	defer cancel()
	for _, ip := range Configs.Endpoints {
		_, err := etcdClient.Status(timeoutCtx, ip)
		if err != nil {
			return false
		}
	}
	return true
}

// 监听etcd的key 并且执行回调函数
func WatchEtcd(option option) {
	watcher := clientv3.NewWatcher(etcdClient)
	wCh := watcher.Watch(context.TODO(), Configs.FullName)
	for wresp := range wCh {
		if len(wresp.Events) == 0 {
			continue
		}
		if err := reloadEtcdConfigs(); err == nil {
			option()
		}
	}
}
