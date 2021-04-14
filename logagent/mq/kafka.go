package mq

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

// kafka 生产者结构体
type kafkaProducer struct {
	hosts       []string                     // kafka地址
	prod        sarama.SyncProducer          // 生产者对象
	sendChan    chan *sarama.ProducerMessage // 发送channel
	kafkaConfig *sarama.Config
	cancel      context.CancelFunc
}

// newKafkaProducer 初始化kafka生产者
func newKafkaProducer(clusters []string) (*kafkaProducer, error) {
	if len(clusters) == 0 {
		return nil, errors.New("conf.Cluster is not exists")
	}

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true

	client, err := sarama.NewSyncProducer(clusters, config)
	if err != nil {
		return nil, err
	}

	kafka := &kafkaProducer{}
	kafka.sendChan = make(chan *sarama.ProducerMessage)
	kafka.hosts = clusters
	kafka.prod = client
	kafka.kafkaConfig = config

	ctx, cancel := context.WithCancel(context.Background())
	kafka.cancel = cancel
	go kafka.work(ctx)

	logrus.Debugf("Connect kafka suc, clusters: %v", clusters)
	return kafka, nil
}

// work kafka生产者开始工作
func (kafka *kafkaProducer) work(ctx context.Context) {
	var prodMsg *sarama.ProducerMessage
	var isBreak bool
	for !isBreak {
		select {
		case <-ctx.Done():
			isBreak = true
		case prodMsg = <-kafka.sendChan:
			if prodMsg == nil {
				isBreak = true
				break
			}
			partition, offset, err := kafka.prod.SendMessage(prodMsg)
			logrus.Debugf("Send kafka message, result: partition %v, offseet %v, err %v", partition, offset, err)
		}
	}
}

// send 将消息发送给channel
func (kafka *kafkaProducer) produce(mqMsg MessageQueueMessage) {
	// 去除首尾空格
	mqMsg["message"] = strings.Trim(mqMsg["message"], " ")
	mqMsg["topic"] = strings.Trim(mqMsg["topic"], " ")
	if mqMsg["message"] == "" || mqMsg["topic"] == "" {
		return
	}

	kafkaMsg := &sarama.ProducerMessage{}
	kafkaMsg.Topic = mqMsg["topic"]
	kafkaMsg.Value = sarama.StringEncoder(mqMsg["message"])

	// 进行超时重试 用于处理重新初始化sendChan的情况
	retry := true
	for retry {
		select {
		case <-time.After(time.Second):
			retry = false
		case kafka.sendChan <- kafkaMsg:
			retry = false
		}
	}
}

// close 释放资源
func (kafka *kafkaProducer) close() {
	// 停止生产
	if kafka.cancel != nil {
		kafka.cancel()
		kafka.cancel = nil
	}
	// 释放连接
	if kafka.prod != nil {
		kafka.prod.Close()
		kafka.prod = nil
	}
	// 关闭channel
	if kafka.sendChan != nil {
		close(kafka.sendChan)
		kafka.sendChan = nil
	}
}

// 更新
func (kafka *kafkaProducer) update(clusters ...string) error {
	if len(clusters) == len(kafka.hosts) {
		// slice转换为map
		m := map[string]struct{}{}
		for _, host := range clusters {
			m[host] = struct{}{}
		}
		// 检查是否有变化
		var update bool
		for _, old := range kafka.hosts {
			if _, ok := m[old]; ok {
				update = true
				break
			}
		}
		if !update {
			return nil
		}
	}

	kafka.close()
	kafka.hosts = []string{}

	// 创建新的连接
	client, err := sarama.NewSyncProducer(clusters, kafka.kafkaConfig)
	if err != nil {
		return err
	}
	kafka.hosts = clusters
	kafka.prod = client
	kafka.sendChan = make(chan *sarama.ProducerMessage)

	ctx, cancel := context.WithCancel(context.Background())
	go kafka.work(ctx)
	kafka.cancel = cancel
	return nil
}
