// Package memarch provides data structures to be used.
// It internally wraps memcore/primitives and memforge.
package memarch

import (
	"memcore/primitives"
	"unsafe"
)

// AllocationFn represents the allocation function to use for the creation of a data structure.
// This can be Malloc or Calloc for any given allocator.
//
// For safety it is best to wrap your actual allocation function so you can catch and handle errors
// where they occur.
type AllocationFn func(sizeBytes, alignment uint64) unsafe.Pointer

// ArrayCreate creates an instance of an array for type T using the provided allocation method.
// Do not store Go pointers inside manually managed memory.
func ArrayCreate[T any](allocFn AllocationFn, capacityElements uint64) *primitives.Array[T] {
	arraySize := primitives.ArrayRequiredBytesGet[T](capacityElements)
	arrayAlignment := primitives.ArrayRequiredAlignmentGet[T]()

	addr := allocFn(arraySize, arrayAlignment)
	return primitives.ArrayCreateAt[T](addr, capacityElements)
}

// StackCreate creates an instance of a stack for type T using the provided allocation method.
// Do not store Go pointers inside manually managed memory.
func StackCreate[T any](allocFn AllocationFn, capacityElements uint64) *primitives.Stack[T] {
	stackSize := primitives.StackRequiredBytesGet[T](capacityElements)
	stackAlignment := primitives.StackRequiredAlignmentGet[T]()

	addr := allocFn(stackSize, stackAlignment)
	return primitives.StackCreateAt[T](addr, capacityElements)
}

// FixedOrderedListCreate creates an instance of a fixed ordered for type T using the provided allocation method.
// Do not store Go pointers inside manually managed memory.
func FixedOrderedListCreate[T any](allocFn AllocationFn, capacityElements uint64) *primitives.FixedOrderedList[T] {
	requiredSize := primitives.FixedOrderedListRequiredBytes[T](capacityElements)
	requiredAlignment := primitives.FixedOrderedListRequiredAlignment[T]()

	addr := allocFn(requiredSize, requiredAlignment)
	return primitives.FixedOrderedListCreateAt[T](addr, capacityElements)
}
