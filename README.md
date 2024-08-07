# AGScheduler

[![test](https://github.com/agscheduler/agscheduler/actions/workflows/test.yml/badge.svg)](https://github.com/agscheduler/agscheduler/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/agscheduler/agscheduler/graph/badge.svg?token=CL5P4VYQTU)](https://codecov.io/gh/agscheduler/agscheduler)
[![Go Report Card](https://goreportcard.com/badge/github.com/agscheduler/agscheduler)](https://goreportcard.com/report/github.com/agscheduler/agscheduler)
[![Go Reference](https://pkg.go.dev/badge/github.com/agscheduler/agscheduler.svg)](https://pkg.go.dev/github.com/agscheduler/agscheduler)
![GitHub release (with filter)](https://img.shields.io/github/v/release/agscheduler/agscheduler)
![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/agscheduler/agscheduler)
[![license](https://img.shields.io/github/license/agscheduler/agscheduler)](https://github.com/agscheduler/agscheduler/blob/main/LICENSE)

> Advanced Golang Scheduler (AGScheduler) is a task scheduling library for Golang that supports multiple scheduling types, dynamically changing and persistent jobs, job queues, job result collection, event listening, remote call, and cluster

English | [简体中文](README.zh-CN.md)

## Features

- Supports three scheduling types
  - [x] One-off execution
  - [x] Interval execution
  - [x] Cron-style scheduling
- Supports multiple job store methods
  - [x] Memory (Cluster HA mode is not supported)
  - [x] [GORM](https://gorm.io/) (any RDBMS supported by GORM works)
  - [x] [Redis](https://redis.io/)
  - [x] [MongoDB](https://www.mongodb.com/)
  - [x] [etcd](https://etcd.io/)
  - [x] [Elasticsearch](https://www.elastic.co/elasticsearch)
- Supports multiple job queues
  - [x] Memory (Cluster mode is not supported)
  - [x] [NSQ](https://nsq.io/)
  - [x] [RabbitMQ](https://www.rabbitmq.com/)
  - [x] [Redis](https://redis.io/)
  - [x] [MQTT](https://mqtt.org/) (History jobs are not supported)
  - [x] [Kafka](https://kafka.apache.org/)
- Supports multiple job result backends
  - [x] Memory (Cluster mode is not supported)
  - [x] [GORM](https://gorm.io/) (any RDBMS supported by GORM works)
  - [x] [MongoDB](https://www.mongodb.com/)
- Supports event listening
  - [x] Scheduler event
  - [x] Job event
- Supports remote call
  - [x] [gRPC](https://grpc.io/)
  - [x] HTTP
- Supports cluster
  - [x] Remote worker nodes
  - [x] Scheduler high availability (Experimental)

## Framework

![Framework](assets/framework.png)

## Installation

```bash
go get -u github.com/agscheduler/agscheduler
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/stores"
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

## Register Funcs

> **_Since golang can't serialize functions, you need to register them with `RegisterFuncs` before `scheduler.Start()`_**

## Queue

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

## Result Collection

```go
backend := &backends.MemoryBackend{}
recorder := &agscheduler.Recorder{Backend: backend}

scheduler.SetRecorder(recorder)

job, _ = scheduler.AddJob(job)
records, _ := recorder.GetRecords(job.Id)
```

## Event listening

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

## Cluster

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

## Cluster HA (High Availability, Experimental)

```go
// HA requires the following conditions to be met:
//
// 1. The number of HA nodes in the cluster must be odd
// 2. All HA nodes need to connect to the same Store (excluding `MemoryStore`)
// 3. The `Mode` of the `ClusterNode` needs to be set to `HA`
// 4. The main HA node must be started first

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

## Examples

[Complete example][examples]

## Development

```bash
# Clone code
git clone git@github.com:agscheduler/agscheduler.git

# Working directory
cd agscheduler

# Install dependencies
make install

# Up CI services
make up-ci-services

# Run check
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

## Thanks

[APScheduler](https://github.com/agronholm/apscheduler/tree/3.x)

[simple-raft](https://github.com/chapin666/simple-raft)

[examples]: https://github.com/agscheduler/agscheduler/tree/main/examples
