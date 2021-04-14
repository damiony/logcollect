package mq

import (
	"context"
	"errors"
	"time"

	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

// kafkaConsumerGroup kafka消费者组
type kafkaConsumerGroup struct {
	topic    string               // kafka topic
	hosts    []string             // kafka brokers
	version  string               // kafka version
	group    sarama.ConsumerGroup // kafka consumer group
	sendChan chan string          // consume message channel
	cancel   context.CancelFunc   // context
	config   *sarama.Config       // kafka consumer configs
}

// newKafkaConsumerGroup 初始化
func newKafkaConsumerGroup(topic, groupId string, hosts ...string) (*kafkaConsumerGroup, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_4_0_0 // sarama版本
	config.Consumer.Return.Errors = true
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky // 重平衡策略
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	group, err := sarama.NewConsumerGroup(hosts, groupId, config)
	if err != nil {
		return nil, err
	}

	kConsumerGroup := &kafkaConsumerGroup{
		topic:    topic,                   // kafka topic
		hosts:    hosts,                   // kafka brokers
		version:  config.Version.String(), // kafka consumer version
		group:    group,                   // kafka consumer group
		sendChan: make(chan string),       // message channel
		config:   config,                  // kafka consumer config
	}

	// 监听错误信息
	go func() {
		for err := range group.Errors() {
			logrus.Error("kafka consumer group error: ", err)
		}
		logrus.Debug("kafka group errors group exit.")
	}()

	// 消费消息
	ctx, cancel := context.WithCancel(context.Background())
	go kConsumerGroup.work(ctx)
	kConsumerGroup.cancel = cancel

	return kConsumerGroup, nil
}

// Setup saram 要求的方法
func (*kafkaConsumerGroup) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup sarama 要求的方法
func (*kafkaConsumerGroup) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim sarama 要求的方法
func (kg *kafkaConsumerGroup) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		logrus.Debugf("Message topic:%q partition:%d offset:%d", msg.Topic, msg.Partition, msg.Offset)
		kg.sendChan <- string(msg.Value)
		sess.MarkMessage(msg, "")
	}
	return nil
}

// work 消费者组开始从kafka消费消息
func (kg *kafkaConsumerGroup) work(ctx context.Context) {
	// 错误重试次数
	counts := 0
	for {
		select {
		case <-ctx.Done():
			logrus.Debug("Kafka parents context cancel, work stop.")
			return
		default:
			err := kg.group.Consume(context.Background(), []string{kg.topic}, kg)
			if err != nil {
				logrus.Error("kafka consume messages error: ", err)
				time.Sleep(time.Second) // 如果消费出错 则等待1s
				counts++
			}
		}
		// 错误重试次数为100
		if counts >= 100 {
			break
		}
	}

	logrus.Errorf("kafka consumer group consume message error counts is %d, exit.", counts)
}

// consumer 从channel获取消息
func (kg *kafkaConsumerGroup) consume() (string, error) {
	msg, ok := <-kg.sendChan
	if !ok {
		return "", errors.New("send channel has closed")
	}
	return msg, nil
}

// update 更新消费者组
func (kg *kafkaConsumerGroup) update(topic, groupId string, hosts ...string) error {
	if kg.config == nil {
		return errors.New("kafkaConsumerGroup config must be created.")
	}
	// 先释放资源
	kg.group.Close()
	kg.group = nil
	kg.cancel()
	kg.cancel = nil

	// 重新创建消费者组
	group, err := sarama.NewConsumerGroup(hosts, groupId, kg.config)
	if err != nil {
		return err
	}
	// 更新数据
	kg.hosts = hosts
	kg.group = group
	// 监听错误信息
	go func() {
		for err := range group.Errors() {
			logrus.Error("kafka consumer group error: ", err)
		}
	}()

	// 消费消息
	ctx, cancel := context.WithCancel(context.Background())
	go kg.work(ctx)
	kg.cancel = cancel

	return nil
}

// close 释放资源
func (kg *kafkaConsumerGroup) close() {
	if kg.sendChan != nil {
		close(kg.sendChan)
		kg.sendChan = nil
	}
	if kg.group != nil {
		kg.group.Close()
		kg.group = nil
	}
	if kg.cancel != nil {
		kg.cancel()
		kg.cancel = nil
	}
}
