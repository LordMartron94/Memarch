/*
Package memarch provides factory functions that allocate and initialize memstruct containers in manual memory.

[Context]
Each factory takes an AllocationFn (typically a memforge Malloc or Calloc wrapper) and returns both
a memcore.MarkRaw and a typed pointer view. Containers live outside the Go heap until the allocator
is destroyed or reset.
*/
package memarch

import (
	"foundation"
	"memcore"
	"memstruct"
)

/*
AllocationFn allocates a manual region of sizeBytes with alignment and returns its mark.

[Context]
Pass memforge allocator wrappers (Malloc/Calloc). Wrap the underlying allocator to translate panics
into errors at application boundaries when desired.

[Invariants]
"Do not store Go pointers" on factory functions means pointers to Go heap-managed (GC-scanned) memory.
Manual regions may hold pointer values and addresses that refer only to other manual regions (for
example native string headers or C string *byte values).
*/
type AllocationFn func(sizeBytes, alignment uint64) memcore.MarkRaw

/*
MemArchArrayCreate allocates and initializes a memstruct.Array[T] with the given capacity.

[Parameters]
allocFn - Must not be nil.
capacityElements - Reserved element slots.

[Returns]
Mark and *memstruct.Array[T] view.

[Complexity]
Time: O(1). Space: O(capacityElements * element size) plus header overhead.

[Invariants]
Do not store Go heap pointers in manually managed memory.
*/
func MemArchArrayCreate[T any](allocFn AllocationFn, capacityElements uint64) (memcore.MarkRaw, *memstruct.Array[T]) {
	arraySize := memstruct.ArrayRequiredBytesGet[T](capacityElements)
	arrayAlignment := memstruct.ArrayRequiredAlignmentGet[T]()

	addr := allocFn(arraySize, arrayAlignment)
	memstruct.ArrayInitializeAt[T](addr, capacityElements)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.Array[T]](addr)
}

/*
MemArchArrayCreateFrom allocates an array and copies elements from srcArray.

[Parameters]
capacityElements - Must be >= the length of srcArray.

[Side Effects]
Initializes the new array from srcArray via memstruct.ArrayInitializeFrom.
*/
func MemArchArrayCreateFrom[T any](allocFn AllocationFn, capacityElements uint64, srcArray memcore.MarkRaw) (memcore.MarkRaw, *memstruct.Array[T]) {
	arraySize := memstruct.ArrayRequiredBytesGet[T](capacityElements)
	arrayAlignment := memstruct.ArrayRequiredAlignmentGet[T]()

	addr := allocFn(arraySize, arrayAlignment)
	memstruct.ArrayInitializeFrom[T](addr, srcArray, capacityElements)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.Array[T]](addr)
}

/*
MemArchArrayCreateWithSeparatedData allocates only the array header; data lives at dataAddr.

[Context]
Use when the element storage is mmap-backed, shared memory, or pre-allocated separately from the header.

[Parameters]
headerAllocFn - Allocates the header in RAM (or another allocator).
dataAddr - Mark to the external data region with capacity for capacityElements items.

[Returns]
Mark and *memstruct.Array[T] with header bound to dataAddr.

[Complexity]
Time: O(1). Space: O(header only); data is not allocated here.

[Edge Cases]
dataAddr may lie before or after the header; the data span must hold capacityElements * sizeof(T).

[Invariants]
Caller must keep dataAddr valid and aligned. Do not store Go heap pointers in manual memory.
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
MemArchArrayCreateHeaderOnly allocates an uninitialized array header with no data region.

[Context]
Bind external data later (for example via BinaryStoreFixedBindAt). Useful for mmap files and
cursor-driven header reuse.

[Parameters]
capacityElements - Declared capacity once data is bound.

[Returns]
Mark and *memstruct.Array[T]; header is zeroed until bound.

[Complexity]
Time: O(1). Space: O(header only).

[Invariants]
Must bind data before use. Do not store Go heap pointers in manual memory.
*/
func MemArchArrayCreateHeaderOnly[T any](
	headerAllocFn AllocationFn,
	capacityElements uint64,
) (memcore.MarkRaw, *memstruct.Array[T]) {
	headerSize := memstruct.ArrayHeaderRequiredBytesGet[T]()
	headerAlignment := memstruct.ArrayHeaderRequiredAlignmentGet[T]()

	headerAddr := headerAllocFn(headerSize, headerAlignment)
	return headerAddr, memcore.MemcoreMarkDereferenceObject[memstruct.Array[T]](headerAddr)
}

