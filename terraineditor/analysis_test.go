package terraineditor

import "testing"

func TestPreviewDoesNotModifyDocument(t *testing.T) {
	document := testDocument()
	changeSet, err := document.Preview(Command{Operation: "raise", Parameters: Parameters{X: 1, Y: 1, Amount: 3}})
	if err != nil || len(changeSet.Changes) != 1 {
		t.Fatalf("Preview changes=%d err=%v", len(changeSet.Changes), err)
	}
	if got := document.HeightAt(1, 1); got != 1 {
		t.Fatalf("height after preview = %d, want 1", got)
	}
	if summary := Summarize(changeSet); summary.MaximumDelta != 3 || summary.ChangedCells != 1 {
		t.Fatalf("summary = %+v", summary)
	}
}

func TestApplyConstrainedRollsBack(t *testing.T) {
	document := testDocument()
	maximum := 2
	_, err := document.ApplyConstrained(Command{Operation: "set-height", Parameters: Parameters{X: 1, Y: 1, Height: 5}}, Constraints{MaxHeight: &maximum})
	if err == nil {
		t.Fatal("expected constraint error")
	}
	if got := document.HeightAt(1, 1); got != 1 {
		t.Fatalf("height after rejected edit = %d, want 1", got)
	}
}
