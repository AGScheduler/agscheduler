// go run examples/stores/memory/main.go

package main

import (
	es "github.com/agscheduler/agscheduler/examples/stores"
	"github.com/agscheduler/agscheduler/stores"
)

func main() {
	store := &stores.MemoryStore{}

	es.RunExample(store)
}