/*
MemArchStackCreate allocates and initializes a memstruct.Stack[T].
*/
func MemArchStackCreate[T any](allocFn AllocationFn, capacityElements uint64) (memcore.MarkRaw, *memstruct.Stack[T]) {
	stackSize := memstruct.StackRequiredBytesGet[T](capacityElements)
	stackAlignment := memstruct.StackRequiredAlignmentGet[T]()

	addr := allocFn(stackSize, stackAlignment)
	memstruct.StackInitializeAt[T](addr, capacityElements)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.Stack[T]](addr)
}

/*
MemArchQueueCreate allocates and initializes a memstruct.Queue[T].
*/
func MemArchQueueCreate[T any](allocFn AllocationFn, capacityElements uint64) (memcore.MarkRaw, *memstruct.Queue[T]) {
	queueSize := memstruct.QueueRequiredBytesGet[T](capacityElements)
	queueAlignment := memstruct.QueueRequiredAlignmentGet[T]()

	addr := allocFn(queueSize, queueAlignment)
	memstruct.QueueInitializeAt[T](addr, capacityElements)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.Queue[T]](addr)
}

/*
MemArchPriorityQueueCreate allocates and initializes a memstruct.PriorityQueue[T].
*/
func MemArchPriorityQueueCreate[T any](allocFn AllocationFn, capacityElements uint64) (memcore.MarkRaw, *memstruct.PriorityQueue[T]) {
	queueSize := memstruct.PriorityQueueRequiredBytesGet[T](capacityElements)
	queueAlignment := memstruct.PriorityQueueRequiredAlignmentGet[T]()

	addr := allocFn(queueSize, queueAlignment)
	memstruct.PriorityQueueInitializeAt[T](addr, capacityElements)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.PriorityQueue[T]](addr)
}

/*
MemArchPriorityQueueCreateFrom allocates a priority queue and copies elements from srcQueue.

[Parameters]
capacityElements - Must be >= the length of srcQueue.
*/
func MemArchPriorityQueueCreateFrom[T any](allocFn AllocationFn, capacityElements uint64, srcQueue memcore.MarkRaw) (memcore.MarkRaw, *memstruct.PriorityQueue[T]) {
	queueSize := memstruct.PriorityQueueRequiredBytesGet[T](capacityElements)
	queueAlignment := memstruct.PriorityQueueRequiredAlignmentGet[T]()

	addr := allocFn(queueSize, queueAlignment)
	memstruct.PriorityQueueInitializeFrom[T](addr, srcQueue, capacityElements)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.PriorityQueue[T]](addr)
}

/*
MemArchFixedOrderedListCreate allocates and initializes a memstruct.FixedOrderedList[T].
*/
func MemArchFixedOrderedListCreate[T any](allocFn AllocationFn, capacityElements uint64) (memcore.MarkRaw, *memstruct.FixedOrderedList[T]) {
	requiredSize := memstruct.FixedOrderedListRequiredBytes[T](capacityElements)
	requiredAlignment := memstruct.FixedOrderedListRequiredAlignment[T]()

	addr := allocFn(requiredSize, requiredAlignment)
	memstruct.FixedOrderedListInitializeAt[T](addr, capacityElements)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.FixedOrderedList[T]](addr)
}

/*
MemArchVectorCreate allocates and initializes a memstruct.Vector[T] for numeric T.
*/
func MemArchVectorCreate[T foundation.Numeric](allocFn AllocationFn, capacityElements uint64) (memcore.MarkRaw, *memstruct.Vector[T]) {
	vectorSize := memstruct.VectorRequiredBytesGet[T](capacityElements)
	vectorAlignment := memstruct.VectorRequiredAlignmentGet[T]()

	addr := allocFn(vectorSize, vectorAlignment)
	memstruct.VectorInitializeAt[T](addr, capacityElements)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.Vector[T]](addr)
}

/*
MemArchVectorCreateFrom allocates a vector and copies elements from src.

[Parameters]
capacityElements - Must be >= the length of src.
*/
func MemArchVectorCreateFrom[T foundation.Numeric](allocFn AllocationFn, src memcore.MarkRaw, capacityElements uint64) (memcore.MarkRaw, *memstruct.Vector[T]) {
	vectorSize := memstruct.VectorRequiredBytesGet[T](capacityElements)
	vectorAlignment := memstruct.VectorRequiredAlignmentGet[T]()

	addr := allocFn(vectorSize, vectorAlignment)
	memstruct.VectorInitializeFrom[T](addr, src, capacityElements)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.Vector[T]](addr)
}

