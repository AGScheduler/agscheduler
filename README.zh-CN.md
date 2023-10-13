# AGScheduler

[![test](https://github.com/kwkwc/agscheduler/actions/workflows/test.yml/badge.svg)](https://github.com/kwkwc/agscheduler/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/kwkwc/agscheduler/branch/main/graph/badge.svg)](https://codecov.io/gh/kwkwc/agscheduler)
[![Go Report Card](https://goreportcard.com/badge/github.com/kwkwc/agscheduler)](https://goreportcard.com/report/github.com/kwkwc/agscheduler)
[![Go Reference](https://pkg.go.dev/badge/github.com/kwkwc/agscheduler.svg)](https://pkg.go.dev/github.com/kwkwc/agscheduler)
![GitHub tag (with filter)](https://img.shields.io/github/v/tag/kwkwc/agscheduler)
![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/kwkwc/agscheduler)
[![license](https://img.shields.io/github/license/kwkwc/agscheduler)](https://github.com/kwkwc/agscheduler/blob/main/LICENSE)

> Advanced Golang Scheduler (AGScheduler) 是一款适用于 Golang 的任务调度程序，支持多种调度方式，动态更改和持久化任务

[English](README.md) | 简体中文

## 警示
> **_该库处于实验阶段，请勿用于生产环境_**

## 特性

- 支持三种调度方式
    - [x] 一次性执行
    - [x] 间隔执行
    - [x] Cron 式调度
- 支持多种任务存储方式
    - [x] Memory
    - [x] [GROM](https://gorm.io/)(任何 GROM 支持的 RDBMS 都能运行)
    - [x] [Redis](https://redis.io/)
    - [x] [MongoDB](https://www.mongodb.com/)

## 使用

```golang
package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/stores"
)

func printMsg(j agscheduler.Job) {
	slog.Info(fmt.Sprintf("Run %s %s\n", j.Name, j.Args))
}

func main() {
	agscheduler.RegisterFuncs(printMsg)

	store := &stores.MemoryStore{}
	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	job := agscheduler.Job{
		Name:     "Job",
		Type:     agscheduler.TYPE_INTERVAL,
		Timezone: "UTC",
		Func:     printMsg,
		Args:     []any{"arg1", "arg2", "arg3"},
		Interval: 2 * time.Second,
	}
	jobId := scheduler.AddJob(job)
	job, _ = scheduler.GetJob(jobId)
	slog.Info(fmt.Sprintf("Scheduler add %s %s.\n\n", job.Name, job))

	scheduler.Start()
	slog.Info("Scheduler Start.\n\n")

	select {}
}
```

## 注册函数
> **_由于 golang 无法序列化函数，所以 `scheduler.Start()` 之前需要使用 `RegisterFuncs` 注册函数_**

## 示例

[完整示例][examples]

## 致谢

[APScheduler](https://github.com/agronholm/apscheduler/tree/3.x)

[examples]: https://github.com/kwkwc/agscheduler/tree/main/examples
