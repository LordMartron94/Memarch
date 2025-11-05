package memarch

import (
	"fmt"
	"foundation"
	foundationtesting "foundation/testing"
	"math"
	"memcore"
	"memforge"
	"memstruct"
	"testing"
)

// ────────────────────────────────────────────────────────────────
//   GROWTH STRATEGY (Same as user provided)
// ────────────────────────────────────────────────────────────────

func doubleGrowth(currentCap, needed uint64) uint64 {
	newSize := currentCap * 2
	if newSize < needed {
		newSize = needed
	}

	if newSize > uint64(1*memcore.GigaByte) {
		panic("way too much memory for a simple test")
	}

	return newSize
}

var growthFnID memcore.FunctionID = memcore.MemcoreFunctionRegisterTyped[memforge.GrowthStrategy](doubleGrowth)

// ────────────────────────────────────────────────────────────────
//   MAIN TEST SUITE
// ────────────────────────────────────────────────────────────────

func TestVector(t *testing.T) {
	allocator := memforge.DynamicLinearAllocatorCreate(uint64(2*memcore.MegaByte), growthFnID)

	err := foundation.WithContext(
		allocator,
		func(data memcore.MarkRaw) error {

			// ─────────────────────────────
			// 1. Create uint8 vector and set values
			// ─────────────────────────────
			vectorMark, _ := MemArchVectorCreate[uint8](
				func(sizeBytes, alignment uint64) memcore.MarkRaw {
					return memforge.DynamicLinearAllocatorCalloc(data, sizeBytes, alignment)
				},
				3, // capacity
			)

			memstruct.VectorSetAtUnsafe[uint8](vectorMark, 0, 5)
			memstruct.VectorSetAtUnsafe[uint8](vectorMark, 1, 15)
			memstruct.VectorSetAtUnsafe[uint8](vectorMark, 2, 250)

			// ─────────────────────────────
			// 2. Validate basic read-back
			// ─────────────────────────────
			v0 := memstruct.VectorItemGetAtUnsafe[uint8](vectorMark, 0)
			v1 := memstruct.VectorItemGetAtUnsafe[uint8](vectorMark, 1)
			v2 := memstruct.VectorItemGetAtUnsafe[uint8](vectorMark, 2)

			foundationtesting.Assert(v0 == 5 && v1 == 15 && v2 == 250,
				fmt.Sprintf("unexpected vector values: %v, %v, %v", v0, v1, v2),
				"uint8 VectorSetAtUnsafe / GetAtUnsafe ok", t)

			// ─────────────────────────────
			// 3. Magnitude tests
			// ─────────────────────────────
			magF64 := memstruct.VectorMagnitudeF64[uint8](vectorMark)
			magF32 := memstruct.VectorMagnitudeF32[uint8](vectorMark)

			expected := math.Sqrt(float64(5*5 + 15*15 + 250*250))
			foundationtesting.Assert(
				math.Abs(magF64-expected) < 1e-6,
				fmt.Sprintf("VectorMagnitudeF64 incorrect: got=%v expected=%v", magF64, expected),
				"VectorMagnitudeF64 ok", t,
			)
			foundationtesting.Assert(
				math.Abs(float64(magF32)-expected) < 1e-3,
				fmt.Sprintf("VectorMagnitudeF32 incorrect: got=%v expected≈%v", magF32, expected),
				"VectorMagnitudeF32 ok", t,
			)

			// ─────────────────────────────
			// 4. Normalization tests
			// ─────────────────────────────
			newVecAddr, vec := MemArchVectorCreate[float64](
				func(sizeBytes, alignment uint64) memcore.MarkRaw {
					return memforge.DynamicLinearAllocatorCalloc(data, sizeBytes, alignment)
				},
				3, // capacity
			)

			memstruct.VectorNormalizedF64[uint8](vectorMark, newVecAddr)
			newMag := memstruct.VectorMagnitudeF64[float64](newVecAddr)

			foundationtesting.Assert(
				math.Abs(newMag-1.0) < 1e-9,
				fmt.Sprintf("normalized vector magnitude not 1: %v, %s", newMag, vec.String()),
				"VectorNormalizedF64 ok", t,
			)

			// ─────────────────────────────
			// 5. Snapshot + Clear/Restore test
			// ─────────────────────────────
			snapAddr, _ := MemArchVectorCreate[uint8](
				func(sizeBytes, alignment uint64) memcore.MarkRaw {
					return memforge.DynamicLinearAllocatorCalloc(data, sizeBytes, alignment)
				},
				3, // capacity
			)
			memstruct.VectorSnapshotCreate[uint8](snapAddr, vectorMark)

			// mutate the original vector
			memstruct.VectorSetAtUnsafe[uint8](vectorMark, 0, 255)

			// restore from snapshot
			errRestore := memstruct.VectorSnapshotRestore[uint8](vectorMark, snapAddr)
			foundationtesting.Assert(
				errRestore == nil,
				fmt.Sprintf("VectorSnapshotRestore failed: %v", errRestore),
				"VectorSnapshotRestore ok", t,
			)
			restoredVal := memstruct.VectorItemGetAtUnsafe[uint8](vectorMark, 0)
			foundationtesting.Assert(
				restoredVal == 5,
				fmt.Sprintf("VectorSnapshotRestore data incorrect: got=%v expected=5", restoredVal),
				"VectorSnapshotRestore data ok", t,
			)

			// ─────────────────────────────
			// 6. Clear test
			// ─────────────────────────────
			memstruct.VectorClear[uint8](vectorMark)
			clearedVal := memstruct.VectorItemGetAtUnsafe[uint8](vectorMark, 1)
			foundationtesting.Assert(
				clearedVal == 0,
				fmt.Sprintf("VectorClear failed: index 1 = %v", clearedVal),
				"VectorClear ok", t,
			)

			return nil
		},
		func(data memcore.MarkRaw) error {
			memforge.DynamicLinearAllocatorDestroy(data)
			return nil
		},
	)

	if err != nil {
		t.Fatalf("Vector test failed: %v", err)
	}
}
