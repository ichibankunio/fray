package fray

import (
	"math"

	"github.com/ichibankunio/fvec/vec3"
)

// TerrainPlacementRule derives deterministic object placements from the
// inferred height surface. Positions and ranges use world grid blocks.
type TerrainPlacementRule struct {
	Seed        uint64
	CellSize    int
	Density     float64
	EdgePadding int
	MinHeight   float64
	MaxHeight   float64
	MaxSlope    float64
	MinScale    float64
	MaxScale    float64
}

type TerrainPlacement struct {
	Position vec3.Vec3
	Normal   vec3.Vec3
	Rotation float64
	Scale    float64
}

func (w *World) GenerateTerrainPlacements(rule TerrainPlacementRule) []TerrainPlacement {
	if rule.CellSize < 1 {
		rule.CellSize = 1
	}
	rule.Density = max(0, min(1, rule.Density))
	if rule.MaxHeight <= rule.MinHeight {
		rule.MaxHeight = float64(w.canvasDepth)
	}
	if rule.MaxSlope <= 0 {
		rule.MaxSlope = math.Inf(1)
	}
	if rule.MinScale <= 0 {
		rule.MinScale = 1
	}
	if rule.MaxScale < rule.MinScale {
		rule.MaxScale = rule.MinScale
	}

	placements := make([]TerrainPlacement, 0)
	for y := rule.EdgePadding; y < w.canvasHeight-rule.EdgePadding; y += rule.CellSize {
		for x := rule.EdgePadding; x < w.canvasWidth-rule.EdgePadding; x += rule.CellSize {
			hash := terrainPlacementHash(rule.Seed, uint64(x), uint64(y))
			if hashUnit(hash) > rule.Density {
				continue
			}
			jitterX := (hashUnit(hash>>16) - 0.5) * float64(rule.CellSize) * 0.82
			jitterY := (hashUnit(hash>>32) - 0.5) * float64(rule.CellSize) * 0.82
			px := max(0.001, min(float64(w.canvasWidth)-1.001, float64(x)+float64(rule.CellSize)*0.5+jitterX))
			py := max(0.001, min(float64(w.canvasHeight)-1.001, float64(y)+float64(rule.CellSize)*0.5+jitterY))
			surface := w.SampleTerrain(px, py)
			height := surface.Position.Z
			if height < rule.MinHeight || height > rule.MaxHeight {
				continue
			}
			normal := surface.Normal
			slope := surface.Slope
			if slope > rule.MaxSlope {
				continue
			}
			scaleUnit := hashUnit(hash >> 8)
			placements = append(placements, TerrainPlacement{
				Position: vec3.New(px, py, height),
				Normal:   normal,
				Rotation: hashUnit(hash>>24) * math.Pi * 2,
				Scale:    rule.MinScale + (rule.MaxScale-rule.MinScale)*scaleUnit,
			})
		}
	}
	return placements
}

func (w *World) terrainNormalAtBlocks(x, y float64) vec3.Vec3 {
	return w.SampleTerrain(x, y).Normal
}

func terrainPlacementHash(seed, x, y uint64) uint64 {
	value := seed ^ x*0x9e3779b97f4a7c15 ^ y*0xbf58476d1ce4e5b9
	value ^= value >> 30
	value *= 0xbf58476d1ce4e5b9
	value ^= value >> 27
	value *= 0x94d049bb133111eb
	return value ^ (value >> 31)
}

func hashUnit(value uint64) float64 {
	return float64(value&0xffff) / 65535
}
