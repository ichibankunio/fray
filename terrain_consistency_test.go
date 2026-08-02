package fray

import (
	"bytes"
	"math"
	"math/rand"
	"testing"
)

func TestTerrainAnalyticGradientMatchesScalarHermiteSurface(t *testing.T) {
	w := benchmarkTerrainWorld(32)
	random := rand.New(rand.NewSource(0x46524159))
	const epsilon = 1e-5
	for i := 0; i < 256; i++ {
		x := 2 + random.Float64()*27
		y := 2 + random.Float64()*27
		sample := w.SampleTerrain(x, y)
		dx := (w.heightAtBlocks(x+epsilon, y) - w.heightAtBlocks(x-epsilon, y)) / (2 * epsilon)
		dy := (w.heightAtBlocks(x, y+epsilon) - w.heightAtBlocks(x, y-epsilon)) / (2 * epsilon)
		if math.Abs(sample.Gradient.X-dx) > 2e-5 || math.Abs(sample.Gradient.Y-dy) > 2e-5 {
			t.Fatalf("sample %d at (%.4f, %.4f): analytic=(%.8f, %.8f) scalar=(%.8f, %.8f)", i, x, y, sample.Gradient.X, sample.Gradient.Y, dx, dy)
		}
	}
}

func TestStandardTerrainShaderUsesSharedHermiteGradient(t *testing.T) {
	required := [][]byte{
		[]byte("func MonotonicCubicGradient"),
		[]byte("func TerrainHeightAndGradient"),
		[]byte("surface := TerrainHeightAndGradient(hitPos)"),
		[]byte("normal := normalize(vec3(-surface.y, -surface.z, 1.0))"),
	}
	for _, token := range required {
		if !bytes.Contains(terrainShaderByte, token) {
			t.Fatalf("standard terrain shader lost CPU/GPU interpolation contract token %q", token)
		}
	}
}

func benchmarkTerrainWorld(size int) *World {
	heights := make([]uint8, size*size)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			heights[y*size+x] = uint8(4 + (x*x+3*y*y+x*y)%19)
		}
	}
	return &World{HeightMap: heights, TerrainInterpolation: TerrainInterpolationMonotonic, canvasWidth: size, canvasHeight: size, canvasDepth: 32}
}
