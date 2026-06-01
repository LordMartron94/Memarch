package memarch

import (
	"memcore"
)

/*
MemArchValueRequiredBytesGet returns the byte size of a single value of type T.

[Context]
Use with AllocationFn to size manual allocations for plain structs, numeric out-parameters,
and other standard types that are not memstruct containers.

[Complexity]
Time: O(1). Space: O(1).

[Side Effects]
Pure function.
*/
func MemArchValueRequiredBytesGet[T any]() uint64 {
	return memcore.SizeOf[T]()
}

/*
MemArchValueRequiredAlignmentGet returns the required alignment of a single value of type T.

[Complexity]
Time: O(1). Space: O(1).

[Side Effects]
Pure function.
*/
func MemArchValueRequiredAlignmentGet[T any]() uint64 {
	return memcore.AlignOf[T]()
}

/*
MemArchValueCreate allocates manual memory for one value of type T.

[Parameters]
allocFn - Allocator callback; must not be nil. Use a Calloc wrapper for zeroed storage.

[Returns]
A memcore.MarkRaw and a *T view of the same storage.

[Complexity]
Time: O(1). Space: O(sizeof(T)).

[Errors]
Panics when allocFn is nil.

[Invariants]
Do not store Go heap pointers inside T when T resides in manual memory. The mark is valid until
the owning allocator is destroyed or the region is reclaimed.
*/
func MemArchValueCreate[T any](allocFn AllocationFn) (memcore.MarkRaw, *T) {
	if allocFn == nil {
		panic("memarch: nil allocation function")
	}
	sizeBytes := memcore.SizeOf[T]()
	alignment := memcore.AlignOf[T]()
	addr := allocFn(sizeBytes, alignment)
	return addr, memcore.MemcoreMarkDereferenceObject[T](addr)
}
