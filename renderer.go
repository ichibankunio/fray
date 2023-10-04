package fray

import (
	_ "embed"
	"image"
	"image/color"
	"runtime"

	"github.com/hajimehoshi/bitmapfont/v2"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/ichibankunio/flib"
	"github.com/ichibankunio/flib/ui"
	"github.com/ichibankunio/fvec/vec2"
	"github.com/ichibankunio/fray/spriteeditor"
	"github.com/ichibankunio/fvec/vec3"
)

//go:embed shaders/renderer.kage
var shaderByte []byte

//go:embed shaders/renderWithNoTextures.kage
var shaderByte2 []byte

type Renderer struct {
	Cam *Camera
	Stk *Stick
	Wld *World

	screenWidth  float64
	screenHeight float64

	shader  *ebiten.Shader
	shader2 *ebiten.Shader

	textures [4]*ebiten.Image

	texSize int

	levelWidth  int
	levelHeight int

	// playerAnimationIndex int
	counter int

	jumpButton *ui.Button
	jumpCounter int
	JumpCountMax int

	SpriteParameterNum int
	SpriteParameters   []float32
}

func (r *Renderer) Init(screenWidth, screenHeight float64, texSize int) {
	r.Cam = &Camera{}
	r.Cam.Init(screenWidth, screenHeight)

	r.Stk = &Stick{}
	r.Stk.Init(screenWidth, screenHeight)

	r.Wld = &World{}
	r.Wld.Init(screenWidth, screenHeight)

	r.texSize = texSize

	r.screenWidth = screenWidth
	r.screenHeight = screenHeight

	var err error
	r.shader, err = ebiten.NewShader(shaderByte)
	if err != nil {
		panic(err)
	}

	r.shader2, err = ebiten.NewShader(shaderByte2)
	if err != nil {
		panic(err)
	}

	r.counter = 0
	// r.playerAnimationIndex = 16
	// r.textures = textures

	r.SpriteParameterNum = 9

	r.jumpButton = ui.NewButton("ジャンプ", int(r.screenWidth)-100, 60, 100, 100, bitmapfont.Face, ui.ThemeRect, color.White, color.RGBA{20, 20, 20, 100}, color.RGBA{20, 20, 20, 100})

	r.jumpCounter = 0
	r.JumpCountMax = 5
}

func (r *Renderer) SetTextures(textures [4]*ebiten.Image) {
	r.textures = textures
}

func (r *Renderer) NewTextureSheet(src []*ebiten.Image) *ebiten.Image {
	sheet := ebiten.NewImage(int(r.screenWidth), int(r.screenHeight))
	for i, s := range src {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64((i%(int(r.screenWidth)/r.texSize))*r.texSize), float64((i/(int(r.screenWidth)/r.texSize))*r.texSize))

		sheet.DrawImage(s, op)
	}

	return sheet
}

func (r *Renderer) NewTextureSheet1x2(src []*ebiten.Image) *ebiten.Image {
	sheet := ebiten.NewImage(int(r.screenWidth), int(r.screenHeight))
	for i, s := range src {
		op := &ebiten.DrawImageOptions{}
		// op.GeoM.Translate(float64((i%(int(r.screenWidth)/r.texSize))*r.texSize), float64((i/(int(r.screenWidth)/(r.texSize*2)))*r.texSize*2))
		op.GeoM.Translate(float64((i%(int(r.screenWidth)/r.texSize))*r.texSize), float64((i/(int(r.screenWidth)/(r.texSize)))*r.texSize))

		sheet.DrawImage(s.SubImage(image.Rect(0, 0, r.texSize, r.texSize)).(*ebiten.Image), op)
	}

	return sheet
}

// func (r *Renderer) SetShader(b []byte) error {
// 	var err error
// 	r.shader, err = ebiten.NewShader(b)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func (r *Renderer) SetShader(s *ebiten.Shader) {
	r.shader = s
}

func (r *Renderer) SetShaderFromBytes(b []byte) error {
	var err error
	r.shader, err = ebiten.NewShader(b)
	if err != nil {
		return err
	}

	return nil
}

func (r *Renderer) SetLevel(level [][]float32, width, height int) {
	r.Wld.level = level
	r.levelWidth = width
	r.levelHeight = height
}

