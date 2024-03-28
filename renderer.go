package fray

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"math"
	"runtime"

	"github.com/hajimehoshi/bitmapfont/v2"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/ichibankunio/fray/spriteeditor"
	"github.com/ichibankunio/fui"
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

	canvasWidth  int
	canvasHeight int

	shader  *ebiten.Shader
	shader2 *ebiten.Shader

	Textures [4]*ImageSrc

	aimPos        vec3.Vec3
	aimDirection  AimDirection
	HandTextureID int

	texSize int

	// levelWidth  int
	// levelHeight int

	// playerAnimationIndex int
	counter int

	jumpButton   *fui.Button
	jumpCounter  int
	JumpCountMax int

	SpriteParameterNum int
	SpriteParameters   []float32
}

type AimDirection int

const (
	AIM_DIR_NORTH AimDirection = iota
	AIM_DIR_SOUTH
	AIM_DIR_EAST
	AIM_DIR_WEST
	AIM_DIR_TOP
)

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

	r.canvasWidth = canvasWidth
	r.canvasHeight = canvasHeight

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

	r.jumpButton = fui.NewButton("Jump", vec2.New((r.screenWidth)-120, 40), vec2.New(80, 80), bitmapfont.Face, fui.ThemeRect, color.Black, color.RGBA{0, 0, 0, 32}, color.Transparent)

	r.jumpCounter = 0
	r.JumpCountMax = 100

	r.aimPos = vec3.New(0, 0, 0)

	r.HandTextureID = 0
}

