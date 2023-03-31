package fray

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type World struct {
	// level    [][]int
	// level          []float32
	level      [][]float32
	levelUint8 [4][]uint8

	// floorLevel     []float32
	// gridSize       int
	topImage *ebiten.Image
	// baseLightValue float32

	// renderMap     [SCREEN_WIDTH]float32
	// floorTexture  *ebiten.Image
	// wallTexture   *ebiten.Image
	// spriteTexture *ebiten.Image

	// width  int
	// height int

	Sprites            []*Sprite
}

func (w *World) Init(screenWidth, screenHeight float64) {
	// w.width = 10
	// w.height = 10
}

// func (w *World) NewSprite(pos vec3.Vec3, texID int) {
// 	if len(w.Sprites) < 3 {
// 		w.Sprites = append(w.Sprites, &Sprite{
// 			Pos:              pos,
// 			ID:               len(w.Sprites),
// 			TexID:            texID,
// 			Size:             vec2.New(0, 0),
// 			DistanceToCamera: 0,
// 			PosOnScreen:      vec2.New(0, 0),
// 		})
// 	}

// }

// func (w *World) NewTopView() {
// 	w.topImage = ebiten.NewImage(w.texSize*w.width, w.texSize*w.height)
// 	grid1 := ebiten.NewImage(w.texSize-2, w.texSize-2)
// 	grid1.Fill(color.RGBA{120, 120, 255, 120})
// 	grid2 := ebiten.NewImage(w.texSize-2, w.texSize-2)
// 	grid2.Fill(color.RGBA{120, 120, 120, 120})

// 	for y := 0; y < w.height; y++ {
// 		for x := 0; x < w.width; x++ {
// 			op := &ebiten.DrawImageOptions{}
// 			op.GeoM.Translate(float64(x*w.texSize+1), float64(y*w.texSize+1))
// 			switch w.level[0][y*w.width+x] {
// 			case 0:
// 				w.topImage.DrawImage(grid2, op)
// 			case 1:
// 				w.topImage.DrawImage(grid1, op)
// 			}
// 		}
// 	}

// }


func (w *World) DrawTopView(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(0.5, 0.5)
	op.GeoM.Translate(0, 0)
	screen.DrawImage(w.topImage, op)
}

