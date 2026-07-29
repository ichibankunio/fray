package fray

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type World struct {
	// levelUint8 [4][]uint8

	WorldMap   [][]uint8 //texture ID map
	HeightMap  []uint8   //height map
	HeightBase []float32
	SlopeX     []float32
	SlopeY     []float32

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
	w.HeightBase = make([]float32, canvasWidth*canvasHeight)
	w.SlopeX = make([]float32, canvasWidth*canvasHeight)
	w.SlopeY = make([]float32, canvasWidth*canvasHeight)
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

func (w *World) CanvasWidth() int {
	return w.canvasWidth
}

func (w *World) CanvasHeight() int {
	return w.canvasHeight
}

func (w *World) CanvasDepth() int {
	return w.canvasDepth
}

func (w *World) BuildHeightMapFromWorldMap() {
	for i := 0; i < len(w.HeightMap); i++ {
		height := 0
		for z := w.canvasDepth - 1; z >= 0; z-- {
			if w.WorldMap[z][i] != 0 {
				height = z + 1
				break
			}
		}
		w.HeightMap[i] = uint8(height)
	}
	w.SyncHeightPlanesFromHeightMap()
}

func (w *World) SyncHeightPlanesFromHeightMap() {
	hAt := func(x, y int) float32 {
		if x < 0 {
			x = 0
		} else if x >= w.canvasWidth {
			x = w.canvasWidth - 1
		}
		if y < 0 {
			y = 0
		} else if y >= w.canvasHeight {
			y = w.canvasHeight - 1
		}
		return float32(w.HeightMap[y*w.canvasWidth+x])
	}
	clampSlope := func(v float32) float32 {
		if v < -1 {
			return -1
		}
		if v > 1 {
			return 1
		}
		return v
	}

	for y := 0; y < w.canvasHeight; y++ {
		for x := 0; x < w.canvasWidth; x++ {
			idx := y*w.canvasWidth + x
			h := hAt(x, y)
			dx := (hAt(x+1, y) - hAt(x-1, y)) * 0.5
			dy := (hAt(x, y+1) - hAt(x, y-1)) * 0.5
			if x == 0 {
				dx = hAt(x+1, y) - h
			} else if x == w.canvasWidth-1 {
				dx = h - hAt(x-1, y)
			}
			if y == 0 {
				dy = hAt(x, y+1) - h
			} else if y == w.canvasHeight-1 {
				dy = h - hAt(x, y-1)
			}

			w.HeightBase[idx] = h
			w.SlopeX[idx] = clampSlope(dx)
			w.SlopeY[idx] = clampSlope(dy)
		}
	}
}

func (w *World) SetTextureID(x, y, z int, texID uint8) bool {
	if x < 0 || y < 0 || z < 0 || x >= w.canvasWidth || y >= w.canvasHeight || z >= w.canvasDepth {
		return false
	}
	idx := y*w.canvasWidth + x
	if w.WorldMap[z][idx] == 0 {
		return false
	}
	w.WorldMap[z][idx] = texID + 1
	return true
}

func (w *World) DeleteValue(x, y, z int) {
	if z == int(w.HeightMap[y*w.canvasWidth+x]) && z != 0 {
		// fmt.Println("OK", z, int(w.HeightMap[y*w.canvasWidth+x]))
		w.WorldMap[z-1][y*w.canvasWidth+x] = 0
		w.HeightMap[y*w.canvasWidth+x] = uint8(z - 1)
		w.HeightBase[y*w.canvasWidth+x] = float32(z - 1)
		w.SlopeX[y*w.canvasWidth+x] = 0
		w.SlopeY[y*w.canvasWidth+x] = 0
	} else {
		// fmt.Println("NG", z ,int(w.HeightMap[y*w.canvasWidth+x]))
	}
}

func (w *World) SetValue(x, y, z int, value uint8) {
	if z-1 == int(w.HeightMap[y*w.canvasWidth+x]) {
		// fmt.Println("OK", z-1, int(w.HeightMap[y*w.canvasWidth+x]))
		w.WorldMap[z-1][y*w.canvasWidth+x] = value + 1
		if z > int(w.HeightMap[y*w.canvasWidth+x]) {
			w.HeightMap[y*w.canvasWidth+x] = uint8(z)
			w.HeightBase[y*w.canvasWidth+x] = float32(z)
			w.SlopeX[y*w.canvasWidth+x] = 0
			w.SlopeY[y*w.canvasWidth+x] = 0
		}
	} else {
		// fmt.Println("NG", z-1 ,int(w.HeightMap[y*w.canvasWidth+x]))
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

func (w *World) SetHeightPlane(x, y int, base, slopeX, slopeY float32) bool {
	if x < 0 || y < 0 || x >= w.canvasWidth || y >= w.canvasHeight {
		return false
	}
	idx := y*w.canvasWidth + x
	w.HeightBase[idx] = base
	w.SlopeX[idx] = slopeX
	w.SlopeY[idx] = slopeY
	return true
}

/*
func (w *World) WriteWorldMapFromHeightMap() {
	for i := 0; i < len(w.WorldMap); i++ { //レイヤーの数
		for j := 0; j < len(w.WorldMap[0]); j++ { //長さ128*128のスライス
			if w.HeightMap[j] > uint8(i) {
				if j%4 == 3 { //alpha must be larger than r, g, and b
					w.WorldMap[i][j] = 1
				} else {
					w.WorldMap[i][j] = 1
				}
			}
		}
	}

	img := ebiten.NewImage(w.canvasWidth/2, w.canvasHeight/2)
	img.WritePixels(w.WorldMap[0])

	savefile, err := os.Create("1stLayer.png")
	if err != nil {
		fmt.Println("保存するためのファイルが作成できませんでした。")
		os.Exit(1)
	}
	defer savefile.Close()
	// PNG形式で保存する
	png.Encode(savefile, img)
}
*/

func (w *World) DrawTopView(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(0.5, 0.5)
	op.GeoM.Translate(0, 0)
	screen.DrawImage(w.topImage, op)
}
