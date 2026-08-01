package fray

import "testing"

func TestSetTerrainAtmosphereConfigClampsDensities(t *testing.T) {
	r := &Renderer{}
	r.SetTerrainAtmosphereConfig(TerrainAtmosphereConfig{
		DistanceFog: -1,
		HeightFog:   -2,
		SunGlow:     3,
	})
	config := r.terrainAtmosphereConfig
	if config.DistanceFog != 0 || config.HeightFog != 0 || config.SunGlow != 1 {
		t.Fatalf("unexpected normalized atmosphere: %+v", config)
	}
}

func TestDefaultTerrainAtmosphereConfigHasVisibleRange(t *testing.T) {
	config := DefaultTerrainAtmosphereConfig()
	if config.DistanceFog <= 0 || config.FogBaseHeight <= 0 {
		t.Fatalf("unexpected default atmosphere: %+v", config)
	}
}
