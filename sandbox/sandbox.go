package sandbox

import (
	"fmt"
	"image/png"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/ichibankunio/fray"
	"github.com/ichibankunio/fray/mapeditor"
	"github.com/ichibankunio/fray/ui"
)

type SandboxManager struct {
	Renderer  *fray.Renderer
	MapEditor *mapeditor.MapEditor
	UIManager *ui.UIManager

	HandTextureID int
}

func (sm *SandboxManager) Init() {
	sm.HandTextureID = 0

	sm.Renderer = &fray.Renderer{}
	sm.MapEditor = &mapeditor.MapEditor{}
	sm.UIManager = &ui.UIManager{}
}

func (sm *SandboxManager) ExportCanvas(filepath string) {
	// 保存するファイル名
	// savefile, err := os.Create("./game/map/map2.png")
	savefile, err := os.Create(filepath)
	if err != nil {
		fmt.Println("保存するためのファイルが作成できませんでした。")
		os.Exit(1)
	}
	defer savefile.Close()
	// PNG形式で保存する
	png.Encode(savefile, sm.MapEditor.GetCanvas())
	fmt.Println("canvas exported")
}

func (sm *SandboxManager) ExportHeightMapImage(filepath string) {
	img := sm.MapEditor.WriteHeightMapImage(sm.Renderer.Wld.HeightMap)
	// 保存するファイル名
	// savefile, err := os.Create("./game/map/heightmapflat.png")
	savefile, err := os.Create(filepath)
	if err != nil {
		fmt.Println("保存するためのファイルが作成できませんでした。")
		os.Exit(1)
	}
	defer savefile.Close()
	// PNG形式で保存する
	png.Encode(savefile, img)
	fmt.Println("heightmap exported")
}

func (sm *SandboxManager) ExportWorldMapImage(filepath string) {
	img := sm.MapEditor.WriteWorldMapImage(sm.Renderer.Wld.WorldMap, sm.Renderer.Wld.HeightMap)

	// 保存するファイル名
	// savefile, err := os.Create("./game/map/worldmapflat.png")
	savefile, err := os.Create(filepath)
	if err != nil {
		fmt.Println("保存するためのファイルが作成できませんでした。")
		os.Exit(1)
	}
	defer savefile.Close()
	// PNG形式で保存する
	png.Encode(savefile, img)
	fmt.Println("worldmap exported")
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

func (sm *SandboxManager) SetHandTextureID(id int) {
	sm.HandTextureID = id
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
			sm.Renderer.Wld.SetValue(x, y-1, z, uint8(sm.HandTextureID))
		case fray.AIM_DIR_SOUTH:
			sm.Renderer.Wld.SetValue(x, y+1, z, uint8(sm.HandTextureID))
		case fray.AIM_DIR_EAST:
			sm.Renderer.Wld.SetValue(x-1, y, z, uint8(sm.HandTextureID))
		case fray.AIM_DIR_WEST:
			sm.Renderer.Wld.SetValue(x+1, y, z, uint8(sm.HandTextureID))
		case fray.AIM_DIR_TOP:
			sm.Renderer.Wld.SetValue(x, y, z+1, uint8(sm.HandTextureID))
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

	sm.Renderer.SetHandTextureID(sm.HandTextureID)
	sm.UpdateMapEdit()

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
