package fray

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNormalizedBillboardConfigRestoresInvalidValues(t *testing.T) {
	got := normalizedBillboardConfig(BillboardConfig{NearDistance: -1, FarDistance: -2})
	if got.NearDistance != 0 || got.FarDistance <= got.NearDistance || got.ProjectionScale <= 0 {
		t.Fatalf("invalid normalized config: %+v", got)
	}
}

func TestClipBillboardTriangleClipsHiddenBottom(t *testing.T) {
	got := clipBillboardTriangle([3]ebiten.Vertex{{DstX: 0, DstY: 0}, {DstX: 2, DstY: 2}, {DstX: 0, DstY: 2}}, 1)
	if len(got) != 3 {
		t.Fatalf("vertices = %d, want 3", len(got))
	}
	for _, vertex := range got {
		if vertex.DstY > 1 {
			t.Fatalf("unclipped vertex: %+v", vertex)
		}
	}
}

func TestBillboardPixelHeightShrinksWithDistance(t *testing.T) {
	config := DefaultBillboardConfig()
	config.MinPixelHeight = .5
	near := billboardPixelHeight(1, 2, 960, config)
	far := billboardPixelHeight(1, 12, 960, config)
	if far >= near || far != 80 {
		t.Fatalf("near=%v far=%v, want distance scaling", near, far)
	}
}

func TestMultiplyBillboardColor(t *testing.T) {
	got := multiplyBillboardColor(rgba(200, 100, 50, 255), rgba(128, 255, 128, 128))
	if got.R != 100 || got.G != 100 || got.B != 25 || got.A != 128 {
		t.Fatalf("multiplied color = %+v", got)
	}
}

func rgba(r, g, b, a uint8) color.RGBA { return color.RGBA{r, g, b, a} }
