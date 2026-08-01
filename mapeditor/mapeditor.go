package mapeditor

import (
	"encoding/json"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type worldMapJSON struct {
	CanvasWidth  int       `json:"canvasWidth"`
	CanvasHeight int       `json:"canvasHeight"`
	CanvasDepth  int       `json:"canvasDepth"`
	Layers       [][]uint8 `json:"layers"`
}

type heightMapJSON struct {
	CanvasWidth  int     `json:"canvasWidth"`
	CanvasHeight int     `json:"canvasHeight"`
	Data         []uint8 `json:"data"`
}

type MapEditor struct {
	bytes []byte
	// data    [4][]uint8
	texture *ebiten.Image
	canvas  *ebiten.Image

	screenWidth  int
	screenHeight int
	canvasWidth  int
	canvasHeight int
	canvasDepth  int

	heightMapBuffer   []uint8
	imageSrcBuffer    []uint8
	heightPlaneBuffer []uint8

	heightMapInitialized bool
	worldMapCanvas       *ebiten.Image
	worldMapBuffer       []uint8
}

func (me *MapEditor) GetCanvas() *ebiten.Image {
	return me.canvas
}

func (me *MapEditor) GetTexture() *ebiten.Image {
	return me.texture
}

// func (me *MapEditor) GetData() [4][]uint8 {
// 	return me.data
// }

func NewMapEditor(screenWidth, screenHeight int, canvasWidth int, canvasHeight int, canvasDepth int) *MapEditor {
	arr := [4][]uint8{}
	for i := range arr {
		arr[i] = make([]uint8, canvasWidth*canvasHeight)
	}

	return &MapEditor{
		bytes: make([]uint8, canvasWidth*canvasHeight*4),
		// data:    arr,
		texture: ebiten.NewImage(screenWidth, screenHeight),
		canvas:  ebiten.NewImage(canvasWidth, canvasHeight),

		screenWidth:       screenWidth,
		screenHeight:      screenHeight,
		canvasWidth:       canvasWidth,
		canvasHeight:      canvasHeight,
		canvasDepth:       canvasDepth, //temporary value
		heightMapBuffer:   make([]uint8, canvasWidth*canvasHeight*4),
		imageSrcBuffer:    make([]uint8, screenWidth*screenHeight*4),
		heightPlaneBuffer: make([]uint8, screenWidth*screenHeight*4),
	}
}

// func (me *MapEditor) SetValue(x, y int, layer int, value uint8) {
// 	me.data[layer][y*me.canvas.Bounds().Dx()+x] = value
// 	me.bytes[4*(y*me.canvas.Bounds().Dx()+x)+layer] = value

// 	me.canvas.WritePixels(me.bytes)

// 	op := &ebiten.DrawImageOptions{}
// 	me.texture.DrawImage(me.canvas, op)
// }

// func (me *MapEditor) GetValue(x, y int, layer int) uint8 {
// 	return me.data[layer][y*me.canvas.Bounds().Dx()+x]
// }

// height mapをCPUにコピー(衝突判定に使う)
func (me *MapEditor) CopyHeightMap(heightMapData []uint8) {
	// me.heightMapData = heightMapData
	// canvas.ReadPixels(me.bytes)
	// for i := 0; i < len(me.bytes)/4; i++ {
	// 	me.heightMapData[i] = me.bytes[4*i+3+offset]//rgbaのaをコピー
	// }
}

// src: canvasWidth*canvasHeight, buffer: screenWidth*screenHeight*4
func (me *MapEditor) PrintHeightMapOnAlphaLayer(src []uint8, dst *ebiten.Image) {
	if !me.heightMapInitialized {
		target := dst
		if b := dst.Bounds(); b.Dx() != me.screenWidth || b.Dy() != me.screenHeight {
			target = dst.SubImage(image.Rect(0, 0, me.screenWidth, me.screenHeight)).(*ebiten.Image)
		}
		target.ReadPixels(me.imageSrcBuffer)
		me.heightMapInitialized = true
	}
	for i := 0; i < len(src); i++ {
		me.imageSrcBuffer[((i/me.canvasWidth)*me.screenWidth+(i%me.canvasWidth))*4+3] = src[i] //rgbaのaをコピー
	}

	target := dst
	if b := dst.Bounds(); b.Dx() != me.screenWidth || b.Dy() != me.screenHeight {
		target = dst.SubImage(image.Rect(0, 0, me.screenWidth, me.screenHeight)).(*ebiten.Image)
	}
	target.WritePixels(me.imageSrcBuffer)

	//保存するファイル名
	// savefile, err := os.Create("heightmap.png")
	// if err != nil {
	// 	fmt.Println("保存するためのファイルが作成できませんでした。")
	// 	os.Exit(1)
	// }
	// defer savefile.Close()
	// // PNG形式で保存する
	// png.Encode(savefile, dst)
}

func (me *MapEditor) PrintHeightPlaneMap(base, slopeX, slopeY []float32, dst *ebiten.Image) {
	me.PrintHeightPlaneMapWithVisibility(base, slopeX, slopeY, nil, dst)
}

// PrintHeightPlaneMapWithVisibility stores inferred height data and optional
// terrain visibility in one GPU lookup texture. Visibility occupies alpha.
func (me *MapEditor) PrintHeightPlaneMapWithVisibility(base, slopeX, slopeY, visibility []float32, dst *ebiten.Image) {
	if cap(me.heightPlaneBuffer) < me.screenWidth*me.screenHeight*4 {
		me.heightPlaneBuffer = make([]uint8, me.screenWidth*me.screenHeight*4)
	} else {
		me.heightPlaneBuffer = me.heightPlaneBuffer[:me.screenWidth*me.screenHeight*4]
	}
	buffer := me.heightPlaneBuffer
	for i := range buffer {
		buffer[i] = 0
	}

	n := len(base)
	if len(slopeX) < n {
		n = len(slopeX)
	}
	if len(slopeY) < n {
		n = len(slopeY)
	}
	if n > me.canvasWidth*me.canvasHeight {
		n = me.canvasWidth * me.canvasHeight
	}

	for i := 0; i < n; i++ {
		baseV := base[i]
		if baseV < 0 {
			baseV = 0
		} else if baseV > 255 {
			baseV = 255
		}
		slopeXv := slopeX[i]
		slopeYv := slopeY[i]
		encSlopeX := slopeXv*128 + 128
		if encSlopeX < 0 {
			encSlopeX = 0
		} else if encSlopeX > 255 {
			encSlopeX = 255
		}
		encSlopeY := slopeYv*128 + 128
		if encSlopeY < 0 {
			encSlopeY = 0
		} else if encSlopeY > 255 {
			encSlopeY = 255
		}

		dstIdx := ((i/me.canvasWidth)*me.screenWidth + (i % me.canvasWidth)) * 4
		buffer[dstIdx] = uint8(baseV + 0.5)
		buffer[dstIdx+1] = uint8(encSlopeX + 0.5)
		buffer[dstIdx+2] = uint8(encSlopeY + 0.5)
		visibilityV := float32(1)
		if i < len(visibility) {
			visibilityV = max(0, min(1, visibility[i]))
		}
		buffer[dstIdx+3] = uint8(visibilityV*255 + 0.5)
	}

	target := dst
	if b := dst.Bounds(); b.Dx() != me.screenWidth || b.Dy() != me.screenHeight {
		target = dst.SubImage(image.Rect(0, 0, me.screenWidth, me.screenHeight)).(*ebiten.Image)
	}
	target.WritePixels(buffer)
}

func (me *MapEditor) WriteWorldMapImage(worldMap [][]uint8, heightMap []uint8) *ebiten.Image {
	//math.Ceil(float64(me.canvasDepth)/float64((me.screenWidth/me.canvasWidth)*3))個、canvasを横に並べられる
	dst := ebiten.NewImage(me.screenWidth, me.canvasHeight*int(math.Ceil(float64(me.canvasDepth)/float64((me.screenWidth/me.canvasWidth)*3))))
	canvas := ebiten.NewImage(me.canvasWidth, me.canvasHeight)
	buffer := make([]uint8, me.canvasWidth*me.canvasHeight*4)
	// for i := 0; i < len(src); i++ {
	// 	for j := 0; j < len(src[i]); j++ {
	// 		buffer[4*j] = src[i][j]
	// 		buffer[4*j+3] = (src[i][j]/1)*255
	// 	}

	// 	canvas.WritePixels(buffer)
	// 	// canvas.Fill(color.Black)

	// 	op := &ebiten.DrawImageOptions{}
	// 	println(i%(me.screenWidth/me.canvasWidth)*me.canvasWidth, i/(me.screenWidth/me.canvasWidth)*me.canvasHeight)
	// 	op.GeoM.Translate(float64(i%(me.screenWidth/me.canvasWidth)*me.canvasWidth), float64(i/(me.screenWidth/me.canvasWidth)*me.canvasHeight))
	// 	dst.DrawImage(canvas, op)
	// }

	for i := 0; i < int(math.Ceil(float64(len(worldMap))/3)); i++ {
		for j := 0; j < len(worldMap[i]); j++ {
			if 3*i < len(worldMap) {
				buffer[4*j] = worldMap[3*i][j]
				if worldMap[3*i][j] > 0 {
					println(worldMap[3*i][j], 3*i, j)
				}
			} else {
				buffer[4*j] = 0
			}

			if 3*i+1 < len(worldMap) {
				buffer[4*j+1] = worldMap[3*i+1][j]
				if worldMap[3*i+1][j] > 0 {
					println(worldMap[3*i+1][j], 3*i+1, j)
				}
			} else {
				buffer[4*j+1] = 0
			}
			if 3*i+2 < len(worldMap) {
				buffer[4*j+2] = worldMap[3*i+2][j]
			} else {
				buffer[4*j+2] = 0
			}

			//世界のデータではないけど、画像で出力したときに目に見えるようにするために透明度を設定
			// buffer[4*j+3] = (worldMap[3*i][j] / 1) * 255
			buffer[4*j+3] = 255
		}

		canvas.WritePixels(buffer)
		// canvas.Fill(color.Black)

		op := &ebiten.DrawImageOptions{}
		// println(i%(me.screenWidth/me.canvasWidth)*me.canvasWidth, i/(me.screenWidth/me.canvasWidth)*me.canvasHeight)
		op.GeoM.Translate(float64(i%(me.screenWidth/me.canvasWidth)*me.canvasWidth), float64(i/(me.screenWidth/me.canvasWidth)*me.canvasHeight))
		dst.DrawImage(canvas, op)
	}

	return dst
}

func (me *MapEditor) PrintWorldMap(src [][]uint8, dst *ebiten.Image) {
	dst.Clear()
	if me.worldMapCanvas == nil {
		me.worldMapCanvas = ebiten.NewImage(me.canvasWidth, me.canvasHeight)
	}
	if cap(me.worldMapBuffer) < me.canvasWidth*me.canvasHeight*4 {
		me.worldMapBuffer = make([]uint8, me.canvasWidth*me.canvasHeight*4)
	} else {
		me.worldMapBuffer = me.worldMapBuffer[:me.canvasWidth*me.canvasHeight*4]
	}
	canvas := me.worldMapCanvas
	buffer := me.worldMapBuffer

	for i := 0; i < int(math.Ceil(float64(len(src))/4)); i++ {
		for j := 0; j < len(src[i]); j++ {
			buffer[4*j] = src[4*i][j]
			buffer[4*j+1] = src[4*i+1][j]
			buffer[4*j+2] = src[4*i+2][j]
			buffer[4*j+3] = src[4*i+3][j]

			/*
				if 4*i < len(src) {
					buffer[4*j] = src[4*i][j]
				} else {
					buffer[4*j] = 0
				}

				if 4*i+1 < len(src) {
					buffer[4*j+1] = src[4*i+1][j]
				} else {
					buffer[4*j+1] = 0
				}
				if 4*i+2 < len(src) {
					buffer[4*j+2] = src[4*i+2][j]
				} else {
					buffer[4*j+2] = 0
				}
				if 4*i+3 < len(src) {
					buffer[4*j+3] = src[4*i+3][j]
				} else {
					buffer[4*j+3] = 0
				}
			*/

		}

		canvas.WritePixels(buffer)
		// canvas.Fill(color.Black)

		op := &ebiten.DrawImageOptions{}
		// println(i%(me.screenWidth/me.canvasWidth)*me.canvasWidth, i/(me.screenWidth/me.canvasWidth)*me.canvasHeight)
		op.GeoM.Translate(float64(i%(me.screenWidth/me.canvasWidth)*me.canvasWidth), float64(i/(me.screenWidth/me.canvasWidth)*me.canvasHeight))
		dst.DrawImage(canvas, op)
	}

}

func (me *MapEditor) PrintWorldMap2(src [][]uint8, dst *ebiten.Image) {
	for z := 0; z < len(src); z++ {
		for i := 0; i < len(src[0]); i++ {
			// me.imageSrcBuffer[z*len(src[0]) + i] = src[z][i]
			me.imageSrcBuffer[z*len(src[0])+i] = src[z][i]
		}
	}
	// buf := make([]uint8, me.screenWidth*me.screenHeight*4)
	// buf := me.imageSrcBuffer
	// for i := 0; i < len(src[0]); i++ {
	// 	// me.imageSrcBuffer[z*len(src[0]) + i] = src[z][i]
	// 	buf[i] = src[0][i]
	// 	// buf[i] = 10
	// }

	// buf := make([]byte, me.screenWidth*me.screenHeight*4)
	// // for i := 0; i < 100; i++ {
	// // 	buf[4*i] = byte(i)
	// // }
	// buf[0] = 255
	// buf[1] = 0
	// buf[2] = 0
	// buf[3] = 255

	// buf[4] = 255
	// buf[5] = 255
	// buf[6] = 0
	// buf[7] = 255

	dst.WritePixels(me.imageSrcBuffer)
	// dst.WritePixels(buf)

	//保存するファイル名
	// savefile, err := os.Create("worldmap.png")
	// if err != nil {
	// 	fmt.Println("保存するためのファイルが作成できませんでした。")
	// 	os.Exit(1)
	// }
	// defer savefile.Close()
	// // PNG形式で保存する
	// png.Encode(savefile, dst)
}

// img: canvasWidth*canvasHeight px, len(dst) = canvasWidth*canvasHeight
func (me *MapEditor) LoadHeightMapFromImage(img *ebiten.Image, dst []uint8) {
	img.ReadPixels(me.heightMapBuffer)
	for i := 0; i < len(dst); i++ {
		dst[i] = me.heightMapBuffer[4*i]
	}
}

// JSON format:
//
//	{
//	  "canvasWidth": 128,
//	  "canvasHeight": 128,
//	  "data": [...]
//	}
func (me *MapEditor) LoadHeightMapFromJson(data []byte, dst []uint8) {
	var hm heightMapJSON
	if err := json.Unmarshal(data, &hm); err != nil {
		return
	}
	if len(hm.Data) == 0 {
		return
	}
	n := len(dst)
	if len(hm.Data) < n {
		n = len(hm.Data)
	}
	copy(dst[:n], hm.Data[:n])
}

func (me *MapEditor) LoadWorldMapFromImage(img *ebiten.Image, dst [][]uint8) {
	buffer := make([]uint8, me.canvasWidth*me.canvasHeight*4)

	b := img.Bounds()
	if b.Dx() < me.canvasWidth || b.Dy() < me.canvasHeight {
		return
	}

	tilesX := b.Dx() / me.canvasWidth
	tilesY := b.Dy() / me.canvasHeight
	if tilesX <= 0 || tilesY <= 0 {
		return
	}

	maxTiles := tilesX * tilesY
	needTiles := int(math.Ceil(float64(len(dst)) / 3))
	if needTiles > maxTiles {
		needTiles = maxTiles
	}

	for i := 0; i < needTiles; i++ {
		x0 := (i % tilesX) * me.canvasWidth
		x1 := x0 + me.canvasWidth
		y0 := (i / tilesX) * me.canvasHeight
		y1 := y0 + me.canvasHeight
		img.SubImage(image.Rect(x0, y0, x1, y1)).(*ebiten.Image).ReadPixels(buffer)

		for j := 0; j < len(dst[0]); j++ {
			if 3*i < len(dst) {
				dst[3*i][j] = buffer[4*j]
				// dst[3*i][j] = uint8(rand.Intn(5))
			}

			if 3*i+1 < len(dst) {
				dst[3*i+1][j] = buffer[4*j+1]
				// dst[3*i+1][j] = uint8(rand.Intn(5))
			}

			if 3*i+2 < len(dst) {
				dst[3*i+2][j] = buffer[4*j+2]
				// dst[3*i+2][j] = uint8(rand.Intn(5))
			}
		}
	}

}

// JSON format:
//
//	{
//	  "canvasWidth": 128,
//	  "canvasHeight": 128,
//	  "canvasDepth": 56,
//	  "layers": [[...], ...]
//	}
func (me *MapEditor) LoadWorldMapFromJson(data []byte, dst [][]uint8) {
	var wm worldMapJSON
	if err := json.Unmarshal(data, &wm); err != nil {
		return
	}
	if len(wm.Layers) == 0 {
		return
	}

	layerCount := len(dst)
	if wm.CanvasDepth > 0 && wm.CanvasDepth < layerCount {
		layerCount = wm.CanvasDepth
	}
	if len(wm.Layers) < layerCount {
		layerCount = len(wm.Layers)
	}

	for i := 0; i < layerCount; i++ {
		if len(wm.Layers[i]) == 0 {
			continue
		}
		n := len(dst[i])
		if len(wm.Layers[i]) < n {
			n = len(wm.Layers[i])
		}
		copy(dst[i][:n], wm.Layers[i][:n])
	}
}

func (me *MapEditor) WriteHeightMapImage(src []uint8) *ebiten.Image {
	dst := ebiten.NewImage(me.canvasWidth, me.canvasHeight)
	for i := 0; i < len(src); i++ {
		me.heightMapBuffer[4*i] = src[i] //どこでもいいけどalphaに保存しておいて取り出す
		me.heightMapBuffer[4*i+1] = 0    //どこでもいいけどalphaに保存しておいて取り出す
		me.heightMapBuffer[4*i+2] = 0    //どこでもいいけどalphaに保存しておいて取り出す
		if src[i] > 0 {
			me.heightMapBuffer[4*i+3] = 255 //どこでもいいけどalphaに保存しておいて取り出す
		} else {
			me.heightMapBuffer[4*i+3] = 0
		}
	}
	dst.WritePixels(me.heightMapBuffer)

	return dst
}

// func (me *MapEditor) LoadMapFromImage(canvas *ebiten.Image) {
// 	me.canvas = canvas
// 	me.canvas.ReadPixels(me.bytes)
// 	for i := 0; i < len(me.bytes)/4; i++ {
// 		me.data[0][i] = me.bytes[4*i]
// 		me.data[1][i] = me.bytes[4*i+1]
// 		me.data[2][i] = me.bytes[4*i+2]
// 		me.data[3][i] = me.bytes[4*i+3]
// 	}

// 	op := &ebiten.DrawImageOptions{}
// 	me.texture.DrawImage(me.canvas, op)
// }
