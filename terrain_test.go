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
			HeightMap:            heightMap,
			TerrainInterpolation: TerrainInterpolationLinear,
		},
		canvasWidth:  2,
		canvasHeight: 2,
		texSize:      64,
	}
	r.Cam.pos = vec3.New(32, 32, r.Cam.shooterHeight+32)
	r.Cam.subjectPos = r.Cam.pos
	return r
}

func TestHeightAtBlocksSupportsInterpolationModes(t *testing.T) {
	r := newTerrainTestRenderer([]uint8{
		0, 1,
		0, 1,
	})

	r.Wld.TerrainInterpolation = TerrainInterpolationFlat
	if got := r.heightAtBlocks(0.25, 0.5); got != 0 {
		t.Fatalf("flat height = %v, want 0", got)
	}

	r.Wld.TerrainInterpolation = TerrainInterpolationLinear
	if got := r.heightAtBlocks(0.25, 0.5); math.Abs(got-0.25) > 1e-9 {
		t.Fatalf("linear height = %v, want 0.25", got)
	}

	r.Wld.TerrainInterpolation = TerrainInterpolationSmooth
	if got := r.heightAtBlocks(0.25, 0.5); math.Abs(got-0.15625) > 1e-9 {
		t.Fatalf("smooth height = %v, want 0.15625", got)
	}
}

func TestLoadTerrainJSONV2(t *testing.T) {
	w := &World{}
	w.Init(64, 64, 2, 2, 2)

	err := w.LoadTerrainJSON([]byte(`{
		"version": 2,
		"canvasWidth": 2,
		"canvasHeight": 2,
		"canvasDepth": 2,
		"terrain": {"interpolation": "smooth"},
		"layers": [[1, 1, 1, 1], [0, 0, 2, 0]]
	}`))
	if err != nil {
		t.Fatalf("LoadTerrainJSON: %v", err)
	}
	if w.TerrainInterpolation != TerrainInterpolationSmooth {
		t.Fatalf("interpolation = %v, want smooth", w.TerrainInterpolation)
	}
	wantHeights := []uint8{1, 1, 2, 1}
	for i, want := range wantHeights {
		if w.HeightMap[i] != want {
			t.Fatalf("height[%d] = %d, want %d", i, w.HeightMap[i], want)
		}
	}
}

func TestLoadTerrainJSONRejectsUnknownInterpolation(t *testing.T) {
	w := &World{}
	w.Init(64, 64, 1, 1, 1)

	err := w.LoadTerrainJSON([]byte(`{
		"version": 2,
		"canvasWidth": 1,
		"canvasHeight": 1,
		"canvasDepth": 1,
		"terrain": {"interpolation": "roundish"},
		"layers": [[1]]
	}`))
	if err == nil {
		t.Fatal("LoadTerrainJSON accepted an unknown interpolation")
	}
}

func TestTerrainTilePrimitiveHeight(t *testing.T) {
	r := newTerrainTestRenderer([]uint8{
		1, 1,
		1, 1,
	})
	r.Wld.TerrainTileShapes = make([]uint8, 4)
	r.Wld.TerrainTileBase = make([]float32, 4)
	r.Wld.TerrainTileRise = make([]float32, 4)
	r.Wld.TerrainTileShapes[0] = uint8(TerrainTileSlopeEast)
	r.Wld.TerrainTileBase[0] = 1
	r.Wld.TerrainTileRise[0] = 1

	if got := r.heightAtBlocks(0.25, 0.5); math.Abs(got-1.25) > 1e-9 {
		t.Fatalf("east slope height = %v, want 1.25", got)
	}
	if got := r.heightAtBlocks(0.75, 0.5); math.Abs(got-1.75) > 1e-9 {
		t.Fatalf("east slope height = %v, want 1.75", got)
	}
}

func TestCollisionAllowsPrimitiveSlope(t *testing.T) {
	r := newTerrainTestRenderer([]uint8{
		1, 1,
		1, 1,
	})
	r.Wld.TerrainTileShapes = make([]uint8, 4)
	r.Wld.TerrainTileBase = make([]float32, 4)
	r.Wld.TerrainTileRise = make([]float32, 4)
	r.Wld.TerrainTileShapes[0] = uint8(TerrainTileSlopeEast)
	r.Wld.TerrainTileRise[0] = 1
	ground := r.GetGroundHeightUnderCollisionBox(vec2.New(r.Cam.subjectPos.X, r.Cam.subjectPos.Y))
	r.Cam.pos.Z = ground + r.Cam.shooterHeight
	r.Cam.subjectPos = r.Cam.pos

	got := r.collisionCheckedDelta(r.Cam.subjectPos, vec2.New(10, 0), r.Cam.collisionDistance)
	if got.X != 10 {
		t.Fatalf("primitive slope movement X = %v, want 10", got.X)
	}
}

func TestLoadTerrainJSONV3TilePrimitive(t *testing.T) {
	w := &World{}
	w.Init(64, 64, 2, 2, 2)

	err := w.LoadTerrainJSON([]byte(`{
		"version": 3,
		"canvasWidth": 2,
		"canvasHeight": 2,
		"canvasDepth": 2,
		"terrain": {
			"interpolation": "linear",
			"tiles": [
				{"x": 0, "y": 0, "shape": "slope_east", "baseHeight": 1, "rise": 1}
			]
		},
		"layers": [[1, 1, 1, 1]]
	}`))
	if err != nil {
		t.Fatalf("LoadTerrainJSON: %v", err)
	}
	if got := TerrainTileShape(w.TerrainTileShapes[0]); got != TerrainTileSlopeEast {
		t.Fatalf("shape = %v, want slope east", got)
	}
	if w.TerrainTileBase[0] != 1 || w.TerrainTileRise[0] != 1 {
		t.Fatalf("primitive data = (%v,%v), want (1,1)", w.TerrainTileBase[0], w.TerrainTileRise[0])
	}
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

func TestCollisionAllowsSmoothOneToOneSlope(t *testing.T) {
	r := newTerrainTestRenderer([]uint8{
		0, 1,
		0, 1,
	})
	r.Wld.TerrainInterpolation = TerrainInterpolationSmooth
	ground := r.GetGroundHeightUnderCollisionBox(vec2.New(r.Cam.subjectPos.X, r.Cam.subjectPos.Y))
	r.Cam.pos.Z = ground + r.Cam.shooterHeight
	r.Cam.subjectPos = r.Cam.pos

	got := r.collisionCheckedDelta(r.Cam.subjectPos, vec2.New(10, 0), r.Cam.collisionDistance)
	if got.X != 10 {
		t.Fatalf("smooth slope movement X = %v, want 10", got.X)
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
