package fray

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type TerrainInterpolation uint8

const (
	TerrainInterpolationFlat TerrainInterpolation = iota
	TerrainInterpolationLinear
	TerrainInterpolationSmooth
)

type TerrainTileShape uint8

const (
	TerrainTileAuto TerrainTileShape = iota
	TerrainTileFlat
	TerrainTileSlopeNorth
	TerrainTileSlopeSouth
	TerrainTileSlopeEast
	TerrainTileSlopeWest
	TerrainTileCornerNorthEast
	TerrainTileCornerSouthEast
	TerrainTileCornerSouthWest
	TerrainTileCornerNorthWest
	TerrainTileRidgeNorthSouth
	TerrainTileRidgeEastWest
	TerrainTileValleyNorthSouth
	TerrainTileValleyEastWest
)

type terrainJSONOptions struct {
	Interpolation     string            `json:"interpolation"`
	StrictConnections bool              `json:"strictConnections"`
	Tiles             []terrainJSONTile `json:"tiles"`
}

type terrainJSONTile struct {
	X          int      `json:"x"`
	Y          int      `json:"y"`
	Shape      string   `json:"shape"`
	Rotation   int      `json:"rotation,omitempty"`
	BaseHeight *float32 `json:"baseHeight,omitempty"`
	Rise       *float32 `json:"rise,omitempty"`
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
	// levelUint8 [4][]uint8

	WorldMap   [][]uint8 //texture ID map
	HeightMap  []uint8   //height map
	HeightBase []float32
	SlopeX     []float32
	SlopeY     []float32

	TerrainInterpolation TerrainInterpolation
	TerrainTileShapes    []uint8
	TerrainTileBase      []float32
	TerrainTileRise      []float32

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
	w.TerrainTileShapes = make([]uint8, canvasWidth*canvasHeight)
	w.TerrainTileBase = make([]float32, canvasWidth*canvasHeight)
	w.TerrainTileRise = make([]float32, canvasWidth*canvasHeight)
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
	if err := json.Unmarshal(data, &document); err != nil {
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
	layerCount := len(document.Layers)
	if layerCount > len(w.WorldMap) {
		layerCount = len(w.WorldMap)
	}
	for z := 0; z < layerCount; z++ {
		if len(document.Layers[z]) > len(w.WorldMap[z]) {
			return fmt.Errorf("terrain layer %d has %d cells, maximum is %d", z, len(document.Layers[z]), len(w.WorldMap[z]))
		}
		copy(w.WorldMap[z], document.Layers[z])
	}
	w.BuildHeightMapFromWorldMap()
	copy(w.TerrainTileBase, w.HeightBase)
	clear(w.TerrainTileShapes)
	clear(w.TerrainTileRise)
	for i, tile := range document.Terrain.Tiles {
		if tile.X < 0 || tile.Y < 0 || tile.X >= w.canvasWidth || tile.Y >= w.canvasHeight {
			return fmt.Errorf("terrain tile %d coordinate (%d,%d) is outside the world", i, tile.X, tile.Y)
		}
		shape, err := parseTerrainTileShape(tile.Shape, tile.Rotation)
		if err != nil {
			return fmt.Errorf("terrain tile %d: %w", i, err)
		}
		idx := tile.Y*w.canvasWidth + tile.X
		w.TerrainTileShapes[idx] = uint8(shape)
		if tile.BaseHeight != nil {
			w.TerrainTileBase[idx] = *tile.BaseHeight
		}
		if tile.Rise != nil {
			w.TerrainTileRise[idx] = *tile.Rise
		} else if shape != TerrainTileAuto && shape != TerrainTileFlat {
			w.TerrainTileRise[idx] = 1
		}
	}
	if document.Terrain.StrictConnections {
		if err := w.ValidateTerrainConnections(1.0 / 32.0); err != nil {
			return err
		}
	}
	return nil
}

func parseTerrainInterpolation(value string) (TerrainInterpolation, error) {
	switch value {
	case "", "linear":
		return TerrainInterpolationLinear, nil
	case "flat":
		return TerrainInterpolationFlat, nil
	case "smooth":
		return TerrainInterpolationSmooth, nil
	default:
		return TerrainInterpolationLinear, fmt.Errorf("unknown terrain interpolation %q", value)
	}
}

func parseTerrainTileShape(value string, rotation int) (TerrainTileShape, error) {
	if rotation < 0 || rotation >= 360 || rotation%90 != 0 {
		return TerrainTileAuto, fmt.Errorf("rotation must be 0, 90, 180, or 270, got %d", rotation)
	}
	switch value {
	case "", "auto":
		return TerrainTileAuto, nil
	case "flat":
		return TerrainTileFlat, nil
	case "slope_north":
		return TerrainTileSlopeNorth, nil
	case "slope_south":
		return TerrainTileSlopeSouth, nil
	case "slope_east":
		return TerrainTileSlopeEast, nil
	case "slope_west":
		return TerrainTileSlopeWest, nil
	case "corner_slope":
		return TerrainTileShape(int(TerrainTileCornerNorthEast) + rotation/90), nil
	case "ridge":
		if rotation%180 == 0 {
			return TerrainTileRidgeNorthSouth, nil
		}
		return TerrainTileRidgeEastWest, nil
	case "valley":
		if rotation%180 == 0 {
			return TerrainTileValleyNorthSouth, nil
		}
		return TerrainTileValleyEastWest, nil
	default:
		return TerrainTileAuto, fmt.Errorf("unknown terrain tile shape %q", value)
	}
}

func (w *World) heightAtBlocks(x, y float64) float64 {
	maxX := float64(w.canvasWidth) - 1e-6
	maxY := float64(w.canvasHeight) - 1e-6
	x = math.Max(0, math.Min(maxX, x))
	y = math.Max(0, math.Min(maxY, y))
	cellX := int(math.Floor(x))
	cellY := int(math.Floor(y))
	return w.heightInCell(cellX, cellY, x-float64(cellX), y-float64(cellY))
}

func (w *World) heightInCell(cellX, cellY int, fracX, fracY float64) float64 {
	idx := cellY*w.canvasWidth + cellX
	if idx >= 0 && idx < len(w.TerrainTileShapes) {
		shape := TerrainTileShape(w.TerrainTileShapes[idx])
		if shape != TerrainTileAuto {
			base := float64(w.TerrainTileBase[idx])
			rise := float64(w.TerrainTileRise[idx])
			switch shape {
			case TerrainTileFlat:
				return base
			case TerrainTileSlopeNorth:
				return base + rise*(1-fracY)
			case TerrainTileSlopeSouth:
				return base + rise*fracY
			case TerrainTileSlopeEast:
				return base + rise*fracX
			case TerrainTileSlopeWest:
				return base + rise*(1-fracX)
			case TerrainTileCornerNorthEast:
				return base + rise*fracX*(1-fracY)
			case TerrainTileCornerSouthEast:
				return base + rise*fracX*fracY
			case TerrainTileCornerSouthWest:
				return base + rise*(1-fracX)*fracY
			case TerrainTileCornerNorthWest:
				return base + rise*(1-fracX)*(1-fracY)
			case TerrainTileRidgeNorthSouth:
				return base + rise*(1-math.Abs(fracX*2-1))
			case TerrainTileRidgeEastWest:
				return base + rise*(1-math.Abs(fracY*2-1))
			case TerrainTileValleyNorthSouth:
				return base + rise*math.Abs(fracX*2-1)
			case TerrainTileValleyEastWest:
				return base + rise*math.Abs(fracY*2-1)
			}
		}
	}

	heightSample := func(x, y int) float64 {
		if x < 0 || y < 0 || x >= w.canvasWidth || y >= w.canvasHeight {
			return 0
		}
		return float64(w.HeightMap[y*w.canvasWidth+x])
	}
	h00 := heightSample(cellX, cellY)
	if w.TerrainInterpolation == TerrainInterpolationFlat {
		return h00
	}
	if w.TerrainInterpolation == TerrainInterpolationSmooth {
		fracX = fracX * fracX * (3 - 2*fracX)
		fracY = fracY * fracY * (3 - 2*fracY)
	}
	h10 := heightSample(cellX+1, cellY)
	h01 := heightSample(cellX, cellY+1)
	h11 := heightSample(cellX+1, cellY+1)
	top := h00 + (h10-h00)*fracX
	bottom := h01 + (h11-h01)*fracX
	return top + (bottom-top)*fracY
}

func (w *World) ValidateTerrainConnections(tolerance float64) error {
	const samples = 8
	for y := 0; y < w.canvasHeight; y++ {
		for x := 0; x < w.canvasWidth; x++ {
			if x+1 < w.canvasWidth {
				for i := 0; i <= samples; i++ {
					t := float64(i) / samples
					left := w.heightInCell(x, y, 1, t)
					right := w.heightInCell(x+1, y, 0, t)
					if math.Abs(left-right) > tolerance {
						return fmt.Errorf("terrain seam between (%d,%d) and (%d,%d): %.3f vs %.3f", x, y, x+1, y, left, right)
					}
				}
			}
			if y+1 < w.canvasHeight {
				for i := 0; i <= samples; i++ {
					t := float64(i) / samples
					top := w.heightInCell(x, y, t, 1)
					bottom := w.heightInCell(x, y+1, t, 0)
					if math.Abs(top-bottom) > tolerance {
						return fmt.Errorf("terrain seam between (%d,%d) and (%d,%d): %.3f vs %.3f", x, y, x, y+1, top, bottom)
					}
				}
			}
		}
	}
	return nil
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
