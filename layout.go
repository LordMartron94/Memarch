package memarch

import (
	"fmt"
	"foundation"
	"memcore"
	"memstruct"
)

/*
LayoutBuilder plans aligned component offsets inside a single manual allocation block.

[Context]
Use when one mmap slab contains several sub-regions (structs, arrays, vectors) at fixed offsets.
Push components to record each offset, then MemArchLayoutBuilderAllocate once for the full span.
Reuse the same builder (or MemArchLayoutBuilderClone) to stamp out many instances with identical layout.

[Invariants]
componentOffsets is stored in manual memory via MemArchArrayCreate. Do not store Go heap pointers
inside the composed structure. Component indices are the return values from Push calls, in push order.
*/
type LayoutBuilder struct {
	componentOffsets memcore.MarkRaw // Array[uint64]

	maxAlignment    uint64
	byteCursor      uint64
	componentCursor uint64
	maxComponents   uint64
}

/*
MemArchLayoutBuilderCreate allocates offset metadata for a layout with up to numComponents slots.

[Parameters]
layoutAllocator - Allocates the uint64 offset table (typically a small scratch allocator).
numComponents - Maximum number of MemArchLayoutBuilderPush calls before panic.

[Returns]
A heap-allocated LayoutBuilder; the offset table itself lives in manual memory.

[Side Effects]
Creates an uninitialized uint64 array mark for component offsets.
*/
func MemArchLayoutBuilderCreate(
	layoutAllocator AllocationFn,
	numComponents uint64,
) *LayoutBuilder {
	componentOffsets, _ := MemArchArrayCreate[uint64](layoutAllocator, numComponents)

	return &LayoutBuilder{
		componentOffsets: componentOffsets,
		maxAlignment:     0,
		byteCursor:       0,
		componentCursor:  0,
		maxComponents:    numComponents,
	}
}

/*
MemArchLayoutBuilderPush reserves a component region and records its byte offset.

[Parameters]
bytes - Size of the component in bytes.
alignment - Required alignment (power of two); advances byteCursor with padding.

[Returns]
The component index (0-based, in push order) for use with MarkGet and Bind helpers.

[Errors]
Panics when the number of pushes exceeds maxComponents from Create.

[Complexity]
Time: O(1). Space: O(1).

[Side Effects]
Updates byteCursor and maxAlignment on the builder.
*/
func MemArchLayoutBuilderPush(
	layout *LayoutBuilder,
	bytes uint64,
	alignment uint64,
) uint64 {
	if layout.componentCursor >= layout.maxComponents {
		panic(fmt.Errorf("memarch: layout push beyond max components (%d)", layout.maxComponents))
	}

	alignedCursor := memcore.AlignUp(layout.byteCursor, alignment)

	memstruct.ArraySetAtUnsafe(layout.componentOffsets, layout.componentCursor, alignedCursor)

	layout.byteCursor = alignedCursor + bytes

	if alignment > layout.maxAlignment {
		layout.maxAlignment = alignment
	}

	index := layout.componentCursor
	layout.componentCursor++
	return index
}

/*
MemArchLayoutBuilderClone copies layout geometry into a new offset table.

[Context]
Use when the same component layout is applied to many allocations but each builder needs its own
offset array mark (for example per-thread scratch tables).

[Returns]
A new LayoutBuilder sharing byteCursor, maxAlignment, and componentCursor from state.
*/
func MemArchLayoutBuilderClone(
	state *LayoutBuilder,
	layoutAllocator AllocationFn,
) *LayoutBuilder {
	componentOffsets, _ := MemArchArrayCreate[uint64](layoutAllocator, state.maxComponents)
	memstruct.ArrayCopyFrom[uint64](componentOffsets, state.componentOffsets, 0)

	return &LayoutBuilder{
		maxAlignment:     state.maxAlignment,
		byteCursor:       state.byteCursor,
		componentCursor:  state.componentCursor,
		maxComponents:    state.maxComponents,
		componentOffsets: componentOffsets,
	}
}

