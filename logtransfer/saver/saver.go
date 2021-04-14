package saver

import (
	"errors"
	"fmt"
)

// 存储器的客户端接口
type dbclt interface {
	insert(index string, content string) error
	update(hosts ...string) error
	close()
}

// 存储器
type Saver struct {
	index string
	hosts []string
	clt   dbclt
}

// 存储器的类型
type FLAG int

// 存储器配置
type SaverConf struct {
	Flag  FLAG
	Index string
	Hosts []string
}

// 指定支持的存储器类型
const (
	ElasticSearch FLAG = iota
)

// NewSaver 初始化存储器
func NewSaver(configs SaverConf) (*Saver, error) {
	var clt dbclt
	var err error

	switch configs.Flag {
	case ElasticSearch:
		clt, err = newElasticsearch(configs.Hosts...)
	default:
		err = fmt.Errorf("wrong saver flag: %d", configs.Flag)
	}

	if err != nil {
		return nil, err
	}

	return &Saver{
		index: configs.Index,
		hosts: configs.Hosts,
		clt:   clt,
	}, nil
}

// Insert 插入资源
func (s *Saver) Insert(content string) error {
	if s.clt != nil {
		return s.clt.insert(s.index, content)
	}

	return errors.New("insert() not implemented")
}

// Update 更新资源
func (s *Saver) Update(config SaverConf) error {
	if s.clt != nil {
		return s.clt.update(config.Hosts...)
	}
	return nil
}

// Close 释放资源
func (s *Saver) Close() {
	if s.clt != nil {
		s.clt.close()
	}
}
