package memarch

import (
	"fmt"
	"foundation"
	"foundation/benchmarking"
	"math"
	"math/rand"
	"memcore"
	"memforge"
	"memstruct"
	"runtime/debug"
	"testing"
)

// ────────────────────────────────────────────────────────────────
//   HELPER: random data
// ────────────────────────────────────────────────────────────────

func fillVectorRandom[T foundation.Numeric](mark memcore.MarkRaw, scale float64) {
	rnd := rand.New(rand.NewSource(42))
	for i := uint64(0); i < memstruct.VectorCapacityGet[T](mark); i++ {
		val := T(rnd.Float64() * scale)
		memstruct.VectorSetAtUnsafe[T](mark, i, val)
	}
}

// ────────────────────────────────────────────────────────────────
//   BENCHMARK SUITE
// ────────────────────────────────────────────────────────────────

func BenchmarkVectorSuite(b *testing.B) {
	sizes := []uint64{64, 256, 1024, 4096, 16384}
	scales := []float64{1, 10, 100, 1000}

	for _, n := range sizes {
		for _, scale := range scales {
			groupName := fmt.Sprintf("N=%d/scale=%.0f", n, scale)

			// Float64
			b.Run(groupName+"/Float64", func(b *testing.B) {
				runVectorBench[float64](b, n, scale)
			})

			// Float32
			b.Run(groupName+"/Float32", func(b *testing.B) {
				runVectorBench[float32](b, n, scale)
			})

			// Int64
			b.Run(groupName+"/Int64", func(b *testing.B) {
				runVectorBench[int64](b, n, scale)
			})

			// Uint64
			b.Run(groupName+"/Uint64", func(b *testing.B) {
				runVectorBench[uint64](b, n, scale)
			})
		}
	}
}

// ────────────────────────────────────────────────────────────────
//   HELPER: benchmark each type
// ────────────────────────────────────────────────────────────────

func runVectorBench[T foundation.Numeric](b *testing.B, capacity uint64, scale float64) {
	type benchData struct {
		allocator memcore.MarkRaw
		vector    memcore.MarkRaw
		oldGC     int
	}

	benchmarking.BenchmarkWithMetrics(b,
		// SETUP
		func(b *testing.B) benchData {
			old := debug.SetGCPercent(-1)
			allocator := memforge.DynamicLinearAllocatorCreate(uint64(2*memcore.MegaByte), growthFnID)
			vector, _ := MemArchVectorCreate[T](
				func(sizeBytes, alignment uint64) memcore.MarkRaw {
					return memforge.DynamicLinearAllocatorCalloc(allocator, sizeBytes, alignment)
				},
				capacity,
			)
			fillVectorRandom[T](vector, scale)
			return benchData{allocator, vector, old}
		},

		// RUN
		func(d benchData, b *testing.B) {
			for i := 0; i < b.N; i++ {
				switch i % 6 {
				case 0:
					memstruct.VectorSum[T](d.vector)
				case 1:
					memstruct.VectorSumSquared[T](d.vector)
				case 2:
					memstruct.VectorMagnitudeF64[T](d.vector)
				case 3:
					memstruct.VectorMagnitudeF32[T](d.vector)
				case 4:
					idx := uint64(i % int(capacity))
					memstruct.VectorItemGetAtUnsafe[T](d.vector, idx)
				case 5:
					idx := uint64(i % int(capacity))
					val := T(math.Mod(float64(i), scale))
					memstruct.VectorSetAtUnsafe[T](d.vector, idx, val)
				}
			}
		},

		// CLEANUP
		func(d benchData, b *testing.B) {
			memforge.DynamicLinearAllocatorDestroy(d.allocator)
			debug.SetGCPercent(d.oldGC)
		},
	)
}

// ────────────────────────────────────────────────────────────────
//   NORMALIZATION BENCHMARKS
// ────────────────────────────────────────────────────────────────

func BenchmarkVectorSuite_Normalization(b *testing.B) {
	sizes := []uint64{128, 512, 2048}
	for _, n := range sizes {
		b.Run(fmt.Sprintf("Normalize/N=%d", n), func(b *testing.B) {
			type benchData struct {
				allocator memcore.MarkRaw
				srcVec    memcore.MarkRaw
				dstVec    memcore.MarkRaw
				oldGC     int
			}

			benchmarking.BenchmarkWithMetrics(b,
				func(b *testing.B) benchData {
					old := debug.SetGCPercent(-1)
					a := memforge.DynamicLinearAllocatorCreate(uint64(2*memcore.MegaByte), growthFnID)
					src, _ := MemArchVectorCreate[float64](
						func(sizeBytes, alignment uint64) memcore.MarkRaw {
							return memforge.DynamicLinearAllocatorCalloc(a, sizeBytes, alignment)
						},
						n,
					)
					dst, _ := MemArchVectorCreate[float64](
						func(sizeBytes, alignment uint64) memcore.MarkRaw {
							return memforge.DynamicLinearAllocatorCalloc(a, sizeBytes, alignment)
						},
						n,
					)
					fillVectorRandom[float64](src, 100)
					return benchData{allocator: a, srcVec: src, dstVec: dst, oldGC: old}
				},

				func(d benchData, b *testing.B) {
					for i := 0; i < b.N; i++ {
						memstruct.VectorNormalizedF64[float64](d.srcVec, d.dstVec)
					}
				},

				func(d benchData, b *testing.B) {
					memforge.DynamicLinearAllocatorDestroy(d.allocator)
					debug.SetGCPercent(d.oldGC)
				},
			)
		})
	}
}
