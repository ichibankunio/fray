package fray

import "testing"

func TestBuildTerrainVisibilityDarkensEnclosedValley(t *testing.T) {
	w := &World{}
	w.Init(32, 32, 9, 9, 16)
	for y := 0; y < 9; y++ {
		for x := 0; x < 9; x++ {
			w.HeightMap[y*9+x] = 2
		}
	}
	for y := 2; y <= 6; y++ {
		for x := 2; x <= 6; x++ {
			if x == 2 || x == 6 || y == 2 || y == 6 {
				w.HeightMap[y*9+x] = 10
			}
		}
	}
	w.BuildTerrainVisibility()

	valley := w.TerrainVisibility[4*9+4]
	open := w.TerrainVisibility[8*9+8]
	if valley >= open {
		t.Fatalf("valley visibility %v must be lower than open terrain %v", valley, open)
	}
}

func TestBuildTerrainVisibilityIsNormalized(t *testing.T) {
	w := &World{}
	w.Init(32, 32, 5, 5, 8)
	w.BuildTerrainVisibility()
	for i, visibility := range w.TerrainVisibility {
		if visibility < 0 || visibility > 1 {
			t.Fatalf("visibility[%d] = %v, want [0, 1]", i, visibility)
		}
	}
}
