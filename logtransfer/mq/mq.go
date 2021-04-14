package mq

import (
	"fmt"
)

// 消息队列客户端接口
type consumer interface {
	consume() (string, error)
	update(string, string, ...string) error
	close()
}

type MqConf struct {
	Flag  Flag
	Topic string
	Hosts []string
}

// 消息队列
type MessageQueue struct {
	Topic    string
	Hosts    []string
	Consumer consumer
}

// 消息队列的类型
type Flag int

// 支持的消息队列类型
const (
	KAFKA Flag = iota
)

// group 暂时写死
var GROUP_ID = "my-group"

// NewMessageQueue 初始化消息队列
func NewMessageQueue(mqconf MqConf) (*MessageQueue, error) {
	var c consumer
	var err error

	switch mqconf.Flag {
	case KAFKA:
		// groupId 暂时写死
		c, err = newKafkaConsumerGroup(mqconf.Topic, GROUP_ID, mqconf.Hosts...)
	default:
		err = fmt.Errorf("Not implement flag: %d", mqconf.Flag)
	}

	if err != nil {
		return nil, err
	}

	mq := &MessageQueue{
		Topic:    mqconf.Topic,
		Hosts:    mqconf.Hosts,
		Consumer: c,
	}
	return mq, nil
}

// Consume 消费消息
func (mq *MessageQueue) Consume() (string, error) {
	return mq.Consumer.consume()
}

// Update 更新资源
func (mq *MessageQueue) Update(config MqConf) error {
	return mq.Consumer.update(config.Topic, GROUP_ID, config.Hosts...)
}

// Close 释放资源
func (mq *MessageQueue) Close() {
	if mq.Consumer != nil {
		mq.Consumer.close()
	}
}
