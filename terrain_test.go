package fray

import (
	"math"
	"testing"

	"github.com/ichibankunio/fvec/vec2"
	"github.com/ichibankunio/fvec/vec3"
)

func newTerrainTestRenderer(heightMap []uint8) *Renderer {
	r := &Renderer{
		Cam: &Camera{shooterHeight: 96, collisionDistance: 0.25},
		Wld: &World{
			HeightMap:            heightMap,
			TerrainInterpolation: TerrainInterpolationLinear,
			canvasWidth:          2,
			canvasHeight:         2,
		},
		canvasWidth: 2, canvasHeight: 2, texSize: 64,
	}
	r.Cam.pos = vec3.New(32, 32, r.Cam.shooterHeight+32)
	r.Cam.subjectPos = r.Cam.pos
	return r
}

func TestHeightAtBlocksSupportsInterpolationModes(t *testing.T) {
	r := newTerrainTestRenderer([]uint8{0, 1, 0, 1})

	r.Wld.TerrainInterpolation = TerrainInterpolationFlat
	if got := r.heightAtBlocks(0.25, 0.5); got != 0 {
		t.Fatalf("flat height = %v, want 0", got)
	}
	r.Wld.TerrainInterpolation = TerrainInterpolationLinear
	if got := r.heightAtBlocks(0.25, 0.5); math.Abs(got-0.25) > 1e-9 {
		t.Fatalf("linear height = %v, want 0.25", got)
	}
	r.Wld.TerrainInterpolation = TerrainInterpolationMonotonic
	if got := r.heightAtBlocks(0.25, 0.5); math.Abs(got-0.15625) > 1e-9 {
		t.Fatalf("monotonic height = %v, want 0.15625", got)
	}
}

func TestMonotonicCubicDoesNotOvershoot(t *testing.T) {
	for i := 0; i <= 100; i++ {
		got := monotonicCubic(0, 0, 3, 3, float64(i)/100)
		if got < 0 || got > 3 {
			t.Fatalf("sample %d overshot: %v", i, got)
		}
	}
}

func TestMonotonicCubicPreservesLinearGradient(t *testing.T) {
	for i := 0; i <= 10; i++ {
		x := float64(i) / 10
		if got := monotonicCubic(0, 1, 2, 3, x); math.Abs(got-(1+x)) > 1e-9 {
			t.Fatalf("sample %.2f = %v, want %v", x, got, 1+x)
		}
	}
}

func TestMonotonicTerrainHasContinuousCellBoundaries(t *testing.T) {
	w := &World{
		HeightMap: []uint8{
			0, 0, 1, 2, 2,
			0, 1, 2, 3, 3,
			1, 2, 3, 3, 4,
			1, 2, 2, 3, 4,
		},
		TerrainInterpolation: TerrainInterpolationMonotonic,
		canvasWidth:          5,
		canvasHeight:         4,
	}
	const epsilon = 1e-5
	left := w.heightAtBlocks(2-epsilon, 1.4)
	right := w.heightAtBlocks(2+epsilon, 1.4)
	if math.Abs(left-right) > 1e-3 {
		t.Fatalf("height jumps at cell boundary: left=%v right=%v", left, right)
	}
	leftSlope := (w.heightAtBlocks(2, 1.4) - w.heightAtBlocks(2-epsilon, 1.4)) / epsilon
	rightSlope := (w.heightAtBlocks(2+epsilon, 1.4) - w.heightAtBlocks(2, 1.4)) / epsilon
	if math.Abs(leftSlope-rightSlope) > 1e-3 {
		t.Fatalf("slope jumps at cell boundary: left=%v right=%v", leftSlope, rightSlope)
	}
}

func TestLoadTerrainJSONMonotonic(t *testing.T) {
	w := &World{}
	w.Init(64, 64, 2, 2, 2)
	err := w.LoadTerrainJSON([]byte(`{
		"version": 3,
		"canvasWidth": 2,
		"canvasHeight": 2,
		"canvasDepth": 2,
		"terrain": {"interpolation": "monotonic"},
		"layers": [[1, 1, 1, 1], [0, 0, 2, 0]]
	}`))
	if err != nil {
		t.Fatalf("LoadTerrainJSON: %v", err)
	}
	if w.TerrainInterpolation != TerrainInterpolationMonotonic {
		t.Fatalf("interpolation = %v, want monotonic", w.TerrainInterpolation)
	}
	wantHeights := []uint8{1, 1, 2, 1}
	for i, want := range wantHeights {
		if w.HeightMap[i] != want {
			t.Fatalf("height[%d] = %d, want %d", i, w.HeightMap[i], want)
		}
	}
}

func TestLoadTerrainJSONWaterLevel(t *testing.T) {
	w := &World{}
	w.Init(64, 64, 1, 1, 8)
	err := w.LoadTerrainJSON([]byte(`{"version":3,"canvasWidth":1,"canvasHeight":1,"canvasDepth":8,"terrain":{"interpolation":"monotonic","waterLevel":2.5},"layers":[[1]]}`))
	if err != nil {
		t.Fatalf("LoadTerrainJSON: %v", err)
	}
	if !w.HasWater || w.WaterLevel != 2.5 {
		t.Fatalf("water = %v at %v", w.HasWater, w.WaterLevel)
	}
}

func TestLoadTerrainJSONRejectsWaterOutsideWorld(t *testing.T) {
	w := &World{}
	w.Init(64, 64, 1, 1, 4)
	if err := w.LoadTerrainJSON([]byte(`{"version":3,"terrain":{"waterLevel":5},"layers":[[1]]}`)); err == nil {
		t.Fatal("LoadTerrainJSON accepted water above world")
	}
}

