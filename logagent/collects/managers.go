package collects

import (
	"logagent/conf"

	"github.com/sirupsen/logrus"
)

var logManagers = map[string]*FileManager{}

func UpdateManagers() {
	// 移除旧的收集器
	m := map[string]struct{}{}
	for _, v := range conf.EtcdInfos {
		m[v.Name] = struct{}{}
	}
	for k := range logManagers {
		if _, ok := m[k]; !ok {
			logManagers[k].close()
			delete(logManagers, k)
		}
	}

	// 更新收集器
	var err error
	for _, config := range conf.EtcdInfos {
		fm := logManagers[config.Name]
		if fm == nil {
			fm, err = newFileManager(config.Name, config.Path, config.MqHosts...)
			if err != nil {
				logrus.Errorf("new file manager error: %v, config: %v", err, config)
				continue
			}
		} else {
			err = fm.update(config.Name, config.Path, config.MqHosts...)
			if err != nil {
				logrus.Errorf("update file manager error: %v, config: %v", err, config)
				logManagers[config.Name].close()
				delete(logManagers, config.Name)
				continue
			}
		}

		logManagers[config.Name] = fm
	}
}

func CloseManagers() {
	for _, m := range logManagers {
		m.close()
	}
	logrus.Debug("Closed all managers")
}
