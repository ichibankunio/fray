package fray

import (
	"image"
	"math"
	"math/rand"
	"testing"

	"github.com/ichibankunio/fvec/vec3"
)

func TestSweepTerrainSpherePreventsVerticalTunneling(t *testing.T) {
	w := flatQueryWorld(2)
	hit, ok := w.SweepTerrainSphere(vec3.New(1, 1, 8), vec3.New(0, 0, -12), .5, DefaultTerrainRaycastConfig())
	if !ok {
		t.Fatal("fast sphere sweep missed terrain")
	}
	if math.Abs(hit.Position.Z-2.5) > .001 || math.Abs(hit.Fraction-5.5/12) > .001 {
		t.Fatalf("hit = %+v", hit)
	}
}

func TestAdaptiveSweepMatchesFixedSweep(t *testing.T) {
	w := benchmarkTerrainWorld(32)
	w.expandTerrainGradientBound(image.Rect(0, 0, 32, 32))
	adaptive := DefaultTerrainRaycastConfig()
	adaptive.Step = .02
	fixed := adaptive
	fixed.MaxStep = fixed.Step
	random := rand.New(rand.NewSource(42))
	for i := 0; i < 100; i++ {
		origin := vec3.New(8+random.Float64()*16, 8+random.Float64()*16, 30)
		movement := vec3.New(random.Float64()*6-3, random.Float64()*6-3, -35)
		want, wantOK := w.SweepTerrainSphere(origin, movement, .35, fixed)
		got, gotOK := w.SweepTerrainSphere(origin, movement, .35, adaptive)
		if gotOK != wantOK || gotOK && math.Abs(got.Distance-want.Distance) > .001 {
			t.Fatalf("case %d: adaptive=(%+v,%t), fixed=(%+v,%t)", i, got, gotOK, want, wantOK)
		}
	}
}

func TestSweepTerrainCapsuleUsesLowerSphere(t *testing.T) {
	w := flatQueryWorld(2)
	hit, ok := w.SweepTerrainCapsule(vec3.New(1, 1, 8), vec3.New(0, 0, -12), .5, 1.5, DefaultTerrainRaycastConfig())
	if !ok || math.Abs(hit.Position.Z-4) > .001 {
		t.Fatalf("hit = %+v, ok=%t", hit, ok)
	}
}

func TestSweepTerrainSphereMissesMovementAboveGround(t *testing.T) {
	w := flatQueryWorld(2)
	if _, ok := w.SweepTerrainSphere(vec3.New(.5, .5, 5), vec3.New(1, 1, 0), .5, DefaultTerrainRaycastConfig()); ok {
		t.Fatal("horizontal sphere sweep above ground reported a hit")
	}
}
