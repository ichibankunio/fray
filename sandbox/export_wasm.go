//go:build wasm
// +build wasm

package sandbox

import "fmt"

func (sm *SandboxManager) ExportCanvas(filepath string) {
	fmt.Println("ExportCanvas: WASM環境ではファイル出力はサポートされていません。（", filepath, "）")
}

func (sm *SandboxManager) ExportHeightMapImage(filepath string) {
	fmt.Println("ExportHeightMapImage: WASM環境ではファイル出力はサポートされていません。（", filepath, "）")
}

func (sm *SandboxManager) ExportWorldMapImage(filepath string) {
	fmt.Println("ExportWorldMapImage: WASM環境ではファイル出力はサポートされていません。（", filepath, "）")
}
