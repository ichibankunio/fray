package fray

import (
	"math"
	"testing"

	"github.com/ichibankunio/fvec/vec2"
	"github.com/ichibankunio/fvec/vec3"
)

func newTerrainTestRenderer(heightMap []uint8) *Renderer {
	r := &Renderer{
		Cam: &Camera{
			shooterHeight:     96,
			collisionDistance: 0.25,
		},
		Wld: &World{
			HeightMap: heightMap,
		},
		canvasWidth:  2,
		canvasHeight: 2,
		texSize:      64,
	}
	r.Cam.pos = vec3.New(32, 32, r.Cam.shooterHeight+32)
	r.Cam.subjectPos = r.Cam.pos
	return r
}

func TestHeightAtBlocksUsesBilinearInterpolation(t *testing.T) {
	r := newTerrainTestRenderer([]uint8{
		0, 1,
		1, 2,
	})

	got := r.heightAtBlocks(0.5, 0.5)
	if math.Abs(got-1.0) > 1e-9 {
		t.Fatalf("heightAtBlocks(0.5, 0.5) = %v, want 1", got)
	}
}

func TestGroundedCameraFollowsInterpolatedSlope(t *testing.T) {
	r := newTerrainTestRenderer([]uint8{
		0, 1,
		0, 1,
	})

	r.moveCameraOnTerrain(r.Cam.subjectPos, vec2.New(10, 0))

	wantZ := r.GetGroundHeightUnderCollisionBox(vec2.New(42, 32)) + r.Cam.shooterHeight
	if math.Abs(r.Cam.subjectPos.Z-wantZ) > 1e-9 {
		t.Fatalf("camera Z = %v, want %v", r.Cam.subjectPos.Z, wantZ)
	}
}

func TestCollisionAllowsOneToOneSlope(t *testing.T) {
	r := newTerrainTestRenderer([]uint8{
		0, 1,
		0, 1,
	})

	got := r.collisionCheckedDelta(r.Cam.subjectPos, vec2.New(10, 0), 0.25)
	if got.X != 10 {
		t.Fatalf("slope movement X = %v, want 10", got.X)
	}
}

func TestCollisionRejectsSharperRise(t *testing.T) {
	r := newTerrainTestRenderer([]uint8{
		0, 3,
		0, 3,
	})
	r.Cam.pos.Z = r.Cam.shooterHeight + 96
	r.Cam.subjectPos = r.Cam.pos

	got := r.collisionCheckedDelta(r.Cam.subjectPos, vec2.New(10, 0), 0.25)
	if got.X != 0 {
		t.Fatalf("steep movement X = %v, want 0", got.X)
	}
}

func TestAirborneCameraDoesNotSnapToSlope(t *testing.T) {
	r := newTerrainTestRenderer([]uint8{
		0, 1,
		0, 1,
	})
	r.Cam.pos.Z += 20
	r.Cam.subjectPos.Z += 20
	beforeZ := r.Cam.subjectPos.Z

	r.moveCameraOnTerrain(r.Cam.subjectPos, vec2.New(10, 0))

	if r.Cam.subjectPos.Z != beforeZ {
		t.Fatalf("airborne camera Z = %v, want unchanged %v", r.Cam.subjectPos.Z, beforeZ)
	}
}
