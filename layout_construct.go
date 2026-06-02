package memarch

import (
	"fmt"
	"foundation"
	"memcore"
	"memstruct"
)

func memarchLayoutBlueprintValidate(blueprint *LayoutBlueprint) {
	if blueprint == nil {
		panic("memarch: layout blueprint must not be nil")
	}
}

/*
MemArchLayoutConstructAllocate maps one manual block sized for the blueprint.

[Parameters]
blueprint - Completed layout blueprint.
allocationFn - Parent allocator; receives RequiredBytes and RequiredAlignment from the blueprint.

[Returns]
Base mark for the composite instance.
*/
func MemArchLayoutConstructAllocate(blueprint *LayoutBlueprint, allocationFn AllocationFn) memcore.MarkRaw {
	memarchLayoutBlueprintValidate(blueprint)
	return allocationFn(blueprint.totalBytes, blueprint.maxAlignment)
}

/*
MemArchLayoutConstructMarkGet returns baseMark offset by the blueprint component index.
*/
func MemArchLayoutConstructMarkGet(
	blueprint *LayoutBlueprint,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
) memcore.MarkRaw {
	memarchLayoutBlueprintValidate(blueprint)
	offset := MemArchLayoutBlueprintComponentOffsetGet(blueprint, componentIndex)
	return memcore.MemcoreMarkOffsetFrom(baseMark, uintptr(offset))
}

// ---------------------------------------------------------------------- BIND (RUNTIME INITIALIZATION)

/*
MemArchLayoutConstructBindStruct returns a transient *T view at the component without zeroing.

[Invariants]
Do not store the returned pointer inside manual memory; persist memcore.MarkRaw via
MemArchLayoutConstructBindMarkToField instead.
*/
func MemArchLayoutConstructBindStruct[T any](
	blueprint *LayoutBlueprint,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
) *T {
	memarchManualTypeValidate[T]("MemArchLayoutConstructBindStruct")
	mark := MemArchLayoutConstructMarkGet(blueprint, baseMark, componentIndex)
	value := memcore.MemcoreMarkDereferenceObject[T](mark)
	return value
}

/*
MemArchLayoutConstructBindArray initializes an array at the component.
*/
func MemArchLayoutConstructBindArray[TElement any](
	blueprint *LayoutBlueprint,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
) *memstruct.Array[TElement] {
	mark := MemArchLayoutConstructMarkGet(blueprint, baseMark, componentIndex)
	memstruct.ArrayInitializeAt[TElement](mark, capacityElements)
	value := memcore.MemcoreMarkDereferenceObject[memstruct.Array[TElement]](mark)
	return value
}

/*
MemArchLayoutConstructBindVector initializes a vector at the component.
*/
func MemArchLayoutConstructBindVector[TElement foundation.Numeric](
	blueprint *LayoutBlueprint,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
) *memstruct.Vector[TElement] {
	mark := MemArchLayoutConstructMarkGet(blueprint, baseMark, componentIndex)
	memstruct.VectorInitializeAt[TElement](mark, capacityElements)
	value := memcore.MemcoreMarkDereferenceObject[memstruct.Vector[TElement]](mark)
	return value
}

/*
MemArchLayoutConstructBindStack initializes a stack at the component.
*/
func MemArchLayoutConstructBindStack[TElement any](
	blueprint *LayoutBlueprint,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
) *memstruct.Stack[TElement] {
	mark := MemArchLayoutConstructMarkGet(blueprint, baseMark, componentIndex)
	memstruct.StackInitializeAt[TElement](mark, capacityElements)
	value := memcore.MemcoreMarkDereferenceObject[memstruct.Stack[TElement]](mark)
	return value
}

/*
MemArchLayoutConstructBindQueue initializes a queue at the component.
*/
func MemArchLayoutConstructBindQueue[TElement any](
	blueprint *LayoutBlueprint,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
) *memstruct.Queue[TElement] {
	mark := MemArchLayoutConstructMarkGet(blueprint, baseMark, componentIndex)
	memstruct.QueueInitializeAt[TElement](mark, capacityElements)
	value := memcore.MemcoreMarkDereferenceObject[memstruct.Queue[TElement]](mark)
	return value
}

