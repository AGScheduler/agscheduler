# AGScheduler

[![license](https://img.shields.io/github/license/kwkwc/flask-docs)](https://github.com/kwkwc/flask-docs/blob/master/LICENSE)

> Advanced Golang Scheduler (AGScheduler) 是一款适用于 Golang 的任务调度程序，支持多种调度方式和持久化任务

[English](README.md) | 简体中文

## 警示
> **_该库处于实验阶段，请勿用于生产环境_**

## 特性

- 支持三种调度方式
    - [x] 一次性执行
    - [x] 间隔执行
    - [x] Cron 式调度
- 支持存储工作
    - [x] Memory
    - [x] [GROM](https://gorm.io/)(MySQL | SQLite)
    - [ ] [Redis](https://redis.io/)
    - [ ] [MongoDB](https://www.mongodb.com/)

## 使用

```golang
package main

import (
	"log"
	"time"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/stores"
)

func printMsg(j agscheduler.Job) {
	log.Printf("Run %s %s\n", j.Name, j.Args)
}

func main() {
	agscheduler.RegisterFuncs(printMsg)

	store := &stores.MemoryStore{}
	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	job := agscheduler.Job{
		Name:     "Job",
		Type:     "interval",
		Timezone: "UTC",
		Func:     printMsg,
		Args:     []any{"arg1", "arg2", "arg3"},
		Interval: 2 * time.Second,
	}
	jobId := scheduler.AddJob(job)
	job, _ = scheduler.GetJob(jobId)
	log.Printf("Scheduler add %s %s.\n\n", job.Name, job)

	scheduler.Start()
	log.Print("Scheduler Start.\n\n")

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
