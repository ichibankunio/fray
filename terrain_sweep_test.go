package fray

import (
	"math"
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
