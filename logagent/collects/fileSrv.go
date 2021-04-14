package collects

import (
	"context"
	"logagent/mq"
	"time"

	"github.com/hpcloud/tail"
	"github.com/sirupsen/logrus"
)

type FileManager struct {
	Topic    string
	Path     string
	TailConf tail.Config
	Tail     *tail.Tail
	Producer *mq.MessageQueueProducer
	Cancel   context.CancelFunc
}

func newFileManager(topic string, path string, hosts ...string) (*FileManager, error) {
	config := tail.Config{
		ReOpen:    true,
		MustExist: true,
		Poll:      true,
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 0},
	}
	tail, err := tail.TailFile(path, config)
	if err != nil {
		return nil, err
	}

	producer, err := mq.NewMessageQueueProducer(mq.MqConf{
		Flag:     mq.KAFKA,
		Clusters: hosts,
	})
	if err != nil {
		tail.Done()
		return nil, err
	}

	tm := &FileManager{
		Path:     path,
		Topic:    topic,
		Producer: producer,
		Tail:     tail,
		TailConf: config,
	}
	// 收集数据
	ctx, cancel := context.WithCancel(context.Background())
	tm.Cancel = cancel
	go tm.collect(ctx)

	return tm, nil
}

// collect 收集数据
func (tm *FileManager) collect(ctx context.Context) {
	lines := tm.Tail.Lines
	var line *tail.Line
	var ok bool
	for {
		select {
		case <-ctx.Done(): // 使用ctx控制channel监听
			return
		case line, ok = <-lines:
			if !ok {
				logrus.Warnf("closed channel %s, wait some time...", tm.Path)
				time.Sleep(1 * time.Second)
				continue
			}
			if line.Text == "" || line.Text == "\n" || line.Text == "\r\n" || line.Text == "\r" {
				logrus.Debugf("read invaild content %v from path %s", line.Text, tm.Path)
				continue
			}
			tm.Producer.Produce(mq.MessageQueueMessage{
				"topic":   tm.Topic,
				"message": line.Text,
			})
		}
	}
}

func (tm *FileManager) update(topic string, path string, hosts ...string) error {
	// 关闭日志收集
	if tm.Cancel != nil {
		tm.Cancel()
	}

	// 更新消息队列
	err := tm.Producer.Update(mq.MqConf{
		Flag:     mq.KAFKA,
		Clusters: hosts,
	})
	if err != nil {
		return err
	}

	// 更新收集器信息
	if tm.Path != path {
		if tm.Tail != nil {
			tm.Tail.Stop()
		}
		tail, err := tail.TailFile(path, tm.TailConf)
		if err != nil {
			tm.Producer.Close()
			return err
		}
		// 更新结构体信息
		tm.Tail = tail
		tm.Path = path
	}
	// 重新开始收集数据
	ctx, cancel := context.WithCancel(context.Background())
	tm.Cancel = cancel
	go tm.collect(ctx)
	return nil
}

func (tm *FileManager) close() {
	if tm == nil {
		return
	}
	if tm.Cancel != nil {
		tm.Cancel()
		tm.Cancel = nil
	}
	if tm.Tail != nil {
		tm.Tail.Stop()
		tm.Tail = nil
	}
	if tm.Producer != nil {
		tm.Producer.Close()
		tm.Producer = nil
	}
	return
}
