package stores

import "testing"

func TestMemoryStore(t *testing.T) {
	store := &MemoryStore{}

	runTest(t, store)
}
