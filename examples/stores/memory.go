// go run examples/stores/base.go examples/stores/memory.go

package main

import "github.com/agscheduler/agscheduler/stores"

func main() {
	store := &stores.MemoryStore{}

	runExample(store)
}
