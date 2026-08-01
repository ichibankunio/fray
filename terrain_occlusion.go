package fray

import "math"

// TerrainOcclusionConfig controls CPU-side terrain visibility queries.
// All positions and heights are expressed in terrain grid blocks.
type TerrainOcclusionConfig struct {
	SampleSpacing float64
	Clearance     float64
}

func DefaultTerrainOcclusionConfig() TerrainOcclusionConfig {
	return TerrainOcclusionConfig{SampleSpacing: .5, Clearance: .08}
}

// TerrainVisibleHeight returns the lowest target height visible from the
// camera. Objects can clip their lower edge to this height instead of popping
// out when only their base is hidden by terrain.
func (w *World) TerrainVisibleHeight(cameraX, cameraY, cameraZ, targetX, targetY float64, config TerrainOcclusionConfig) float64 {
	if config.SampleSpacing <= 0 {
		config.SampleSpacing = DefaultTerrainOcclusionConfig().SampleSpacing
	}
	if config.Clearance < 0 {
		config.Clearance = 0
	}
	dx := targetX - cameraX
	dy := targetY - cameraY
	distance := math.Hypot(dx, dy)
	steps := max(1, int(math.Ceil(distance/config.SampleSpacing)))
	visibleHeight := math.Inf(-1)
	for step := 1; step < steps; step++ {
		t := float64(step) / float64(steps)
		height := w.SampleTerrainHeight(cameraX+dx*t, cameraY+dy*t) + config.Clearance
		// Solve cameraZ+(targetZ-cameraZ)*t >= height for targetZ.
		requiredTargetHeight := cameraZ + (height-cameraZ)/t
		visibleHeight = max(visibleHeight, requiredTargetHeight)
	}
	return visibleHeight
}

// TerrainLineOfSight reports whether a point is visible above the interpolated
// terrain surface.
func (w *World) TerrainLineOfSight(cameraX, cameraY, cameraZ, targetX, targetY, targetZ float64, config TerrainOcclusionConfig) bool {
	return targetZ >= w.TerrainVisibleHeight(cameraX, cameraY, cameraZ, targetX, targetY, config)
}
