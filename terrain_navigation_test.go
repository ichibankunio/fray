package fray

import "testing"

func TestTerrainNavigationFindsPathAcrossFlatTerrain(t *testing.T) {
	w := &World{}
	w.Init(32, 32, 8, 8, 8)
	for i := range w.HeightMap {
		w.HeightMap[i] = 2
	}
	w.TerrainInterpolation = TerrainInterpolationMonotonic
	navigation := w.BuildTerrainNavigation(TerrainNavigationConfig{AllowDiagonal: true})
	path, ok := navigation.FindPath(1, 1, 6, 6)
	if !ok || len(path) < 2 {
		t.Fatalf("expected path, got %v, %v", path, ok)
	}
	if path[0].X != 1.5 || path[len(path)-1].X != 6.5 {
		t.Fatalf("unexpected endpoints: %v -> %v", path[0], path[len(path)-1])
	}
}

func TestTerrainNavigationRejectsUnwalkableGoal(t *testing.T) {
	w := &World{}
	w.Init(32, 32, 5, 5, 8)
	for i := range w.HeightMap {
		w.HeightMap[i] = 2
	}
	w.HeightMap[2*5+2] = 7
	w.TerrainInterpolation = TerrainInterpolationFlat
	navigation := w.BuildTerrainNavigation(TerrainNavigationConfig{MaxSlope: .2, MaxStep: 1})
	if _, ok := navigation.FindPath(0, 0, 2, 2); ok {
		t.Fatal("expected raised goal to be unwalkable")
	}
}
