package spriteeditor

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

var writeRowBuffer []byte

type SpriteEditor struct {
	bytes   []byte
	texture *ebiten.Image
}

func NewSpriteEditor(screenWidth, screenHeight int) *SpriteEditor {
	return &SpriteEditor{
		bytes:   make([]uint8, screenWidth*screenHeight*4),
		texture: ebiten.NewImage(screenWidth, screenHeight),
	}
}

func (se *SpriteEditor) GetTexture() *ebiten.Image {
	return se.texture
}

func WriteTexture(dst *ebiten.Image, data []float32, offset int) *ebiten.Image {
	if dst == nil || len(data) == 0 {
		return dst
	}

	bounds := dst.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return dst
	}

	if offset < 0 || offset >= width*height {
		return dst
	}

	// Write only the affected region to avoid full-frame allocations/ReadPixels in WASM.
	startX := offset % width
	startY := offset / width

	remaining := len(data)
	dataIndex := 0
	if cap(writeRowBuffer) < width*4 {
		writeRowBuffer = make([]byte, width*4)
	}
	rowBuffer := writeRowBuffer[:width*4]
	y := startY
	x := startX
	for remaining > 0 && y < height {
		rowWidth := width - x
		if remaining < rowWidth {
			rowWidth = remaining
		}

		for i := 0; i < rowWidth; i++ {
			rgba := Float32ToRGBA(data[dataIndex])
			base := i * 4
			rowBuffer[base] = rgba[0]
			rowBuffer[base+1] = rgba[1]
			rowBuffer[base+2] = rgba[2]
			rowBuffer[base+3] = rgba[3]
			dataIndex++
		}

		sub := dst.SubImage(image.Rect(x, y, x+rowWidth, y+1)).(*ebiten.Image)
		sub.WritePixels(rowBuffer[:rowWidth*4])

		remaining -= rowWidth
		x = 0
		y++
	}

	// savefile, err := os.Create("./game/texturesheet.png")
	// if err != nil {
	// 	fmt.Println("保存するためのファイルが作成できませんでした。")
	// 	os.Exit(1)
	// }
	// defer savefile.Close()
	// // PNG形式で保存する
	// png.Encode(savefile, dst)

	return dst
}

// https://qiita.com/edo_m18/items/4b23846b8a97ec2a21de
// [0, 1]の値を変換できる
func Float32ToRGBA(f float32) [4]byte {
	tmp := f * 255

	ri := float32(int(tmp))

	tmp = (tmp - ri) * 255

	gi := float32(int(tmp))

	tmp = (tmp - gi) * 255

	bi := float32(int(tmp))

	tmp = (tmp - bi) * 255

	ai := float32(int(tmp))

	tmp = (tmp - ai) * 255

	return [4]byte{byte(ri), byte(gi), byte(bi), byte(ai)}
}
