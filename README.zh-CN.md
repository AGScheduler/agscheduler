# AGScheduler

[![test](https://github.com/agscheduler/agscheduler/actions/workflows/test.yml/badge.svg)](https://github.com/agscheduler/agscheduler/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/agscheduler/agscheduler/graph/badge.svg?token=CL5P4VYQTU)](https://codecov.io/gh/agscheduler/agscheduler)
[![Go Report Card](https://goreportcard.com/badge/github.com/agscheduler/agscheduler)](https://goreportcard.com/report/github.com/agscheduler/agscheduler)
[![Go Reference](https://pkg.go.dev/badge/github.com/agscheduler/agscheduler.svg)](https://pkg.go.dev/github.com/agscheduler/agscheduler)
![GitHub release (with filter)](https://img.shields.io/github/v/release/agscheduler/agscheduler)
![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/agscheduler/agscheduler)
[![license](https://img.shields.io/github/license/agscheduler/agscheduler)](https://github.com/agscheduler/agscheduler/blob/main/LICENSE)

> Advanced Golang Scheduler (AGScheduler) 是一款适用于 Golang 的任务调度库，支持多种调度类型，支持动态更改和持久化作业，支持作业队列，支持作业结果回收，支持事件监听，支持远程调用，支持集群

[English](README.md) | 简体中文

## 特性

- 支持三种调度类型
  - [x] 一次性执行
  - [x] 间隔执行
  - [x] Cron 式调度
- 支持多种作业存储方式
  - [x] Memory (不支持集群 HA 模式)
  - [x] [GORM](https://gorm.io/) (任何 GORM 支持的 RDBMS 都能运行)
  - [x] [Redis](https://redis.io/)
  - [x] [MongoDB](https://www.mongodb.com/)
  - [x] [etcd](https://etcd.io/)
  - [x] [Elasticsearch](https://www.elastic.co/elasticsearch)
- 支持多种作业队列
  - [x] Memory (不支持集群模式)
  - [x] [NSQ](https://nsq.io/)
  - [x] [RabbitMQ](https://www.rabbitmq.com/)
  - [x] [Redis](https://redis.io/)
  - [x] [MQTT](https://mqtt.org/) (不支持历史作业)
  - [x] [Kafka](https://kafka.apache.org/)
- 支持多种作业结果后端
  - [x] Memory (不支持集群模式)
  - [x] [GORM](https://gorm.io/) (任何 GORM 支持的 RDBMS 都能运行)
  - [x] [MongoDB](https://www.mongodb.com/)
- 支持事件监听
  - [x] 调度器事件
  - [x] 作业事件
- 支持远程调用
  - [x] [gRPC](https://grpc.io/)
  - [x] HTTP
- 支持集群
  - [x] 远程工作节点
  - [x] 调度器高可用 (实验性)

## 架构

![Framework](assets/framework.png)

## 安装

```bash
go get -u github.com/agscheduler/agscheduler
```

可选集成现在使用独立 module，只会引入对应实现所需的依赖：

```bash
go get github.com/agscheduler/agscheduler/stores/memory
go get github.com/agscheduler/agscheduler/queues/kafka
go get github.com/agscheduler/agscheduler/backends/mongodb
go get github.com/agscheduler/agscheduler/services/grpc
```

## 从 v0.12 及更早版本迁移

`v0.20.0` 将聚合包拆分为按具体实现组织的独立 module。核心包名保持不变，只需要将 import 更新为实际使用的实现：

各实现的 package 名仍为 `stores`、`queues`、`backends` 和 `services`，因此 `stores.MemoryStore` 等类型引用不需要改名。

| v0.12 及更早版本 | v0.20.0 |
| --- | --- |
| `github.com/agscheduler/agscheduler/stores` | `github.com/agscheduler/agscheduler/stores/memory`、`.../stores/gorm`、`.../stores/redis`、`.../stores/mongodb`、`.../stores/etcd` 或 `.../stores/elasticsearch` |
| `github.com/agscheduler/agscheduler/queues` | `github.com/agscheduler/agscheduler/queues/memory`、`.../queues/nsq`、`.../queues/rabbitmq`、`.../queues/redis`、`.../queues/mqtt` 或 `.../queues/kafka` |
| `github.com/agscheduler/agscheduler/backends` | `github.com/agscheduler/agscheduler/backends/memory`、`.../backends/gorm` 或 `.../backends/mongodb` |
| `github.com/agscheduler/agscheduler/services` | `github.com/agscheduler/agscheduler/services/http`、`.../services/grpc` 或 `.../services/cluster` |

`JobToPbJobPtr` 等 Protobuf 转换函数已从核心包迁移到 `github.com/agscheduler/agscheduler/services/grpc/convert`。

## Module 开发

Module 目录和依赖边界分别声明在 `scripts/modules.txt` 与 `scripts/module-deps.txt`。

```bash
# 创建并登记新的可选 module。
make add-module MODULE=stores/cassandra DEPS='github.com/gocql/gocql@v1.6.0'

# 统一更新所有 module、workspace replacement、版本常量和迁移文档。
make bump-version VERSION=0.21.0

# 添加源码或修改依赖后，刷新 go.mod/go.sum。
make tidy
make check-deps
```

添加源码后运行受影响 module 的测试。如果测试需要某个第三方 driver，而该依赖不应出现在适配器 module 的 `go.mod` 中，应把测试放到 `tests/integration`。

## 使用

```go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/stores/memory"
)

func printMsg(ctx context.Context, j agscheduler.Job) (result string) {
	slog.Info(fmt.Sprintf("Run job `%s` %s\n\n", j.FullName(), j.Args))
	return
}

func main() {
	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: printMsg},
	)

	scheduler := &agscheduler.Scheduler{}

	store := &stores.MemoryStore{}
	scheduler.SetStore(store)

	job1 := agscheduler.Job{
		Name:     "Job1",
		Type:     agscheduler.JOB_TYPE_INTERVAL,
		Interval: "2s",
		Timezone: "UTC",
		Func:     printMsg,
		Args:     map[string]any{"arg1": "1", "arg2": "2", "arg3": "3"},
	}
	job1, _ = scheduler.AddJob(job1)
	slog.Info(fmt.Sprintf("%s.\n\n", job1))

	job2 := agscheduler.Job{
		Name:     "Job2",
		Type:     agscheduler.JOB_TYPE_CRON,
		CronExpr: "*/1 * * * *",
		Timezone: "Asia/Shanghai",
		FuncName: "main.printMsg",
		Args:     map[string]any{"arg4": "4", "arg5": "5", "arg6": "6", "arg7": "7"},
	}
	job2, _ = scheduler.AddJob(job2)
	slog.Info(fmt.Sprintf("%s.\n\n", job2))

	job3 := agscheduler.Job{
		Name:     "Job3",
		Type:     agscheduler.JOB_TYPE_DATETIME,
		StartAt:  "2023-09-22 07:30:08",
		Timezone: "America/New_York",
		Func:     printMsg,
		Args:     map[string]any{"arg8": "8", "arg9": "9"},
	}
	job3, _ = scheduler.AddJob(job3)
	slog.Info(fmt.Sprintf("%s.\n\n", job3))

	jobs, _ := scheduler.GetAllJobs()
	slog.Info(fmt.Sprintf("Scheduler get all jobs %s.\n\n", jobs))

	scheduler.Start()

	select {}
}
```

## 优雅退出

`Scheduler.Stop()` 会停止调度新任务，并默认最多等待 `DefaultShutdownTimeout`，让正在执行的任务完成。需要处理系统信号时使用 `WaitForShutdown`：

```go
import (
	"fmt"
	"log/slog"
	"time"
)

scheduler.Start()
if err := scheduler.WaitForShutdown(30*time.Second, service.Stop); err != nil {
	slog.Warn(fmt.Sprintf("Scheduler graceful shutdown failed: %s", err))
}
```

`WaitForShutdown` 处理 SIGINT 和 SIGTERM。第一次信号调用 `service.Stop` 并优雅关闭调度器，第二次信号立即退出。宽限期到期后，正在执行的任务会收到已取消的 Context，并记录为 `RECORD_STATUS_CANCELLED`。任务函数应监听 `ctx.Done()`，以便及时退出；没有需要预先关闭的服务时，回调传 `nil`。

## 注册函数

> **_由于 golang 无法序列化函数，所以 `scheduler.Start()` 之前需要使用 `RegisterFuncs` 注册函数_**

## 队列

```go
mq := &queues.MemoryQueue{}
broker := &agscheduler.Broker{
	Queues: map[string]agscheduler.QueuePkg{
		"default": {
			Queue:   mq,
			Workers: 2,
		},
	},
}

scheduler.SetBroker(broker)
```

## 结果回收

```go
backend := &backends.MemoryBackend{}
recorder := &agscheduler.Recorder{Backend: backend}

scheduler.SetRecorder(recorder)

job, _ = scheduler.AddJob(job)
records, _ := recorder.GetRecords(job.Id)
```

## 事件监听
```go
func jobCallback(ep agscheduler.EventPkg) {
	slog.Info(fmt.Sprintf("Event code: `%d`, job `%s`.\n\n", ep.Event, ep.JobId))
}

......

listener := &agscheduler.Listener{
	Callbacks: []agscheduler.CallbackPkg{
		{
			Callback: jobCallback,
			Event:    agscheduler.EVENT_JOB_ADDED | agscheduler.EVENT_JOB_DELETED,
		},
	},
}

scheduler.SetListener(listener)
```

## gRPC

```go
// Server
grservice := services.GRPCService{
	Scheduler: scheduler,
	Address:   "127.0.0.1:36360",
	// PasswordSha2: "xxxxxx",
}
grservice.Start()

// Client
conn, _ := grpc.NewClient("127.0.0.1:36360", grpc.WithTransportCredentials(insecure.NewCredentials()))
client := pb.NewSchedulerClient(conn)
client.AddJob(ctx, job)
```

## HTTP

```go
// Server
hservice := services.HTTPService{
	Scheduler: scheduler,
	Address:   "127.0.0.1:36370",
	// PasswordSha2: "xxxxxx",
}
hservice.Start()

// Client
mJob := map[string]any{...}
bJob, _ := json.Marshal(mJob)
resp, _ := http.Post("http://127.0.0.1:36370/scheduler/job", "application/json", bytes.NewReader(bJob))
```

## 集群

```go
// Main Node
cnMain := &agscheduler.ClusterNode{
	Endpoint:     "127.0.0.1:36380",
	EndpointGRPC: "127.0.0.1:36360",
	EndpointHTTP: "127.0.0.1:36370",
	Queue:        "default",
}
schedulerMain.SetClusterNode(ctx, cnMain)
cserviceMain := &services.ClusterService{Cn: cnMain}
cserviceMain.Start()

// Worker Node
cnNode := &agscheduler.ClusterNode{
	EndpointMain: "127.0.0.1:36380",
	Endpoint:     "127.0.0.1:36381",
	EndpointGRPC: "127.0.0.1:36361",
	EndpointHTTP: "127.0.0.1:36371",
	Queue:        "worker",
}
schedulerNode.SetClusterNode(ctx, cnNode)
cserviceNode := &services.ClusterService{Cn: cnNode}
cserviceNode.Start()
```

## 集群 HA (高可用，实验性)

```go
// HA 需要满足以下条件：
//
// 1. 集群中 HA 节点的数量必须为奇数
// 2. 所有 HA 节点都需要连接到同一个存储（不包含 MemoryStore）
// 3. ClusterNode 的 Mode 属性需要设置为 `HA`
// 4. HA 主节点必须先启动

// Main HA Node
cnMain := &agscheduler.ClusterNode{..., Mode: "HA"}

// HA Node
cnNode1 := &agscheduler.ClusterNode{..., Mode: "HA"}
cnNode2 := &agscheduler.ClusterNode{..., Mode: "HA"}

// Worker Node
cnNode3 := &agscheduler.ClusterNode{...}
```

## Base API

| gRPC Function | HTTP Method | HTTP Path                 |
|---------------|-------------|---------------------------|
| GetInfo       | GET         | /info                     |
| GetFuncs      | GET         | /funcs                    |

## Scheduler API

| gRPC Function | HTTP Method | HTTP Path                 |
|---------------|-------------|---------------------------|
| AddJob        | POST        | /scheduler/job            |
| GetJob        | GET         | /scheduler/job/:id        |
| GetAllJobs    | GET         | /scheduler/jobs           |
| UpdateJob     | PUT         | /scheduler/job            |
| DeleteJob     | DELETE      | /scheduler/job/:id        |
| DeleteAllJobs | DELETE      | /scheduler/jobs           |
| PauseJob      | POST        | /scheduler/job/:id/pause  |
| ResumeJob     | POST        | /scheduler/job/:id/resume |
| RunJob        | POST        | /scheduler/job/run        |
| ScheduleJob   | POST        | /scheduler/job/schedule   |
| Start         | POST        | /scheduler/start          |
| Stop          | POST        | /scheduler/stop           |

## Broker API

| gRPC Function | HTTP Method | HTTP Path                 |
|---------------|-------------|---------------------------|
| GetQueues     | GET         | /broker/queues            |

## Recorder API

| gRPC Function | HTTP Method | HTTP Path                 |
|---------------|-------------|---------------------------|
| GetRecords    | GET         | /recorder/records/:job_id |
| GetAllRecords | GET         | /recorder/records         |
| DeleteRecords | DELETE      | /recorder/records/:job_id |
| DeleteAllRecords | DELETE   | /recorder/records         |

## Cluster API

| gRPC Function | HTTP Method | HTTP Path                 |
|---------------|-------------|---------------------------|
| GetNodes      | GET         | /cluster/nodes            |

## 示例

[完整示例][examples]

## 开发

```bash
# 克隆代码
git clone git@github.com:agscheduler/agscheduler.git

# 工作目录
cd agscheduler

# 安装依赖
make install

# 启动 CI 服务
make up-ci-services

# 运行检查
make check-all
```

## [Cli](https://github.com/AGScheduler/agscheduler-cli)

```bash
cargo install agscheduler-cli
```

## [Web](https://github.com/AGScheduler/agscheduler-web)

```bash
docker run --rm -p 8080:80 ghcr.io/agscheduler/agscheduler-web:latest
```

## 致谢

[APScheduler](https://github.com/agronholm/apscheduler/tree/3.x)

[simple-raft](https://github.com/chapin666/simple-raft)

[examples]: https://github.com/agscheduler/agscheduler/tree/main/examples
