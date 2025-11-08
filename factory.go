// Package memarch provides data structures to be used.
// It internally wraps memstruct and memforge.
package memarch

import (
	"foundation"
	"memcore"
	"memstruct"
)

// AllocationFn represents the allocation function to use for the creation of a data structure.
// This can be Malloc or Calloc for any given allocator.
//
// For safety it is best to wrap your actual allocation function so you can catch and handle errors
// where they occur.
type AllocationFn func(sizeBytes, alignment uint64) memcore.MarkRaw

// MemArchArrayCreate creates an instance of an array for type T using the provided allocation method.
// Do not store Go pointers inside manually managed memory.
func MemArchArrayCreate[T any](allocFn AllocationFn, capacityElements uint64) (memcore.MarkRaw, *memstruct.Array[T]) {
	arraySize := memstruct.ArrayRequiredBytesGet[T](capacityElements)
	arrayAlignment := memstruct.ArrayRequiredAlignmentGet[T]()

	addr := allocFn(arraySize, arrayAlignment)
	memstruct.ArrayInitializeAt[T](addr, capacityElements)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.Array[T]](addr)
}

// MemArchStackCreate creates an instance of a stack for type T using the provided allocation method.
// Do not store Go pointers inside manually managed memory.
func MemArchStackCreate[T any](allocFn AllocationFn, capacityElements uint64) (memcore.MarkRaw, *memstruct.Stack[T]) {
	stackSize := memstruct.StackRequiredBytesGet[T](capacityElements)
	stackAlignment := memstruct.StackRequiredAlignmentGet[T]()

	addr := allocFn(stackSize, stackAlignment)
	memstruct.StackInitializeAt[T](addr, capacityElements)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.Stack[T]](addr)
}

// MemArchFixedOrderedListCreate creates an instance of a fixed ordered for type T using the provided allocation method.
// Do not store Go pointers inside manually managed memory.
func MemArchFixedOrderedListCreate[T any](allocFn AllocationFn, capacityElements uint64) (memcore.MarkRaw, *memstruct.FixedOrderedList[T]) {
	requiredSize := memstruct.FixedOrderedListRequiredBytes[T](capacityElements)
	requiredAlignment := memstruct.FixedOrderedListRequiredAlignment[T]()

	addr := allocFn(requiredSize, requiredAlignment)
	memstruct.FixedOrderedListInitializeAt[T](addr, capacityElements)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.FixedOrderedList[T]](addr)
}

// MemArchVectorCreate creates an instance of a vector for type T using the provided allocation method.
// Do not store Go pointers inside manually managed memory.
func MemArchVectorCreate[T foundation.Numeric](allocFn AllocationFn, capacityElements uint64) (memcore.MarkRaw, *memstruct.Vector[T]) {
	vectorSize := memstruct.VectorRequiredBytesGet[T](capacityElements)
	vectirAlignment := memstruct.VectorRequiredAlignmentGet[T]()

	addr := allocFn(vectorSize, vectirAlignment)
	memstruct.ArrayInitializeAt[T](addr, capacityElements)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.Vector[T]](addr)
}
