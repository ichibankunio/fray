package terraineditor

import "testing"

func BenchmarkPreviewBrush(b *testing.B) {
	d := benchmarkDocument(128, 128, 32)
	command := Command{Operation: "smooth", Parameters: Parameters{X: 64, Y: 64, Radius: 8, Blend: .5, Shape: "circle"}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := d.Preview(command); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkDocument(width, height, depth int) *Document {
	layer := make([]uint8, width*height)
	for i := range layer {
		layer[i] = 1
	}
	return &Document{Version: CurrentVersion, CanvasWidth: width, CanvasHeight: height, CanvasDepth: depth, Layers: [][]uint8{layer}, Terrain: TerrainOptions{Interpolation: "monotonic"}}
}
