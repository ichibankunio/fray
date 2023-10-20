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
	"github.com/ichibankunio/fray/spriteeditor"
	"github.com/ichibankunio/fvec/vec2"
	"github.com/ichibankunio/fvec/vec3"
)

//go:embed shaders/renderer3dmap.kage
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

	Textures [4]*ImageSrc

	texSize int

	levelWidth  int
	levelHeight int

	// playerAnimationIndex int
	counter int

	jumpButton   *ui.Button
	jumpCounter  int
	JumpCountMax int

	SpriteParameterNum int
	SpriteParameters   []float32
}

func (r *Renderer) Init(screenWidth, screenHeight float64, canvasWidth, canvasHeight, canvasDepth int, texSize int) {
	r.Cam = &Camera{}
	r.Cam.Init(screenWidth, screenHeight)

	r.Stk = &Stick{}
	r.Stk.Init(screenWidth, screenHeight)

	r.Wld = &World{}
	r.Wld.Init(int(screenWidth), int(screenHeight), canvasWidth, canvasHeight, canvasDepth)

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

func (r *Renderer) SetTextures(textures [4]*ImageSrc) {
	r.Textures = textures
}

func (r *Renderer) NewTextureSheet(src []*ebiten.Image) *ImageSrc {
	textureSheet := &ImageSrc{
		Src:    ebiten.NewImage(int(r.screenWidth), int(r.screenHeight)),
		Offset: int(r.screenWidth) * int(r.screenHeight-1),
	}
	for i, s := range src {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64((i%(int(r.screenWidth)/r.texSize))*r.texSize), float64((i/(int(r.screenWidth)/r.texSize))*r.texSize))

		textureSheet.Src.DrawImage(s, op)
	}

	return textureSheet
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

// func (r *Renderer) SetLevel(level [][]float32, width, height int) {
// 	r.Wld.level = level
// 	r.levelWidth = width
// 	r.levelHeight = height
// }

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
	r.UpdateCamPos(r.Cam.subjectPos)
	r.UpdateCamPosZ()
	// r.UpdateCameraPos()

	r.GetDistanceInterferenceFromSubject()

	r.updateSpriteParameters()
	r.sortSpriteParameters()

	spriteeditor.WriteTexture(r.Textures[0].Src, r.SpriteParameters, r.Textures[0].Offset)

	r.counter++
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

		"PosZ":                  float32(r.Cam.pos.Z / float64(r.texSize)),
		"Pitch":                 r.Cam.pitch,
		"SpriteNum":             len(r.Wld.Sprites),
		"SpriteParameterNum":    r.SpriteParameterNum,

		"TexSize": float32(r.texSize),
	}

	op.Images[0] = r.Textures[0].Src //wall(texture), sprite(texture)
	op.Images[1] = r.Textures[1].Src //floor(texture)
	op.Images[2] = r.Textures[2].Src //sprite(data)
	op.Images[3] = r.Textures[3].Src //map(data)
	screen.DrawRectShader(int(r.screenWidth), int(r.screenHeight), r.shader, op)
}

func (r *Renderer) UpdateCamPos(playerPos vec3.Vec3) {
	// delta := vec3.New(0, 0, 0)
	r.Cam.v = vec3.New(0, 0, 0)
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) || r.Stk.Input[0] == STICK_UP {
		// if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.GamepadAxisValue(0, 1) < -0.1 || r.Stk.Input[0] == STICK_UP {
		r.Cam.v = r.collisionCheckedDelta(playerPos, r.Cam.dir.Scale(r.Cam.Speed), r.Cam.collisionDistance)

		// r.Cam.pos = r.GetValidPos(r.Cam.por.X + r.Cam.dir.X*v, r.Cam.por.Y + r.Cam.dir.Y*v)
	} else if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) || r.Stk.Input[0] == STICK_DOWN {
		// } else if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.GamepadAxisValue(0, 1) > 0.1 || r.Stk.Input[0] == STICK_DOWN {
		r.Cam.v = r.collisionCheckedDelta(playerPos, r.Cam.dir.Scale(-r.Cam.Speed), r.Cam.collisionDistance)
		// r.Cam.pos = r.Cam.pos.Add(delta)
		// r.Cam.subjectPos = r.Cam.subjectPos.Add(delta)

	} else if ebiten.IsKeyPressed(ebiten.KeyD) || r.Stk.Input[0] == STICK_RIGHT {
		// } else if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.GamepadAxisValue(0, 0) > 0.1 || r.Stk.Input[0] == STICK_RIGHT {
		r.Cam.v = r.collisionCheckedDelta(playerPos, r.Cam.plane.Scale(-r.Cam.Speed), r.Cam.collisionDistance)
		// r.Cam.pos = r.Cam.pos.Add(delta)
		// r.Cam.subjectPos = r.Cam.subjectPos.Add(delta)

	} else if ebiten.IsKeyPressed(ebiten.KeyA) || r.Stk.Input[0] == STICK_LEFT {
		// } else if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.GamepadAxisValue(0, 0) < -0.1 || r.Stk.Input[0] == STICK_LEFT {
		r.Cam.v = r.collisionCheckedDelta(playerPos, r.Cam.plane.Scale(r.Cam.Speed), r.Cam.collisionDistance)
		// r.Cam.pos = r.Cam.pos.Add(delta)
		// r.Cam.subjectPos = r.Cam.subjectPos.Add(delta)

	}

	// blockLastFrame := r.Cam.subjectPos.Scale(1/float64(r.texSize)).Floor()

	r.Cam.pos = r.Cam.pos.Add(r.Cam.v)
	r.Cam.subjectPos = r.Cam.subjectPos.Add(r.Cam.v)
}

func (r *Renderer) IsRunningOnGround() bool {
	return r.Cam.v.SquaredLength() > 0 && r.Cam.subjectPos.Z-(r.GetGroundHeight(r.Cam.subjectPos)+r.Cam.shooterHeight) == 0
}

func (r *Renderer) UpdateCamPosZ() {
	// if inpututil.IsKeyJustReleased(ebiten.KeySpace) || flib.IsThereJustReleasedTouch(r.jumpButton.Spr.Pos, vec2.New(float64(r.jumpButton.Spr.Img.Bounds().Dx()), float64(r.jumpButton.Spr.Img.Bounds().Dy()))) {
	// if (inpututil.IsKeyJustReleased(ebiten.KeySpace) || flib.IsThereJustReleasedTouch(r.jumpButton.Spr.Pos, vec2.New(100, 100))) && r.Cam.subjectPos.Z - (r.GetGroundHeight(r.Cam.subjectPos)+r.Cam.shooterHeight) == 0 {
	if (inpututil.IsKeyJustReleased(ebiten.KeySpace) || flib.IsThereJustReleasedTouch(r.jumpButton.Spr.Pos, vec2.New(100, 100))) && r.jumpCounter < r.JumpCountMax {
		r.Cam.vZ = 4
		r.jumpCounter++
	}

	r.Cam.vZ += GRAVITY

	delta := r.collisionCheckedDeltaZ(r.Cam.subjectPos, r.Cam.vZ)
	r.Cam.subjectPos.Z += delta
	if delta == 0 {
		r.Cam.vZ = 0
		r.Cam.subjectPos.Z = (r.GetGroundHeight(r.Cam.subjectPos) + r.Cam.shooterHeight)
	}

	if r.Cam.subjectPos.Z-(r.GetGroundHeight(r.Cam.subjectPos)+r.Cam.shooterHeight) == 0 {
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
