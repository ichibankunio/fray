package sandbox

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/ichibankunio/fray"
	"github.com/ichibankunio/fray/mapeditor"
	"github.com/ichibankunio/fray/ui"
	"github.com/ichibankunio/fui"
)

type SandboxManager struct {
	Renderer  *fray.Renderer
	MapEditor *mapeditor.MapEditor
	UIManager *ui.UIManager

	place  *fui.Button
	delete *fui.Button
}

func (sm *SandboxManager) Init() {
	sm.Renderer = &fray.Renderer{}
	sm.MapEditor = &mapeditor.MapEditor{}
	sm.UIManager = &ui.UIManager{}
}

func (sm *SandboxManager) PrintHeightMap(screen *ebiten.Image) {
	for i := 0; i < len(sm.Renderer.Wld.HeightMap); i++ {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", sm.Renderer.Wld.HeightMap[i]), (i%128)*16, (i/128)*16)
	}
}

func (sm *SandboxManager) PrintWorldMap(screen *ebiten.Image) {
	for i := 0; i < len(sm.Renderer.Wld.WorldMap[0]); i++ {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", sm.Renderer.Wld.WorldMap[0][i]), (i%128)*16, (i/128)*16)
	}
}

func (sm *SandboxManager) PlaceBlock() {
	aimPos := sm.Renderer.GetAimPosition()
	if aimPos.X < 0 {
		return
	}

	x := int(aimPos.X)
	y := int(aimPos.Y)
	z := int(aimPos.Z)

	switch sm.Renderer.GetAimDirection() {
	case fray.AIM_DIR_NORTH:
		sm.Renderer.Wld.SetValue(x, y-1, z, uint8(sm.Renderer.HandTextureID))
	case fray.AIM_DIR_SOUTH:
		sm.Renderer.Wld.SetValue(x, y+1, z, uint8(sm.Renderer.HandTextureID))
	case fray.AIM_DIR_EAST:
		sm.Renderer.Wld.SetValue(x-1, y, z, uint8(sm.Renderer.HandTextureID))
	case fray.AIM_DIR_WEST:
		sm.Renderer.Wld.SetValue(x+1, y, z, uint8(sm.Renderer.HandTextureID))
	case fray.AIM_DIR_TOP:
		sm.Renderer.Wld.SetValue(x, y, z+1, uint8(sm.Renderer.HandTextureID))
	}

	sm.MapEditor.PrintHeightMapOnAlphaLayer(sm.Renderer.Wld.HeightMap, sm.Renderer.Textures[1].Src)
	sm.MapEditor.PrintWorldMap(sm.Renderer.Wld.WorldMap, sm.Renderer.Textures[2].Src)
}

func (sm *SandboxManager) DeleteBlock() {
	aimPos := sm.Renderer.GetAimPosition()
	if aimPos.X < 0 {
		return
	}

	x := int(aimPos.X)
	y := int(aimPos.Y)
	z := int(aimPos.Z)

	sm.Renderer.Wld.DeleteValue(x, y, z)
	sm.MapEditor.PrintHeightMapOnAlphaLayer(sm.Renderer.Wld.HeightMap, sm.Renderer.Textures[1].Src)
	sm.MapEditor.PrintWorldMap(sm.Renderer.Wld.WorldMap, sm.Renderer.Textures[2].Src)
}

func (sm *SandboxManager) UpdateMapEdit() {
	aimPos := sm.Renderer.GetAimPosition()
	if aimPos.X < 0 {
		return
	}

	x := int(aimPos.X)
	y := int(aimPos.Y)
	z := int(aimPos.Z)

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) { //左クリック: 置く
		switch sm.Renderer.GetAimDirection() {
		case fray.AIM_DIR_NORTH:
			sm.Renderer.Wld.SetValue(x, y-1, z, uint8(sm.Renderer.HandTextureID))
		case fray.AIM_DIR_SOUTH:
			sm.Renderer.Wld.SetValue(x, y+1, z, uint8(sm.Renderer.HandTextureID))
		case fray.AIM_DIR_EAST:
			sm.Renderer.Wld.SetValue(x-1, y, z, uint8(sm.Renderer.HandTextureID))
		case fray.AIM_DIR_WEST:
			sm.Renderer.Wld.SetValue(x+1, y, z, uint8(sm.Renderer.HandTextureID))
		case fray.AIM_DIR_TOP:
			sm.Renderer.Wld.SetValue(x, y, z+1, uint8(sm.Renderer.HandTextureID))
		}

		sm.MapEditor.PrintHeightMapOnAlphaLayer(sm.Renderer.Wld.HeightMap, sm.Renderer.Textures[1].Src)
		sm.MapEditor.PrintWorldMap(sm.Renderer.Wld.WorldMap, sm.Renderer.Textures[2].Src)
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		sm.Renderer.Wld.DeleteValue(x, y, z)
		sm.MapEditor.PrintHeightMapOnAlphaLayer(sm.Renderer.Wld.HeightMap, sm.Renderer.Textures[1].Src)
		sm.MapEditor.PrintWorldMap(sm.Renderer.Wld.WorldMap, sm.Renderer.Textures[2].Src)
	}
}

func (sm *SandboxManager) Update() error {
	sm.Renderer.Update()

	// sm.UpdateMapEdit()

	if inpututil.IsKeyJustPressed(ebiten.KeyK) {
		sm.Renderer.Cam.Speed = 20
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyO) {
		sm.Renderer.Cam.Speed = 2.0
	}

	return nil
}

func (sm *SandboxManager) Draw(screen *ebiten.Image) {
	sm.Renderer.Draw(screen)

	if ebiten.IsKeyPressed(ebiten.KeyH) {
		sm.PrintHeightMap(screen)
	} else if ebiten.IsKeyPressed(ebiten.KeyM) {
		sm.PrintWorldMap(screen)
	}
}
