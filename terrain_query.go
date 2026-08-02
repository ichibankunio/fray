package fray

import (
	"math"

	"github.com/ichibankunio/fvec/vec2"
	"github.com/ichibankunio/fvec/vec3"
)

type terrainGradient struct {
	value float64
	dx    float64
	dy    float64
}

// TerrainSample describes the inferred surface at a position in grid blocks.
type TerrainSample struct {
	Position vec3.Vec3
	Normal   vec3.Vec3
	Gradient vec2.Vec2
	Slope    float64
	Inside   bool
}

// TerrainContact describes the vertical relationship between a point and the terrain.
type TerrainContact struct {
	TerrainSample
	Distance float64
	Grounded bool
}

// TerrainRaycastHit is an intersection with the inferred height surface.
type TerrainRaycastHit struct {
	TerrainSample
	Distance float64
}

// TerrainRaycastConfig controls CPU ray traversal in grid blocks.
type TerrainRaycastConfig struct {
	Step            float64
	RefinementSteps int
}

func DefaultTerrainRaycastConfig() TerrainRaycastConfig {
	return TerrainRaycastConfig{Step: 0.05, RefinementSteps: 12}
}

// SampleTerrain returns height, exact interpolation gradient, normal, and slope together.
func (w *World) SampleTerrain(x, y float64) TerrainSample {
	inside := x >= 0 && y >= 0 && x < float64(w.canvasWidth) && y < float64(w.canvasHeight)
	surface := w.terrainGradientAtBlocks(x, y)
	normal := vec3.New(-surface.dx, -surface.dy, 1)
	normal = normal.Scale(1 / max(0.001, normal.Length()))
	return TerrainSample{
		Position: vec3.New(x, y, surface.value),
		Normal:   normal,
		Gradient: vec2.New(surface.dx, surface.dy),
		Slope:    math.Hypot(surface.dx, surface.dy),
		Inside:   inside,
	}
}

// QueryTerrainContact compares a point with the surface directly below it.
func (w *World) QueryTerrainContact(position vec3.Vec3, groundTolerance float64) TerrainContact {
	sample := w.SampleTerrain(position.X, position.Y)
	distance := position.Z - sample.Position.Z
	return TerrainContact{TerrainSample: sample, Distance: distance, Grounded: sample.Inside && distance >= 0 && distance <= max(0, groundTolerance)}
}

// RaycastTerrain intersects a normalized ray with the inferred height surface.
func (w *World) RaycastTerrain(origin, direction vec3.Vec3, maxDistance float64, config TerrainRaycastConfig) (TerrainRaycastHit, bool) {
	length := direction.Length()
	if length <= 1e-12 || maxDistance < 0 {
		return TerrainRaycastHit{}, false
	}
	direction = direction.Scale(1 / length)
	if config.Step <= 0 {
		config.Step = DefaultTerrainRaycastConfig().Step
	}
	if config.RefinementSteps <= 0 {
		config.RefinementSteps = DefaultTerrainRaycastConfig().RefinementSteps
	}

	previousDistance := 0.0
	previousSample := w.SampleTerrain(origin.X, origin.Y)
	previousClearance := origin.Z - previousSample.Position.Z
	if previousSample.Inside && previousClearance <= 0 {
		return TerrainRaycastHit{TerrainSample: previousSample}, true
	}
	for distance := min(config.Step, maxDistance); distance <= maxDistance+1e-12; distance = min(distance+config.Step, maxDistance) {
		point := origin.Add(direction.Scale(distance))
		sample := w.SampleTerrain(point.X, point.Y)
		clearance := point.Z - sample.Position.Z
		if sample.Inside && previousSample.Inside && previousClearance > 0 && clearance <= 0 {
			near, far := previousDistance, distance
			for i := 0; i < config.RefinementSteps; i++ {
				mid := (near + far) * 0.5
				midPoint := origin.Add(direction.Scale(mid))
				midSample := w.SampleTerrain(midPoint.X, midPoint.Y)
				if midPoint.Z <= midSample.Position.Z {
					far = mid
				} else {
					near = mid
				}
			}
			hitPoint := origin.Add(direction.Scale(far))
			hitSample := w.SampleTerrain(hitPoint.X, hitPoint.Y)
			return TerrainRaycastHit{TerrainSample: hitSample, Distance: far}, true
		}
		if distance >= maxDistance {
			break
		}
		previousDistance = distance
		previousSample = sample
		previousClearance = clearance
	}
	return TerrainRaycastHit{}, false
}