func (r *Renderer) SetLevelUint8(level [4][]uint8, width, height int) {
	r.Wld.levelUint8 = level
	r.levelWidth = width
	r.levelHeight = height
}

func (r *Renderer) GetScreenWidth() float64 {
	return r.screenWidth
}

func (r *Renderer) GetScreenHeight() float64 {
	return r.screenHeight
}

func (r *Renderer) Update() {
	// r.updateCamera()
	// r.UpdateCamRotationByMouse()
	// r.UpdateCamRotationByTouch()

	if runtime.GOOS == "darwin" {
		r.UpdateCamRotationAroundSubjectByMouse()
	} else {
		r.UpdateCamRotationAroundSubjectByTouch()
	}
	r.UpdateCamPos()
	r.UpdateCamPosZ()
	// r.UpdateCameraPos()

	r.GetDistanceInterferenceFromSubject()

	r.updateSpriteParameters()
	r.sortSpriteParameters()

	spriteeditor.WriteTexture(r.textures[2], r.SpriteParameters)

	// if ebiten.IsKeyPressed(ebiten.KeyG) {
	// 	r.Cam.RotateHorizontalAroundSubject(0.021)
	// }
	// if ebiten.IsKeyPressed(ebiten.KeyH) {
	// 	r.Cam.RotateHorizontalAroundSubject(-0.021)
	// }
	// fmt.Printf("cam%f, sub%f\n", r.Cam.pos, r.Cam.subjectPos)

	r.counter++
	// if r.Cam.v.SquaredLength() > 0 {
	// 	if r.counter%2 == 0 {
	// 		r.playerAnimationIndex = (r.playerAnimationIndex + 1) % 16
	// 	}
	// } else {
	// 	r.playerAnimationIndex = 16
	// }
}

func (r *Renderer) Draw(screen *ebiten.Image) {
	r.renderWall(screen)
	// r.renderWithNoTextures(screen)

	// r.Wld.DrawTopView(screen)

	// ebitenutil.DrawRect(screen, r.Cam.pos.X/2-2, r.Cam.pos.Y/2-2, 4, 4, color.RGBA{255, 0, 0, 255})

	// ebitenutil.DrawLine(screen, r.Cam.pos.X/2, r.Cam.pos.Y/2, r.Cam.pos.X/2+r.Cam.dir.X*200, r.Cam.pos.Y/2+r.Cam.dir.Y*200, color.RGBA{255, 0, 0, 255})

	// s.fps.Draw(screen)
	// s.debug.Draw(screen)

	// r.Stk.Draw(screen)
	// screen.DrawImage(r.textures[0], nil)

	r.jumpButton.Draw(screen)
}

func (r *Renderer) DrawTopView(screen *ebiten.Image) {
	r.Wld.DrawTopView(screen)
}

func (r *Renderer) renderWall(screen *ebiten.Image) {

	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]interface{}{
		"ScreenSize": []float32{float32(r.screenWidth), float32(r.screenHeight)},
		"Pos":        []float32{float32(r.Cam.pos.X / float64(r.texSize)), float32(r.Cam.pos.Y / float64(r.texSize))},
		"Dir":        []float32{float32(r.Cam.dir.X), float32(r.Cam.dir.Y)},
		"Plane":      []float32{float32(r.Cam.plane.X), float32(r.Cam.plane.Y)},
		// "TexSize":    float32(r.Wld.texSize),
		// "MapSize":    []float32{float32(len(r.Wld.level[0])), float32(len(r.Wld.level))},
		// "WorldSize":   []float32{float32(r.Wld.width), float32(r.Wld.height)},
		// "PosDirPlane": []float32{
		// 	float32(r.Cam.pos.X / float64(r.texSize)), float32(r.Cam.pos.Y / float64(r.texSize)),
		// 	float32(r.Cam.dir.X), float32(r.Cam.dir.Y),
		// 	float32(r.Cam.plane.X), float32(r.Cam.plane.Y),
		// },

		"PosZ": float32(r.Cam.pos.Z / float64(r.texSize)),
		// "ShooterFloatingHeight": float32(r.Cam.pos.Z - r.GetGroundHeight(r.Cam.pos) - r.Cam.shooterHeight),
		// "SubjectPos":            []float32{float32(r.Cam.subjectPos.X), float32(r.Cam.subjectPos.Y)},
		"Pitch":              r.Cam.pitch,
		"SpriteNum":          len(r.Wld.Sprites),
		"SpriteParameterNum": r.SpriteParameterNum,

		"TexSize": float32(r.texSize),

		// "PlayerAnimationIndex": float32(r.playerAnimationIndex),
		// "Level":       r.Wld.level[0],
		// "FloorLevel":  r.Wld.level[1],

		// "SpriteParam": r.Wld.SpriteRenderParam,
	}

	op.Images[0] = r.textures[0] //wall(texture), sprite(texture)
	op.Images[1] = r.textures[1] //floor(texture)
	op.Images[2] = r.textures[2] //sprite(data)
	op.Images[3] = r.textures[3] //map(data)
	screen.DrawRectShader(int(r.screenWidth), int(r.screenHeight), r.shader, op)
}

