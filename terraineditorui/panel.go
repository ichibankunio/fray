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
	Shape     string

	heightmap   *ebiten.Image
	baselineMap *ebiten.Image
	baseline    *terraineditor.Document
	review      bool
	diffSummary terraineditor.ChangeSummary
	pixels      []byte
	lastCell    image.Point
	hasCell     bool
	lastError   error
	watcher     *terraineditor.Watcher
	preview     terraineditor.ChangeSet
	previewKey  string
	commandJSON string
	selection   image.Rectangle
	selectStart image.Point
	selecting   bool
	section     bool
	cursorCell  image.Point
}

func NewPanel(session *terraineditorruntime.Session, bounds image.Rectangle) *Panel {
	panel := &Panel{Session: session, Bounds: bounds, Operation: "raise", Radius: 2, Amount: 1, Height: 4, Shape: "circle"}
	if session != nil && session.Document != nil {
		panel.baseline, _ = session.Document.Clone()
		panel.heightmap = ebiten.NewImage(session.Document.CanvasWidth, session.Document.CanvasHeight)
		panel.baselineMap = ebiten.NewImage(session.Document.CanvasWidth, session.Document.CanvasHeight)
		panel.updateHeightmap(image.Rect(0, 0, session.Document.CanvasWidth, session.Document.CanvasHeight))
		panel.updateBaselineMap()
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
	if inside {
		p.cursorCell = cell
	}
	if p.updateSelection(cell, inside) {
		return p.lastError
	}
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
	command := terraineditor.Command{Operation: operation, Parameters: terraineditor.Parameters{X: cell.X, Y: cell.Y, Radius: p.Radius, Amount: p.Amount, Height: p.Height, Blend: .5, Falloff: "smooth", Shape: p.Shape}}
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
	p.baselineMap = ebiten.NewImage(document.CanvasWidth, document.CanvasHeight)
	p.baseline, _ = document.Clone()
	p.updateHeightmap(image.Rect(0, 0, document.CanvasWidth, document.CanvasHeight))
	p.updateBaselineMap()
	p.lastError = p.Session.SyncGPU()
}

func (p *Panel) Draw(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, float64(p.Bounds.Min.X), float64(p.Bounds.Min.Y), float64(p.Bounds.Dx()), float64(p.Bounds.Dy()), color.RGBA{28, 32, 29, 255})
	mapBounds := p.mapBounds()
	if p.review && p.baselineMap != nil {
		before := p.beforeMapBounds()
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(before.Dx())/float64(p.baselineMap.Bounds().Dx()), float64(before.Dy())/float64(p.baselineMap.Bounds().Dy()))
		op.GeoM.Translate(float64(before.Min.X), float64(before.Min.Y))
		screen.DrawImage(p.baselineMap, op)
		ebitenutil.DebugPrintAt(screen, "BEFORE", before.Min.X, before.Min.Y)
		ebitenutil.DebugPrintAt(screen, "AFTER", mapBounds.Min.X, mapBounds.Min.Y)
	}
	if p.heightmap != nil && !mapBounds.Empty() {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(mapBounds.Dx())/float64(p.heightmap.Bounds().Dx()), float64(mapBounds.Dy())/float64(p.heightmap.Bounds().Dy()))
		op.GeoM.Translate(float64(mapBounds.Min.X), float64(mapBounds.Min.Y))
		screen.DrawImage(p.heightmap, op)
	}
	p.drawPreview(screen, mapBounds)
	p.drawSelection(screen, mapBounds)
	p.drawSection(screen, mapBounds)
	s := p.diffSummary
	diffText := fmt.Sprintf(" Changed:%d Delta:%d..%d", s.ChangedCells, s.MinimumDelta, s.MaximumDelta)
	text := fmt.Sprintf("TERRAIN%s\n1 Set 2 Raise 3 Lower 4 Flat 5 Smooth\nTool:%s R:%d A:%d H:%d Shape:%s\nB shape; R review; X section\nShift+drag select; Enter apply\nH/V flip; Alt/Ctrl+arrow move/copy\nZ undo Y redo\nCommand: %s", diffText, p.Operation, p.Radius, p.Amount, p.Height, p.Shape, p.commandJSON)
	if p.lastError != nil {
		text += "\nERROR: " + p.lastError.Error()
	}
	ebitenutil.DebugPrintAt(screen, text, p.Bounds.Min.X+8, mapBounds.Max.Y+8)
}