func (w *World) terrainGradientAtBlocks(x, y float64) terrainGradient {
	x = max(0, min(float64(w.canvasWidth)-1e-6, x))
	y = max(0, min(float64(w.canvasHeight)-1e-6, y))
	cellX, cellY := int(math.Floor(x)), int(math.Floor(y))
	fx, fy := x-float64(cellX), y-float64(cellY)
	h00 := w.heightSample(cellX, cellY)
	if w.TerrainInterpolation == TerrainInterpolationFlat {
		return terrainGradient{value: h00}
	}
	if w.TerrainInterpolation == TerrainInterpolationLinear {
		h10 := w.heightSample(cellX+1, cellY)
		h01 := w.heightSample(cellX, cellY+1)
		h11 := w.heightSample(cellX+1, cellY+1)
		top, bottom := h00+(h10-h00)*fx, h01+(h11-h01)*fx
		return terrainGradient{value: top + (bottom-top)*fy, dx: (h10-h00)*(1-fy) + (h11-h01)*fy, dy: bottom - top}
	}

	rows := [4]terrainGradient{}
	for row := -1; row <= 2; row++ {
		rows[row+1] = monotonicCubicGradient(
			terrainGradient{value: w.heightSample(cellX-1, cellY+row)}, terrainGradient{value: w.heightSample(cellX, cellY+row)},
			terrainGradient{value: w.heightSample(cellX+1, cellY+row)}, terrainGradient{value: w.heightSample(cellX+2, cellY+row)},
			terrainGradient{value: fx, dx: 1},
		)
	}
	return monotonicCubicGradient(rows[0], rows[1], rows[2], rows[3], terrainGradient{value: fy, dy: 1})
}

func monotonicSlopeGradient(before, after terrainGradient) terrainGradient {
	if before.value*after.value <= 0 {
		return terrainGradient{}
	}
	sum := before.value + after.value
	product := before.value * after.value
	dxProduct := before.dx*after.value + before.value*after.dx
	dyProduct := before.dy*after.value + before.value*after.dy
	return terrainGradient{
		value: 2 * product / sum,
		dx:    2 * (dxProduct*sum - product*(before.dx+after.dx)) / (sum * sum),
		dy:    2 * (dyProduct*sum - product*(before.dy+after.dy)) / (sum * sum),
	}
}

func monotonicCubicGradient(p0, p1, p2, p3, t terrainGradient) terrainGradient {
	d0, d1, d2 := subtractGradient(p1, p0), subtractGradient(p2, p1), subtractGradient(p3, p2)
	m1, m2 := monotonicSlopeGradient(d0, d1), monotonicSlopeGradient(d1, d2)
	t2, t3 := t.value*t.value, t.value*t.value*t.value
	b0, b1, b2, b3 := 2*t3-3*t2+1, t3-2*t2+t.value, -2*t3+3*t2, t3-t2
	db0, db1 := 6*t2-6*t.value, 3*t2-4*t.value+1
	db2, db3 := -6*t2+6*t.value, 3*t2-2*t.value
	return terrainGradient{
		value: b0*p1.value + b1*m1.value + b2*p2.value + b3*m2.value,
		dx:    db0*t.dx*p1.value + b0*p1.dx + db1*t.dx*m1.value + b1*m1.dx + db2*t.dx*p2.value + b2*p2.dx + db3*t.dx*m2.value + b3*m2.dx,
		dy:    db0*t.dy*p1.value + b0*p1.dy + db1*t.dy*m1.value + b1*m1.dy + db2*t.dy*p2.value + b2*p2.dy + db3*t.dy*m2.value + b3*m2.dy,
	}
}

func subtractGradient(a, b terrainGradient) terrainGradient {
	return terrainGradient{value: a.value - b.value, dx: a.dx - b.dx, dy: a.dy - b.dy}
}