/*
MemArchLayoutBuilderRequiredBytesGet returns the total byte span after all Push calls.

[Returns]
layout.byteCursor (unaligned tail; Allocate uses this with maxAlignment).
*/
func MemArchLayoutBuilderRequiredBytesGet(layout *LayoutBuilder) uint64 {
	return layout.byteCursor
}

/*
MemArchLayoutBuilderRequiredAlignmentGet returns the strictest alignment required by any component.
*/
func MemArchLayoutBuilderRequiredAlignmentGet(layout *LayoutBuilder) uint64 {
	return layout.maxAlignment
}

/*
MemArchLayoutBuilderComponentCountGet returns how many components have been pushed so far.
*/
func MemArchLayoutBuilderComponentCountGet(layout *LayoutBuilder) uint64 {
	return layout.componentCursor
}

/*
MemArchLayoutBuilderAllocate maps a single block sized for the completed layout.

[Parameters]
allocationFn - Typically the parent arena allocator; receives byteCursor and maxAlignment.

[Returns]
A mark to the base of the composite block; use MarkGet/Bind to reach each component.
*/
func MemArchLayoutBuilderAllocate(
	layout *LayoutBuilder,
	allocationFn AllocationFn,
) memcore.MarkRaw {
	return allocationFn(layout.byteCursor, layout.maxAlignment)
}

/*
MemArchLayoutBuilderMarkGet returns baseMark advanced by the stored offset for componentIndex.
*/
func MemArchLayoutBuilderMarkGet(
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
) memcore.MarkRaw {
	offset := memstruct.ArrayItemGetAtUnsafe[uint64](layout.componentOffsets, componentIndex)
	return memcore.MemcoreMarkOffsetFrom(baseMark, uintptr(offset))
}

// ---------------------------------------------------------------------- PUSH (SIZE PLANNING)

/*
MemArchLayoutBuilderPushStruct reserves space for a value of type T.
*/
func MemArchLayoutBuilderPushStruct[T any](layout *LayoutBuilder) uint64 {
	return MemArchLayoutBuilderPush(layout, memcore.SizeOf[T](), memcore.AlignOf[T]())
}

/*
MemArchLayoutBuilderPushArray reserves space for a memstruct.Array[TElement] with capacityElements.
*/
func MemArchLayoutBuilderPushArray[TElement any](layout *LayoutBuilder, capacityElements uint64) uint64 {
	arraySize := memstruct.ArrayRequiredBytesGet[TElement](capacityElements)
	arrayAlignment := memstruct.ArrayRequiredAlignmentGet[TElement]()
	return MemArchLayoutBuilderPush(layout, arraySize, arrayAlignment)
}

/*
MemArchLayoutBuilderPushVector reserves space for a memstruct.Vector[TElement].
*/
func MemArchLayoutBuilderPushVector[TElement foundation.Numeric](layout *LayoutBuilder, capacityElements uint64) uint64 {
	vectorSize := memstruct.VectorRequiredBytesGet[TElement](capacityElements)
	vectorAlignment := memstruct.VectorRequiredAlignmentGet[TElement]()
	return MemArchLayoutBuilderPush(layout, vectorSize, vectorAlignment)
}

/*
MemArchLayoutBuilderPushStack reserves space for a memstruct.Stack[TElement].
*/
func MemArchLayoutBuilderPushStack[TElement any](layout *LayoutBuilder, capacityElements uint64) uint64 {
	stackSize := memstruct.StackRequiredBytesGet[TElement](capacityElements)
	stackAlignment := memstruct.StackRequiredAlignmentGet[TElement]()
	return MemArchLayoutBuilderPush(layout, stackSize, stackAlignment)
}

