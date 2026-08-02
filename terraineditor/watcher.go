package terraineditor

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
)

type Watcher struct {
	path        string
	hash        [sha256.Size]byte
	pendingHash [sha256.Size]byte
}

func NewWatcher(path string) (*Watcher, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("watch terrain document: %w", err)
	}
	return &Watcher{path: path, hash: sha256.Sum256(data)}, nil
}

// Poll loads and validates the document only when its content has changed.
func (w *Watcher) Poll() (*Document, bool, error) {
	data, err := os.ReadFile(w.path)
	if err != nil {
		return nil, false, fmt.Errorf("poll terrain document: %w", err)
	}
	hash := sha256.Sum256(data)
	if bytes.Equal(hash[:], w.hash[:]) {
		return nil, false, nil
	}
	document, err := Parse(data)
	if err != nil {
		return nil, false, err
	}
	w.pendingHash = hash
	return document, true, nil
}

// Accept marks the most recently polled document as applied.
func (w *Watcher) Accept() { w.hash = w.pendingHash }
