// Package memarch provides data structures to be used.
// It internally wraps memcore and memforge.
package memarch

import (
	"memcore"
	"memforge"
)

// ArrayCreate creates an instance of an array for type T using the provided allocator.
// Ensure the allocator has enough memory capacity.
// Do NOT store this inside custom allocated memory.
func ArrayCreate[T any](allocatorInstance *memforge.FixedLinearAllocator, capacity uint64) *memcore.Array[T] {
	itemSize := memcore.SizeOf[T]()

	addr := memforge.FixedLinearAllocatorMalloc(allocatorInstance, capacity*itemSize, memcore.AlignOf[T]())
	return memcore.ArrayCreateAt[T](addr, capacity)
}
