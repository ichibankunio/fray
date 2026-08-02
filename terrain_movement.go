package fray

import (
	"math"

	"github.com/ichibankunio/fvec/vec3"
)

type TerrainMovementConfig struct {
	Radius             float64
	HalfSegment        float64
	MaxSlope           float64
	MaxStepHeight      float64
	GroundTolerance    float64
	GroundSnapDistance float64
	Skin               float64
	Sweep              TerrainRaycastConfig
}

func DefaultTerrainMovementConfig() TerrainMovementConfig {
	return TerrainMovementConfig{
		Radius:             .3,
		HalfSegment:        .6,
		MaxSlope:           .8,
		MaxStepHeight:      .6,
		GroundTolerance:    .08,
		GroundSnapDistance: .25,
		Skin:               .002,
		Sweep:              DefaultTerrainRaycastConfig(),
	}
}

type TerrainMovementResult struct {
	Position        vec3.Vec3
	AppliedMovement vec3.Vec3
	GroundNormal    vec3.Vec3
	Grounded        bool
	Blocked         bool
	Slope           float64
}

// MoveOnTerrain resolves one input-independent movement step for a vertical
// capsule. It supports character, AI, and vehicle controllers without reading input.
func (w *World) MoveOnTerrain(origin, desiredMovement vec3.Vec3, config TerrainMovementConfig) TerrainMovementResult {
	config = normalizeTerrainMovementConfig(config)
	position := origin
	horizontalRejected := false
	startSupport := w.terrainCapsuleSupport(origin, config)
	startGrounded := startSupport.Inside && startSupport.Distance >= -config.Skin && startSupport.Distance <= config.GroundTolerance

	// Reject non-walkable rises before allowing the capsule to tunnel into a
	// cliff face. Airborne movement remains unconstrained until swept contact.
	if startGrounded && (desiredMovement.X != 0 || desiredMovement.Y != 0) {
		targetSurface := w.SampleTerrain(origin.X+desiredMovement.X, origin.Y+desiredMovement.Y)
		rise := targetSurface.Position.Z - startSupport.TerrainSample.Position.Z
		horizontalDistance := math.Hypot(desiredMovement.X, desiredMovement.Y)
		if targetSurface.Slope > config.MaxSlope || rise > config.MaxStepHeight+config.MaxSlope*horizontalDistance {
			desiredMovement.X = 0
			desiredMovement.Y = 0
			horizontalRejected = true
		}
	}

	hit, blocked := w.SweepTerrainCapsule(position, desiredMovement, config.Radius, config.HalfSegment, config.Sweep)
	if !blocked {
		position = position.Add(desiredMovement)
	} else {
		position = safeTerrainSweepPosition(position, desiredMovement, hit, config.Skin)
		remaining := desiredMovement.Scale(max(0, 1-hit.Fraction))
		inward := remaining.X*hit.Normal.X + remaining.Y*hit.Normal.Y + remaining.Z*hit.Normal.Z
		if inward < 0 {
			remaining = remaining.Add(hit.Normal.Scale(-inward))
		}
		if remaining.Length() > config.Skin {
			if slideHit, slideBlocked := w.SweepTerrainCapsule(position, remaining, config.Radius, config.HalfSegment, config.Sweep); slideBlocked {
				position = safeTerrainSweepPosition(position, remaining, slideHit, config.Skin)
			} else {
				position = position.Add(remaining)
			}
		}
	}

	support := w.terrainCapsuleSupport(position, config)
	if support.Inside && support.Slope <= config.MaxSlope && support.Distance >= -config.Skin && support.Distance <= config.GroundSnapDistance {
		position.Z -= support.Distance
		support.Distance = 0
	}
	grounded := support.Inside && support.Slope <= config.MaxSlope && support.Distance >= -config.Skin && support.Distance <= config.GroundTolerance
	return TerrainMovementResult{
		Position:        position,
		AppliedMovement: position.Sub(origin),
		GroundNormal:    support.Normal,
		Grounded:        grounded,
		Blocked:         blocked || horizontalRejected,
		Slope:           support.Slope,
	}
}

func normalizeTerrainMovementConfig(config TerrainMovementConfig) TerrainMovementConfig {
	defaults := DefaultTerrainMovementConfig()
	if config.Radius <= 0 {
		config.Radius = defaults.Radius
	}
	if config.HalfSegment < 0 {
		config.HalfSegment = 0
	}
	if config.MaxSlope <= 0 {
		config.MaxSlope = defaults.MaxSlope
	}
	if config.MaxStepHeight <= 0 {
		config.MaxStepHeight = defaults.MaxStepHeight
	}
	if config.GroundTolerance <= 0 {
		config.GroundTolerance = defaults.GroundTolerance
	}
	if config.GroundSnapDistance < config.GroundTolerance {
		config.GroundSnapDistance = max(defaults.GroundSnapDistance, config.GroundTolerance)
	}
	if config.Skin <= 0 {
		config.Skin = defaults.Skin
	}
	return config
}

func (w *World) terrainCapsuleSupport(center vec3.Vec3, config TerrainMovementConfig) TerrainContact {
	surface := w.SampleTerrain(center.X, center.Y)
	requiredCenterZ := surface.Position.Z + config.HalfSegment + config.Radius/max(.001, surface.Normal.Z)
	return TerrainContact{TerrainSample: surface, Distance: center.Z - requiredCenterZ}
}

func safeTerrainSweepPosition(origin, movement vec3.Vec3, hit TerrainSweepHit, skin float64) vec3.Vec3 {
	distance := movement.Length()
	if distance <= 1e-12 {
		return origin
	}
	travel := max(0, hit.Distance-skin)
	return origin.Add(movement.Scale(travel / distance))
}
