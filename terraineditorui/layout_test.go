package terraineditorui

import (
	"image"
	"testing"
)

func TestSplitKeepsGameAndEditorInsideWindow(t *testing.T) {
	layout := Split(image.Rect(0, 0, 960, 540), 288)
	if layout.Game != image.Rect(0, 0, 672, 540) || layout.Editor != image.Rect(672, 0, 960, 540) {
		t.Fatalf("layout = %+v", layout)
	}
	if layout.Game.Intersect(layout.Editor).Dx() != 0 {
		t.Fatalf("layout overlaps: %+v", layout)
	}
}
