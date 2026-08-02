// Package terraineditorui provides an optional Ebitengine editor panel.
package terraineditorui

import "image"

type SplitLayout struct {
	Game   image.Rectangle
	Editor image.Rectangle
}

func Split(bounds image.Rectangle, editorWidth int) SplitLayout {
	editorWidth = max(160, min(editorWidth, bounds.Dx()/2))
	division := bounds.Max.X - editorWidth
	return SplitLayout{Game: image.Rect(bounds.Min.X, bounds.Min.Y, division, bounds.Max.Y), Editor: image.Rect(division, bounds.Min.Y, bounds.Max.X, bounds.Max.Y)}
}
