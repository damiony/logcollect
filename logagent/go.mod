module logagent

go 1.14

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require (
	github.com/Shopify/sarama v1.19.0
	github.com/coreos/etcd v3.3.25+incompatible // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/hpcloud/tail v1.0.0
	github.com/prometheus/client_golang v1.10.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/viper v1.7.1
	go.etcd.io/etcd v3.3.25+incompatible
	go.uber.org/zap v1.16.0 // indirect
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2 // indirect
)
