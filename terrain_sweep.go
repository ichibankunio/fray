package fray

import "github.com/ichibankunio/fvec/vec3"

// TerrainSweepHit describes the first contact along a swept volume movement.
type TerrainSweepHit struct {
	TerrainSample
	Position vec3.Vec3
	Distance float64
	Fraction float64
}

// SweepTerrainSphere sweeps a sphere through the inferred terrain. Positions,
// radius, and movement use terrain grid blocks.
func (w *World) SweepTerrainSphere(origin, movement vec3.Vec3, radius float64, config TerrainRaycastConfig) (TerrainSweepHit, bool) {
	return w.sweepTerrainVerticalCapsule(origin, movement, max(0, radius), 0, config)
}

// SweepTerrainCapsule sweeps a vertical capsule. halfSegment is the distance
// from its center to either sphere center and excludes the radius.
func (w *World) SweepTerrainCapsule(origin, movement vec3.Vec3, radius, halfSegment float64, config TerrainRaycastConfig) (TerrainSweepHit, bool) {
	return w.sweepTerrainVerticalCapsule(origin, movement, max(0, radius), max(0, halfSegment), config)
}

func (w *World) sweepTerrainVerticalCapsule(origin, movement vec3.Vec3, radius, halfSegment float64, config TerrainRaycastConfig) (TerrainSweepHit, bool) {
	distance := movement.Length()
	if config.Step <= 0 {
		config.Step = DefaultTerrainRaycastConfig().Step
	}
	if config.RefinementSteps <= 0 {
		config.RefinementSteps = DefaultTerrainRaycastConfig().RefinementSteps
	}
	direction := vec3.New(0, 0, 0)
	if distance > 1e-12 {
		direction = movement.Scale(1 / distance)
	}
	clearanceAt := func(travel float64) (float64, TerrainSample, vec3.Vec3) {
		center := origin.Add(direction.Scale(travel))
		lowerCenter := center.Add(vec3.New(0, 0, -halfSegment))
		surface := w.SampleTerrain(lowerCenter.X, lowerCenter.Y)
		// For the local tangent plane, vertical separation times normal.Z is
		// signed sphere-to-plane distance.
		clearance := (lowerCenter.Z-surface.Position.Z)*surface.Normal.Z - radius
		return clearance, surface, center
	}
	initial, surface, center := clearanceAt(0)
	if surface.Inside && initial <= 0 {
		return TerrainSweepHit{TerrainSample: surface, Position: center}, true
	}
	if distance <= 1e-12 {
		return TerrainSweepHit{}, false
	}
	previous := 0.0
	previousClearance := initial
	for travel := min(config.Step, distance); travel <= distance+1e-12; travel = min(travel+config.Step, distance) {
		clearance, currentSurface, _ := clearanceAt(travel)
		if currentSurface.Inside && previousClearance > 0 && clearance <= 0 {
			near, far := previous, travel
			for i := 0; i < config.RefinementSteps; i++ {
				mid := (near + far) * .5
				midClearance, _, _ := clearanceAt(mid)
				if midClearance <= 0 {
					far = mid
				} else {
					near = mid
				}
			}
			_, hitSurface, hitCenter := clearanceAt(far)
			return TerrainSweepHit{TerrainSample: hitSurface, Position: hitCenter, Distance: far, Fraction: far / distance}, true
		}
		if travel >= distance {
			break
		}
		previous, previousClearance = travel, clearance
	}
	return TerrainSweepHit{}, false
}
