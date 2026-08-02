package visualregression

import (
	"image"
	"image/color"
	"testing"
)

func TestCompare(t *testing.T) {
	baseline := image.NewRGBA(image.Rect(0, 0, 10, 10))
	actual := image.NewRGBA(image.Rect(0, 0, 10, 10))
	baseline.Set(5, 5, color.RGBA{100, 100, 100, 255})
	actual.Set(5, 5, color.RGBA{104, 96, 102, 255})
	if result := Compare(baseline, actual, DefaultOptions()); !result.Passed {
		t.Fatalf("minor rounding difference failed: %+v", result)
	}
	actual.Set(5, 5, color.RGBA{255, 255, 255, 255})
	options := DefaultOptions()
	options.MaxChangedRatio = 0
	if result := Compare(baseline, actual, options); result.Passed {
		t.Fatalf("visible difference passed: %+v", result)
	}
}

func TestCompareRejectsDifferentBounds(t *testing.T) {
	if result := Compare(image.NewRGBA(image.Rect(0, 0, 1, 1)), image.NewRGBA(image.Rect(0, 0, 2, 1)), DefaultOptions()); result.Passed {
		t.Fatal("different bounds passed")
	}
}
