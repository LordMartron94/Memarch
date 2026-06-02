package memarch

import (
	"fmt"
	"foundation"
	"memcore"
	"memstruct"
)

/*
LayoutBlueprint describes aligned component offsets for a composite manual-memory block.

[Context]
The blueprint header lives on the Go heap; component offsets are stored in a manual memstruct.Array[uint64]
allocated via layoutAllocator (typically a small scratch arena). Reuse the same blueprint for many runtime instances via MemArchLayoutConstructAllocate and
MemArchLayoutConstructBind* (pass the blueprint pointer directly).

[Invariants]
Do not mutate a blueprint concurrently with Push. Do not store Go heap pointers inside composed
manual regions or inside the offset table.
*/
type LayoutBlueprint struct {
	componentOffsets memcore.MarkRaw // Array[uint64]

	totalBytes     uint64
	maxAlignment   uint64
	componentCount uint64
	maxComponents  uint64
}

/*
MemArchLayoutBlueprintCreate reserves an offset table for up to maxComponents Push operations.

[Parameters]
layoutAllocator - Allocates the uint64 offset array (typically a scratch linear allocator).
maxComponents - Maximum number of Push calls before panic.

[Returns]
A heap-allocated blueprint; the offset table lives in manual memory.

[Errors]
Panics when maxComponents is zero.
*/
func MemArchLayoutBlueprintCreate(
	layoutAllocator AllocationFn,
	maxComponents uint64,
) *LayoutBlueprint {
	if maxComponents == 0 {
		panic("memarch: layout blueprint requires maxComponents > 0")
	}

	componentOffsets, _ := MemArchArrayCreate[uint64](layoutAllocator, maxComponents)

	return &LayoutBlueprint{
		componentOffsets: componentOffsets,
		maxComponents:    maxComponents,
	}
}

/*
MemArchLayoutBlueprintPush records a component size and alignment in the blueprint.

[Parameters]
bytes - Payload size in bytes.
alignment - Required alignment; must be a power of two greater than zero.

[Returns]
Component index for use with construct/bind helpers.

[Errors]
Panics when Push exceeds maxComponents from Create.
*/
func MemArchLayoutBlueprintPush(blueprint *LayoutBlueprint, bytes, alignment uint64) uint64 {
	if blueprint.componentCount >= blueprint.maxComponents {
		panic(fmt.Errorf("memarch: layout blueprint push beyond max components (%d)", blueprint.maxComponents))
	}

	alignedCursor := memcore.AlignUp(blueprint.totalBytes, alignment)

	memstruct.ArraySetAtUnsafe(blueprint.componentOffsets, blueprint.componentCount, alignedCursor)

	blueprint.totalBytes = alignedCursor + bytes

	if alignment > blueprint.maxAlignment {
		blueprint.maxAlignment = alignment
	}

	index := blueprint.componentCount
	blueprint.componentCount++
	return index
}

/*
MemArchLayoutBlueprintClone duplicates blueprint geometry into a new offset table.

[Parameters]
layoutAllocator - Allocates the cloned uint64 offset array.

[Returns]
A new blueprint with copied offsets and the same totalBytes, maxAlignment, and componentCount.
*/
func MemArchLayoutBlueprintClone(
	source *LayoutBlueprint,
	layoutAllocator AllocationFn,
) *LayoutBlueprint {
	componentOffsets, _ := MemArchArrayCreate[uint64](layoutAllocator, source.maxComponents)
	memstruct.ArrayCopyFrom[uint64](componentOffsets, source.componentOffsets, 0)

	return &LayoutBlueprint{
		componentOffsets: componentOffsets,
		totalBytes:       source.totalBytes,
		maxAlignment:     source.maxAlignment,
		componentCount:   source.componentCount,
		maxComponents:    source.maxComponents,
	}
}

/*
MemArchLayoutBlueprintRequiredBytesGet returns the total byte span of all pushed components.
*/
func MemArchLayoutBlueprintRequiredBytesGet(blueprint *LayoutBlueprint) uint64 {
	return blueprint.totalBytes
}