func (p *Panel) updateSelection(cell image.Point, inside bool) bool {
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)
	left := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	if shift && left && inside {
		if !p.selecting {
			p.selectStart, p.selecting = cell, true
		}
		p.selection = inclusiveRect(p.selectStart, cell)
		p.preview = terraineditor.ChangeSet{}
		return true
	}
	if p.selecting {
		p.selecting = false
		return true
	}
	return false
}

func inclusiveRect(a, b image.Point) image.Rectangle {
	return image.Rect(min(a.X, b.X), min(a.Y, b.Y), max(a.X, b.X)+1, max(a.Y, b.Y)+1)
}

func (p *Panel) drawSelection(screen *ebiten.Image, bounds image.Rectangle) {
	if p.selection.Empty() || bounds.Empty() {
		return
	}
	d := p.Session.Document
	x0 := bounds.Min.X + p.selection.Min.X*bounds.Dx()/d.CanvasWidth
	y0 := bounds.Min.Y + p.selection.Min.Y*bounds.Dy()/d.CanvasHeight
	x1 := bounds.Min.X + p.selection.Max.X*bounds.Dx()/d.CanvasWidth
	y1 := bounds.Min.Y + p.selection.Max.Y*bounds.Dy()/d.CanvasHeight
	clr := color.RGBA{75, 170, 255, 255}
	ebitenutil.DrawRect(screen, float64(x0), float64(y0), float64(x1-x0), 1, clr)
	ebitenutil.DrawRect(screen, float64(x0), float64(y1-1), float64(x1-x0), 1, clr)
	ebitenutil.DrawRect(screen, float64(x0), float64(y0), 1, float64(y1-y0), clr)
	ebitenutil.DrawRect(screen, float64(x1-1), float64(y0), 1, float64(y1-y0), clr)
}

func (p *Panel) applySelectionCommand(operation string, to image.Point, axis string) {
	if p.selection.Empty() {
		return
	}
	params := terraineditor.Parameters{X: p.selection.Min.X, Y: p.selection.Min.Y, Width: p.selection.Dx(), Rows: p.selection.Dy(), Height: p.Height, Amount: p.Amount, Blend: .5, Falloff: "smooth", ToX: to.X, ToY: to.Y, Axis: axis}
	changeSet, err := p.Session.Apply(terraineditor.Command{Operation: operation, Parameters: params})
	if err == nil {
		p.updateHeightmap(changeSet.Region)
		err = p.Session.SyncGPU()
	}
	p.lastError = err
}

func (p *Panel) updatePreview(cell image.Point, inside bool, operation string) {
	if !inside {
		p.preview, p.previewKey, p.commandJSON = terraineditor.ChangeSet{}, "", ""
		return
	}
	command := terraineditor.Command{Operation: operation, Parameters: terraineditor.Parameters{X: cell.X, Y: cell.Y, Radius: p.Radius, Amount: p.Amount, Height: p.Height, Blend: .5, Falloff: "smooth", Shape: p.Shape}}
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
	if inpututil.IsKeyJustPressed(ebiten.KeyX) {
		p.section = !p.section
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		shapes := []string{"circle", "square", "diamond", "noise"}
		for i, shape := range shapes {
			if shape == p.Shape {
				p.Shape = shapes[(i+1)%len(shapes)]
				break
			}
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		p.review = !p.review
	}
	if !p.selection.Empty() {
		if inpututil.IsKeyJustPressed(ebiten.KeyC) {
			p.selection = image.Rectangle{}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			p.applySelectionCommand(p.Operation, image.Point{}, "")
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyH) {
			p.applySelectionCommand("flip-region", image.Point{}, "horizontal")
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyV) {
			p.applySelectionCommand("flip-region", image.Point{}, "vertical")
		}
		for _, move := range []struct {
			key   ebiten.Key
			delta image.Point
		}{{ebiten.KeyArrowLeft, image.Pt(-1, 0)}, {ebiten.KeyArrowRight, image.Pt(1, 0)}, {ebiten.KeyArrowUp, image.Pt(0, -1)}, {ebiten.KeyArrowDown, image.Pt(0, 1)}} {
			if !inpututil.IsKeyJustPressed(move.key) {
				continue
			}
			to := p.selection.Min.Add(move.delta)
			if ebiten.IsKeyPressed(ebiten.KeyAlt) {
				p.applySelectionCommand("move-region", to, "")
				if p.lastError == nil {
					p.selection = p.selection.Add(move.delta)
				}
			}
			if ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta) {
				p.applySelectionCommand("copy-region", to, "")
				if p.lastError == nil {
					p.selection = p.selection.Add(move.delta)
				}
			}
		}
	}
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
	if p.selection.Empty() && inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		p.Height = min(p.Session.Document.CanvasDepth, p.Height+1)
	}
	if p.selection.Empty() && inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
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