func (r *Renderer) SetHandTextureID(id int) {
	r.HandTextureID = id
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

// func (r *Renderer) SetLevelUint8(level [4][]uint8, width, height int) {
// 	r.Wld.levelUint8 = level
// 	r.levelWidth = width
// 	r.levelHeight = height
// }

func (r *Renderer) GetAimPosition() vec3.Vec3 {
	return r.aimPos
}

func (r *Renderer) GetAimDirection() AimDirection {
	return r.aimDirection
}

func (r *Renderer) GetAimPositionFromScreen(screen *ebiten.Image) {
	fragment := screen.At(screen.Bounds().Dx()/2, screen.Bounds().Dy()/2)
	x, y, z, direction := fragment.RGBA()
	// fmt.Println(uint8(x), uint8(y), uint8(z), uint8(a))

	r.aimPos.X = float64(uint8(x))
	r.aimPos.Y = float64(uint8(y))
	r.aimPos.Z = float64(uint8(z))
	r.aimDirection = AimDirection(uint8(direction))
}

func (r *Renderer) CalculateAimPosition() {
	aimDistance := 0.0
	origin := r.Cam.subjectPos.Add(vec3.New(0, 0, -r.Cam.shooterHeight))
	overallDistance := 0.0
	for i := 0; i < 1; i++ {
		fmt.Println("i", i)
		ray := r.castRayMultiHeight(r.Cam.dir, r.Cam.plane, origin, RAY_HIT_UP)

		//カメラのz座標が128なら128/64 = 2倍aimDistanceを長くする
		aimDistance = math.Abs(-r.screenHeight/(float64(r.Cam.pitch))) * (1 + ((r.Cam.pos.Z - origin.Z) / float64(r.texSize)))

		fmt.Printf("初期: aimDistance: %.2f, detectedWallHeight: %d, perpWallDist: %.2f\n", aimDistance, ray.detectedWallHeight, ray.perpWallDist)

		if ray.detectedWallHeight > 0 && ray.perpWallDist < aimDistance && aimDistance > 0 { //遮蔽物があればaimPos.x, aimPos.yはより近いところにあるはずなのでaimDistanceをより近いところに
			aimDistance = ray.perpWallDist
			if r.Cam.dir.X < 0 || r.Cam.dir.Y < 0 {
				aimDistance += 0.001
			}
			fmt.Printf("近いところに衝突したら: aimDistance: %.2f, detectedWallHeight: %d, perpWallDist: %.2f\n", aimDistance, ray.detectedWallHeight, ray.perpWallDist)

		}

		if ray.detectedWallHeight > 0 { //遮蔽物があるとき
			lineHeight := r.screenHeight / ray.perpWallDist * float64(ray.detectedWallHeight)
			fmt.Println("lineHeight: ", lineHeight)

			// aimがどの高さのブロックを指しているか(pointedZ=0(基準平面)は常に同じ高さになるようにした)
			pointedZ := math.Ceil((lineHeight+float64(r.Cam.pitch))/lineHeight*float64(ray.detectedWallHeight)-float64(ray.detectedWallHeight)+1) + (r.Cam.pos.Z-float64(r.texSize))/float64(r.texSize)

			fmt.Println("pointedZ: ", pointedZ)

			if pointedZ > float64(ray.detectedWallHeight) {
				fmt.Println("高すぎ")
				/*
					aimDistance = math.Abs(-r.screenHeight / (float64(r.Cam.pitch))) * (1 + ((r.Cam.pos.Z - float64(ray.detectedWallHeight) * float64(r.texSize) - float64(r.texSize)) / float64(r.texSize)))
					fmt.Println("aimDistance: ", aimDistance)


					hitPos := vec2.New(origin.X + r.Cam.dir.X*aimDistance*float64(r.texSize), origin.Y + r.Cam.dir.Y*aimDistance*float64(r.texSize)).Scale(1/float64(r.texSize)).Floor()
					r.aimPos.X = hitPos.X
					r.aimPos.Y = hitPos.Y
					fmt.Println("hitPos: ", hitPos)

					r.aimPos.Z = float64(r.Wld.HeightMap[int(hitPos.Y)*r.canvasWidth+int(hitPos.X)])
					fmt.Println("r.aimPos.Z: ", r.aimPos.Z, "ray.detectedWallHeight", ray.detectedWallHeight)
					if r.aimPos.Z != float64(ray.detectedWallHeight) {

						fmt.Println("無効")
						r.aimPos.X = -1
						r.aimPos.Y = -1
						r.aimPos.Z = -1
						return
					}
					fmt.Println("hitPos: ", hitPos, "r.aimPos.Z: ", r.aimPos.Z)
					return
				*/

				origin.Z = float64(ray.detectedWallHeight) * float64(r.texSize)
				origin.X = ray.hitPos.X
				origin.Y = ray.hitPos.Y
				ray2 := r.castRayMultiHeight(r.Cam.dir, r.Cam.plane, origin, RAY_HIT_UP)
				fmt.Println("ray2.detectedwallHeight", ray2.detectedWallHeight)

				aimDistance = math.Abs(-r.screenHeight/(float64(r.Cam.pitch))) * (1 + ((r.Cam.pos.Z - float64(ray.detectedWallHeight)*float64(r.texSize) - float64(r.texSize)) / float64(r.texSize)))
				fmt.Println("aimDistance: ", aimDistance)
				// if ray2.detectedWallHeight > ray.detectedWallHeight {
				// }

				//aimDistanceからaimPos.xyを決定
				playerIsAt := r.Cam.GetSubjectPos().Scale(1 / float64(r.GetTexSize()))
				dir := vec3.NewFromVec2(r.Cam.GetDir())
				aimPos := playerIsAt.Add(dir.Scale(aimDistance))
				r.aimPos.X = math.Floor(aimPos.X)
				r.aimPos.Y = math.Floor(aimPos.Y)
				if aimDistance > ray2.perpWallDist+ray.perpWallDist {
					r.aimPos.Z = float64(ray2.detectedWallHeight)
				}
				println(int(r.aimPos.X), int(r.aimPos.Y), int(r.aimPos.Z), "----------------------------")

				return

			} else if pointedZ <= 0 {
				fmt.Println("低すぎ")
				// origin.Z = float64(ray.detectedWallHeight) * float64(r.texSize)
				// continue
				r.aimPos.Z = 0
				// aimDistance = math.Abs(-r.screenHeight / (float64(r.Cam.pitch)))
				aimDistance = math.Abs(-r.screenHeight/(float64(r.Cam.pitch))) * (1 + ((r.Cam.pos.Z - origin.Z - float64(r.texSize)) / float64(r.texSize)))

				// aimDistance += aimDistance * (origin.Z - r.aimPos.Z*float64(r.texSize)) / float64(r.texSize)
				fmt.Println("aimDistance: ", aimDistance)
				playerIsAt := r.Cam.GetSubjectPos().Scale(1 / float64(r.GetTexSize()))
				dir := vec3.NewFromVec2(r.Cam.GetDir())
				aimPos := playerIsAt.Add(dir.Scale(aimDistance))
				r.aimPos.X = math.Floor(aimPos.X)
				r.aimPos.Y = math.Floor(aimPos.Y)
				fmt.Println("r.aimPos: ", r.aimPos)
				return

			} else {
				r.aimPos.Z = pointedZ

				break
			}
		} else { //遮蔽物ないとき1(= aimが床を直接指すとき)、aimPos.zはその遮蔽物の高さ
			// r.aimPos.Z = math.Floor(r.GetGroundHeight(r.aimPos.Scale(float64(r.texSize))) / float64(r.texSize))
			// r.aimPos.Z = r.Cam.pos.Z/float64(r.texSize) - 1
			// r.aimPos.Z = 0
			downRay := r.castRayMultiHeight(r.Cam.dir, r.Cam.plane, origin, RAY_HIT_DOWN)
			fmt.Printf("downRay.perpwalldist: %.2f, downray.detectedwallHeight: %d\n", downRay.perpWallDist, downRay.detectedWallHeight)
			if overallDistance+downRay.perpWallDist < aimDistance {
				r.aimPos.Z = float64(r.Wld.HeightMap[int(downRay.hitPosOnMap.Y)*r.canvasWidth+int(downRay.hitPosOnMap.X)]) * float64(r.texSize)
				fmt.Println("降りる")
			} else {
				r.aimPos.Z = origin.Z / float64(r.texSize)
				fmt.Println("フラット")
			}
			fmt.Println("遮蔽物なし")
			// origin.Z -= float64(r.texSize)
			// continue

			fmt.Println("diff: ", origin.Z-r.aimPos.Z*float64(r.texSize), origin.Z, r.aimPos.Z*float64(r.texSize))

			// additionalDistance := r.screenHeight * (origin.Z - r.aimPos.Z*float64(r.texSize))/float64(r.texSize)
			// aimDistance += (origin.Z - r.aimPos.Z*float64(r.texSize))/float64(r.texSize)
			// fmt.Println(additionalDistance)

			//aimDistanceを伸ばす(相似な図形の考え方)
			aimDistance += aimDistance * (origin.Z - r.aimPos.Z*float64(r.texSize)) / float64(r.texSize)
		}

	}

	//aimが遠すぎるところを指していたら無効とする
	if aimDistance < 0 || math.Abs(aimDistance) > 5 {
		r.aimPos.X = -1
		r.aimPos.Y = -1
		r.aimPos.Z = -1
		return
	}

	//aimDistanceからaimPos.xyを決定
	playerIsAt := r.Cam.GetSubjectPos().Scale(1 / float64(r.GetTexSize()))
	dir := vec3.NewFromVec2(r.Cam.GetDir())
	aimPos := playerIsAt.Add(dir.Scale(aimDistance))
	r.aimPos.X = math.Floor(aimPos.X)
	r.aimPos.Y = math.Floor(aimPos.Y)

	println(int(r.aimPos.X), int(r.aimPos.Y), int(r.aimPos.Z), "----------------------------")

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
		r.UpdateCamRotationAroundSubjectByTouch(true)
	}
	r.UpdateCamPos(r.Cam.subjectPos)
	r.UpdateCamPosZ()
	// r.UpdateCameraPos()

	r.GetDistanceInterferenceFromSubject()

	r.updateSpriteParameters()
	r.sortSpriteParameters()

	spriteeditor.WriteTexture(r.Textures[0].Src, r.SpriteParameters, r.Textures[0].Offset)

	// r.CalculateAimPosition()
	r.jumpButton.SimpleUpdate()

	r.counter++
}