/*
MemArchLayoutConstructBindPriorityQueue initializes a priority queue at the component.
*/
func MemArchLayoutConstructBindPriorityQueue[TElement any](
	blueprint *LayoutBlueprint,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
) *memstruct.PriorityQueue[TElement] {
	mark := MemArchLayoutConstructMarkGet(blueprint, baseMark, componentIndex)
	memstruct.PriorityQueueInitializeAt[TElement](mark, capacityElements)
	value := memcore.MemcoreMarkDereferenceObject[memstruct.PriorityQueue[TElement]](mark)
	return value
}

/*
MemArchLayoutConstructBindFixedOrderedList initializes a fixed ordered list at the component.
*/
func MemArchLayoutConstructBindFixedOrderedList[TElement any](
	blueprint *LayoutBlueprint,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
) *memstruct.FixedOrderedList[TElement] {
	mark := MemArchLayoutConstructMarkGet(blueprint, baseMark, componentIndex)
	memstruct.FixedOrderedListInitializeAt[TElement](mark, capacityElements)
	value := memcore.MemcoreMarkDereferenceObject[memstruct.FixedOrderedList[TElement]](mark)
	return value
}

/*
MemArchLayoutConstructBindCircularBuffer initializes a circular buffer at the component.
*/
func MemArchLayoutConstructBindCircularBuffer[TElement any](
	blueprint *LayoutBlueprint,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
) *memstruct.CircularBuffer[TElement] {
	mark := MemArchLayoutConstructMarkGet(blueprint, baseMark, componentIndex)
	memstruct.CircularBufferInitializeAt[TElement](mark, capacityElements)
	value := memcore.MemcoreMarkDereferenceObject[memstruct.CircularBuffer[TElement]](mark)
	return value
}

/*
MemArchLayoutConstructBindMatrix initializes a matrix at the component.
*/
func MemArchLayoutConstructBindMatrix[TElement foundation.Numeric](
	blueprint *LayoutBlueprint,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityRows, capacityCols uint64,
) *memstruct.Matrix[TElement] {
	mark := MemArchLayoutConstructMarkGet(blueprint, baseMark, componentIndex)
	memstruct.MatrixInitializeAt[TElement](mark, capacityRows, capacityCols)
	value := memcore.MemcoreMarkDereferenceObject[memstruct.Matrix[TElement]](mark)
	return value
}

/*
MemArchLayoutConstructBindHashMap initializes a hash map at the component.
*/
func MemArchLayoutConstructBindHashMap[TKey, TValue any](
	blueprint *LayoutBlueprint,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
	keyComparisonFunc memstruct.KeyComparer[TKey],
	keyMarkFunc memstruct.KeyMarkRetriever[TKey],
) *memstruct.HashMap[TKey, TValue] {
	mark := MemArchLayoutConstructMarkGet(blueprint, baseMark, componentIndex)
	memstruct.HashMapInitializeAt[TKey, TValue](mark, capacityElements, keyComparisonFunc, keyMarkFunc)
	value := memcore.MemcoreMarkDereferenceObject[memstruct.HashMap[TKey, TValue]](mark)
	return value
}

/*
MemArchLayoutConstructBindString initializes a memstruct.String at the component.
*/
func MemArchLayoutConstructBindString(
	blueprint *LayoutBlueprint,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	content string,
) *memstruct.String {
	mark := MemArchLayoutConstructMarkGet(blueprint, baseMark, componentIndex)
	memstruct.StringInitializeAt(mark, content)
	value := memcore.MemcoreMarkDereferenceObject[memstruct.String](mark)
	return value
}

/*
MemArchLayoutConstructBindGoString initializes a native Go string at the component.
*/
func MemArchLayoutConstructBindGoString(
	blueprint *LayoutBlueprint,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	content string,
) *string {
	mark := MemArchLayoutConstructMarkGet(blueprint, baseMark, componentIndex)
	memstruct.GoStringInitializeAt(mark, content)
	value := memcore.MemcoreMarkDereferenceObject[string](mark)
	return value
}

/*
MemArchLayoutConstructBindCString initializes a C string at the component.

[Returns]
Transient *byte view; do not persist in manual memory.
*/
func MemArchLayoutConstructBindCString(
	blueprint *LayoutBlueprint,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	content string,
) *byte {
	mark := MemArchLayoutConstructMarkGet(blueprint, baseMark, componentIndex)
	memstruct.CStringInitializeAt(mark, content)
	return memstruct.CStringPointerGet(mark)
}

