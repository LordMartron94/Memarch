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
//
// "No Go pointers" warnings on factory functions mean pointers to Go heap-managed memory (GC-scanned
// allocations). Manual memory may store pointer values and addresses that refer only to manual
// regions (for example native Go string headers or C string *byte values).
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

/*
MemArchArrayCreateWithSeparatedData creates an array instance with the header allocated from RAM
and the data region provided at a separate memory address.

This function is designed for use cases where the data region comes from a different memory source
than the header, such as memory-mapped files, shared memory, or pre-allocated data buffers. The
header is allocated using the provided allocation function (typically from RAM), while the data
address is provided directly by the caller.

Use cases:
- Memory-mapped files where data is mapped from disk and header is in RAM
- Shared memory scenarios where data is in a shared region
- Pre-allocated data buffers that need array metadata management
- Memory pools with separate header and data allocation strategies

Time complexity: O(1) - constant time allocation and initialization
Space complexity: O(1) - only header is allocated (data is provided externally)

Prerequisites:
- headerAllocFn must be a valid allocation function
- dataAddr must point to a valid, properly aligned memory address for the data region
- The data region must have sufficient capacity for the specified number of elements
- dataAddr must be within memory managed by memcore

Edge cases:
- dataAddr can be located before or after the header in memory (offset can be negative or positive)
- The header and data can be in completely separate memory regions
- The data region size must match capacityElements * sizeof(T)

Additional notes:
- Only the header is allocated using headerAllocFn; data is not allocated
- The caller is responsible for ensuring dataAddr is valid and properly aligned
- This allows for flexible memory layouts where header and data are managed separately
- Do not store Go pointers inside manually managed memory
*/
func MemArchArrayCreateWithSeparatedData[T any](
	headerAllocFn AllocationFn,
	dataAddr memcore.MarkRaw,
	capacityElements uint64,
) (memcore.MarkRaw, *memstruct.Array[T]) {
	headerSize := memstruct.ArrayHeaderRequiredBytesGet[T]()
	headerAlignment := memstruct.ArrayHeaderRequiredAlignmentGet[T]()

	headerAddr := headerAllocFn(headerSize, headerAlignment)
	memstruct.ArrayInitializeWithSeparatedHeaderAndData[T](headerAddr, dataAddr, capacityElements)
	return headerAddr, memcore.MemcoreMarkDereferenceObject[memstruct.Array[T]](headerAddr)
}

