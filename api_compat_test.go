package fray_test

import (
	"image"

	"github.com/ichibankunio/fray"
	"github.com/ichibankunio/fvec/vec3"
)

// Compile-time API checks protect the public terrain surface used by games.
func ExampleWorld_terrainAPICompatibility() {
	var world *fray.World
	_ = world.SampleTerrain(0, 0)
	_ = world.QueryTerrainContact(vec3.New(0, 0, 0), .1)
	_, _ = world.RaycastTerrain(vec3.New(0, 0, 1), vec3.New(0, 0, -1), 10, fray.DefaultTerrainRaycastConfig())
	_ = world.RebuildTerrainRegion(image.Rect(0, 0, 1, 1))
	_ = world.TerrainRevision()
	_ = world.ConsumeTerrainDirtyRegion()
}
