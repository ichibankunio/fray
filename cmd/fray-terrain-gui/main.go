package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/ichibankunio/fray"
	"github.com/ichibankunio/fray/terraineditor"
	"github.com/ichibankunio/fray/terraineditorruntime"
	"github.com/ichibankunio/fray/terraineditorui"
	"github.com/ichibankunio/fvec/vec2"
	"github.com/ichibankunio/fvec/vec3"
)

const (
	windowWidth  = 1100
	windowHeight = 700
	editorWidth  = 300
	tileSize     = 32
)

type editorGame struct {
	path      string
	renderer  *fray.Renderer
	session   *terraineditorruntime.Session
	panel     *terraineditorui.Panel
	gameImage *ebiten.Image
	layout    terraineditorui.SplitLayout
	status    string
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: fray-terrain-gui <terrain.json>")
		os.Exit(2)
	}
	game, err := newEditorGame(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "fray-terrain-gui:", err)
		os.Exit(1)
	}
	ebiten.SetWindowSize(windowWidth, windowHeight)
	ebiten.SetWindowTitle("fray terrain editor")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(true)
	if err := ebiten.RunGame(game); err != nil {
		fmt.Fprintln(os.Stderr, "fray-terrain-gui:", err)
		os.Exit(1)
	}
}

func newEditorGame(path string) (*editorGame, error) {
	document, err := terraineditor.Load(path)
	if err != nil {
		return nil, err
	}
	layout := terraineditorui.Split(imageRect(windowWidth, windowHeight), editorWidth)
	renderer := &fray.Renderer{}
	if err := renderer.InitWithError(float64(layout.Game.Dx()), float64(layout.Game.Dy()), document.CanvasWidth, document.CanvasHeight, document.CanvasDepth, tileSize); err != nil {
		return nil, err
	}
	if err := renderer.UseTerrainShader(); err != nil {
		return nil, fmt.Errorf("compile terrain shader: %w", err)
	}
	renderer.SetTerrainRenderScale(.33)
	renderer.SetPOVScale(1.5)
	renderer.SetCeilingHeight(0)
	renderer.SetCeilingTextureID(-1)
	renderer.SetAimHighlightEnabled(false)
	renderer.SetJumpButtonVisible(false)
	renderer.SetVerticalMovementEnabled(false)
	renderer.SetCameraInterferenceEnabled(false)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := renderer.Wld.LoadTerrainJSON(data); err != nil {
		return nil, err
	}
	gameWidth, gameHeight := layout.Game.Dx(), layout.Game.Dy()
	renderer.Textures = [4]*fray.ImageSrc{
		{Src: ebiten.NewImage(gameWidth, gameHeight)},
		renderer.NewTextureSheet(terrainTextures(tileSize)),
		{Src: ebiten.NewImage(gameWidth, gameHeight)},
		{Src: ebiten.NewImage(gameWidth, gameHeight)},
	}
	if err := renderer.SyncTerrainGPU(); err != nil {
		return nil, err
	}
	configureCamera(renderer, document)

	session, err := terraineditorruntime.New(document, renderer)
	if err != nil {
		return nil, err
	}
	panel := terraineditorui.NewPanel(session, layout.Editor)
	if err := panel.Watch(path); err != nil {
		return nil, err
	}
	return &editorGame{path: path, renderer: renderer, session: session, panel: panel, gameImage: ebiten.NewImage(gameWidth, gameHeight), layout: layout, status: "Ctrl+S save | WASD move | mouse look"}, nil
}

func (g *editorGame) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	x, y := ebiten.CursorPosition()
	g.renderer.SetControlInputEnabled(imagePoint(x, y).In(g.layout.Game))
	g.renderer.Update()
	if err := g.panel.Update(); err != nil {
		g.status = err.Error()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) && (ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)) {
		if err := g.session.Document.Save(g.path); err != nil {
			g.status = "save failed: " + err.Error()
		} else {
			g.session.History.MarkSaved()
			g.status = "saved " + g.path
		}
	}
	return nil
}

func (g *editorGame) Draw(screen *ebiten.Image) {
	g.gameImage.Clear()
	g.renderer.Draw(g.gameImage)
	screen.DrawImage(g.gameImage, nil)
	g.panel.Draw(screen)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%.1f FPS | %s", ebiten.ActualFPS(), g.status), 8, 8)
}

func (g *editorGame) Layout(_, _ int) (int, int) { return windowWidth, windowHeight }

func configureCamera(renderer *fray.Renderer, document *terraineditor.Document) {
	cameraX := max(1.0, float64(document.CanvasWidth)/2)
	cameraY := max(1.0, float64(document.CanvasHeight)-2)
	targetX := float64(document.CanvasWidth) / 2
	targetY := float64(document.CanvasHeight) / 2
	eyeHeight := float64(tileSize) * 2.5
	position := vec3.New(cameraX*tileSize, cameraY*tileSize, 0)
	position.Z = renderer.GetGroundHeight(position) + eyeHeight
	direction := vec2.New(targetX-cameraX, targetY-cameraY)
	direction = direction.Scale(1 / max(.001, direction.Length()))
	plane := vec2.New(direction.Y*.48, -direction.X*.48)
	renderer.Cam.SetDafaultDistanceBetweenSubjectCamera(0)
	renderer.Cam.SetShooterHeight(eyeHeight)
	renderer.Cam.SetPose(position, direction, plane, -70)
}

func terrainTextures(size int) []*ebiten.Image {
	colors := []color.RGBA{{74, 123, 75, 255}, {118, 99, 67, 255}, {101, 105, 102, 255}, {215, 222, 218, 255}}
	textures := make([]*ebiten.Image, len(colors))
	for index, clr := range colors {
		texture := ebiten.NewImage(size, size)
		texture.Fill(clr)
		textures[index] = texture
	}
	return textures
}

func imageRect(width, height int) image.Rectangle { return image.Rect(0, 0, width, height) }
func imagePoint(x, y int) image.Point             { return image.Pt(x, y) }