/*
MemArchLayoutBuilderPushQueue reserves space for a memstruct.Queue[TElement].
*/
func MemArchLayoutBuilderPushQueue[TElement any](layout *LayoutBuilder, capacityElements uint64) uint64 {
	queueSize := memstruct.QueueRequiredBytesGet[TElement](capacityElements)
	queueAlignment := memstruct.QueueRequiredAlignmentGet[TElement]()
	return MemArchLayoutBuilderPush(layout, queueSize, queueAlignment)
}

/*
MemArchLayoutBuilderPushPriorityQueue reserves space for a memstruct.PriorityQueue[TElement].
*/
func MemArchLayoutBuilderPushPriorityQueue[TElement any](layout *LayoutBuilder, capacityElements uint64) uint64 {
	queueSize := memstruct.PriorityQueueRequiredBytesGet[TElement](capacityElements)
	queueAlignment := memstruct.PriorityQueueRequiredAlignmentGet[TElement]()
	return MemArchLayoutBuilderPush(layout, queueSize, queueAlignment)
}

/*
MemArchLayoutBuilderPushFixedOrderedList reserves space for a memstruct.FixedOrderedList[TElement].
*/
func MemArchLayoutBuilderPushFixedOrderedList[TElement any](layout *LayoutBuilder, capacityElements uint64) uint64 {
	listSize := memstruct.FixedOrderedListRequiredBytes[TElement](capacityElements)
	listAlignment := memstruct.FixedOrderedListRequiredAlignment[TElement]()
	return MemArchLayoutBuilderPush(layout, listSize, listAlignment)
}

/*
MemArchLayoutBuilderPushCircularBuffer reserves space for a memstruct.CircularBuffer[TElement].
*/
func MemArchLayoutBuilderPushCircularBuffer[TElement any](layout *LayoutBuilder, capacityElements uint64) uint64 {
	bufferSize := memstruct.CircularBufferRequiredBytesGet[TElement](capacityElements)
	bufferAlignment := memstruct.CircularBufferRequiredAlignmentGet[TElement]()
	return MemArchLayoutBuilderPush(layout, bufferSize, bufferAlignment)
}

/*
MemArchLayoutBuilderPushMatrix reserves space for a memstruct.Matrix[TElement].
*/
func MemArchLayoutBuilderPushMatrix[TElement foundation.Numeric](layout *LayoutBuilder, capacityRows, capacityCols uint64) uint64 {
	matrixSize := memstruct.MatrixRequiredBytesGet[TElement](capacityRows, capacityCols)
	matrixAlignment := memstruct.MatrixRequiredAlignmentGet[TElement]()
	return MemArchLayoutBuilderPush(layout, matrixSize, matrixAlignment)
}

/*
MemArchLayoutBuilderPushHashMap reserves space for a memstruct.HashMap[TKey, TValue].

[Context]
Binding still requires key comparer and mark retriever callbacks via MemArchLayoutBuilderBindHashMap.
*/
func MemArchLayoutBuilderPushHashMap[TKey, TValue any](layout *LayoutBuilder, capacityElements uint64) uint64 {
	hashMapSize := memstruct.HashMapRequiredBytesGet[TKey, TValue](capacityElements)
	hashMapAlignment := memstruct.HashMapRequiredAlignmentGet[TKey, TValue]()
	return MemArchLayoutBuilderPush(layout, hashMapSize, hashMapAlignment)
}

/*
MemArchLayoutBuilderPushString reserves space for a memstruct.String holding content.
*/
func MemArchLayoutBuilderPushString(layout *LayoutBuilder, content string) uint64 {
	stringSize := memstruct.StringRequiredBytesGet(content)
	stringAlignment := memstruct.StringRequiredAlignmentGet()
	return MemArchLayoutBuilderPush(layout, stringSize, stringAlignment)
}

