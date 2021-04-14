package saver

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

type ElasticMessage struct {
	Time int64  `json:"time"`
	Msg  string `json:"msg"`
}

type ElasticSaver struct {
	client *elastic.Client
	hosts  []string
}

// newLeasticsearch
func newElasticsearch(urls ...string) (*ElasticSaver, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(urls...),
		elastic.SetSniff(false),
	)
	if err != nil {
		return nil, err
	}

	return &ElasticSaver{
		client: client,
		hosts:  urls,
	}, nil
}

// insert 插入数据
func (es *ElasticSaver) insert(index string, content string) error {
	if content == "" {
		logrus.Warn("es: Insert invalid message")
		return nil
	}
	if strings.HasSuffix(content, "\r") {
		content = strings.TrimSuffix(content, "\r")
	}
	if strings.HasSuffix(content, "\r\n") {
		content = strings.TrimSuffix(content, "\r\b")
	}
	if strings.HasSuffix(content, "\n") {
		content = strings.TrimSuffix(content, "\n")
	}
	esMsg := ElasticMessage{
		Time: time.Now().Unix(),
		Msg:  content,
	}
	b, err := json.Marshal(esMsg)
	if err != nil {
		return err
	}
	_, err = es.client.Index().Index(index).BodyJson(string(b)).Do(context.Background())
	if err != nil {
		return err
	}

	return nil
}

// update 更新连接
func (es *ElasticSaver) update(urls ...string) error {
	// 检查配置是否有变化
	if len(urls) == len(es.hosts) {
		m := map[string]bool{}
		for _, url := range urls {
			m[url] = true
		}

		var update bool
		for _, old := range es.hosts {
			if !m[old] {
				update = true
				break
			}
		}

		if !update {
			return nil
		}
	}

	// 更新连接
	client, err := elastic.NewClient(
		elastic.SetURL(urls...),
		elastic.SetSniff(false),
	)
	if err != nil {
		return err
	}
	es.hosts = urls
	es.client = client
	return nil
}

// close 释放资源
func (es *ElasticSaver) close() {
}