/*
MemArchArrayCreateHeaderOnly creates an array instance with only the header allocated, no data region.

This function allocates only the header structure for an array, without allocating the data region.
The header is left uninitialized and must be bound to data using BinaryStoreFixedBindAt or similar
before use. This is useful when the data is stored in memory-mapped files, shared memory, or other
external storage where the data region is not managed by the allocator.

Use cases:
- Memory-mapped files where data is mapped from disk and header is in RAM
- Cursor-based iteration where headers are reused and rebound to different data locations
- Memory-efficient scenarios where data is stored separately from headers
- Avoiding unnecessary data region allocation when data is external

Time complexity: O(1) - constant time allocation only
Space complexity: O(1) - only header is allocated (data is provided externally)

Prerequisites:
- headerAllocFn must be a valid allocation function
- Header must be bound to data using BinaryStoreFixedBindAt or similar before use
- The data region must exist externally (memory-mapped file, shared memory, etc.)

Edge cases:
- Header is uninitialized (zero values) and must be bound before use
- No data region is allocated - data must be provided externally
- Capacity is specified but data region is not allocated

Additional notes:
- Only the header is allocated using headerAllocFn; no data region is allocated
- The header must be bound to external data before it can be used
- More memory-efficient than MemArchArrayCreate when data is stored separately
- Designed for use with BinaryStoreFixedBindAt and cursor-based iteration
- Do not store Go pointers inside manually managed memory
*/
func MemArchArrayCreateHeaderOnly[T any](
	headerAllocFn AllocationFn,
	capacityElements uint64,
) (memcore.MarkRaw, *memstruct.Array[T]) {
	headerSize := memstruct.ArrayHeaderRequiredBytesGet[T]()
	headerAlignment := memstruct.ArrayHeaderRequiredAlignmentGet[T]()

	headerAddr := headerAllocFn(headerSize, headerAlignment)
	// Header is left uninitialized - it will be properly initialized when bound
	// via BinaryStoreFixedBindAt or similar functions
	return headerAddr, memcore.MemcoreMarkDereferenceObject[memstruct.Array[T]](headerAddr)
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

/*
MemArchVectorCreateWithSeparatedData creates a vector instance with the header allocated from RAM
and the data region provided at a separate memory address.

This function is designed for use cases where the data region comes from a different memory source
than the header, such as memory-mapped files, shared memory, or pre-allocated data buffers. The
header is allocated using the provided allocation function (typically from RAM), while the data
address is provided directly by the caller.

Use cases:
- Memory-mapped files where data is mapped from disk and header is in RAM
- Shared memory scenarios where data is in a shared region
- Pre-allocated data buffers that need vector metadata management
- Memory pools with separate header and data allocation strategies
- Numerical computing scenarios with large datasets in memory-mapped files

Time complexity: O(1) - constant time allocation and initialization
Space complexity: O(1) - only header is allocated (data is provided externally)

Prerequisites:
- headerAllocFn must be a valid allocation function
- dataAddr must point to a valid, properly aligned memory address for the data region
- The data region must have sufficient capacity for the specified number of elements
- dataAddr must be within memory managed by memcore
- Type T must be a numeric type (foundation.Numeric)

Edge cases:
- dataAddr can be located before or after the header in memory (offset can be negative or positive)
- The header and data can be in completely separate memory regions
- The data region size must match capacityElements * sizeof(T)

Additional notes:
- Only the header is allocated using headerAllocFn; data is not allocated
- The caller is responsible for ensuring dataAddr is valid and properly aligned
- This allows for flexible memory layouts where header and data are managed separately
- Do not store Go pointers inside manually managed memory
- Internally wraps MemArchArrayCreateWithSeparatedData for numeric types
*/
func MemArchVectorCreateWithSeparatedData[T foundation.Numeric](
	headerAllocFn AllocationFn,
	dataAddr memcore.MarkRaw,
	capacityElements uint64,
) (memcore.MarkRaw, *memstruct.Vector[T]) {
	headerSize := memstruct.VectorHeaderRequiredBytesGet[T]()
	headerAlignment := memstruct.VectorHeaderRequiredAlignmentGet[T]()

	headerAddr := headerAllocFn(headerSize, headerAlignment)
	memstruct.VectorInitializeWithSeparatedHeaderAndData[T](headerAddr, dataAddr, capacityElements)
	return headerAddr, memcore.MemcoreMarkDereferenceObject[memstruct.Vector[T]](headerAddr)
}

/*
MemArchVectorCreateHeaderOnly creates a vector instance with only the header allocated, no data region.

This function allocates only the header structure for a vector, without allocating the data region.
The header is left uninitialized and must be bound to data using BinaryStoreFixedBindAt or similar
before use. This is useful when the data is stored in memory-mapped files, shared memory, or other
external storage where the data region is not managed by the allocator.

Use cases:
- Memory-mapped files where data is mapped from disk and header is in RAM
- Cursor-based iteration where headers are reused and rebound to different data locations
- Memory-efficient scenarios where data is stored separately from headers
- Avoiding unnecessary data region allocation when data is external
- Numerical computing scenarios with large datasets in memory-mapped files

Time complexity: O(1) - constant time allocation only
Space complexity: O(1) - only header is allocated (data is provided externally)

Prerequisites:
- headerAllocFn must be a valid allocation function
- Header must be bound to data using BinaryStoreFixedBindAt or similar before use
- The data region must exist externally (memory-mapped file, shared memory, etc.)
- Type T must be a numeric type (foundation.Numeric)

Edge cases:
- Header is uninitialized (zero values) and must be bound before use
- No data region is allocated - data must be provided externally
- Capacity is specified but data region is not allocated

Additional notes:
- Only the header is allocated using headerAllocFn; no data region is allocated
- The header must be bound to external data before it can be used
- More memory-efficient than MemArchVectorCreate when data is stored separately
- Designed for use with BinaryStoreFixedBindAt and cursor-based iteration
- Internally wraps MemArchArrayCreateHeaderOnly for numeric types
- Do not store Go pointers inside manually managed memory
*/
func MemArchVectorCreateHeaderOnly[T foundation.Numeric](
	headerAllocFn AllocationFn,
	capacityElements uint64,
) (memcore.MarkRaw, *memstruct.Vector[T]) {
	headerAddr, _ := MemArchArrayCreateHeaderOnly[T](headerAllocFn, capacityElements)
	return headerAddr, memcore.MemcoreMarkDereferenceObject[memstruct.Vector[T]](headerAddr)
}

// MemArchMatrixCreate creates an instance of a matrix for type T using the provided allocation method.
// Do not store Go pointers inside manually managed memory.
func MemArchMatrixCreate[T foundation.Numeric](allocFn AllocationFn, capacityRows, capacityCols uint64) (memcore.MarkRaw, *memstruct.Matrix[T]) {
	matrixSize := memstruct.MatrixRequiredBytesGet[T](capacityRows, capacityCols)
	matrixAlignment := memstruct.MatrixRequiredAlignmentGet[T]()

	addr := allocFn(matrixSize, matrixAlignment)
	memstruct.MatrixInitializeAt[T](addr, capacityRows, capacityCols)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.Matrix[T]](addr)
}

