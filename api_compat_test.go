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
	_, _ = world.SweepTerrainSphere(vec3.New(0, 0, 1), vec3.New(0, 0, -1), .25, fray.DefaultTerrainRaycastConfig())
	_, _ = world.SweepTerrainCapsule(vec3.New(0, 0, 1), vec3.New(0, 0, -1), .25, .5, fray.DefaultTerrainRaycastConfig())
	_ = world.MoveOnTerrain(vec3.New(0, 0, 1), vec3.New(1, 0, 0), fray.DefaultTerrainMovementConfig())
	_ = world.RebuildTerrainRegion(image.Rect(0, 0, 1, 1))
	_ = world.TerrainRevision()
	_ = world.ConsumeTerrainDirtyRegion()
	var renderer *fray.Renderer
	_ = renderer.SyncTerrainGPU()
	_ = renderer.InitWithError(640, 480, 16, 16, 8, 16)
	renderer.ReleaseGPUResources()
	_ = renderer.RestoreGPUResources()
}
