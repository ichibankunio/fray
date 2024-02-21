package fray

import (
	"fmt"
	"image/png"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type World struct {
	// levelUint8 [4][]uint8

	WorldMap  [][]uint8 //texture ID map
	HeightMap []uint8   //height map

	screenWidth  int
	screenHeight int
	canvasWidth  int
	canvasHeight int
	canvasDepth  int

	imageSrcBuffer []uint8
	canvasBuffer   []uint8

	topImage *ebiten.Image

	Sprites []*Sprite
}

func (w *World) Init(screenWidth int, screenHeight int, canvasWidth int, canvasHeight int, canvasDepth int) {
	w.imageSrcBuffer = make([]uint8, screenWidth*screenHeight*4)
	w.HeightMap = make([]uint8, canvasWidth*canvasHeight)
	w.WorldMap = make([][]uint8, canvasDepth)
	for i := 0; i < canvasDepth; i++ {
		w.WorldMap[i] = make([]uint8, canvasWidth*canvasHeight)
	}

	w.canvasHeight = canvasHeight
	w.canvasWidth = canvasWidth
	w.canvasDepth = canvasDepth
	w.screenHeight = screenHeight
	w.screenWidth = screenWidth
}

func (w *World) GetValue(x, y, z int) uint8 {
	return w.WorldMap[z][y*w.canvasWidth+x]
}

func (w *World) GetHeight(x, y int) uint8 {
	return w.HeightMap[y*w.canvasWidth+x]
}

func (w *World) DeleteValue(x, y, z int) {
	if z == int(w.HeightMap[y*w.canvasWidth+x]) && z != 0 {
		fmt.Println("OK", z, int(w.HeightMap[y*w.canvasWidth+x]))
		w.WorldMap[z-1][y*w.canvasWidth+x] = 0
		w.HeightMap[y*w.canvasWidth+x] = uint8(z-1)
	}else {
		fmt.Println("NG", z ,int(w.HeightMap[y*w.canvasWidth+x]))
	}
}

func (w *World) SetValue(x, y, z int, value uint8) {
	if z-1 == int(w.HeightMap[y*w.canvasWidth+x]) {
		fmt.Println("OK", z-1, int(w.HeightMap[y*w.canvasWidth+x]))
		w.WorldMap[z-1][y*w.canvasWidth+x] = value+1
		if z > int(w.HeightMap[y*w.canvasWidth+x]) {
			w.HeightMap[y*w.canvasWidth+x] = uint8(z)
		}
	}else {
		fmt.Println("NG", z-1 ,int(w.HeightMap[y*w.canvasWidth+x]))
	}

	// w.WorldMap[z-1][y*w.canvasWidth+x] = value
	// if z > int(w.HeightMap[y*w.canvasWidth+x]) {
	// 	w.HeightMap[y*w.canvasWidth+x] = uint8(z)
	// }



	// me.bytes[4*(y*me.canvas.Bounds().Dx()+x)+layer] = value

	// me.canvas.WritePixels(me.bytes)

	// op := &ebiten.DrawImageOptions{}
	// me.texture.DrawImage(me.canvas, op)
}

func (w *World) WriteWorldMapFromHeightMap() {
	// for i := 0; i < len(w.WorldMap); i++ {
	// 	for j := 0; j < int(w.HeightMap[i]); j++ {
	// 		w.WorldMap[j][i] = 1
	// 	}
	// }

	for i := 0; i < len(w.WorldMap); i++ { //レイヤーの数
		for j := 0; j < len(w.WorldMap[0]); j++ { //長さ128*128のスライス
			// if j > 4 * 65 {
			// 	continue
			// }

			// if w.HeightMap[j] > uint8(i) {
			// 	if j % 4 == 0 {
			// 		w.WorldMap[i][j] = 255
			// 	}else {
			// 		w.WorldMap[i][j] = 1
			// 	}
			// }
			if w.HeightMap[j] > uint8(i) {
				// if j % 4 == 0 {
				// 	w.WorldMap[i][j] = 1
				// }else if j % 4 == 1 {
				// 	w.WorldMap[i][j] = 0
				// }else if j % 4 == 2 {
				// 	w.WorldMap[i][j] = 0
				// }else if j % 4 == 3 {
				// 	w.WorldMap[i][j] = 255
				// }
				if j%4 == 3 { //alpha must be larger than r, g, and b
					// w.WorldMap[i][j] = uint8(255 - 1)
					w.WorldMap[i][j] = 1
				} else {
					w.WorldMap[i][j] = 1
				}
			}
		}
	}

	// for k := 0; k < 100; k++ {
	// 	println(w.WorldMap[0][k])
	// }

	img := ebiten.NewImage(w.canvasWidth/2, w.canvasHeight/2)
	img.WritePixels(w.WorldMap[0])
	// r, g, b, a := img.At(0, 2).RGBA()
	// println(uint8(r), uint8(g), uint8(b), uint8(a))
	// println(r, g, b, a)
	savefile, err := os.Create("1stLayer.png")
	if err != nil {
		fmt.Println("保存するためのファイルが作成できませんでした。")
		os.Exit(1)
	}
	defer savefile.Close()
	// PNG形式で保存する
	png.Encode(savefile, img)
}

func (w *World) DrawTopView(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(0.5, 0.5)
	op.GeoM.Translate(0, 0)
	screen.DrawImage(w.topImage, op)
}