/*
func (r *Renderer) renderWithNoTextures(screen *ebiten.Image) {
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]interface{}{
		"PosDirPlane": []float32{
			float32(r.Cam.pos.X / float64(r.Wld.gridSize)), float32(r.Cam.pos.Y / float64(r.Wld.gridSize)),
			float32(r.Cam.dir.X), float32(r.Cam.dir.Y),
			float32(r.Cam.plane.X), float32(r.Cam.plane.Y),
		},
	}

	op.Images[0] = r.textures[0]
	op.Images[1] = r.textures[1]
	op.Images[2] = r.textures[2]
	// op.Images[3] = r.textures[3]
	screen.DrawRectShader(int(r.screenWidth), int(r.screenHeight), r.shader2, op)
}
*/

func (r *Renderer) UpdateCamPos() {
	// delta := vec3.New(0, 0, 0)
	r.Cam.v = vec3.New(0, 0, 0)
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) || r.Stk.Input[0] == STICK_UP {
		// if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.GamepadAxisValue(0, 1) < -0.1 || r.Stk.Input[0] == STICK_UP {
		r.Cam.v = r.collisionCheckedDelta(r.Cam.subjectPos, r.Cam.dir.Scale(r.Cam.Speed), r.Cam.collisionDistance)

		// r.Cam.pos = r.GetValidPos(r.Cam.por.X + r.Cam.dir.X*v, r.Cam.por.Y + r.Cam.dir.Y*v)
	} else if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) || r.Stk.Input[0] == STICK_DOWN {
		// } else if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.GamepadAxisValue(0, 1) > 0.1 || r.Stk.Input[0] == STICK_DOWN {
		r.Cam.v = r.collisionCheckedDelta(r.Cam.subjectPos, r.Cam.dir.Scale(-r.Cam.Speed), r.Cam.collisionDistance)
		// r.Cam.pos = r.Cam.pos.Add(delta)
		// r.Cam.subjectPos = r.Cam.subjectPos.Add(delta)

	} else if ebiten.IsKeyPressed(ebiten.KeyD) || r.Stk.Input[0] == STICK_RIGHT {
		// } else if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.GamepadAxisValue(0, 0) > 0.1 || r.Stk.Input[0] == STICK_RIGHT {
		r.Cam.v = r.collisionCheckedDelta(r.Cam.subjectPos, r.Cam.plane.Scale(-r.Cam.Speed), r.Cam.collisionDistance)
		// r.Cam.pos = r.Cam.pos.Add(delta)
		// r.Cam.subjectPos = r.Cam.subjectPos.Add(delta)

	} else if ebiten.IsKeyPressed(ebiten.KeyA) || r.Stk.Input[0] == STICK_LEFT {
		// } else if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.GamepadAxisValue(0, 0) < -0.1 || r.Stk.Input[0] == STICK_LEFT {
		r.Cam.v = r.collisionCheckedDelta(r.Cam.subjectPos, r.Cam.plane.Scale(r.Cam.Speed), r.Cam.collisionDistance)
		// r.Cam.pos = r.Cam.pos.Add(delta)
		// r.Cam.subjectPos = r.Cam.subjectPos.Add(delta)

	}

	// blockLastFrame := r.Cam.subjectPos.Scale(1/float64(r.texSize)).Floor()

	r.Cam.pos = r.Cam.pos.Add(r.Cam.v)
	r.Cam.subjectPos = r.Cam.subjectPos.Add(r.Cam.v)

	// if inpututil.IsKeyJustReleased(ebiten.KeyW) {
	// 	r.Cam.v.Y = 0
	// }
	// if inpututil.IsKeyJustReleased(ebiten.KeyS) {
	// 	r.Cam.v.Y = 0
	// }

	/*
		dist := 16.0

		if r.Cam.subjectPos.Z > r.GetGroundHeight(r.Cam.subjectPos) + r.Cam.shooterHeight && r.Cam.vZ < 0 {
			blockThisFrame := r.Cam.subjectPos.Scale(1/float64(r.texSize)).Floor()
			if blockLastFrame.X != blockThisFrame.X {
				if delta.X < 0 {
					r.Cam.subjectPos.X = r.Cam.subjectPos.X - dist
					r.Cam.pos.X = r.Cam.pos.X - dist
				}else if delta.X > 0 {
					r.Cam.subjectPos.X = r.Cam.subjectPos.X + dist
					r.Cam.pos.X = r.Cam.pos.X + dist

				}
			}else if blockLastFrame.Y != blockThisFrame.Y {
				if delta.Y < 0 {
					r.Cam.subjectPos.Y = r.Cam.subjectPos.Y - dist
					r.Cam.pos.Y = r.Cam.pos.Y - dist

				}else if delta.Y > 0 {
					r.Cam.subjectPos.Y = r.Cam.subjectPos.Y + dist
					r.Cam.pos.Y = r.Cam.pos.Y + dist

				}
			}

		}
	*/
}

