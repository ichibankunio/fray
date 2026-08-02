package fray

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStandardTerrainScenes(t *testing.T) {
	tests := []struct {
		name          string
		width, height int
		centerHeight  uint8
		hasWater      bool
	}{
		{name: "flat", width: 3, height: 3, centerHeight: 2},
		{name: "ridge", width: 5, height: 5, centerHeight: 3},
		{name: "valley", width: 5, height: 5, centerHeight: 1, hasWater: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", "terrain", test.name+".json"))
			if err != nil {
				t.Fatal(err)
			}
			w := &World{}
			w.Init(16, 16, test.width, test.height, 5)
			if err := w.LoadTerrainJSON(data); err != nil {
				t.Fatal(err)
			}
			if got := w.GetHeight(test.width/2, test.height/2); got != test.centerHeight {
				t.Fatalf("center height = %d, want %d", got, test.centerHeight)
			}
			if w.HasWater != test.hasWater {
				t.Fatalf("HasWater = %t", w.HasWater)
			}
		})
	}
}