// ---------------------------------------------------------------------- FIELD WIRING

/*
MemArchLayoutConstructBindMarkToField resolves a component mark and stores it in a manual-memory field.

[Context]
Use when one region inside a composite block must refer to another (nested struct, array header, etc.).
The field must be memcore.MarkRaw so references survive MemcoreRegionBaseUpdate and remap.

[Parameters]
outTargetField - Address of the MarkRaw field inside manual memory; must not be nil.

[Returns]
The component mark written to outTargetField.
*/
func MemArchLayoutConstructBindMarkToField(
	blueprint *LayoutBlueprint,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	outTargetField *memcore.MarkRaw,
) memcore.MarkRaw {
	memarchLayoutBlueprintValidate(blueprint)
	if outTargetField == nil {
		panic(fmt.Errorf("memarch: layout bind mark to field requires non-nil outTargetField"))
	}
	mark := MemArchLayoutConstructMarkGet(blueprint, baseMark, componentIndex)
	*outTargetField = mark
	return mark
}

/*
MemArchLayoutConstructBindArrayToField initializes an array and wires its mark into outTargetField.
*/
func MemArchLayoutConstructBindArrayToField[TElement any](
	blueprint *LayoutBlueprint,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
	outTargetField *memcore.MarkRaw,
) *memstruct.Array[TElement] {
	mark := MemArchLayoutConstructBindMarkToField(blueprint, baseMark, componentIndex, outTargetField)
	memstruct.ArrayInitializeAt[TElement](mark, capacityElements)
	value := memcore.MemcoreMarkDereferenceObject[memstruct.Array[TElement]](mark)
	return value
}

/*
MemArchLayoutConstructBindVectorToField initializes a vector and wires its mark into outTargetField.
*/
func MemArchLayoutConstructBindVectorToField[TElement foundation.Numeric](
	blueprint *LayoutBlueprint,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
	outTargetField *memcore.MarkRaw,
) *memstruct.Vector[TElement] {
	mark := MemArchLayoutConstructBindMarkToField(blueprint, baseMark, componentIndex, outTargetField)
	memstruct.VectorInitializeAt[TElement](mark, capacityElements)
	value := memcore.MemcoreMarkDereferenceObject[memstruct.Vector[TElement]](mark)
	return value
}

/*
MemArchLayoutConstructBindStackToField initializes a stack and wires its mark into outTargetField.
*/
func MemArchLayoutConstructBindStackToField[TElement any](
	blueprint *LayoutBlueprint,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
	outTargetField *memcore.MarkRaw,
) *memstruct.Stack[TElement] {
	mark := MemArchLayoutConstructBindMarkToField(blueprint, baseMark, componentIndex, outTargetField)
	memstruct.StackInitializeAt[TElement](mark, capacityElements)
	value := memcore.MemcoreMarkDereferenceObject[memstruct.Stack[TElement]](mark)
	return value
}

/*
MemArchLayoutConstructBindQueueToField initializes a queue and wires its mark into outTargetField.
*/
func MemArchLayoutConstructBindQueueToField[TElement any](
	blueprint *LayoutBlueprint,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityElements uint64,
	outTargetField *memcore.MarkRaw,
) *memstruct.Queue[TElement] {
	mark := MemArchLayoutConstructBindMarkToField(blueprint, baseMark, componentIndex, outTargetField)
	memstruct.QueueInitializeAt[TElement](mark, capacityElements)
	value := memcore.MemcoreMarkDereferenceObject[memstruct.Queue[TElement]](mark)
	return value
}

/*
MemArchLayoutConstructBindMatrixToField initializes a matrix and wires its mark into outTargetField.
*/
func MemArchLayoutConstructBindMatrixToField[TElement foundation.Numeric](
	blueprint *LayoutBlueprint,
	baseMark memcore.MarkRaw,
	componentIndex uint64,
	capacityRows, capacityCols uint64,
	outTargetField *memcore.MarkRaw,
) *memstruct.Matrix[TElement] {
	mark := MemArchLayoutConstructBindMarkToField(blueprint, baseMark, componentIndex, outTargetField)
	memstruct.MatrixInitializeAt[TElement](mark, capacityRows, capacityCols)
	value := memcore.MemcoreMarkDereferenceObject[memstruct.Matrix[TElement]](mark)
	return value
}
