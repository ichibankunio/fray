package fray

import "testing"

func TestSetTerrainRaymarchConfigNormalizesInvalidValues(t *testing.T) {
	r := &Renderer{}
	r.SetTerrainRaymarchConfig(TerrainRaymarchConfig{
		NearStep:      -1,
		MidStep:       0.01,
		FarStep:       0.01,
		MidDistance:   5,
		FarDistance:   2,
		MaxDistance:   1,
		SurfaceBand:   0,
		FlatStepBoost: 9,
	})

	config := r.terrainRaymarchConfig
	if config.NearStep <= 0 || config.MidStep < config.NearStep || config.FarStep < config.MidStep {
		t.Fatalf("invalid normalized steps: %+v", config)
	}
	if config.FarDistance < config.MidDistance || config.MaxDistance < config.FarDistance {
		t.Fatalf("invalid normalized distances: %+v", config)
	}
	if config.SurfaceBand <= 0 || config.FlatStepBoost != 2 {
		t.Fatalf("invalid adaptive parameters: %+v", config)
	}
}

func TestDefaultTerrainRaymarchConfigPreservesTwentyBlockRange(t *testing.T) {
	config := DefaultTerrainRaymarchConfig()
	if config.MaxDistance != 20 {
		t.Fatalf("MaxDistance = %v, want 20", config.MaxDistance)
	}
	if config.NearStep > config.MidStep || config.MidStep > config.FarStep {
		t.Fatalf("steps are not monotonic: %+v", config)
	}
}
