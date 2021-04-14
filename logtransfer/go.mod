module logtransfer

go 1.15

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require (
	github.com/Shopify/sarama v1.28.0
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/olivere/elastic/v7 v7.0.23
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v1.7.1
	sigs.k8s.io/yaml v1.2.0 // indirect
)