/*
MemArchLayoutBuilderPushGoString reserves space for a native Go string in manual memory.
*/
func MemArchLayoutBuilderPushGoString(layout *LayoutBuilder, content string) uint64 {
	stringSize := memstruct.GoStringRequiredBytesGet(content)
	stringAlignment := memstruct.GoStringRequiredAlignmentGet()
	return MemArchLayoutBuilderPush(layout, stringSize, stringAlignment)
}

/*
MemArchLayoutBuilderPushCString reserves space for a NUL-terminated C string in manual memory.
*/
func MemArchLayoutBuilderPushCString(layout *LayoutBuilder, content string) uint64 {
	stringSize := memstruct.CStringRequiredBytesGet(content)
	stringAlignment := memstruct.CStringRequiredAlignmentGet()
	return MemArchLayoutBuilderPush(layout, stringSize, stringAlignment)
}

// ---------------------------------------------------------------------- BIND (INITIALIZE IN PLACE)

/*
MemArchLayoutBuilderBindStruct returns a *T view at the component offset without zeroing.
*/
func MemArchLayoutBuilderBindStruct[T any](
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
) *T {
	mark := MemArchLayoutBuilderMarkGet(layout, baseMark, componentIndex)
	return memcore.MemcoreMarkDereferenceObject[T](mark)
}

/*
MemArchLayoutBuilderBindArray initializes an array at the component and returns its view.
*/
func MemArchLayoutBuilderBindArray[TElement any](
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
) *memstruct.Array[TElement] {
	mark := MemArchLayoutBuilderMarkGet(layout, baseMark, componentIndex)
	memstruct.ArrayInitializeAt[TElement](mark, capacityElements)
	return memcore.MemcoreMarkDereferenceObject[memstruct.Array[TElement]](mark)
}

/*
MemArchLayoutBuilderBindVector initializes a vector at the component and returns its view.
*/
func MemArchLayoutBuilderBindVector[TElement foundation.Numeric](
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
) *memstruct.Vector[TElement] {
	mark := MemArchLayoutBuilderMarkGet(layout, baseMark, componentIndex)
	memstruct.VectorInitializeAt[TElement](mark, capacityElements)
	return memcore.MemcoreMarkDereferenceObject[memstruct.Vector[TElement]](mark)
}

/*
MemArchLayoutBuilderBindStack initializes a stack at the component and returns its view.
*/
func MemArchLayoutBuilderBindStack[TElement any](
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
) *memstruct.Stack[TElement] {
	mark := MemArchLayoutBuilderMarkGet(layout, baseMark, componentIndex)
	memstruct.StackInitializeAt[TElement](mark, capacityElements)
	return memcore.MemcoreMarkDereferenceObject[memstruct.Stack[TElement]](mark)
}

/*
MemArchLayoutBuilderBindQueue initializes a queue at the component and returns its view.
*/
func MemArchLayoutBuilderBindQueue[TElement any](
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
) *memstruct.Queue[TElement] {
	mark := MemArchLayoutBuilderMarkGet(layout, baseMark, componentIndex)
	memstruct.QueueInitializeAt[TElement](mark, capacityElements)
	return memcore.MemcoreMarkDereferenceObject[memstruct.Queue[TElement]](mark)
}

/*
MemArchLayoutBuilderBindPriorityQueue initializes a priority queue at the component.
*/
func MemArchLayoutBuilderBindPriorityQueue[TElement any](
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
) *memstruct.PriorityQueue[TElement] {
	mark := MemArchLayoutBuilderMarkGet(layout, baseMark, componentIndex)
	memstruct.PriorityQueueInitializeAt[TElement](mark, capacityElements)
	return memcore.MemcoreMarkDereferenceObject[memstruct.PriorityQueue[TElement]](mark)
}