/*
MemArchLayoutBlueprintRequiredAlignmentGet returns the strictest alignment among pushed components.
*/
func MemArchLayoutBlueprintRequiredAlignmentGet(blueprint *LayoutBlueprint) uint64 {
	return blueprint.maxAlignment
}

/*
MemArchLayoutBlueprintComponentCountGet returns how many components have been pushed.
*/
func MemArchLayoutBlueprintComponentCountGet(blueprint *LayoutBlueprint) uint64 {
	return blueprint.componentCount
}

/*
MemArchLayoutBlueprintComponentOffsetGet returns the byte offset of a component within the block.
*/
func MemArchLayoutBlueprintComponentOffsetGet(blueprint *LayoutBlueprint, componentIndex uint64) uint64 {
	if componentIndex >= blueprint.componentCount {
		panic(fmt.Errorf("memarch: layout blueprint component index %d out of range (count %d)", componentIndex, blueprint.componentCount))
	}
	return memstruct.ArrayItemGetAtUnsafe[uint64](blueprint.componentOffsets, componentIndex)
}

// ---------------------------------------------------------------------- PUSH (BLUEPRINT HELPERS)

/*
MemArchLayoutBlueprintPushStruct reserves space for a value of type T.
*/
func MemArchLayoutBlueprintPushStruct[T any](blueprint *LayoutBlueprint) uint64 {
	memarchManualTypeValidate[T]("MemArchLayoutBlueprintPushStruct")
	return MemArchLayoutBlueprintPush(blueprint, memcore.SizeOf[T](), memcore.AlignOf[T]())
}

/*
MemArchLayoutBlueprintPushArray reserves space for a memstruct.Array[TElement].
*/
func MemArchLayoutBlueprintPushArray[TElement any](blueprint *LayoutBlueprint, capacityElements uint64) uint64 {
	arraySize := memstruct.ArrayRequiredBytesGet[TElement](capacityElements)
	arrayAlignment := memstruct.ArrayRequiredAlignmentGet[TElement]()
	return MemArchLayoutBlueprintPush(blueprint, arraySize, arrayAlignment)
}

/*
MemArchLayoutBlueprintPushVector reserves space for a memstruct.Vector[TElement].
*/
func MemArchLayoutBlueprintPushVector[TElement foundation.Numeric](blueprint *LayoutBlueprint, capacityElements uint64) uint64 {
	vectorSize := memstruct.VectorRequiredBytesGet[TElement](capacityElements)
	vectorAlignment := memstruct.VectorRequiredAlignmentGet[TElement]()
	return MemArchLayoutBlueprintPush(blueprint, vectorSize, vectorAlignment)
}

/*
MemArchLayoutBlueprintPushStack reserves space for a memstruct.Stack[TElement].
*/
func MemArchLayoutBlueprintPushStack[TElement any](blueprint *LayoutBlueprint, capacityElements uint64) uint64 {
	stackSize := memstruct.StackRequiredBytesGet[TElement](capacityElements)
	stackAlignment := memstruct.StackRequiredAlignmentGet[TElement]()
	return MemArchLayoutBlueprintPush(blueprint, stackSize, stackAlignment)
}

/*
MemArchLayoutBlueprintPushQueue reserves space for a memstruct.Queue[TElement].
*/
func MemArchLayoutBlueprintPushQueue[TElement any](blueprint *LayoutBlueprint, capacityElements uint64) uint64 {
	queueSize := memstruct.QueueRequiredBytesGet[TElement](capacityElements)
	queueAlignment := memstruct.QueueRequiredAlignmentGet[TElement]()
	return MemArchLayoutBlueprintPush(blueprint, queueSize, queueAlignment)
}

/*
MemArchLayoutBlueprintPushPriorityQueue reserves space for a memstruct.PriorityQueue[TElement].
*/
func MemArchLayoutBlueprintPushPriorityQueue[TElement any](blueprint *LayoutBlueprint, capacityElements uint64) uint64 {
	queueSize := memstruct.PriorityQueueRequiredBytesGet[TElement](capacityElements)
	queueAlignment := memstruct.PriorityQueueRequiredAlignmentGet[TElement]()
	return MemArchLayoutBlueprintPush(blueprint, queueSize, queueAlignment)
}

