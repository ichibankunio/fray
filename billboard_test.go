package fray

import (
	"image/color"
	"testing"
)

func TestNormalizedBillboardConfigRestoresInvalidValues(t *testing.T) {
	got := normalizedBillboardConfig(BillboardConfig{NearDistance: -1, FarDistance: -2})
	if got.NearDistance != 0 || got.FarDistance <= got.NearDistance || got.ProjectionScale <= 0 {
		t.Fatalf("invalid normalized config: %+v", got)
	}
}

func TestMultiplyBillboardColor(t *testing.T) {
	got := multiplyBillboardColor(rgba(200, 100, 50, 255), rgba(128, 255, 128, 128))
	if got.R != 100 || got.G != 100 || got.B != 25 || got.A != 128 {
		t.Fatalf("multiplied color = %+v", got)
	}
}

func rgba(r, g, b, a uint8) color.RGBA { return color.RGBA{r, g, b, a} }
