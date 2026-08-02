package terraineditor

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestParseAndMarshalAreDeterministic(t *testing.T) {
	input := []byte(`{"version":3,"terrain":{"interpolation":"monotonic"},"canvasWidth":2,"canvasHeight":2,"canvasDepth":3,"layers":["AQEBAQ=="]}`)
	document, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	first, err := document.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	secondDocument, err := Parse(first)
	if err != nil {
		t.Fatal(err)
	}
	second, _ := secondDocument.Marshal()
	if !bytes.Equal(first, second) {
		t.Fatalf("marshal is not deterministic:\n%s\n%s", first, second)
	}
}

func TestSaveAtomicallyReplacesDocument(t *testing.T) {
	document := testDocument()
	path := filepath.Join(t.TempDir(), "terrain.json")
	if err := document.Save(path); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.HeightAt(1, 1) != 1 {
		t.Fatalf("loaded height = %d", loaded.HeightAt(1, 1))
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Fatalf("temporary files remained: %v", entries)
	}
}

func TestValidateRejectsMalformedLayer(t *testing.T) {
	document := testDocument()
	document.Layers[0] = document.Layers[0][:2]
	if err := document.Validate(); err == nil {
		t.Fatal("Validate accepted short layer")
	}
}

func testDocument() *Document {
	return &Document{Version: 3, Terrain: TerrainOptions{Interpolation: "monotonic"}, CanvasWidth: 3, CanvasHeight: 3, CanvasDepth: 8, Layers: [][]uint8{{1, 1, 1, 1, 1, 1, 1, 1, 1}}}
}
