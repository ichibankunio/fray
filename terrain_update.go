package fray

import "image"

const terrainVisibilityRadius = 10

// RebuildTerrainRegion refreshes heightmap-derived CPU data after WorldMap was
// changed. WorldMap remains authoritative. The returned rectangle is the area
// whose GPU lookup data must be uploaded again.
func (w *World) RebuildTerrainRegion(region image.Rectangle) image.Rectangle {
	bounds := image.Rect(0, 0, w.canvasWidth, w.canvasHeight)
	region = region.Intersect(bounds)
	if region.Empty() {
		return image.Rectangle{}
	}
	for y := region.Min.Y; y < region.Max.Y; y++ {
		for x := region.Min.X; x < region.Max.X; x++ {
			index := y*w.canvasWidth + x
			height := 0
			for z := w.canvasDepth - 1; z >= 0; z-- {
				if w.WorldMap[z][index] != 0 {
					height = z + 1
					break
				}
			}
			w.HeightMap[index] = uint8(height)
		}
	}

	slopeRegion := expandTerrainRegion(region, 1).Intersect(bounds)
	visibilityRegion := expandTerrainRegion(region, terrainVisibilityRadius).Intersect(bounds)
	w.syncHeightPlanesRegion(slopeRegion)
	w.buildTerrainVisibilityRegion(visibilityRegion)
	w.terrainRevision++
	w.terrainDirty = unionTerrainRegion(w.terrainDirty, visibilityRegion)
	return visibilityRegion
}

// RebuildTerrain refreshes all data inferred from WorldMap.
func (w *World) RebuildTerrain() {
	w.RebuildTerrainRegion(image.Rect(0, 0, w.canvasWidth, w.canvasHeight))
}

// TerrainRevision increases whenever inferred terrain data is rebuilt.
func (w *World) TerrainRevision() uint64 { return w.terrainRevision }

// ConsumeTerrainDirtyRegion returns the region requiring GPU synchronization
// and clears it. An empty rectangle means there is no pending update.
func (w *World) ConsumeTerrainDirtyRegion() image.Rectangle {
	dirty := w.terrainDirty
	w.terrainDirty = image.Rectangle{}
	return dirty
}

func expandTerrainRegion(region image.Rectangle, amount int) image.Rectangle {
	return image.Rect(region.Min.X-amount, region.Min.Y-amount, region.Max.X+amount, region.Max.Y+amount)
}

func unionTerrainRegion(a, b image.Rectangle) image.Rectangle {
	if a.Empty() {
		return b
	}
	if b.Empty() {
		return a
	}
	return a.Union(b)
}
