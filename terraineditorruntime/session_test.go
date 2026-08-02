package terraineditorruntime

import (
	"testing"

	"github.com/ichibankunio/fray"
	"github.com/ichibankunio/fray/terraineditor"
)

func TestSessionAppliesAndUndoesLiveTerrain(t *testing.T) {
	document := &terraineditor.Document{Version: 3, Terrain: terraineditor.TerrainOptions{Interpolation: "monotonic"}, CanvasWidth: 3, CanvasHeight: 3, CanvasDepth: 4, Layers: [][]uint8{{1, 1, 1, 1, 1, 1, 1, 1, 1}}}
	world := &fray.World{}
	world.Init(3, 3, 3, 3, 4)
	copy(world.WorldMap[0], document.Layers[0])
	world.RebuildTerrain()
	renderer := &fray.Renderer{Wld: world}
	session, err := New(document, renderer)
	if err != nil {
		t.Fatal(err)
	}
	changeSet, err := session.Apply(terraineditor.Command{Operation: "set-height", Parameters: terraineditor.Parameters{X: 1, Y: 1, Height: 3}})
	if err != nil || len(changeSet.Changes) != 1 || world.GetHeight(1, 1) != 3 {
		t.Fatalf("apply changes=%d height=%d err=%v", len(changeSet.Changes), world.GetHeight(1, 1), err)
	}
	if _, ok, err := session.Undo(); err != nil || !ok || world.GetHeight(1, 1) != 1 {
		t.Fatalf("undo ok=%t height=%d err=%v", ok, world.GetHeight(1, 1), err)
	}
}

func TestSessionReloadReplacesWorldAndHistory(t *testing.T) {
	document := &terraineditor.Document{Version: 3, CanvasWidth: 2, CanvasHeight: 2, CanvasDepth: 3, Layers: [][]uint8{{1, 1, 1, 1}}}
	world := &fray.World{}
	world.Init(2, 2, 2, 2, 3)
	renderer := &fray.Renderer{Wld: world}
	session, err := New(document, renderer)
	if err != nil {
		t.Fatal(err)
	}
	reloaded := &terraineditor.Document{Version: 3, CanvasWidth: 2, CanvasHeight: 2, CanvasDepth: 3, Layers: [][]uint8{{1, 1, 1, 1}, {0, 2, 0, 0}}}
	if err := session.Reload(reloaded); err != nil {
		t.Fatal(err)
	}
	if world.GetHeight(1, 0) != 2 || session.History.CanUndo() {
		t.Fatalf("height=%d canUndo=%t", world.GetHeight(1, 0), session.History.CanUndo())
	}
}

func BenchmarkSessionBrushEdit(b *testing.B) {
	const size = 128
	cells := size * size
	document := &terraineditor.Document{Version: 3, CanvasWidth: size, CanvasHeight: size, CanvasDepth: 56, Layers: [][]uint8{make([]uint8, cells)}}
	for index := range document.Layers[0] {
		document.Layers[0][index] = 1
	}
	world := &fray.World{}
	world.Init(size, size, size, size, 56)
	copy(world.WorldMap[0], document.Layers[0])
	world.RebuildTerrain()
	session, _ := New(document, &fray.Renderer{Wld: world})
	command := terraineditor.Command{Operation: "raise", Parameters: terraineditor.Parameters{X: 64, Y: 64, Radius: 2, Amount: 1, Falloff: "smooth"}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = session.Apply(command)
		_, _, _ = session.Undo()
	}
}
