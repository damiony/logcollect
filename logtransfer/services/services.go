package services

import (
	"context"
	"fmt"
	"logtransfer/conf"
	"logtransfer/mq"
	"logtransfer/saver"
	"time"

	"github.com/sirupsen/logrus"
)

type manager struct {
	topic    string
	consumer *mq.MessageQueue
	saver    *saver.Saver
	cancel   context.CancelFunc
}

var logManagers map[string]*manager

func newManager(title string, dbhosts []string, mqhosts []string) (*manager, error) {
	if title == "" || len(dbhosts) == 0 || len(mqhosts) == 0 {
		return nil, fmt.Errorf("Wrong parameters: title %s, dbhosts %v, mqhosts %v", title, dbhosts, mqhosts)
	}

	consumer, err := mq.NewMessageQueue(mq.MqConf{
		Flag:  mq.KAFKA,
		Topic: title,
		Hosts: mqhosts,
	})
	if err != nil {
		return nil, err
	}
	saver, err := saver.NewSaver(saver.SaverConf{
		Flag:  saver.ElasticSearch,
		Index: title,
		Hosts: dbhosts,
	})
	if err != nil {
		consumer.Close()
		return nil, err
	}

	return &manager{
		topic:    title,
		consumer: consumer,
		saver:    saver,
	}, nil
}

func InitManagers() {
	logManagers = map[string]*manager{}
	for _, eInfo := range conf.EtcdInfos {
		m, err := newManager(eInfo.Title, eInfo.DbHosts, eInfo.MqHosts)
		if err != nil {
			logrus.Error("new manager error: ", err)
			continue
		}
		logManagers[eInfo.Title] = m
		go m.work()
	}
}

func UpdateManagers() {
	// 转换为map
	m := map[string]bool{}
	for i := range conf.EtcdInfos {
		m[conf.EtcdInfos[i].Title] = true
	}
	// 删除旧数据
	for i := range logManagers {
		if !m[i] {
			logManagers[i].consumer.Close()
			logManagers[i].saver.Close()
			delete(logManagers, i)
		}
	}
	// 更新数据
	for _, eInfo := range conf.EtcdInfos {
		if eInfo.Title == "" || len(eInfo.DbHosts) == 0 || len(eInfo.MqHosts) == 0 {
			logrus.Errorf("Wrong etcd info: %v, manager exit.", eInfo)
			delete(logManagers, eInfo.Title)
			continue
		}

		var err error
		manager := logManagers[eInfo.Title]
		if manager == nil {
			manager, err = newManager(eInfo.Title, eInfo.DbHosts, eInfo.MqHosts)
			if err != nil {
				logrus.Error("Create new manager error: ", err)
			} else {
				logManagers[eInfo.Title] = manager
				go manager.work()
			}
			continue
		}

		err = manager.update(eInfo.Title, eInfo.DbHosts, eInfo.MqHosts)
		if err != nil {
			logrus.Error("Update manager err: ", err)
			delete(logManagers, eInfo.Title)
		}
	}
	return
}

func CloseManagers() {
	for _, m := range logManagers {
		m.close()
	}
	logrus.Debug("Closed all managers")
}

func (m *manager) work() {
	for {
		message, err := m.consumer.Consume()
		if err != nil {
			logrus.Errorf("consume messages from mq error: %s, exit.", err.Error())
			return
		}

		err = m.saver.Insert(message)
		if err != nil {
			logrus.Errorf("Save error: %s, message: %s, wait one second.", err.Error(), message)
			time.Sleep(time.Second)
		}
	}
}

func (m *manager) update(title string, dbhosts []string, mqhosts []string) error {
	if m.consumer == nil || m.saver == nil {
		return fmt.Errorf("Manager must implement consumer and saver.")
	}
	// 更新消费者
	err := m.consumer.Update(mq.MqConf{
		Flag:  mq.KAFKA,
		Topic: title,
		Hosts: mqhosts,
	})
	if err != nil {
		return err
	}
	// 更新存储器
	err = m.saver.Update(saver.SaverConf{
		Flag:  saver.ElasticSearch,
		Index: title,
		Hosts: dbhosts,
	})
	if err != nil {
		m.consumer.Close()
		return err
	}

	return nil
}

func (m *manager) close() {
	m.consumer.Close()
	m.saver.Close()
}