func (p *Panel) drawSection(screen *ebiten.Image, bounds image.Rectangle) {
	if !p.section || bounds.Empty() || p.Session == nil {
		return
	}
	graph := image.Rect(bounds.Min.X, bounds.Max.Y-90, bounds.Max.X, bounds.Max.Y)
	ebitenutil.DrawRect(screen, float64(graph.Min.X), float64(graph.Min.Y), float64(graph.Dx()), float64(graph.Dy()), color.RGBA{8, 12, 10, 220})
	d := p.Session.Document
	samples := max(2, graph.Dx())
	previous := image.Point{}
	for i := 0; i < samples; i++ {
		x := float64(i) * float64(d.CanvasWidth-1) / float64(samples-1)
		h := p.Session.Renderer.Wld.SampleTerrainHeight(x, float64(p.cursorCell.Y)+.5)
		point := image.Pt(graph.Min.X+i, graph.Max.Y-1-int(h/float64(d.CanvasDepth)*float64(graph.Dy()-12)))
		if i > 0 {
			ebitenutil.DrawLine(screen, float64(previous.X), float64(previous.Y), float64(point.X), float64(point.Y), color.RGBA{100, 235, 145, 255})
		}
		previous = point
	}
	for x := 0; x < d.CanvasWidth; x++ {
		px := graph.Min.X + x*graph.Dx()/max(1, d.CanvasWidth-1)
		h := d.HeightAt(x, p.cursorCell.Y)
		py := graph.Max.Y - 1 - h*(graph.Dy()-12)/d.CanvasDepth
		ebitenutil.DrawRect(screen, float64(px), float64(py), 2, 2, color.White)
	}
}

func (p *Panel) mapBounds() image.Rectangle {
	size := min(p.Bounds.Dx()-16, p.Bounds.Dy()-150)
	if size <= 0 {
		return image.Rectangle{}
	}
	if p.review {
		size = (p.Bounds.Dx() - 24) / 2
		return image.Rect(p.Bounds.Min.X+16+size, p.Bounds.Min.Y+8, p.Bounds.Min.X+16+size*2, p.Bounds.Min.Y+8+size)
	}
	return image.Rect(p.Bounds.Min.X+8, p.Bounds.Min.Y+8, p.Bounds.Min.X+8+size, p.Bounds.Min.Y+8+size)
}

func (p *Panel) beforeMapBounds() image.Rectangle {
	size := (p.Bounds.Dx() - 24) / 2
	return image.Rect(p.Bounds.Min.X+8, p.Bounds.Min.Y+8, p.Bounds.Min.X+8+size, p.Bounds.Min.Y+8+size)
}

func (p *Panel) MarkSaved() {
	p.baseline, _ = p.Session.Document.Clone()
	p.updateBaselineMap()
	p.diffSummary = terraineditor.ChangeSummary{}
}

func (p *Panel) refreshDiffSummary() {
	if p.baseline == nil {
		return
	}
	changes, err := terraineditor.Diff(p.baseline, p.Session.Document)
	if err == nil {
		p.diffSummary = terraineditor.Summarize(changes)
	}
}

func (p *Panel) updateBaselineMap() {
	if p.baselineMap == nil || p.baseline == nil {
		return
	}
	pixels := make([]byte, p.baseline.CanvasWidth*p.baseline.CanvasHeight*4)
	for y := 0; y < p.baseline.CanvasHeight; y++ {
		for x := 0; x < p.baseline.CanvasWidth; x++ {
			h := p.baseline.HeightAt(x, y)
			value := uint8(35 + 190*h/max(1, p.baseline.CanvasDepth))
			i := (y*p.baseline.CanvasWidth + x) * 4
			pixels[i], pixels[i+1], pixels[i+2], pixels[i+3] = value/2, value, value/2, 255
		}
	}
	p.baselineMap.WritePixels(pixels)
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
	p.refreshDiffSummary()
}