/*
MemArchVectorCreateWithSeparatedData allocates a vector header bound to external numeric data at dataAddr.

[Context]
Same layout strategy as MemArchArrayCreateWithSeparatedData; T must satisfy foundation.Numeric.

[Complexity]
Time: O(1). Space: O(header only).
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
MemArchVectorCreateHeaderOnly allocates an uninitialized vector header for later data binding.

[Context]
Wraps MemArchArrayCreateHeaderOnly for numeric T. Bind external data before use.
*/
func MemArchVectorCreateHeaderOnly[T foundation.Numeric](
	headerAllocFn AllocationFn,
	capacityElements uint64,
) (memcore.MarkRaw, *memstruct.Vector[T]) {
	headerAddr, _ := MemArchArrayCreateHeaderOnly[T](headerAllocFn, capacityElements)
	return headerAddr, memcore.MemcoreMarkDereferenceObject[memstruct.Vector[T]](headerAddr)
}

/*
MemArchMatrixCreate allocates and initializes a memstruct.Matrix[T] with the given dimensions.
*/
func MemArchMatrixCreate[T foundation.Numeric](allocFn AllocationFn, capacityRows, capacityCols uint64) (memcore.MarkRaw, *memstruct.Matrix[T]) {
	matrixSize := memstruct.MatrixRequiredBytesGet[T](capacityRows, capacityCols)
	matrixAlignment := memstruct.MatrixRequiredAlignmentGet[T]()

	addr := allocFn(matrixSize, matrixAlignment)
	memstruct.MatrixInitializeAt[T](addr, capacityRows, capacityCols)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.Matrix[T]](addr)
}

/*
MemArchMatrixCreateFrom allocates a matrix and copies elements from src.

[Parameters]
capacityRows, capacityCols - Must be >= the corresponding dimensions of src.
*/
func MemArchMatrixCreateFrom[T foundation.Numeric](allocFn AllocationFn, src memcore.MarkRaw, capacityRows, capacityCols uint64) (memcore.MarkRaw, *memstruct.Matrix[T]) {
	matrixSize := memstruct.MatrixRequiredBytesGet[T](capacityRows, capacityCols)
	matrixAlignment := memstruct.MatrixRequiredAlignmentGet[T]()

	addr := allocFn(matrixSize, matrixAlignment)
	memstruct.MatrixInitializeFrom[T](addr, src, capacityRows, capacityCols)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.Matrix[T]](addr)
}

/*
MemArchStringCreate allocates a memstruct.String containing content.
*/
func MemArchStringCreate(allocFn AllocationFn, content string) (memcore.MarkRaw, *memstruct.String) {
	stringSize := memstruct.StringRequiredBytesGet(content)
	stringAlignment := memstruct.StringRequiredAlignmentGet()

	addr := allocFn(stringSize, stringAlignment)
	memstruct.StringInitializeAt(addr, content)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.String](addr)
}

/*
MemArchGoStringCreate allocates a native Go string header and data in manual memory.

[Context]
Use when APIs require a Go string value rather than memstruct.String. If the backing region moves,
update the string data pointer accordingly.

[Parameters]
content - UTF-8 bytes copied into the manual allocation.

[Returns]
Mark and *string view backed by manual memory.

[Complexity]
Time: O(len(content)). Space: O(len(content)).

[Invariants]
The string's data pointer refers to manual memory, not the Go heap.
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

[Context]
Suitable for C-ABI and FFI (*byte / char*). For a native Go string in manual memory, use MemArchGoStringCreate.

[Returns]
Mark and *byte pointing at the first byte of the C string.

[Complexity]
Time: O(len(content)). Space: O(len(content) + 1) for the terminator.
*/
func MemArchCStringCreate(allocFn AllocationFn, content string) (memcore.MarkRaw, *byte) {
	stringSize := memstruct.CStringRequiredBytesGet(content)
	stringAlignment := memstruct.CStringRequiredAlignmentGet()

	addr := allocFn(stringSize, stringAlignment)
	memstruct.CStringInitializeAt(addr, content)
	return addr, memstruct.CStringPointerGet(addr)
}

/*
MemArchHashMapCreate allocates and initializes a memstruct.HashMap[TKey, TValue].

[Parameters]
keyComparisonFunc, keyMarkFunc - Required memstruct key callbacks for lookup and storage.
capacityElements - Bucket/table capacity passed to memstruct.HashMapInitializeAt.
*/
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

/*
MemArchCircularBufferCreate allocates and initializes a memstruct.CircularBuffer[T].
*/
func MemArchCircularBufferCreate[T any](allocFn AllocationFn, capacityElements uint64) (memcore.MarkRaw, *memstruct.CircularBuffer[T]) {
	bufferSize := memstruct.CircularBufferRequiredBytesGet[T](capacityElements)
	bufferAlignment := memstruct.CircularBufferRequiredAlignmentGet[T]()

	addr := allocFn(bufferSize, bufferAlignment)
	memstruct.CircularBufferInitializeAt[T](addr, capacityElements)
	return addr, memcore.MemcoreMarkDereferenceObject[memstruct.CircularBuffer[T]](addr)
}
