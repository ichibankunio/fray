// Package terraineditor provides deterministic, headless terrain editing used
// by both graphical tools and automation such as Codex CLI.
package terraineditor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const CurrentVersion = 3

type TerrainOptions struct {
	Interpolation string   `json:"interpolation,omitempty"`
	WaterLevel    *float64 `json:"waterLevel,omitempty"`
}

type Document struct {
	Version      int            `json:"version"`
	Terrain      TerrainOptions `json:"terrain"`
	CanvasWidth  int            `json:"canvasWidth"`
	CanvasHeight int            `json:"canvasHeight"`
	CanvasDepth  int            `json:"canvasDepth"`
	Layers       [][]uint8      `json:"layers"`
}

func Parse(data []byte) (*Document, error) {
	var document Document
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("decode terrain document: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("decode terrain document: multiple JSON documents")
		}
		return nil, fmt.Errorf("decode terrain document trailing data: %w", err)
	}
	if err := document.Validate(); err != nil {
		return nil, err
	}
	document.normalize()
	return &document, nil
}

func Load(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read terrain document %s: %w", path, err)
	}
	return Parse(data)
}

func (d *Document) Validate() error {
	if d == nil {
		return fmt.Errorf("terrain document is nil")
	}
	if d.Version < 0 || d.Version > CurrentVersion {
		return fmt.Errorf("unsupported terrain document version %d", d.Version)
	}
	if d.CanvasWidth <= 0 || d.CanvasHeight <= 0 || d.CanvasDepth <= 0 || d.CanvasDepth > 255 {
		return fmt.Errorf("terrain dimensions must be positive and depth must not exceed 255")
	}
	if len(d.Layers) == 0 || len(d.Layers) > d.CanvasDepth {
		return fmt.Errorf("terrain has %d layers, expected 1..%d", len(d.Layers), d.CanvasDepth)
	}
	cells := d.CanvasWidth * d.CanvasHeight
	for index, layer := range d.Layers {
		if len(layer) != cells {
			return fmt.Errorf("terrain layer %d has %d cells, expected %d", index, len(layer), cells)
		}
	}
	switch d.Terrain.Interpolation {
	case "", "flat", "linear", "smooth", "monotonic":
	default:
		return fmt.Errorf("unknown terrain interpolation %q", d.Terrain.Interpolation)
	}
	if d.Terrain.WaterLevel != nil && (*d.Terrain.WaterLevel < 0 || *d.Terrain.WaterLevel > float64(d.CanvasDepth)) {
		return fmt.Errorf("terrain water level %.2f is outside depth 0..%d", *d.Terrain.WaterLevel, d.CanvasDepth)
	}
	return nil
}

func (d *Document) Marshal() ([]byte, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode terrain document: %w", err)
	}
	return append(data, '\n'), nil
}

// Save writes a validated document using atomic replacement in the same directory.
func (d *Document) Save(path string) error {
	data, err := d.Marshal()
	if err != nil {
		return err
	}
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".fray-terrain-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary terrain document: %w", err)
	}
	temporaryPath := temporary.Name()
	committed := false
	defer func() {
		temporary.Close()
		if !committed {
			_ = os.Remove(temporaryPath)
		}
	}()
	if _, err := temporary.Write(data); err != nil {
		return fmt.Errorf("write temporary terrain document: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("sync temporary terrain document: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary terrain document: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace terrain document: %w", err)
	}
	committed = true
	return nil
}

func (d *Document) HeightAt(x, y int) int {
	if x < 0 || y < 0 || x >= d.CanvasWidth || y >= d.CanvasHeight {
		return 0
	}
	index := y*d.CanvasWidth + x
	for z := len(d.Layers) - 1; z >= 0; z-- {
		if d.Layers[z][index] != 0 {
			return z + 1
		}
	}
	return 0
}

func (d *Document) normalize() {
	if d.Version < 3 && d.Terrain.Interpolation == "" {
		d.Terrain.Interpolation = "linear"
	}
}
