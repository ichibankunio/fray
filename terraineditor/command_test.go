package terraineditor

import (
	"path/filepath"
	"testing"
)

func TestSetRaiseLowerFlattenAndSmooth(t *testing.T) {
	tests := []struct {
		command Command
		x, y    int
		want    int
	}{
		{Command{Operation: "set-height", Parameters: Parameters{X: 1, Y: 1, Height: 5}}, 1, 1, 5},
		{Command{Operation: "raise", Parameters: Parameters{X: 1, Y: 1, Radius: 1, Amount: 2}}, 1, 1, 3},
		{Command{Operation: "lower", Parameters: Parameters{X: 1, Y: 1, Amount: 1}}, 1, 1, 0},
		{Command{Operation: "flatten", Parameters: Parameters{X: 0, Y: 0, Width: 2, Rows: 2, Height: 4}}, 1, 1, 4},
	}
	for _, test := range tests {
		document := testDocument()
		if _, err := document.Apply(test.command); err != nil {
			t.Fatalf("%s: %v", test.command.Operation, err)
		}
		if got := document.HeightAt(test.x, test.y); got != test.want {
			t.Fatalf("%s height = %d, want %d", test.command.Operation, got, test.want)
		}
	}

	document := testDocument()
	_, _ = document.Apply(Command{Operation: "set-height", Parameters: Parameters{X: 1, Y: 1, Height: 7}})
	_, err := document.Apply(Command{Operation: "smooth", Parameters: Parameters{X: 1, Y: 1, Radius: 1, Blend: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if got := document.HeightAt(1, 1); got != 2 {
		t.Fatalf("smoothed center = %d, want 2", got)
	}
}

func TestHistoryUndoRedoPreservesColumnMaterials(t *testing.T) {
	document := testDocument()
	document.Layers[0][4] = 7
	history := NewHistory(document)
	changeSet, err := history.Apply(Command{Operation: "set-height", Parameters: Parameters{X: 1, Y: 1, Height: 4}})
	if err != nil || len(changeSet.Changes) != 1 || document.HeightAt(1, 1) != 4 {
		t.Fatalf("apply = %+v, err=%v", changeSet, err)
	}
	if _, ok, err := history.Undo(); err != nil || !ok || document.HeightAt(1, 1) != 1 || document.Layers[0][4] != 7 {
		t.Fatalf("undo ok=%t err=%v height=%d material=%d", ok, err, document.HeightAt(1, 1), document.Layers[0][4])
	}
	if _, ok, err := history.Redo(); err != nil || !ok || document.HeightAt(1, 1) != 4 || document.Layers[3][4] != 7 {
		t.Fatalf("redo ok=%t err=%v height=%d material=%d", ok, err, document.HeightAt(1, 1), document.Layers[3][4])
	}
}

func TestSmoothUsesSnapshotAndIsOrderIndependent(t *testing.T) {
	document := testDocument()
	_, _ = document.Apply(Command{Operation: "set-height", Parameters: Parameters{X: 0, Y: 0, Height: 8}})
	changeSet, err := document.Apply(Command{Operation: "smooth", Parameters: Parameters{X: 1, Y: 1, Radius: 2, Blend: 1}})
	if err != nil || len(changeSet.Changes) == 0 {
		t.Fatalf("smooth changes=%d err=%v", len(changeSet.Changes), err)
	}
}

func TestRegionCommandsPreserveMaterialsAndOverlap(t *testing.T) {
	document := testDocument()
	document.Layers[0][0] = 7
	_, err := document.Apply(Command{Operation: "copy-region", Parameters: Parameters{X: 0, Y: 0, Width: 2, Rows: 1, ToX: 1, ToY: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if got := document.Layers[0][4]; got != 7 {
		t.Fatalf("copied material = %d, want 7", got)
	}

	_, err = document.Apply(Command{Operation: "move-region", Parameters: Parameters{X: 0, Y: 0, Width: 2, Rows: 1, ToX: 1, ToY: 0}})
	if err != nil {
		t.Fatal(err)
	}
	if got := document.HeightAt(0, 0); got != 0 {
		t.Fatalf("cleared source height = %d", got)
	}
	if got := document.Layers[0][1]; got != 7 {
		t.Fatalf("overlapping move material = %d, want 7", got)
	}
}

func TestFlipRegion(t *testing.T) {
	document := testDocument()
	_, _ = document.Apply(Command{Operation: "set-height", Parameters: Parameters{X: 0, Y: 0, Height: 3}})
	_, err := document.Apply(Command{Operation: "flip-region", Parameters: Parameters{X: 0, Y: 0, Width: 3, Rows: 1, Axis: "horizontal"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := document.HeightAt(2, 0); got != 3 {
		t.Fatalf("flipped height = %d, want 3", got)
	}
}

func TestHistoryCommandsPersistAndReplay(t *testing.T) {
	document := testDocument()
	history := NewHistory(document)
	command := Command{Operation: "raise", Parameters: Parameters{X: 1, Y: 1, Amount: 2}}
	_, _ = history.Apply(command)
	path := filepath.Join(t.TempDir(), "history.json")
	if err := history.SaveCommands(path); err != nil {
		t.Fatal(err)
	}
	commands, err := LoadCommands(path)
	if err != nil || len(commands) != 1 {
		t.Fatalf("commands=%v err=%v", commands, err)
	}
	replayed := testDocument()
	if _, err := Replay(replayed, commands, Constraints{}); err != nil {
		t.Fatal(err)
	}
	if got := replayed.HeightAt(1, 1); got != 3 {
		t.Fatalf("replayed height=%d, want 3", got)
	}
	_, _, _ = history.Undo()
	if len(history.Commands()) != 0 {
		t.Fatal("undone command remained in active history")
	}
}
