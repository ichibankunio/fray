//go:build !wasm
// +build !wasm

package sandbox

import (
	"fmt"
	"image/png"
	"os"
)

func (sm *SandboxManager) ExportCanvas(filepath string) {
	savefile, err := os.Create(filepath)
	if err != nil {
		fmt.Println("保存するためのファイルが作成できませんでした。")
		os.Exit(1)
	}
	defer savefile.Close()
	png.Encode(savefile, sm.MapEditor.GetCanvas())
	fmt.Println("canvas exported")
}

func (sm *SandboxManager) ExportHeightMapImage(filepath string) {
	img := sm.MapEditor.WriteHeightMapImage(sm.Renderer.Wld.HeightMap)
	savefile, err := os.Create(filepath)
	if err != nil {
		fmt.Println("保存するためのファイルが作成できませんでした。")
		os.Exit(1)
	}
	defer savefile.Close()
	png.Encode(savefile, img)
	fmt.Println("heightmap exported")
}

func (sm *SandboxManager) ExportWorldMapImage(filepath string) {
	img := sm.MapEditor.WriteWorldMapImage(sm.Renderer.Wld.WorldMap, sm.Renderer.Wld.HeightMap)
	savefile, err := os.Create(filepath)
	if err != nil {
		fmt.Println("保存するためのファイルが作成できませんでした。")
		os.Exit(1)
	}
	defer savefile.Close()
	png.Encode(savefile, img)
	fmt.Println("worldmap exported")
}
