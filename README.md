# memarch

Factory layer for creating `memstruct` data structures using `memforge` allocators.

## Overview

`memarch` provides a convenient API for creating data structures from the `memstruct` package, handling memory allocation, sizing, and initialization automatically. It bridges `memforge` allocators and `memstruct` data structures with a clean, easy-to-use interface.

## Design Philosophy

- **Convenience First**: Simplifies the creation of manually managed data structures
- **Type Safe**: Leverages Go generics for compile-time type safety
- **Allocator Agnostic**: Works with any allocator that provides an `AllocationFn`
- **Error Handling**: Encourages wrapping allocation functions for error handling

## Core Concept

Instead of manually calculating sizes, alignments, allocating memory, and initializing structures, `memarch` does it all in one call:

```go
// Without memarch (verbose):
size := memstruct.ArrayRequiredBytesGet[int](100)
align := memstruct.ArrayRequiredAlignmentGet[int]()
mark := allocator.Malloc(size, align)
memstruct.ArrayInitializeAt[int](mark, 100)
array := memcore.MemcoreMarkDereferenceObject[memstruct.Array[int]](mark)

// With memarch (simple):
allocFn := func(sizeBytes, alignment uint64) memcore.MarkRaw {
    return allocator.Malloc(sizeBytes, alignment)
}
mark, array := memarch.MemArchArrayCreate[int](allocFn, 100)
```

## Allocation Functions

`memarch` uses an `AllocationFn` type that abstracts allocation:

```go
type AllocationFn func(sizeBytes, alignment uint64) memcore.MarkRaw
```

You can create this from any allocator:

```go
// From FixedLinearAllocator
allocFn := func(sizeBytes, alignment uint64) memcore.MarkRaw {
    return memforge.FixedLinearAllocatorMalloc(allocator, sizeBytes, alignment)
}

// From FixedManualAllocator
allocFn := func(sizeBytes, alignment uint64) memcore.MarkRaw {
    return memforge.FixedManualAllocatorMalloc(allocator, sizeBytes, alignment)
}

// With error handling
allocFn := func(sizeBytes, alignment uint64) memcore.MarkRaw {
    mark := memforge.FixedLinearAllocatorMalloc(allocator, sizeBytes, alignment)
    if mark.regionID == 0 { // Check for allocation failure
        panic("allocation failed")
    }
    return mark
}
```

## Available Factory Functions

### Array

```go
// Create empty array
mark, array := memarch.MemArchArrayCreate[T](allocFn, capacity)

// Create array from existing array
mark, array := memarch.MemArchArrayCreateFrom[T](allocFn, capacity, srcArrayMark)
```

### Stack

```go
mark, stack := memarch.MemArchStackCreate[T](allocFn, capacity)
```

### Queue

```go
mark, queue := memarch.MemArchQueueCreate[T](allocFn, capacity)
```

### PriorityQueue

```go
// Create empty priority queue
mark, pq := memarch.MemArchPriorityQueueCreate[T](allocFn, capacity)

// Create from existing priority queue
mark, pq := memarch.MemArchPriorityQueueCreateFrom[T](allocFn, capacity, srcQueueMark)
```

### FixedOrderedList

```go
mark, list := memarch.MemArchFixedOrderedListCreate[T](allocFn, capacity)
```

### Vector (Numeric Types Only)

```go
// Create empty vector
mark, vector := memarch.MemArchVectorCreate[T](allocFn, capacity)

// Create vector from existing vector
mark, vector := memarch.MemArchVectorCreateFrom[T](allocFn, capacity, srcVectorMark)
```

### Matrix (Numeric Types Only)

```go
// Create empty matrix
mark, matrix := memarch.MemArchMatrixCreate[T](allocFn, rows, cols)

// Create matrix from existing matrix
mark, matrix := memarch.MemArchMatrixCreateFrom[T](allocFn, rows, cols, srcMatrixMark)
```

### String

