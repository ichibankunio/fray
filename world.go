package fray

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type World struct {
	levelUint8 [4][]uint8

	WorldMap []uint8//texture ID map
	HeightMap []uint8//height map

	screenWidth int 
	screenHeight int
	canvasWidth int
	canvasHeight int
	canvasDepth int

	imageSrcBuffer []uint8
	canvasBuffer []uint8

	topImage *ebiten.Image

	Sprites            []*Sprite
}

func (w *World) Init(screenWidth int, screenHeight int, canvasWidth int, canvasHeight int, canvasDepth int) {
	w.imageSrcBuffer = make([]uint8, screenWidth*screenHeight*4)
	w.HeightMap = make([]uint8, canvasWidth*canvasHeight)
	w.WorldMap = make([]uint8, canvasDepth*canvasWidth*canvasHeight)

	w.canvasHeight = canvasHeight
	w.canvasWidth = canvasWidth
	w.canvasDepth = canvasDepth
	w.screenHeight = screenHeight
	w.screenWidth = screenWidth
}

func (w *World) GetValue(x, y, z int) uint8 {
	return w.WorldMap[z*w.canvasWidth*w.canvasHeight + y*w.canvasWidth+x]
}

func (w *World) GetHeight(x, y int) uint8 {
	return w.HeightMap[y*w.canvasWidth+x]
}

func (w *World) SetValue(x, y, z int, value uint8) {
	w.WorldMap[z*w.canvasWidth*w.canvasHeight + y*w.canvasWidth+x] = value
	if z > int(w.HeightMap[y*w.canvasWidth+x]) {
		w.HeightMap[y*w.canvasWidth+x] = uint8(z)
	}

	// me.bytes[4*(y*me.canvas.Bounds().Dx()+x)+layer] = value

	// me.canvas.WritePixels(me.bytes)

	// op := &ebiten.DrawImageOptions{}
	// me.texture.DrawImage(me.canvas, op)
}

func (w *World) GenerateWorldMapFromHeightMap() {
	for i := 0; i < len(w.HeightMap); i++ {
		for j := 0; j < int(w.HeightMap[i]); j++ {
			w.WorldMap[j*w.canvasWidth*w.canvasHeight + i] = 1
		}
	}
}




func (w *World) DrawTopView(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(0.5, 0.5)
	op.GeoM.Translate(0, 0)
	screen.DrawImage(w.topImage, op)
}

