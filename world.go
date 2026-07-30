package fray

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type TerrainInterpolation uint8

const (
	TerrainInterpolationFlat TerrainInterpolation = iota
	TerrainInterpolationLinear
	TerrainInterpolationMonotonic

	// TerrainInterpolationSmooth is kept as a source-compatible alias.
	TerrainInterpolationSmooth = TerrainInterpolationMonotonic
)

type terrainJSONOptions struct {
	Interpolation string `json:"interpolation"`
}

type terrainJSONDocument struct {
	Version      int                `json:"version"`
	CanvasWidth  int                `json:"canvasWidth"`
	CanvasHeight int                `json:"canvasHeight"`
	CanvasDepth  int                `json:"canvasDepth"`
	Layers       [][]uint8          `json:"layers"`
	Terrain      terrainJSONOptions `json:"terrain"`
}

type World struct {
	WorldMap   [][]uint8 // texture ID map
	HeightMap  []uint8   // height inferred from WorldMap
	HeightBase []float32
	SlopeX     []float32
	SlopeY     []float32

	TerrainInterpolation TerrainInterpolation

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
	w.TerrainInterpolation = TerrainInterpolationLinear
}

func (w *World) LoadTerrainJSON(data []byte) error {
	var document terrainJSONDocument
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return fmt.Errorf("decode terrain JSON: %w", err)
	}
	if document.Version < 0 || document.Version > 3 {
		return fmt.Errorf("unsupported terrain JSON version %d", document.Version)
	}
	if document.CanvasWidth != 0 && document.CanvasWidth != w.canvasWidth {
		return fmt.Errorf("terrain width %d does not match world width %d", document.CanvasWidth, w.canvasWidth)
	}
	if document.CanvasHeight != 0 && document.CanvasHeight != w.canvasHeight {
		return fmt.Errorf("terrain height %d does not match world height %d", document.CanvasHeight, w.canvasHeight)
	}
	if document.CanvasDepth != 0 && document.CanvasDepth > w.canvasDepth {
		return fmt.Errorf("terrain depth %d exceeds world depth %d", document.CanvasDepth, w.canvasDepth)
	}
	if len(document.Layers) == 0 {
		return fmt.Errorf("terrain JSON has no layers")
	}

	interpolation, err := parseTerrainInterpolation(document.Terrain.Interpolation)
	if err != nil {
		return err
	}
	w.TerrainInterpolation = interpolation

	for z := range w.WorldMap {
		clear(w.WorldMap[z])
	}
	layerCount := min(len(document.Layers), len(w.WorldMap))
	for z := 0; z < layerCount; z++ {
		if len(document.Layers[z]) > len(w.WorldMap[z]) {
			return fmt.Errorf("terrain layer %d has %d cells, maximum is %d", z, len(document.Layers[z]), len(w.WorldMap[z]))
		}
		copy(w.WorldMap[z], document.Layers[z])
	}
	w.BuildHeightMapFromWorldMap()
	return nil
}

func parseTerrainInterpolation(value string) (TerrainInterpolation, error) {
	switch value {
	case "", "linear":
		return TerrainInterpolationLinear, nil
	case "flat":
		return TerrainInterpolationFlat, nil
	case "smooth", "monotonic":
		return TerrainInterpolationMonotonic, nil
	default:
		return TerrainInterpolationLinear, fmt.Errorf("unknown terrain interpolation %q", value)
	}
}

func (w *World) heightAtBlocks(x, y float64) float64 {
	maxX := float64(w.canvasWidth) - 1e-6
	maxY := float64(w.canvasHeight) - 1e-6
	x = math.Max(0, math.Min(maxX, x))
	y = math.Max(0, math.Min(maxY, y))
	cellX := int(math.Floor(x))
	cellY := int(math.Floor(y))
	fracX := x - float64(cellX)
	fracY := y - float64(cellY)

	if w.TerrainInterpolation == TerrainInterpolationFlat {
		return w.heightSample(cellX, cellY)
	}
	if w.TerrainInterpolation == TerrainInterpolationLinear {
		h00 := w.heightSample(cellX, cellY)
		h10 := w.heightSample(cellX+1, cellY)
		h01 := w.heightSample(cellX, cellY+1)
		h11 := w.heightSample(cellX+1, cellY+1)
		top := h00 + (h10-h00)*fracX
		bottom := h01 + (h11-h01)*fracX
		return top + (bottom-top)*fracY
	}

	var rows [4]float64
	for row := -1; row <= 2; row++ {
		rows[row+1] = monotonicCubic(
			w.heightSample(cellX-1, cellY+row),
			w.heightSample(cellX, cellY+row),
			w.heightSample(cellX+1, cellY+row),
			w.heightSample(cellX+2, cellY+row),
			fracX,
		)
	}
	return monotonicCubic(rows[0], rows[1], rows[2], rows[3], fracY)
}

func (w *World) heightSample(x, y int) float64 {
	x = max(0, min(w.canvasWidth-1, x))
	y = max(0, min(w.canvasHeight-1, y))
	return float64(w.HeightMap[y*w.canvasWidth+x])
}

func monotonicCubic(p0, p1, p2, p3, t float64) float64 {
	d0 := p1 - p0
	d1 := p2 - p1
	d2 := p3 - p2
	m1 := monotonicSlope(d0, d1)
	m2 := monotonicSlope(d1, d2)
	t2 := t * t
	t3 := t2 * t
	return (2*t3-3*t2+1)*p1 + (t3-2*t2+t)*m1 + (-2*t3+3*t2)*p2 + (t3-t2)*m2
}

func monotonicSlope(before, after float64) float64 {
	if before*after <= 0 {
		return 0
	}
	return 2 * before * after / (before + after)
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
