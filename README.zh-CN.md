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

> **_该库处于实验阶段，不建议用于生产环境_**

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
	slog.Info(fmt.Sprintf("Run job `%s` %s\n", j.FullName(), j.Args))
}

func main() {
	agscheduler.RegisterFuncs(printMsg)

	store := &stores.MemoryStore{}
	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	job1 := agscheduler.Job{
		Name:     "Job1",
		Type:     agscheduler.TYPE_INTERVAL,
		Interval: "2s",
		Timezone: "UTC",
		Func:     printMsg,
		Args:     map[string]any{"arg1": "1", "arg2": "2", "arg3": "3"},
	}
	job1, _ = scheduler.AddJob(job1)
	slog.Info(fmt.Sprintf("Scheduler add job `%s` %s.\n\n", job1.FullName(), job1))

	job2 := agscheduler.Job{
		Name:     "Job2",
		Type:     agscheduler.TYPE_CRON,
		CronExpr: "*/1 * * * *",
		Timezone: "Asia/Shanghai",
		FuncName: "main.printMsg",
		Args:     map[string]any{"arg4": "4", "arg5": "5", "arg6": "6", "arg7": "7"},
	}
	job2, _ = s.AddJob(job2)
	slog.Info(fmt.Sprintf("Scheduler add job `%s` %s.\n\n", job2.FullName(), job2))

	job3 := agscheduler.Job{
		Name:     "Job3",
		Type:     agscheduler.TYPE_DATETIME,
		StartAt:  "2023-09-22 07:30:08",
		Timezone: "America/New_York",
		Func:     printMsg,
		Args:     map[string]any{"arg8": "8", "arg9": "9"},
	}
	job3, _ = s.AddJob(job3)
	slog.Info(fmt.Sprintf("Scheduler add job `%s` %s.\n\n", job3.FullName(), job3))

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
