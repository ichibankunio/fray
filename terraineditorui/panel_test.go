package terraineditorui

import (
	"image"
	"testing"
)

func TestInclusiveRectNormalizesDragDirection(t *testing.T) {
	want := image.Rect(2, 3, 8, 10)
	for _, points := range [][2]image.Point{{image.Pt(2, 3), image.Pt(7, 9)}, {image.Pt(7, 9), image.Pt(2, 3)}} {
		if got := inclusiveRect(points[0], points[1]); got != want {
			t.Fatalf("inclusiveRect = %v, want %v", got, want)
		}
	}
}
