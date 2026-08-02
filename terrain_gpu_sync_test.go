package fray

import (
	"image"
	"testing"
)

func TestWriteTerrainLookupPixels(t *testing.T) {
	w := &World{
		HeightBase:        []float32{2, 3, 4, 5},
		SlopeX:            []float32{-1, 0, .5, 1},
		SlopeY:            []float32{1, .5, 0, -1},
		TerrainVisibility: []float32{0, .25, .5, 1},
		canvasWidth:       2,
		canvasHeight:      2,
	}
	pixels := make([]byte, 8)
	if !w.writeTerrainLookupPixels(image.Rect(1, 0, 2, 2), pixels) {
		t.Fatal("encoding failed")
	}
	want := []byte{3, 128, 192, 64, 5, 255, 0, 255}
	for i := range want {
		if pixels[i] != want[i] {
			t.Fatalf("pixels[%d] = %d, want %d; pixels=%v", i, pixels[i], want[i], pixels)
		}
	}
}

func TestSyncTerrainGPUKeepsDirtyRegionOnError(t *testing.T) {
	w := &World{terrainDirty: image.Rect(1, 1, 2, 2)}
	r := &Renderer{Wld: w}
	if err := r.SyncTerrainGPU(); err == nil {
		t.Fatal("SyncTerrainGPU succeeded without texture")
	}
	if got := w.TerrainDirtyRegion(); got != image.Rect(1, 1, 2, 2) {
		t.Fatalf("dirty region after error = %v", got)
	}
}
