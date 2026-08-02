package terraineditorui

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/ichibankunio/fray/terraineditor"
	"github.com/ichibankunio/fray/terraineditorruntime"
)

type Panel struct {
	Session   *terraineditorruntime.Session
	Bounds    image.Rectangle
	Operation string
	Radius    int
	Amount    int
	Height    int

	heightmap   *ebiten.Image
	pixels      []byte
	lastCell    image.Point
	hasCell     bool
	lastError   error
	watcher     *terraineditor.Watcher
	preview     terraineditor.ChangeSet
	previewKey  string
	commandJSON string
}

func NewPanel(session *terraineditorruntime.Session, bounds image.Rectangle) *Panel {
	panel := &Panel{Session: session, Bounds: bounds, Operation: "raise", Radius: 2, Amount: 1, Height: 4}
	if session != nil && session.Document != nil {
		panel.heightmap = ebiten.NewImage(session.Document.CanvasWidth, session.Document.CanvasHeight)
		panel.updateHeightmap(image.Rect(0, 0, session.Document.CanvasWidth, session.Document.CanvasHeight))
	}
	return panel
}

func (p *Panel) Update() error {
	if p.Session == nil || p.Session.Document == nil {
		return nil
	}
	p.updateShortcuts()
	p.pollExternalDocument()
	mouse := image.Pt(ebiten.CursorPosition())
	cell, inside := p.cellAt(mouse)
	pressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	if !pressed {
		p.hasCell = false
		p.updatePreview(cell, inside, p.Operation)
		return p.lastError
	}
	if !inside || p.hasCell && cell == p.lastCell {
		return p.lastError
	}
	operation := p.Operation
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		operation = "lower"
	}
	command := terraineditor.Command{Operation: operation, Parameters: terraineditor.Parameters{X: cell.X, Y: cell.Y, Radius: p.Radius, Amount: p.Amount, Height: p.Height, Blend: .5, Falloff: "smooth"}}
	p.preview = terraineditor.ChangeSet{}
	p.previewKey = ""
	changeSet, err := p.Session.Apply(command)
	if err == nil {
		p.updateHeightmap(changeSet.Region)
		err = p.Session.SyncGPU()
	}
	p.lastCell, p.hasCell, p.lastError = cell, true, err
	return err
}

// Watch reloads valid external CLI edits. Unsaved GUI history prevents an
// automatic reload so human edits are never silently overwritten.
func (p *Panel) Watch(path string) error {
	watcher, err := terraineditor.NewWatcher(path)
	if err != nil {
		return err
	}
	p.watcher = watcher
	return nil
}

func (p *Panel) pollExternalDocument() {
	if p.watcher == nil {
		return
	}
	document, changed, err := p.watcher.Poll()
	if err != nil || !changed {
		p.lastError = err
		return
	}
	if p.Session.History.HasUnsavedChanges() {
		p.lastError = fmt.Errorf("external terrain changed while GUI has unsaved edits")
		return
	}
	if err := p.Session.Reload(document); err != nil {
		p.lastError = err
		return
	}
	p.watcher.Accept()
	p.heightmap = ebiten.NewImage(document.CanvasWidth, document.CanvasHeight)
	p.updateHeightmap(image.Rect(0, 0, document.CanvasWidth, document.CanvasHeight))
	p.lastError = p.Session.SyncGPU()
}

func (p *Panel) Draw(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, float64(p.Bounds.Min.X), float64(p.Bounds.Min.Y), float64(p.Bounds.Dx()), float64(p.Bounds.Dy()), color.RGBA{28, 32, 29, 255})
	mapBounds := p.mapBounds()
	if p.heightmap != nil && !mapBounds.Empty() {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(mapBounds.Dx())/float64(p.heightmap.Bounds().Dx()), float64(mapBounds.Dy())/float64(p.heightmap.Bounds().Dy()))
		op.GeoM.Translate(float64(mapBounds.Min.X), float64(mapBounds.Min.Y))
		screen.DrawImage(p.heightmap, op)
	}
	p.drawPreview(screen, mapBounds)
	text := fmt.Sprintf("TERRAIN\n1 Set  2 Raise  3 Lower\n4 Flat 5 Smooth\nTool: %s  R:%d A:%d H:%d\n[ / ] radius   - / = amount\nUp / Down height\nZ undo  Y redo\nLMB apply  RMB lower\nCommand: %s", p.Operation, p.Radius, p.Amount, p.Height, p.commandJSON)
	if p.lastError != nil {
		text += "\nERROR: " + p.lastError.Error()
	}
	ebitenutil.DebugPrintAt(screen, text, p.Bounds.Min.X+8, mapBounds.Max.Y+8)
}

