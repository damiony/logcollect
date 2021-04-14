package mq

import "fmt"

// MqConf mq的配置
type MqConf struct {
	// 消息队列类型
	Flag MqType
	// 消息队列的集群
	Clusters []string
}

// producerInterface 消费者接口
type producerInterface interface {
	produce(msg MessageQueueMessage)
	update(clusters ...string) error
	close()
}

// MessageQueueProducer 消费者结构体
type MessageQueueProducer struct {
	conf     MqConf
	producer producerInterface
}

// MessageQueueMessage 消息
type MessageQueueMessage map[string]string

// MqType 消息队列类型
type MqType uint8

const (
	KAFKA MqType = 1 // mq.KAFKA kafka类型
)

// NewMessageQueueProducer 创建新的生产者
func NewMessageQueueProducer(conf MqConf) (*MessageQueueProducer, error) {
	var p producerInterface
	var err error

	switch conf.Flag {
	case KAFKA:
		p, err = newKafkaProducer(conf.Clusters)
	default:
		err = fmt.Errorf("Not implement for message queue type: %d", conf.Flag)
	}

	if err != nil {
		return nil, err
	}

	return &MessageQueueProducer{
		conf:     conf,
		producer: p,
	}, nil
}

// Produce 发送消息
func (p *MessageQueueProducer) Produce(mqMsg MessageQueueMessage) {
	p.producer.produce(mqMsg)
}

// Close 关闭生产者
func (p *MessageQueueProducer) Close() {
	if p.producer == nil {
		return
	}
	p.producer.close()
	p.producer = nil
}

// 更新生产者的信息
func (p *MessageQueueProducer) Update(conf MqConf) error {
	err := p.producer.update(conf.Clusters...)
	return err
}
