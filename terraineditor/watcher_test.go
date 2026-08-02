package terraineditor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWatcherReportsOnlyAcceptedContentChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "terrain.json")
	document := testDocument()
	if err := document.Save(path); err != nil {
		t.Fatal(err)
	}
	watcher, err := NewWatcher(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, changed, err := watcher.Poll(); err != nil || changed {
		t.Fatalf("unchanged poll changed=%t err=%v", changed, err)
	}
	data, _ := os.ReadFile(path)
	updated, _ := Parse(data)
	_, _ = updated.Apply(Command{Operation: "set-height", Parameters: Parameters{X: 1, Y: 1, Height: 3}})
	if err := updated.Save(path); err != nil {
		t.Fatal(err)
	}
	if _, changed, err := watcher.Poll(); err != nil || !changed {
		t.Fatalf("first changed poll changed=%t err=%v", changed, err)
	}
	if _, changed, _ := watcher.Poll(); !changed {
		t.Fatal("unaccepted change was forgotten")
	}
	watcher.Accept()
	if _, changed, err := watcher.Poll(); err != nil || changed {
		t.Fatalf("accepted poll changed=%t err=%v", changed, err)
	}
}
