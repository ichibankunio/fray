package fray

import (
	"reflect"
	"testing"
)

func TestGenerateTerrainPlacementsIsDeterministic(t *testing.T) {
	w := placementTestWorld()
	rule := TerrainPlacementRule{Seed: 42, CellSize: 2, Density: .8, MinHeight: 1, MaxHeight: 10, MaxSlope: 2}
	first := w.GenerateTerrainPlacements(rule)
	second := w.GenerateTerrainPlacements(rule)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("placements differ for the same seed and heightmap")
	}
	if len(first) == 0 {
		t.Fatal("expected generated placements")
	}
}

func TestGenerateTerrainPlacementsRespectsHeightAndSlope(t *testing.T) {
	w := placementTestWorld()
	rule := TerrainPlacementRule{Seed: 7, CellSize: 1, Density: 1, MinHeight: 3, MaxHeight: 5, MaxSlope: .2}
	for _, placement := range w.GenerateTerrainPlacements(rule) {
		if placement.Position.Z < rule.MinHeight || placement.Position.Z > rule.MaxHeight {
			t.Fatalf("height outside rule: %+v", placement)
		}
		slope := placement.Normal.X*placement.Normal.X + placement.Normal.Y*placement.Normal.Y
		if slope > rule.MaxSlope*rule.MaxSlope {
			t.Fatalf("slope outside rule: %+v", placement)
		}
	}
}

func placementTestWorld() *World {
	w := &World{}
	w.Init(32, 32, 8, 8, 8)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			w.HeightMap[y*8+x] = uint8(2 + x/3)
		}
	}
	w.TerrainInterpolation = TerrainInterpolationMonotonic
	return w
}
