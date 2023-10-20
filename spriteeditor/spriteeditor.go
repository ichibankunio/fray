package spriteeditor

import (
	"github.com/hajimehoshi/ebiten/v2"
)

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
	bytes := make([]byte, 4*dst.Bounds().Dx()*dst.Bounds().Dy())

	dst.ReadPixels(bytes)//これ毎フレーム読んでるので重い　修正必要

	for i := 0; i < len(data); i++ {
		rgba := Float32ToRGBA(data[i])
		bytes[4*i+offset*4] = rgba[0]
		bytes[4*i+1+offset*4] = rgba[1]
		bytes[4*i+2+offset*4] = rgba[2]
		bytes[4*i+3+offset*4] = rgba[3]
	}

	dst.WritePixels(bytes)

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