// MemArchMatrixCreateFrom creates an instance of a matrix for type T using the provided allocation method.
// This variant stores the items of Matrix A into Matrix B upon initialization.
// The capacity must be >= length of Matrix A
// Do not store Go pointers inside manually managed memory.
func MemArchMatrixCreateFrom[T foundation.Numeric](allocFn AllocationFn, src memcore.MarkRaw, capacityRows, capacityCols uint64) (memcore.MarkRaw, *memstruct.Matrix[T]) {
	matrixSize := memstruct.MatrixRequiredBytesGet[T](capacityRows, capacityCols)
	matrixAlignment := memstruct.MatrixRequiredAlignmentGet[T]()

	addr := allocFn(matrixSize, matrixAlignment)
	memstruct.MatrixInitializeFrom[T](addr, src, capacityRows, capacityCols)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.Matrix[T]](addr)
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

/*
MemArchGoStringCreate allocates a native Go string in manual memory.

The returned *string and its internal data pointer refer only to the manual allocation. The data
pointer is a pointer value, not a reference to Go heap-managed memory. Use this when APIs require
a Go string value rather than memstruct.String. If the memory region is relocated, the string data
pointer must be updated.

Time complexity: O(n) where n is len(content)
Space complexity: O(n)
*/
func MemArchGoStringCreate(allocFn AllocationFn, content string) (memcore.MarkRaw, *string) {
	stringSize := memstruct.GoStringRequiredBytesGet(content)
	stringAlignment := memstruct.GoStringRequiredAlignmentGet()

	addr := allocFn(stringSize, stringAlignment)
	memstruct.GoStringInitializeAt(addr, content)
	return addr, memcore.MemcoreMarkDereferenceObject[string](addr)
}

/*
MemArchCStringCreate allocates a NUL-terminated UTF-8 C string in manual memory.

The returned *byte is suitable for C-ABI and FFI (const char*, char*). The allocation contains
only bytes plus a trailing 0x00; no Go string header is stored. For a native Go string in manual
memory, use MemArchGoStringCreate instead.

Time complexity: O(n) where n is len(content)
Space complexity: O(n)
*/
func MemArchCStringCreate(allocFn AllocationFn, content string) (memcore.MarkRaw, *byte) {
	stringSize := memstruct.CStringRequiredBytesGet(content)
	stringAlignment := memstruct.CStringRequiredAlignmentGet()

	addr := allocFn(stringSize, stringAlignment)
	memstruct.CStringInitializeAt(addr, content)
	return addr, memstruct.CStringPointerGet(addr)
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

// MemArchCircularBufferCreate creates an instance of a circular buffer for type T using the provided allocation method.
// Do not store Go pointers inside manually managed memory.
func MemArchCircularBufferCreate[T any](allocFn AllocationFn, capacityElements uint64) (memcore.MarkRaw, *memstruct.CircularBuffer[T]) {
	bufferSize := memstruct.CircularBufferRequiredBytesGet[T](capacityElements)
	bufferAlignment := memstruct.CircularBufferRequiredAlignmentGet[T]()

	addr := allocFn(bufferSize, bufferAlignment)
	memstruct.CircularBufferInitializeAt[T](addr, capacityElements)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.CircularBuffer[T]](addr)
}