func (r *Renderer) IsRunningOnGround() bool {
	return r.Cam.v.SquaredLength() > 0 && r.Cam.subjectPos.Z - (r.GetGroundHeight(r.Cam.subjectPos)+r.Cam.shooterHeight) == 0
}

func (r *Renderer) UpdateCamPosZ() {
	// if inpututil.IsKeyJustReleased(ebiten.KeySpace) || flib.IsThereJustReleasedTouch(r.jumpButton.Spr.Pos, vec2.New(float64(r.jumpButton.Spr.Img.Bounds().Dx()), float64(r.jumpButton.Spr.Img.Bounds().Dy()))) {
	// if (inpututil.IsKeyJustReleased(ebiten.KeySpace) || flib.IsThereJustReleasedTouch(r.jumpButton.Spr.Pos, vec2.New(100, 100))) && r.Cam.subjectPos.Z - (r.GetGroundHeight(r.Cam.subjectPos)+r.Cam.shooterHeight) == 0 {
	if (inpututil.IsKeyJustReleased(ebiten.KeySpace) || flib.IsThereJustReleasedTouch(r.jumpButton.Spr.Pos, vec2.New(100, 100))) && r.jumpCounter < r.JumpCountMax {
		r.Cam.vZ = 4
		r.jumpCounter ++
	}

	r.Cam.vZ += GRAVITY

	delta := r.collisionCheckedDeltaZ(r.Cam.subjectPos, r.Cam.vZ)
	r.Cam.subjectPos.Z += delta
	if delta == 0 {
		r.Cam.vZ = 0
		r.Cam.subjectPos.Z = (r.GetGroundHeight(r.Cam.subjectPos) + r.Cam.shooterHeight)
	}

	if r.Cam.subjectPos.Z - (r.GetGroundHeight(r.Cam.subjectPos)+r.Cam.shooterHeight) == 0 {
		r.jumpCounter = 0
	}

	// if r.GetGroundHeight(r.Cam.subjectPos) < r.GetGroundHeight(r.Cam.pos) {
	// 	r.Cam.pos.Z = r.collisionCheckedDeltaZ(r.Cam.pos, r.Cam.vZ)
	// }else {
	// 	r.Cam.pos.Z = r.Cam.subjectPos.Z
	// }
	r.Cam.pos.Z = r.Cam.subjectPos.Z
}

func (r *Renderer) GetTexSize() int {
	return r.texSize
}
