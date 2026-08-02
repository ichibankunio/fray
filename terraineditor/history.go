package terraineditor

type History struct {
	document *Document
	undo     []ChangeSet
	redo     []ChangeSet
	dirty    bool
}

func NewHistory(document *Document) *History { return &History{document: document} }

func (h *History) Apply(command Command) (ChangeSet, error) {
	changeSet, err := h.document.Apply(command)
	if err != nil {
		return ChangeSet{}, err
	}
	if len(changeSet.Changes) > 0 {
		h.undo = append(h.undo, changeSet)
		h.redo = h.redo[:0]
		h.dirty = true
	}
	return changeSet, nil
}

func (h *History) Undo() (ChangeSet, bool, error) {
	if len(h.undo) == 0 {
		return ChangeSet{}, false, nil
	}
	changeSet := h.undo[len(h.undo)-1]
	h.undo = h.undo[:len(h.undo)-1]
	if err := h.document.ApplyChangeSet(changeSet, false); err != nil {
		return ChangeSet{}, false, err
	}
	h.redo = append(h.redo, changeSet)
	h.dirty = true
	return changeSet, true, nil
}

func (h *History) Redo() (ChangeSet, bool, error) {
	if len(h.redo) == 0 {
		return ChangeSet{}, false, nil
	}
	changeSet := h.redo[len(h.redo)-1]
	h.redo = h.redo[:len(h.redo)-1]
	if err := h.document.ApplyChangeSet(changeSet, true); err != nil {
		return ChangeSet{}, false, err
	}
	h.undo = append(h.undo, changeSet)
	h.dirty = true
	return changeSet, true, nil
}

func (h *History) CanUndo() bool           { return len(h.undo) > 0 }
func (h *History) CanRedo() bool           { return len(h.redo) > 0 }
func (h *History) HasUnsavedChanges() bool { return h.dirty }
func (h *History) MarkSaved()              { h.dirty = false }