func TestLoadTerrainJSONRejectsRemovedTileSchema(t *testing.T) {
	w := &World{}
	w.Init(64, 64, 1, 1, 1)
	err := w.LoadTerrainJSON([]byte(`{
		"version": 3,
		"terrain": {"interpolation": "monotonic", "tiles": []},
		"layers": [[1]]
	}`))
	if err == nil {
		t.Fatal("LoadTerrainJSON accepted removed terrain.tiles")
	}
}

func TestLoadTerrainJSONRejectsUnknownInterpolation(t *testing.T) {
	w := &World{}
	w.Init(64, 64, 1, 1, 1)
	err := w.LoadTerrainJSON([]byte(`{
		"version": 3,
		"terrain": {"interpolation": "roundish"},
		"layers": [[1]]
	}`))
	if err == nil {
		t.Fatal("LoadTerrainJSON accepted an unknown interpolation")
	}
}

func TestGroundedCameraFollowsInterpolatedSlope(t *testing.T) {
	r := newTerrainTestRenderer([]uint8{0, 1, 0, 1})
	r.Wld.TerrainInterpolation = TerrainInterpolationMonotonic
	r.moveCameraOnTerrain(r.Cam.subjectPos, vec2.New(10, 0))
	wantZ := r.GetGroundHeightUnderCollisionBox(vec2.New(42, 32)) + r.Cam.shooterHeight
	if math.Abs(r.Cam.subjectPos.Z-wantZ) > 1e-9 {
		t.Fatalf("camera Z = %v, want %v", r.Cam.subjectPos.Z, wantZ)
	}
}

func TestCollisionAllowsMonotonicSlope(t *testing.T) {
	r := newTerrainTestRenderer([]uint8{0, 1, 0, 1})
	r.Wld.TerrainInterpolation = TerrainInterpolationMonotonic
	ground := r.GetGroundHeightUnderCollisionBox(vec2.New(r.Cam.subjectPos.X, r.Cam.subjectPos.Y))
	r.Cam.pos.Z = ground + r.Cam.shooterHeight
	r.Cam.subjectPos = r.Cam.pos
	got := r.collisionCheckedDelta(r.Cam.subjectPos, vec2.New(10, 0), r.Cam.collisionDistance)
	if got.X != 10 {
		t.Fatalf("slope movement X = %v, want 10", got.X)
	}
}

func TestCollisionRejectsSharperRise(t *testing.T) {
	r := newTerrainTestRenderer([]uint8{0, 3, 0, 3})
	r.Wld.TerrainInterpolation = TerrainInterpolationMonotonic
	r.Cam.pos.Z = r.Cam.shooterHeight + 96
	r.Cam.subjectPos = r.Cam.pos
	got := r.collisionCheckedDelta(r.Cam.subjectPos, vec2.New(10, 0), 0.25)
	if got.X != 0 {
		t.Fatalf("steep movement X = %v, want 0", got.X)
	}
}

func TestAirborneCameraDoesNotSnapToSlope(t *testing.T) {
	r := newTerrainTestRenderer([]uint8{0, 1, 0, 1})
	r.Wld.TerrainInterpolation = TerrainInterpolationMonotonic
	r.Cam.pos.Z += 20
	r.Cam.subjectPos.Z += 20
	beforeZ := r.Cam.subjectPos.Z
	r.moveCameraOnTerrain(r.Cam.subjectPos, vec2.New(10, 0))
	if r.Cam.subjectPos.Z != beforeZ {
		t.Fatalf("airborne camera Z = %v, want unchanged %v", r.Cam.subjectPos.Z, beforeZ)
	}
}

func TestTerrainSamplingAPIUsesSelectedInterpolation(t *testing.T) {
	w := &World{}
	w.Init(32, 32, 3, 3, 8)
	w.HeightMap = []uint8{
		1, 3, 3,
		1, 3, 3,
		1, 3, 3,
	}
	w.TerrainInterpolation = TerrainInterpolationLinear
	if got := w.SampleTerrainHeight(.5, .5); got != 2 {
		t.Fatalf("height = %v, want 2", got)
	}
	normal := w.SampleTerrainNormal(.5, .5)
	if normal.Z <= 0 || w.SampleTerrainSlope(.5, .5) <= 0 {
		t.Fatalf("unexpected normal/slope: %+v, %v", normal, w.SampleTerrainSlope(.5, .5))
	}
}

func TestTerrainRenderScaleSupportsLowResolutionRendering(t *testing.T) {
	r := &Renderer{screenWidth: 540, screenHeight: 960}
	r.SetTerrainRenderScale(.20)
	width, height := r.terrainRenderSize()
	if width != 108 || height != 192 {
		t.Fatalf("render size = %dx%d, want 108x192", width, height)
	}
	r.SetTerrainRenderScale(.01)
	width, height = r.terrainRenderSize()
	if width != 43 || height != 76 {
		t.Fatalf("clamped render size = %dx%d, want 43x76", width, height)
	}
}

func TestTerrainDebugModeRejectsUnknownMode(t *testing.T) {
	r := &Renderer{}
	r.SetTerrainDebugMode(TerrainDebugSlope)
	if r.TerrainDebugMode() != TerrainDebugSlope {
		t.Fatal("debug mode was not retained")
	}
	r.SetTerrainDebugMode(TerrainDebugMode(99))
	if r.TerrainDebugMode() != TerrainDebugOff {
		t.Fatal("unknown debug mode was not disabled")
	}
}