/*
MemArchLayoutBlueprintPushFixedOrderedList reserves space for a memstruct.FixedOrderedList[TElement].
*/
func MemArchLayoutBlueprintPushFixedOrderedList[TElement any](blueprint *LayoutBlueprint, capacityElements uint64) uint64 {
	listSize := memstruct.FixedOrderedListRequiredBytes[TElement](capacityElements)
	listAlignment := memstruct.FixedOrderedListRequiredAlignment[TElement]()
	return MemArchLayoutBlueprintPush(blueprint, listSize, listAlignment)
}

/*
MemArchLayoutBlueprintPushCircularBuffer reserves space for a memstruct.CircularBuffer[TElement].
*/
func MemArchLayoutBlueprintPushCircularBuffer[TElement any](blueprint *LayoutBlueprint, capacityElements uint64) uint64 {
	bufferSize := memstruct.CircularBufferRequiredBytesGet[TElement](capacityElements)
	bufferAlignment := memstruct.CircularBufferRequiredAlignmentGet[TElement]()
	return MemArchLayoutBlueprintPush(blueprint, bufferSize, bufferAlignment)
}

/*
MemArchLayoutBlueprintPushMatrix reserves space for a memstruct.Matrix[TElement].
*/
func MemArchLayoutBlueprintPushMatrix[TElement foundation.Numeric](blueprint *LayoutBlueprint, capacityRows, capacityCols uint64) uint64 {
	matrixSize := memstruct.MatrixRequiredBytesGet[TElement](capacityRows, capacityCols)
	matrixAlignment := memstruct.MatrixRequiredAlignmentGet[TElement]()
	return MemArchLayoutBlueprintPush(blueprint, matrixSize, matrixAlignment)
}

/*
MemArchLayoutBlueprintPushHashMap reserves space for a memstruct.HashMap[TKey, TValue].

[Context]
Binding still requires key callbacks via MemArchLayoutConstructBindHashMap at construction time.
*/
func MemArchLayoutBlueprintPushHashMap[TKey, TValue any](blueprint *LayoutBlueprint, capacityElements uint64) uint64 {
	hashMapSize := memstruct.HashMapRequiredBytesGet[TKey, TValue](capacityElements)
	hashMapAlignment := memstruct.HashMapRequiredAlignmentGet[TKey, TValue]()
	return MemArchLayoutBlueprintPush(blueprint, hashMapSize, hashMapAlignment)
}

/*
MemArchLayoutBlueprintPushString reserves space for a memstruct.String holding content.
*/
func MemArchLayoutBlueprintPushString(blueprint *LayoutBlueprint, content string) uint64 {
	stringSize := memstruct.StringRequiredBytesGet(content)
	stringAlignment := memstruct.StringRequiredAlignmentGet()
	return MemArchLayoutBlueprintPush(blueprint, stringSize, stringAlignment)
}

/*
MemArchLayoutBlueprintPushGoString reserves space for a native Go string in manual memory.
*/
func MemArchLayoutBlueprintPushGoString(blueprint *LayoutBlueprint, content string) uint64 {
	stringSize := memstruct.GoStringRequiredBytesGet(content)
	stringAlignment := memstruct.GoStringRequiredAlignmentGet()
	return MemArchLayoutBlueprintPush(blueprint, stringSize, stringAlignment)
}

/*
MemArchLayoutBlueprintPushCString reserves space for a NUL-terminated C string in manual memory.
*/
func MemArchLayoutBlueprintPushCString(blueprint *LayoutBlueprint, content string) uint64 {
	stringSize := memstruct.CStringRequiredBytesGet(content)
	stringAlignment := memstruct.CStringRequiredAlignmentGet()
	return MemArchLayoutBlueprintPush(blueprint, stringSize, stringAlignment)
}
