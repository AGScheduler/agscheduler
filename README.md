# AGScheduler

[![test](https://github.com/kwkwc/agscheduler/actions/workflows/test.yml/badge.svg)](https://github.com/kwkwc/agscheduler/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/kwkwc/agscheduler/branch/main/graph/badge.svg)](https://codecov.io/gh/kwkwc/agscheduler)
[![license](https://img.shields.io/github/license/kwkwc/flask-docs)](https://github.com/kwkwc/flask-docs/blob/master/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/kwkwc/agscheduler.svg)](https://pkg.go.dev/github.com/kwkwc/agscheduler)
![GitHub tag (with filter)](https://img.shields.io/github/v/tag/kwkwc/agscheduler)
![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/kwkwc/agscheduler)

> Advanced Golang Scheduler (AGScheduler) is a task scheduler for Golang, that supports multiple scheduling types, dynamic changes and persistent tasks.

English | [简体中文](README.zh-CN.md)

## Warning
> **_This library is experimental and should not be used in a production environment_**

## Features

- Support for three scheduling types
    - [x] One-off execution
    - [x] Interval execution
    - [x] Cron-style scheduling
- Support for multiple task storage methods
    - [x] Memory
    - [x] [GROM](https://gorm.io/)(any RDBMS supported by GROM works)
    - [x] [Redis](https://redis.io/)
    - [x] [MongoDB](https://www.mongodb.com/)

## Usage

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
		Type:     agscheduler.TYPE_INTERVAL,
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

## Register Funcs
> **_Since golang can't serialize functions, you need to register them with `RegisterFuncs` before `scheduler.Start()`_**

## Examples

[Complete example][examples]

## Thanks

[APScheduler](https://github.com/agronholm/apscheduler/tree/3.x)

[examples]: https://github.com/kwkwc/agscheduler/tree/main/examples
