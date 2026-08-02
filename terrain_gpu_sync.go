package fray

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// SyncTerrainGPU uploads only the dirty heightmap-derived lookup region.
// It must be called from the Ebitengine main thread.
func (r *Renderer) SyncTerrainGPU() error {
	if r.Wld == nil {
		return fmt.Errorf("sync terrain GPU: renderer has no world")
	}
	region := r.Wld.TerrainDirtyRegion()
	if region.Empty() {
		return nil
	}
	if r.Textures[3] == nil || r.Textures[3].Src == nil {
		return fmt.Errorf("sync terrain GPU: terrain lookup texture is not configured")
	}
	bounds := r.Textures[3].Src.Bounds()
	if !region.In(bounds) {
		return fmt.Errorf("sync terrain GPU: dirty region %v exceeds texture bounds %v", region, bounds)
	}
	required := region.Dx() * region.Dy() * 4
	if cap(r.terrainUploadBuffer) < required {
		r.terrainUploadBuffer = make([]byte, required)
	} else {
		r.terrainUploadBuffer = r.terrainUploadBuffer[:required]
	}
	if !r.Wld.writeTerrainLookupPixels(region, r.terrainUploadBuffer) {
		return fmt.Errorf("sync terrain GPU: could not encode region %v", region)
	}
	target := r.Textures[3].Src.SubImage(region).(*ebiten.Image)
	target.WritePixels(r.terrainUploadBuffer)
	r.Wld.ConsumeTerrainDirtyRegion()
	return nil
}