func (p *Panel) updatePreview(cell image.Point, inside bool, operation string) {
	if !inside {
		p.preview, p.previewKey, p.commandJSON = terraineditor.ChangeSet{}, "", ""
		return
	}
	command := terraineditor.Command{Operation: operation, Parameters: terraineditor.Parameters{X: cell.X, Y: cell.Y, Radius: p.Radius, Amount: p.Amount, Height: p.Height, Blend: .5, Falloff: "smooth"}}
	data, _ := json.Marshal(command)
	key := string(data)
	if key == p.previewKey {
		return
	}
	p.previewKey, p.commandJSON = key, key
	p.preview, p.lastError = p.Session.Document.Preview(command)
}

func (p *Panel) drawPreview(screen *ebiten.Image, bounds image.Rectangle) {
	if bounds.Empty() || p.Session == nil || p.Session.Document == nil {
		return
	}
	document := p.Session.Document
	for _, change := range p.preview.Changes {
		before, after := terraineditor.ColumnHeight(change.Before), terraineditor.ColumnHeight(change.After)
		shade := color.RGBA{235, 196, 55, 150}
		if after > before {
			shade = color.RGBA{80, 230, 115, 150}
		}
		if after < before {
			shade = color.RGBA{235, 85, 65, 150}
		}
		x0 := bounds.Min.X + change.X*bounds.Dx()/document.CanvasWidth
		y0 := bounds.Min.Y + change.Y*bounds.Dy()/document.CanvasHeight
		x1 := bounds.Min.X + (change.X+1)*bounds.Dx()/document.CanvasWidth
		y1 := bounds.Min.Y + (change.Y+1)*bounds.Dy()/document.CanvasHeight
		ebitenutil.DrawRect(screen, float64(x0), float64(y0), float64(max(1, x1-x0)), float64(max(1, y1-y0)), shade)
	}
}

func (p *Panel) updateShortcuts() {
	operations := []struct {
		key       ebiten.Key
		operation string
	}{{ebiten.Key1, "set-height"}, {ebiten.Key2, "raise"}, {ebiten.Key3, "lower"}, {ebiten.Key4, "flatten"}, {ebiten.Key5, "smooth"}}
	for _, item := range operations {
		if inpututil.IsKeyJustPressed(item.key) {
			p.Operation = item.operation
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBracketLeft) {
		p.Radius = max(0, p.Radius-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBracketRight) {
		p.Radius = min(32, p.Radius+1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) {
		p.Amount = max(1, p.Amount-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) {
		p.Amount = min(16, p.Amount+1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		p.Height = min(p.Session.Document.CanvasDepth, p.Height+1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		p.Height = max(0, p.Height-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		if changeSet, ok, err := p.Session.Undo(); err != nil {
			p.lastError = err
		} else if ok {
			p.updateHeightmap(changeSet.Region)
			p.lastError = p.Session.SyncGPU()
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyY) {
		if changeSet, ok, err := p.Session.Redo(); err != nil {
			p.lastError = err
		} else if ok {
			p.updateHeightmap(changeSet.Region)
			p.lastError = p.Session.SyncGPU()
		}
	}
}

func (p *Panel) mapBounds() image.Rectangle {
	size := min(p.Bounds.Dx()-16, p.Bounds.Dy()-150)
	if size <= 0 {
		return image.Rectangle{}
	}
	return image.Rect(p.Bounds.Min.X+8, p.Bounds.Min.Y+8, p.Bounds.Min.X+8+size, p.Bounds.Min.Y+8+size)
}

func (p *Panel) cellAt(point image.Point) (image.Point, bool) {
	bounds := p.mapBounds()
	if !point.In(bounds) {
		return image.Point{}, false
	}
	document := p.Session.Document
	x := (point.X - bounds.Min.X) * document.CanvasWidth / bounds.Dx()
	y := (point.Y - bounds.Min.Y) * document.CanvasHeight / bounds.Dy()
	return image.Pt(min(document.CanvasWidth-1, x), min(document.CanvasHeight-1, y)), true
}

func (p *Panel) updateHeightmap(region image.Rectangle) {
	if p.heightmap == nil || region.Empty() {
		return
	}
	document := p.Session.Document
	region = region.Intersect(image.Rect(0, 0, document.CanvasWidth, document.CanvasHeight))
	required := region.Dx() * region.Dy() * 4
	if cap(p.pixels) < required {
		p.pixels = make([]byte, required)
	} else {
		p.pixels = p.pixels[:required]
	}
	pixels := p.pixels
	for y := region.Min.Y; y < region.Max.Y; y++ {
		for x := region.Min.X; x < region.Max.X; x++ {
			height := document.HeightAt(x, y)
			value := uint8(35 + 190*height/max(1, document.CanvasDepth))
			index := ((y-region.Min.Y)*region.Dx() + x - region.Min.X) * 4
			pixels[index], pixels[index+1], pixels[index+2], pixels[index+3] = value/2, value, value/2, 255
		}
	}
	p.heightmap.SubImage(region).(*ebiten.Image).WritePixels(pixels)
}
