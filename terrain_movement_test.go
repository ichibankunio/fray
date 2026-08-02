package fray

import (
	"math"
	"testing"

	"github.com/ichibankunio/fvec/vec3"
)

func TestMoveOnTerrainMovesGroundedCapsuleAcrossFlatSurface(t *testing.T) {
	w := flatQueryWorld(2)
	config := DefaultTerrainMovementConfig()
	origin := vec3.New(1, 1, 2+config.HalfSegment+config.Radius)
	result := w.MoveOnTerrain(origin, vec3.New(.4, 0, 0), config)
	if result.Blocked || !result.Grounded || math.Abs(result.Position.X-1.4) > 1e-9 || math.Abs(result.Position.Z-origin.Z) > 1e-9 {
		t.Fatalf("result = %+v", result)
	}
}

func TestMoveOnTerrainRejectsSlopeAboveLimit(t *testing.T) {
	w := &World{HeightMap: []uint8{0, 1, 0, 1}, TerrainInterpolation: TerrainInterpolationLinear, canvasWidth: 2, canvasHeight: 2}
	config := DefaultTerrainMovementConfig()
	config.MaxSlope = .5
	surface := w.SampleTerrain(.25, .5)
	origin := vec3.New(.25, .5, surface.Position.Z+config.HalfSegment+config.Radius/surface.Normal.Z)
	result := w.MoveOnTerrain(origin, vec3.New(.4, 0, 0), config)
	if !result.Blocked || math.Abs(result.Position.X-origin.X) > 1e-9 {
		t.Fatalf("result = %+v", result)
	}
}

func TestMoveOnTerrainStopsFastFallAtGround(t *testing.T) {
	w := flatQueryWorld(2)
	config := DefaultTerrainMovementConfig()
	result := w.MoveOnTerrain(vec3.New(1, 1, 8), vec3.New(0, 0, -12), config)
	wantZ := 2 + config.HalfSegment + config.Radius
	if !result.Blocked || !result.Grounded || math.Abs(result.Position.Z-wantZ) > .01 {
		t.Fatalf("result = %+v, want Z %.3f", result, wantZ)
	}
}

func TestMoveOnTerrainHasNoAllocations(t *testing.T) {
	w := flatQueryWorld(2)
	config := DefaultTerrainMovementConfig()
	origin := vec3.New(1, 1, 2+config.HalfSegment+config.Radius)
	if allocations := testing.AllocsPerRun(100, func() { _ = w.MoveOnTerrain(origin, vec3.New(.1, 0, 0), config) }); allocations != 0 {
		t.Fatalf("MoveOnTerrain allocations = %v", allocations)
	}
}
