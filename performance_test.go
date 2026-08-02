package fray

import (
	"testing"

	"github.com/ichibankunio/fvec/vec3"
)

func TestTerrainHotPathAllocations(t *testing.T) {
	w := benchmarkTerrainWorld(32)
	if allocations := testing.AllocsPerRun(1000, func() { _ = w.SampleTerrain(12.25, 18.75) }); allocations != 0 {
		t.Fatalf("SampleTerrain allocations = %v, want 0", allocations)
	}
	config := DefaultTerrainRaycastConfig()
	if allocations := testing.AllocsPerRun(100, func() {
		_, _ = w.RaycastTerrain(vec3.New(12, 12, 31), vec3.New(.2, .1, -1), 40, config)
	}); allocations != 0 {
		t.Fatalf("RaycastTerrain allocations = %v, want 0", allocations)
	}
	if allocations := testing.AllocsPerRun(100, func() {
		_, _ = w.SweepTerrainSphere(vec3.New(12, 12, 31), vec3.New(.2, .1, -20), .4, config)
	}); allocations != 0 {
		t.Fatalf("SweepTerrainSphere allocations = %v, want 0", allocations)
	}
}

func BenchmarkSampleTerrainMonotonic(b *testing.B) {
	w := benchmarkTerrainWorld(128)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		x := 2 + float64(i%123) + .37
		y := 2 + float64((i*17)%123) + .61
		_ = w.SampleTerrain(x, y)
	}
}

func BenchmarkRaycastTerrain(b *testing.B) {
	w := benchmarkTerrainWorld(128)
	config := DefaultTerrainRaycastConfig()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = w.RaycastTerrain(vec3.New(64, 64, 31), vec3.New(.15, .05, -1), 40, config)
	}
}

func BenchmarkTerrainLineOfSight(b *testing.B) {
	w := benchmarkTerrainWorld(128)
	config := DefaultTerrainOcclusionConfig()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = w.TerrainLineOfSight(4, 4, 20, 115, 115, 20, config)
	}
}

func BenchmarkSweepTerrainSphereAdaptive(b *testing.B) {
	w := flatQueryWorld(2)
	config := DefaultTerrainRaycastConfig()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = w.SweepTerrainSphere(vec3.New(1, 1, 30), vec3.New(0, 0, -40), .4, config)
	}
}

func BenchmarkSweepTerrainSphereFixed(b *testing.B) {
	w := flatQueryWorld(2)
	config := DefaultTerrainRaycastConfig()
	config.MaxStep = config.Step
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = w.SweepTerrainSphere(vec3.New(1, 1, 30), vec3.New(0, 0, -40), .4, config)
	}
}
