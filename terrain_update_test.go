package fray

import (
	"image"
	"testing"
)

func TestRebuildTerrainRegionUpdatesDerivedDataAndRevision(t *testing.T) {
	w := &World{}
	w.Init(16, 16, 16, 16, 4)
	w.WorldMap[0][8*16+8] = 1
	dirty := w.RebuildTerrainRegion(image.Rect(8, 8, 9, 9))
	if got := w.GetHeight(8, 8); got != 1 {
		t.Fatalf("height = %d, want 1", got)
	}
	if dirty != image.Rect(0, 0, 16, 16) {
		t.Fatalf("dirty region = %v", dirty)
	}
	if w.TerrainRevision() != 1 {
		t.Fatalf("revision = %d, want 1", w.TerrainRevision())
	}
	if got := w.ConsumeTerrainDirtyRegion(); got != dirty {
		t.Fatalf("consumed dirty region = %v, want %v", got, dirty)
	}
	if got := w.ConsumeTerrainDirtyRegion(); !got.Empty() {
		t.Fatalf("second consumed region = %v, want empty", got)
	}
}

func TestRebuildTerrainRegionMergesPendingRegions(t *testing.T) {
	w := &World{}
	w.Init(32, 32, 32, 32, 2)
	w.RebuildTerrainRegion(image.Rect(1, 1, 2, 2))
	w.RebuildTerrainRegion(image.Rect(30, 30, 31, 31))
	if got := w.ConsumeTerrainDirtyRegion(); got != image.Rect(0, 0, 32, 32) {
		t.Fatalf("merged dirty region = %v", got)
	}
}

func TestSetAndDeleteValueRefreshTerrain(t *testing.T) {
	w := &World{}
	w.Init(8, 8, 8, 8, 3)
	w.SetValue(4, 4, 1, 0)
	if got := w.GetHeight(4, 4); got != 1 {
		t.Fatalf("height after SetValue = %d, want 1", got)
	}
	w.DeleteValue(4, 4, 1)
	if got := w.GetHeight(4, 4); got != 0 {
		t.Fatalf("height after DeleteValue = %d, want 0", got)
	}
}