func (r *Renderer) Draw(screen *ebiten.Image) {
	r.renderWall(screen)

	r.GetAimPositionFromScreen(screen)

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

		"PosZ":               float32(r.Cam.pos.Z / float64(r.texSize)),
		"Pitch":              r.Cam.pitch,
		"SpriteNum":          len(r.Wld.Sprites),
		"SpriteParameterNum": r.SpriteParameterNum,

		"AimPos":        []float32{float32(r.aimPos.X), float32(r.aimPos.Y), float32(r.aimPos.Z)},
		"HandTextureID": float32(r.HandTextureID),
		"TexSize":       float32(r.texSize),
		"WorldSize":     []float32{float32(r.Wld.canvasWidth), float32(r.Wld.canvasHeight)},
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
	if ebiten.IsKeyPressed(ebiten.KeyW) || r.Stk.Input[0] == STICK_UP {
		// if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.GamepadAxisValue(0, 1) < -0.1 || r.Stk.Input[0] == STICK_UP {
		r.Cam.v = r.collisionCheckedDelta(playerPos, r.Cam.dir.Scale(r.Cam.Speed), r.Cam.collisionDistance)

		// r.Cam.pos = r.GetValidPos(r.Cam.por.X + r.Cam.dir.X*v, r.Cam.por.Y + r.Cam.dir.Y*v)
	} else if ebiten.IsKeyPressed(ebiten.KeyS) || r.Stk.Input[0] == STICK_DOWN {
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
	if (inpututil.IsKeyJustReleased(ebiten.KeySpace) || r.jumpButton.IsTouchJustReleased()) && r.jumpCounter < r.JumpCountMax {
		r.Cam.vZ = 6
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

	r.Cam.pos.Z = r.Cam.subjectPos.Z
}

func (r *Renderer) GetTexSize() int {
	return r.texSize
}
