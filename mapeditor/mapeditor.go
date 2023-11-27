package mapeditor

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

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

	heightMapBuffer []uint8
	imageSrcBuffer  []uint8
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

		screenWidth:     screenWidth,
		screenHeight:    screenHeight,
		canvasWidth:     canvasWidth,
		canvasHeight:    canvasHeight,
		canvasDepth:     canvasDepth, //temporary value
		heightMapBuffer: make([]uint8, canvasWidth*canvasHeight*4),
		imageSrcBuffer:  make([]uint8, screenWidth*screenHeight*4),
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
	dst.ReadPixels(me.imageSrcBuffer)
	for i := 0; i < len(src); i++ {
		me.imageSrcBuffer[((i/me.canvasWidth)*me.screenWidth+(i%me.canvasWidth))*4+3] = src[i] //rgbaのaをコピー
	}

	dst.WritePixels(me.imageSrcBuffer)

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

func (me *MapEditor) WriteWorldMapImage(src [][]uint8) *ebiten.Image {
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
	for i := 0; i < int(math.Ceil(float64(len(src))/3)); i++ {
		for j := 0; j < len(src[i]); j++ {
			if 3*i < len(src) {
				buffer[4*j] = src[3*i][j]
			} else {
				buffer[4*j] = 0
			}

			if 3*i+1 < len(src) {
				buffer[4*j+1] = src[3*i+1][j]
			} else {
				buffer[4*j+1] = 0
			}
			if 3*i+2 < len(src) {
				buffer[4*j+2] = src[3*i+2][j]
			} else {
				buffer[4*j+2] = 0
			}

			buffer[4*j+3] = (src[3*i][j] / 1) * 255
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
	canvas := ebiten.NewImage(me.canvasWidth, me.canvasHeight)
	buffer := make([]uint8, me.canvasWidth*me.canvasHeight*4)

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

func (me *MapEditor) LoadWorldMapFromImage(img *ebiten.Image, dst [][]uint8) {
	buffer := make([]uint8, me.canvasWidth*me.canvasHeight*4)

	for i := 0; i < int(math.Ceil(float64(len(dst))/3)); i++ {
		x0 := i % (me.screenWidth / me.canvasWidth) * me.canvasWidth
		x1 := x0 + me.canvasWidth
		y0 := i / (me.screenWidth / me.canvasWidth) * me.canvasHeight
		y1 := y0 + me.canvasHeight
		// println(x0, x1, y0, y1)
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

func (me *MapEditor) WriteHeightMapImage(src []uint8) *ebiten.Image {
	dst := ebiten.NewImage(me.canvasWidth, me.canvasHeight)
	for i := 0; i < len(src); i++ {
		me.heightMapBuffer[4*i] = src[i] //どこでもいいけどalphaに保存しておいて取り出す
		me.heightMapBuffer[4*i+1] = 0 //どこでもいいけどalphaに保存しておいて取り出す
		me.heightMapBuffer[4*i+2] = 0 //どこでもいいけどalphaに保存しておいて取り出す
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
