// go run base.go memory.go

package main

import (
	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/stores"
)

func main() {
	store := &stores.MemoryStore{}

	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	runExample(scheduler)

	select {}
}
