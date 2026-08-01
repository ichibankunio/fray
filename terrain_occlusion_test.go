package fray

import (
	"math"
	"testing"
)

func TestTerrainVisibleHeightClipsAgainstRidge(t *testing.T) {
	w := &World{HeightMap: []uint8{0, 2, 0}, TerrainInterpolation: TerrainInterpolationLinear, canvasWidth: 3, canvasHeight: 1}
	got := w.TerrainVisibleHeight(0, 0, 1, 2, 0, TerrainOcclusionConfig{SampleSpacing: 1, Clearance: 0})
	if math.Abs(got-3) > 1e-9 {
		t.Fatalf("visible height = %v, want 3", got)
	}
}

func TestTerrainLineOfSightUsesVisibleHeight(t *testing.T) {
	w := &World{HeightMap: []uint8{0, 2, 0}, TerrainInterpolation: TerrainInterpolationLinear, canvasWidth: 3, canvasHeight: 1}
	config := TerrainOcclusionConfig{SampleSpacing: 1, Clearance: 0}
	if w.TerrainLineOfSight(0, 0, 1, 2, 0, 2.9, config) {
		t.Fatal("point below ridge horizon is visible")
	}
	if !w.TerrainLineOfSight(0, 0, 1, 2, 0, 3, config) {
		t.Fatal("point on ridge horizon is hidden")
	}
}