```go
// memstruct.String (offset-based header; relocation-safe data pointer)
mark, str := memarch.MemArchStringCreate(allocFn, "Hello, World!")

// Native Go string (runtime string type; data pointer targets manual memory in the same allocation)
mark, goStr := memarch.MemArchGoStringCreate(allocFn, "Hello, World!")
_ = *goStr
```

### HashMap

```go
keyCompare := func(a, b TKey) bool { /* comparison logic */ }
keyMark := func(key TKey) memcore.MarkRaw { /* convert key to mark */ }

mark, hashmap := memarch.MemArchHashMapCreate[TKey, TValue](
    allocFn,
    capacity,
    keyCompare,
    keyMark,
)
```

## Complete Example

```go
package main

import (
    "memcore"
    "memforge"
    "memarch"
    "memstruct"
)

func main() {
    // Create allocator
    allocator := memforge.FixedLinearAllocatorCreate(uint64(memcore.MegaByte))
    defer memforge.FixedLinearAllocatorDestroy(allocator)

    // Wrap as allocation function
    allocFn := func(sizeBytes, alignment uint64) memcore.MarkRaw {
        return memforge.FixedLinearAllocatorMalloc(allocator, sizeBytes, alignment)
    }

    // Create an array
    arrayMark, array := memarch.MemArchArrayCreate[int](allocFn, 100)
    _ = array // Use array...

    // Create a vector
    vectorMark, vector := memarch.MemArchVectorCreate[float64](allocFn, 1000)
    _ = vector // Use vector...

    // Create a matrix
    matrixMark, matrix := memarch.MemArchMatrixCreate[float32](allocFn, 10, 20)
    _ = matrix // Use matrix...

    // Create a stack
    stackMark, stack := memarch.MemArchStackCreate[string](allocFn, 50)
    _ = stack // Use stack...

    // Use memstruct functions on the returned pointers
    memstruct.ArraySetAt[int](arrayMark, 0, 42)
    memstruct.VectorSetAt[float64](vectorMark, 0, 3.14)
    memstruct.MatrixSetAt[float32](matrixMark, 0, 0, 1.5)
    memstruct.StackPush[string](stackMark, "hello")
}
```

## Error Handling

While `memarch` factory functions don't return errors directly, you can wrap your `AllocationFn` to handle allocation failures:

```go
allocFn := func(sizeBytes, alignment uint64) memcore.MarkRaw {
    mark := memforge.FixedLinearAllocatorMalloc(allocator, sizeBytes, alignment)
    
    // Check for allocation failure (implementation depends on allocator)
    if /* allocation failed */ {
        // Handle error - log, return zero mark, or panic
        panic(fmt.Sprintf("failed to allocate %d bytes", sizeBytes))
    }
    
    return mark
}
```

## When to Use memarch

**Use `memarch` when:**
- You want simpler, more readable code
- You're creating many different data structures
- You want to abstract allocation details
- You're prototyping or building higher-level APIs

**Use `memstruct` directly when:**
- You need fine-grained control over allocation
- You're implementing custom allocation strategies
- You're working with existing memory regions
- Performance profiling shows allocation overhead matters

## Integration

`memarch` is the recommended entry point for most users of the manual memory ecosystem:

```
memcore (foundation)
    ↓
memforge (allocators)
    ↓
memarch (factories) ← Start here
    ↓
memstruct (data structures)
    ↓
blaze (numerical computing)
```

For most applications, create allocators with `memforge`, wrap them as `AllocationFn`, and use `memarch` to create your data structures.

## Safety Reminders

⚠️ All safety guidelines from `memstruct` and `memforge` apply:

- Never store Go pointers in manually managed memory
- Ensure allocators outlive their data structures
- Use the returned `MarkRaw` values, not raw pointers, for long-term storage
- Be aware of allocator reset/destroy behavior

## Performance

`memarch` adds minimal overhead:
- One extra function call per creation
- Size/alignment calculations (same as you'd do manually)
- No runtime allocation overhead (same as manual approach)

The convenience is worth the tiny overhead for almost all use cases.