/*
MemArchLayoutBuilderBindFixedOrderedList initializes a fixed ordered list at the component.
*/
func MemArchLayoutBuilderBindFixedOrderedList[TElement any](
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
) *memstruct.FixedOrderedList[TElement] {
	mark := MemArchLayoutBuilderMarkGet(layout, baseMark, componentIndex)
	memstruct.FixedOrderedListInitializeAt[TElement](mark, capacityElements)
	return memcore.MemcoreMarkDereferenceObject[memstruct.FixedOrderedList[TElement]](mark)
}

/*
MemArchLayoutBuilderBindCircularBuffer initializes a circular buffer at the component.
*/
func MemArchLayoutBuilderBindCircularBuffer[TElement any](
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
) *memstruct.CircularBuffer[TElement] {
	mark := MemArchLayoutBuilderMarkGet(layout, baseMark, componentIndex)
	memstruct.CircularBufferInitializeAt[TElement](mark, capacityElements)
	return memcore.MemcoreMarkDereferenceObject[memstruct.CircularBuffer[TElement]](mark)
}

/*
MemArchLayoutBuilderBindMatrix initializes a matrix at the component.
*/
func MemArchLayoutBuilderBindMatrix[TElement foundation.Numeric](
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityRows, capacityCols uint64,
) *memstruct.Matrix[TElement] {
	mark := MemArchLayoutBuilderMarkGet(layout, baseMark, componentIndex)
	memstruct.MatrixInitializeAt[TElement](mark, capacityRows, capacityCols)
	return memcore.MemcoreMarkDereferenceObject[memstruct.Matrix[TElement]](mark)
}

/*
MemArchLayoutBuilderBindHashMap initializes a hash map at the component.

[Parameters]
keyComparisonFunc, keyMarkFunc - Required memstruct key callbacks (same as MemArchHashMapCreate).
*/
func MemArchLayoutBuilderBindHashMap[TKey, TValue any](
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
	keyComparisonFunc memstruct.KeyComparer[TKey],
	keyMarkFunc memstruct.KeyMarkRetriever[TKey],
) *memstruct.HashMap[TKey, TValue] {
	mark := MemArchLayoutBuilderMarkGet(layout, baseMark, componentIndex)
	memstruct.HashMapInitializeAt[TKey, TValue](mark, capacityElements, keyComparisonFunc, keyMarkFunc)
	return memcore.MemcoreMarkDereferenceObject[memstruct.HashMap[TKey, TValue]](mark)
}

/*
MemArchLayoutBuilderBindString initializes a memstruct.String at the component.
*/
func MemArchLayoutBuilderBindString(
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	content string,
) *memstruct.String {
	mark := MemArchLayoutBuilderMarkGet(layout, baseMark, componentIndex)
	memstruct.StringInitializeAt(mark, content)
	return memcore.MemcoreMarkDereferenceObject[memstruct.String](mark)
}

/*
MemArchLayoutBuilderBindGoString initializes a native Go string in manual memory at the component.
*/
func MemArchLayoutBuilderBindGoString(
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	content string,
) *string {
	mark := MemArchLayoutBuilderMarkGet(layout, baseMark, componentIndex)
	memstruct.GoStringInitializeAt(mark, content)
	return memcore.MemcoreMarkDereferenceObject[string](mark)
}

/*
MemArchLayoutBuilderBindCString initializes a C string at the component.

[Returns]
*byte pointing at the first byte of the C string payload.
*/
func MemArchLayoutBuilderBindCString(
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	content string,
) *byte {
	mark := MemArchLayoutBuilderMarkGet(layout, baseMark, componentIndex)
	memstruct.CStringInitializeAt(mark, content)
	return memstruct.CStringPointerGet(mark)
}

// ---------------------------------------------------------------------- FIELD WIRING

func memarchLayoutBuilderBindMarkToField(
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	outTargetField *memcore.MarkRaw,
) memcore.MarkRaw {
	mark := MemArchLayoutBuilderMarkGet(layout, baseMark, componentIndex)
	if outTargetField != nil {
		*outTargetField = mark
	}
	return mark
}

