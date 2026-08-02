// Package terraineditorruntime connects headless editor documents to a live fray renderer.
package terraineditorruntime

import (
	"fmt"
	"time"

	"github.com/ichibankunio/fray"
	"github.com/ichibankunio/fray/terraineditor"
)

type Session struct {
	Document *terraineditor.Document
	History  *terraineditor.History
	Renderer *fray.Renderer
	Metrics  Metrics
}

type Metrics struct {
	Edits        uint64
	ChangedCells uint64
	LastEdit     time.Duration
	MaximumEdit  time.Duration
	LastGPU      time.Duration
	MaximumGPU   time.Duration
}

func New(document *terraineditor.Document, renderer *fray.Renderer) (*Session, error) {
	if document == nil || renderer == nil || renderer.Wld == nil {
		return nil, fmt.Errorf("terrain editor runtime requires document and initialized renderer")
	}
	world := renderer.Wld
	if document.CanvasWidth != world.CanvasWidth() || document.CanvasHeight != world.CanvasHeight() || document.CanvasDepth > world.CanvasDepth() {
		return nil, fmt.Errorf("editor terrain %dx%dx%d does not match world %dx%dx%d", document.CanvasWidth, document.CanvasHeight, document.CanvasDepth, world.CanvasWidth(), world.CanvasHeight(), world.CanvasDepth())
	}
	return &Session{Document: document, History: terraineditor.NewHistory(document), Renderer: renderer}, nil
}

// Apply updates the editor document and live CPU terrain in one transaction.
func (s *Session) Apply(command terraineditor.Command) (terraineditor.ChangeSet, error) {
	started := time.Now()
	changeSet, err := s.History.Apply(command)
	if err != nil {
		return terraineditor.ChangeSet{}, err
	}
	s.applyColumns(changeSet)
	s.recordEdit(time.Since(started), len(changeSet.Changes))
	return changeSet, nil
}

func (s *Session) Undo() (terraineditor.ChangeSet, bool, error) {
	changeSet, ok, err := s.History.Undo()
	if err == nil && ok {
		s.applyColumns(changeSet)
	}
	return changeSet, ok, err
}

func (s *Session) Redo() (terraineditor.ChangeSet, bool, error) {
	changeSet, ok, err := s.History.Redo()
	if err == nil && ok {
		s.applyColumns(changeSet)
	}
	return changeSet, ok, err
}

// SyncGPU uploads the merged dirty region. Call once after all edits in a frame.
func (s *Session) SyncGPU() error {
	started := time.Now()
	err := s.Renderer.SyncTerrainGPU()
	elapsed := time.Since(started)
	s.Metrics.LastGPU = elapsed
	if elapsed > s.Metrics.MaximumGPU {
		s.Metrics.MaximumGPU = elapsed
	}
	return err
}

func (s *Session) recordEdit(elapsed time.Duration, cells int) {
	s.Metrics.Edits++
	s.Metrics.ChangedCells += uint64(cells)
	s.Metrics.LastEdit = elapsed
	if elapsed > s.Metrics.MaximumEdit {
		s.Metrics.MaximumEdit = elapsed
	}
}

// Reload replaces the live document after an externally edited file validates.
func (s *Session) Reload(document *terraineditor.Document) error {
	if document == nil {
		return fmt.Errorf("reload terrain editor: document is nil")
	}
	world := s.Renderer.Wld
	if document.CanvasWidth != world.CanvasWidth() || document.CanvasHeight != world.CanvasHeight() || document.CanvasDepth > world.CanvasDepth() {
		return fmt.Errorf("reloaded terrain dimensions do not match live world")
	}
	s.Document = document
	s.History = terraineditor.NewHistory(document)
	for z := 0; z < world.CanvasDepth(); z++ {
		clear(world.WorldMap[z])
		if z < len(document.Layers) {
			copy(world.WorldMap[z], document.Layers[z])
		}
	}
	world.RebuildTerrain()
	return nil
}

func (s *Session) applyColumns(changeSet terraineditor.ChangeSet) {
	world := s.Renderer.Wld
	for _, change := range changeSet.Changes {
		index := change.Y*world.CanvasWidth() + change.X
		for z := 0; z < world.CanvasDepth(); z++ {
			value := uint8(0)
			if z < len(s.Document.Layers) {
				value = s.Document.Layers[z][change.Y*s.Document.CanvasWidth+change.X]
			}
			world.WorldMap[z][index] = value
		}
	}
	if !changeSet.Region.Empty() {
		world.RebuildTerrainRegion(changeSet.Region)
	}
}
