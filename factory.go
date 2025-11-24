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

// MemArchArrayCreateFrom creates an instance of an array for type T using the provided allocation method.
// This variant stores the items of Array A into Array B upon initialization.
// The capacity must be >= length of Array A
// Do not store Go pointers inside manually managed memory.
func MemArchArrayCreateFrom[T any](allocFn AllocationFn, capacityElements uint64, srcArray memcore.MarkRaw) (memcore.MarkRaw, *memstruct.Array[T]) {
	arraySize := memstruct.ArrayRequiredBytesGet[T](capacityElements)
	arrayAlignment := memstruct.ArrayRequiredAlignmentGet[T]()

	addr := allocFn(arraySize, arrayAlignment)
	memstruct.ArrayInitializeFrom[T](addr, srcArray, capacityElements)
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

// MemArchQueueCreate creates an instance of a queue for type T using the provided allocation method.
// Do not store Go pointers inside manually managed memory.
func MemArchQueueCreate[T any](allocFn AllocationFn, capacityElements uint64) (memcore.MarkRaw, *memstruct.Queue[T]) {
	queueSize := memstruct.QueueRequiredBytesGet[T](capacityElements)
	queueAlignment := memstruct.QueueRequiredAlignmentGet[T]()

	addr := allocFn(queueSize, queueAlignment)
	memstruct.QueueInitializeAt[T](addr, capacityElements)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.Queue[T]](addr)
}

// MemArchPriorityQueueCreate creates an instance of a priority queue for type T using the provided allocation method.
// Do not store Go pointers inside manually managed memory.
func MemArchPriorityQueueCreate[T any](allocFn AllocationFn, capacityElements uint64) (memcore.MarkRaw, *memstruct.PriorityQueue[T]) {
	queueSize := memstruct.PriorityQueueRequiredBytesGet[T](capacityElements)
	queueAlignment := memstruct.PriorityQueueRequiredAlignmentGet[T]()

	addr := allocFn(queueSize, queueAlignment)
	memstruct.PriorityQueueInitializeAt[T](addr, capacityElements)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.PriorityQueue[T]](addr)
}

// MemArchPriorityQueueCreateFrom creates an instance of a priority queue for type T using the provided allocation method.
// This variant stores the items of queue A into queue B upon initialization.
// The capacity must be >= length of queue A
// Do not store Go pointers inside manually managed memory.
func MemArchPriorityQueueCreateFrom[T any](allocFn AllocationFn, capacityElements uint64, srcQueue memcore.MarkRaw) (memcore.MarkRaw, *memstruct.PriorityQueue[T]) {
	queueSize := memstruct.PriorityQueueRequiredBytesGet[T](capacityElements)
	queueAlignment := memstruct.PriorityQueueRequiredAlignmentGet[T]()

	addr := allocFn(queueSize, queueAlignment)
	memstruct.PriorityQueueInitializeFrom[T](addr, srcQueue, capacityElements)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.PriorityQueue[T]](addr)
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
	vectorAlignment := memstruct.VectorRequiredAlignmentGet[T]()

	addr := allocFn(vectorSize, vectorAlignment)
	memstruct.VectorInitializeAt[T](addr, capacityElements)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.Vector[T]](addr)
}

// MemArchVectorCreateFrom creates an instance of a vector for type T using the provided allocation method.
// This variant stores the items of Vector A into Vector B upon initialization.
// The capacity must be >= length of Vector A
// Do not store Go pointers inside manually managed memory.
func MemArchVectorCreateFrom[T foundation.Numeric](allocFn AllocationFn, src memcore.MarkRaw, capacityElements uint64) (memcore.MarkRaw, *memstruct.Vector[T]) {
	vectorSize := memstruct.VectorRequiredBytesGet[T](capacityElements)
	vectorAlignment := memstruct.VectorRequiredAlignmentGet[T]()

	addr := allocFn(vectorSize, vectorAlignment)
	memstruct.VectorInitializeFrom[T](addr, src, capacityElements)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.Vector[T]](addr)
}

// MemArchStringCreate creates an instance of a string using the provided allocation method.
// Do not store Go pointers inside manually managed memory.
func MemArchStringCreate(allocFn AllocationFn, content string) (memcore.MarkRaw, *memstruct.String) {
	stringSize := memstruct.StringRequiredBytesGet(content)
	stringAlignment := memstruct.StringRequiredAlignmentGet()

	addr := allocFn(stringSize, stringAlignment)
	memstruct.StringInitializeAt(addr, content)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.String](addr)
}

// MemArchHashMapCreate creates an instance of a hashmap using the provided allocation method.
// Do not store Go pointers inside manually managed memory.
func MemArchHashMapCreate[TKey, TValue any](
	allocFn AllocationFn,
	capacityElements uint64,
	keyComparisonFunc memstruct.KeyComparer[TKey],
	keyMarkFunc memstruct.KeyMarkRetriever[TKey],
) (memcore.MarkRaw, *memstruct.HashMap[TKey, TValue]) {
	hashMapSize := memstruct.HashMapRequiredBytesGet[TKey, TValue](capacityElements)
	hashMapAlignment := memstruct.HashMapRequiredAlignmentGet[TKey, TValue]()

	addr := allocFn(hashMapSize, hashMapAlignment)
	memstruct.HashMapInitializeAt[TKey, TValue](addr, capacityElements, keyComparisonFunc, keyMarkFunc)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.HashMap[TKey, TValue]](addr)
}
