package mapeditor

import (
	"fmt"
	"image/png"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type MapEditor struct {
	bytes         []byte
	data          [4][]uint8
	texture       *ebiten.Image
	canvas        *ebiten.Image

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

func (me *MapEditor) GetData() [4][]uint8 {
	return me.data
}

func NewMapEditor(screenWidth, screenHeight int, canvasWidth int, canvasHeight int) *MapEditor {
	arr := [4][]uint8{}
	for i := range arr {
		arr[i] = make([]uint8, canvasWidth*canvasHeight)
	}

	return &MapEditor{
		bytes:         make([]uint8, canvasWidth*canvasHeight*4),
		data:          arr,
		texture:       ebiten.NewImage(screenWidth, screenHeight),
		canvas:        ebiten.NewImage(canvasWidth, canvasHeight),

		screenWidth:     screenWidth,
		screenHeight:    screenHeight,
		canvasWidth:     canvasWidth,
		canvasHeight:    canvasHeight,
		canvasDepth:     0, //temporary value
		heightMapBuffer: make([]uint8, canvasWidth*canvasHeight*4),
		imageSrcBuffer:  make([]uint8, screenWidth*screenHeight*4),
	}
}

func (me *MapEditor) SetValue(x, y int, layer int, value uint8) {
	me.data[layer][y*me.canvas.Bounds().Dx()+x] = value
	me.bytes[4*(y*me.canvas.Bounds().Dx()+x)+layer] = value

	me.canvas.WritePixels(me.bytes)

	op := &ebiten.DrawImageOptions{}
	me.texture.DrawImage(me.canvas, op)
}

func (me *MapEditor) GetValue(x, y int, layer int) uint8 {
	return me.data[layer][y*me.canvas.Bounds().Dx()+x]
}

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
	savefile, err := os.Create("heightmap.png")
	if err != nil {
		fmt.Println("保存するためのファイルが作成できませんでした。")
		os.Exit(1)
	}
	defer savefile.Close()
	// PNG形式で保存する
	png.Encode(savefile, dst)
}

// img: canvasWidth*canvasHeight px, len(dst) = canvasWidth*canvasHeight
func (me *MapEditor) LoadHeightMapFromImage(img *ebiten.Image, dst []uint8) {
	img.ReadPixels(me.heightMapBuffer)
	for i := 0; i < len(dst); i++ {
		dst[i] = me.heightMapBuffer[4*i+3] //どこでもいいけどalphaに保存しておいて取り出す
	}
}

func (me *MapEditor) WriteHeightMapImage(src []uint8) *ebiten.Image {
	dst := ebiten.NewImage(me.canvasWidth, me.canvasHeight)
	for i := 0; i < len(src); i++ {
		me.heightMapBuffer[4*i+3] = src[i] //どこでもいいけどalphaに保存しておいて取り出す
	}
	dst.WritePixels(me.heightMapBuffer)

	return dst
}


func (me *MapEditor) LoadMapFromImage(canvas *ebiten.Image) {
	me.canvas = canvas
	me.canvas.ReadPixels(me.bytes)
	for i := 0; i < len(me.bytes)/4; i++ {
		me.data[0][i] = me.bytes[4*i]
		me.data[1][i] = me.bytes[4*i+1]
		me.data[2][i] = me.bytes[4*i+2]
		me.data[3][i] = me.bytes[4*i+3]
	}

	op := &ebiten.DrawImageOptions{}
	me.texture.DrawImage(me.canvas, op)
}

func (me *MapEditor) WriteTexture() {
	for i := 0; i < len(me.data[0]); i++ {
		me.bytes[4*i] = me.data[0][i]
		me.bytes[4*i+1] = me.data[1][i]
		me.bytes[4*i+2] = me.data[2][i]
		me.bytes[4*i+3] = me.data[3][i]
	}

	me.canvas.WritePixels(me.bytes)
	// me.canvas.Fill(color.White)

	op := &ebiten.DrawImageOptions{}
	me.texture.DrawImage(me.canvas, op)

	// // 保存するファイル名
	// savefile, err := os.Create("./map/map.png")
	// if err != nil {
	// 	fmt.Println("保存するためのファイルが作成できませんでした。")
	// 	os.Exit(1)
	// }
	// defer savefile.Close()
	// // PNG形式で保存する
	// png.Encode(savefile, me.canvas)
}
