## `logagent`

日志收集客户端，可以监听指定的日志文件，并且将其中的日志消息实时发送给消息队列。

### 目录结构

```shell
.
├── README.md      
├── collects
├── conf
├── docs
├── go.mod
├── go.sum
├── main.go
├── mq
├── test
└── utils
```

- `collect`: 服务的具体逻辑，负责监听日志文件，并且将日志消息发送给消息队列。
- `conf`: 配置文件管理，即使用了本地文件`configs.yml`，也使用了`etcd`。
- `docs`: 本地配置文件。
- `mq`: 消息队列，代码中使用的是`kafka`，也可以根据需要替换成其他工具，替换时需要改动的代码量很少。
- `test`: 测试文件，只写了几个测试样例。
- `utils`: 通用工具。

### 配置文件说明

本地配置存放的是`etcd`相关信息，`etcd`存放的是日志收集相关信息。

`etcd`的`key`为`/logcollects/{本机ip}/logagent.json`。

`etcd`的`value`是`json`字符串：

```json
[
    {
        "name": "log",
        "mqhosts": ["10.1.3.95:9092"],
        "path": "/root/sub/file.log",
    }
]
```

`name`是日志的类别，`mqhosts`是存放该类别日志的消息队列，`path`是日志文件的绝对地址，注意`windows`下的目录路径依然以`/`分隔。

可以根据需要，在数组中添加多个日志的配置。