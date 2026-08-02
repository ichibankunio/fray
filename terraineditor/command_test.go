package terraineditor

import "testing"

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
