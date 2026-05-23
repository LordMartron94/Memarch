package memarch

import (
	"memcore"
)

/*
MemArchValueRequiredBytesGet returns the size in bytes of a single value of type T.

Use with AllocationFn to size manual allocations for plain structs, numeric out-parameters,
and other standard types that are not memstruct containers.
*/
func MemArchValueRequiredBytesGet[T any]() uint64 {
	return memcore.SizeOf[T]()
}

/*
MemArchValueRequiredAlignmentGet returns the required alignment of a single value of type T.
*/
func MemArchValueRequiredAlignmentGet[T any]() uint64 {
	return memcore.AlignOf[T]()
}

/*
MemArchValueCreate allocates manual memory for one value of type T.

Whether the storage is zero-initialized depends on allocFn (for example a memforge Calloc wrapper
versus Malloc). The returned pointer refers to the manual region until the owning allocator is
destroyed or the region is reclaimed. Do not store Go heap pointers inside T when T is written
into manual memory.

Time complexity: O(1)
Space complexity: O(sizeof(T))
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
