package terraineditor

import (
	"encoding/json"
	"fmt"
	"os"
)

type historyEntry struct {
	command Command
	changes ChangeSet
}

type History struct {
	document *Document
	undo     []historyEntry
	redo     []historyEntry
	dirty    bool
}

func NewHistory(document *Document) *History { return &History{document: document} }

func (h *History) Apply(command Command) (ChangeSet, error) {
	changeSet, err := h.document.Apply(command)
	if err != nil {
		return ChangeSet{}, err
	}
	if len(changeSet.Changes) > 0 {
		h.undo = append(h.undo, historyEntry{command: command, changes: changeSet})
		h.redo = h.redo[:0]
		h.dirty = true
	}
	return changeSet, nil
}

func (h *History) Undo() (ChangeSet, bool, error) {
	if len(h.undo) == 0 {
		return ChangeSet{}, false, nil
	}
	entry := h.undo[len(h.undo)-1]
	h.undo = h.undo[:len(h.undo)-1]
	if err := h.document.ApplyChangeSet(entry.changes, false); err != nil {
		return ChangeSet{}, false, err
	}
	h.redo = append(h.redo, entry)
	h.dirty = true
	return entry.changes, true, nil
}

func (h *History) Redo() (ChangeSet, bool, error) {
	if len(h.redo) == 0 {
		return ChangeSet{}, false, nil
	}
	entry := h.redo[len(h.redo)-1]
	h.redo = h.redo[:len(h.redo)-1]
	if err := h.document.ApplyChangeSet(entry.changes, true); err != nil {
		return ChangeSet{}, false, err
	}
	h.undo = append(h.undo, entry)
	h.dirty = true
	return entry.changes, true, nil
}

func (h *History) Commands() []Command {
	commands := make([]Command, len(h.undo))
	for i, entry := range h.undo {
		commands[i] = entry.command
	}
	return commands
}

func (h *History) SaveCommands(path string) error {
	data, err := json.MarshalIndent(h.Commands(), "", "  ")
	if err != nil {
		return fmt.Errorf("encode terrain history: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write terrain history: %w", err)
	}
	return nil
}

func LoadCommands(path string) ([]Command, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read terrain history: %w", err)
	}
	var commands []Command
	if err := json.Unmarshal(data, &commands); err != nil {
		return nil, fmt.Errorf("decode terrain history: %w", err)
	}
	return commands, nil
}

func Replay(document *Document, commands []Command, constraints Constraints) ([]ChangeSet, error) {
	changes := make([]ChangeSet, 0, len(commands))
	for index, command := range commands {
		changeSet, err := document.ApplyConstrained(command, constraints)
		if err != nil {
			return changes, fmt.Errorf("replay command %d: %w", index, err)
		}
		changes = append(changes, changeSet)
	}
	return changes, nil
}

func (h *History) CanUndo() bool           { return len(h.undo) > 0 }
func (h *History) CanRedo() bool           { return len(h.redo) > 0 }
func (h *History) HasUnsavedChanges() bool { return h.dirty }
func (h *History) MarkSaved()              { h.dirty = false }