/*
MemArchLayoutBuilderBindMarkToField resolves a component mark and optionally stores it in a header field.

[Context]
Use for parent structs that store memcore.MarkRaw references to child regions inside the same block.
*/
func MemArchLayoutBuilderBindMarkToField(
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	outTargetField *memcore.MarkRaw,
) memcore.MarkRaw {
	return memarchLayoutBuilderBindMarkToField(layout, baseMark, componentIndex, outTargetField)
}

/*
MemArchLayoutBuilderBindArrayToField initializes an array component and wires its mark into outTargetField.
*/
func MemArchLayoutBuilderBindArrayToField[TElement any](
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
	outTargetField *memcore.MarkRaw,
) *memstruct.Array[TElement] {
	mark := memarchLayoutBuilderBindMarkToField(layout, baseMark, componentIndex, outTargetField)
	memstruct.ArrayInitializeAt[TElement](mark, capacityElements)
	return memcore.MemcoreMarkDereferenceObject[memstruct.Array[TElement]](mark)
}

/*
MemArchLayoutBuilderBindVectorToField initializes a vector component and wires its mark into outTargetField.
*/
func MemArchLayoutBuilderBindVectorToField[TElement foundation.Numeric](
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
	outTargetField *memcore.MarkRaw,
) *memstruct.Vector[TElement] {
	mark := memarchLayoutBuilderBindMarkToField(layout, baseMark, componentIndex, outTargetField)
	memstruct.VectorInitializeAt[TElement](mark, capacityElements)
	return memcore.MemcoreMarkDereferenceObject[memstruct.Vector[TElement]](mark)
}

/*
MemArchLayoutBuilderBindStackToField initializes a stack component and wires its mark into outTargetField.
*/
func MemArchLayoutBuilderBindStackToField[TElement any](
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
	outTargetField *memcore.MarkRaw,
) *memstruct.Stack[TElement] {
	mark := memarchLayoutBuilderBindMarkToField(layout, baseMark, componentIndex, outTargetField)
	memstruct.StackInitializeAt[TElement](mark, capacityElements)
	return memcore.MemcoreMarkDereferenceObject[memstruct.Stack[TElement]](mark)
}

/*
MemArchLayoutBuilderBindQueueToField initializes a queue component and wires its mark into outTargetField.
*/
func MemArchLayoutBuilderBindQueueToField[TElement any](
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
	outTargetField *memcore.MarkRaw,
) *memstruct.Queue[TElement] {
	mark := memarchLayoutBuilderBindMarkToField(layout, baseMark, componentIndex, outTargetField)
	memstruct.QueueInitializeAt[TElement](mark, capacityElements)
	return memcore.MemcoreMarkDereferenceObject[memstruct.Queue[TElement]](mark)
}

/*
MemArchLayoutBuilderBindMatrixToField initializes a matrix component and wires its mark into outTargetField.
*/
func MemArchLayoutBuilderBindMatrixToField[TElement foundation.Numeric](
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityRows, capacityCols uint64,
	outTargetField *memcore.MarkRaw,
) *memstruct.Matrix[TElement] {
	mark := memarchLayoutBuilderBindMarkToField(layout, baseMark, componentIndex, outTargetField)
	memstruct.MatrixInitializeAt[TElement](mark, capacityRows, capacityCols)
	return memcore.MemcoreMarkDereferenceObject[memstruct.Matrix[TElement]](mark)
}

/*
MemArchLayoutBuilderBindSubStructToField maps a nested struct pointer into a parent **T field.
*/
func MemArchLayoutBuilderBindSubStructToField[T any](
	layout *LayoutBuilder,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	outTargetField **T,
) *T {
	mark := MemArchLayoutBuilderMarkGet(layout, baseMark, componentIndex)
	ptr := memcore.MemcoreMarkDereferenceObject[T](mark)

	if outTargetField != nil {
		*outTargetField = ptr
	}

	return ptr
}
