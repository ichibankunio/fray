package fray

import (
	"math"
	"testing"

	"github.com/ichibankunio/fvec/vec3"
)

func TestSampleTerrainMonotonicAnalyticGradient(t *testing.T) {
	w := &World{
		HeightMap: []uint8{
			0, 1, 2, 3, 4, 5,
			1, 2, 3, 4, 5, 6,
			2, 3, 4, 5, 6, 7,
			3, 4, 5, 6, 7, 8,
			4, 5, 6, 7, 8, 9,
			5, 6, 7, 8, 9, 10,
		},
		TerrainInterpolation: TerrainInterpolationMonotonic,
		canvasWidth:          6,
		canvasHeight:         6,
	}
	sample := w.SampleTerrain(2.25, 2.4)
	if math.Abs(sample.Position.Z-4.65) > 1e-9 {
		t.Fatalf("height = %.12f, want 4.65", sample.Position.Z)
	}
	if math.Abs(sample.Gradient.X-1) > 1e-9 || math.Abs(sample.Gradient.Y-1) > 1e-9 {
		t.Fatalf("gradient = (%v, %v), want (1, 1)", sample.Gradient.X, sample.Gradient.Y)
	}
	wantNormal := 1 / math.Sqrt(3)
	if math.Abs(sample.Normal.X+wantNormal) > 1e-9 || math.Abs(sample.Normal.Y+wantNormal) > 1e-9 || math.Abs(sample.Normal.Z-wantNormal) > 1e-9 {
		t.Fatalf("normal = %+v", sample.Normal)
	}
}

func TestQueryTerrainContact(t *testing.T) {
	w := flatQueryWorld(2)
	contact := w.QueryTerrainContact(vec3.New(1, 1, 2.08), .1)
	if !contact.Grounded || math.Abs(contact.Distance-.08) > 1e-9 {
		t.Fatalf("contact = %+v", contact)
	}
}

func TestRaycastTerrainDownward(t *testing.T) {
	w := flatQueryWorld(2)
	hit, ok := w.RaycastTerrain(vec3.New(1, 1, 5), vec3.New(0, 0, -2), 10, DefaultTerrainRaycastConfig())
	if !ok {
		t.Fatal("downward ray missed terrain")
	}
	if math.Abs(hit.Distance-3) > 0.001 || math.Abs(hit.Position.Z-2) > 1e-9 {
		t.Fatalf("hit = %+v", hit)
	}
}

func TestRaycastTerrainMissesUpward(t *testing.T) {
	w := flatQueryWorld(2)
	if _, ok := w.RaycastTerrain(vec3.New(1, 1, 3), vec3.New(0, 0, 1), 5, DefaultTerrainRaycastConfig()); ok {
		t.Fatal("upward ray unexpectedly hit terrain")
	}
}

func flatQueryWorld(height uint8) *World {
	return &World{
		HeightMap:            []uint8{height, height, height, height, height, height, height, height, height},
		TerrainInterpolation: TerrainInterpolationMonotonic,
		canvasWidth:          3,
		canvasHeight:         3,
	}
}
