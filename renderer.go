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

//go:embed shaders/pseudoShadow.kage
var pseudoShadowShaderByte []byte

type PseudoShadowConfig struct {
	Enabled          bool
	LightDir         [2]float32
	GroundY          float32
	ShadowStrength   float32
	VignetteStrength float32
	TintStrength     float32
}

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

	pseudoShadowShader *ebiten.Shader
	worldBuffer        *ebiten.Image
	pseudoShadowConfig PseudoShadowConfig

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

	SpriteParameterNum     int
	SpriteParameters       []float32
	lastSpriteParameters   []float32
	spriteParametersBuffer []float32

	CeilingHeight             float32
	CeilingTextureID          float32
	DefaultFloorTextureID     float32
	DefaultFloorColor         [3]float32
	POVScale                  float32
	AimHighlightEnabled       bool
	ControlInputEnabled       bool
	JumpButtonVisible         bool
	CameraInterferenceEnabled bool
	VerticalMovementEnabled   bool
	OccludingWallFadeEnabled  bool
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
	r.pseudoShadowShader, _ = ebiten.NewShader(pseudoShadowShaderByte)
	r.worldBuffer = ebiten.NewImage(int(r.screenWidth), int(r.screenHeight))

	r.counter = 0
	// r.playerAnimationIndex = 16
	// r.textures = textures

	r.SpriteParameterNum = 9

	r.jumpButton = fui.NewButton("Jump", vec2.New((r.screenWidth)-120, 40), vec2.New(80, 80), bitmapfont.Face, fui.ThemeRect, color.Black, color.RGBA{0, 0, 0, 32}, color.Transparent)

	r.jumpCounter = 0
	r.JumpCountMax = 100

	r.aimPos = vec3.New(0, 0, 0)

	r.HandTextureID = 0
	r.CeilingHeight = 0
	r.CeilingTextureID = -1
	r.DefaultFloorTextureID = -1
	r.DefaultFloorColor = [3]float32{240.0 / 255.0, 240.0 / 255.0, 240.0 / 255.0}
	r.POVScale = 1
	r.AimHighlightEnabled = true
	r.ControlInputEnabled = true
	r.JumpButtonVisible = true
	r.CameraInterferenceEnabled = true
	r.VerticalMovementEnabled = true
	r.OccludingWallFadeEnabled = false
	r.pseudoShadowConfig = defaultPseudoShadowConfig(float32(r.screenHeight))
	if r.pseudoShadowShader == nil {
		r.pseudoShadowConfig.Enabled = false
	}
}

func defaultPseudoShadowConfig(screenHeight float32) PseudoShadowConfig {
	return PseudoShadowConfig{
		Enabled:          false,
		LightDir:         [2]float32{0.58, -1.0},
		GroundY:          screenHeight * 0.7,
		ShadowStrength:   0.34,
		VignetteStrength: 0.65,
		TintStrength:     0.12,
	}
}

func (r *Renderer) SetCameraInterferenceEnabled(enabled bool) {
	r.CameraInterferenceEnabled = enabled
}

func (r *Renderer) SetOccludingWallFadeEnabled(enabled bool) {
	r.OccludingWallFadeEnabled = enabled
}

func (r *Renderer) SetVerticalMovementEnabled(enabled bool) {
	r.VerticalMovementEnabled = enabled
	if enabled {
		return
	}
	r.Cam.vZ = 0
	r.jumpCounter = 0
}

func (r *Renderer) SetCollisionBoxSize(width, depth, height float64) {
	r.Cam.SetCollisionBoxSize(width, depth, height)
}

func (r *Renderer) SetHandTextureID(id int) {
	r.HandTextureID = id
}

func (r *Renderer) SetTextures(textures [4]*ImageSrc) {
	r.Textures = textures
}

func (r *Renderer) SetCeilingHeight(h float32) {
	r.CeilingHeight = h
}

func (r *Renderer) SetCeilingTextureID(id int) {
	r.SetDefaultCeilingTextureID(id)
}

func (r *Renderer) SetDefaultCeilingTextureID(id int) {
	r.CeilingTextureID = float32(id)
}

func (r *Renderer) SetDefaultFloorTextureID(id int) {
	r.DefaultFloorTextureID = float32(id)
}

func (r *Renderer) SetDefaultFloorColor(clr color.Color) {
	if clr == nil {
		return
	}
	rr, gg, bb, _ := clr.RGBA()
	r.DefaultFloorColor = [3]float32{
		float32(rr) / 65535.0,
		float32(gg) / 65535.0,
		float32(bb) / 65535.0,
	}
}

func (r *Renderer) SetAimHighlightEnabled(enabled bool) {
	r.AimHighlightEnabled = enabled
}

func (r *Renderer) SetPOVScale(scale float32) {
	if scale <= 0 {
		scale = 1
	}
	r.POVScale = scale
}

func (r *Renderer) SetControlInputEnabled(enabled bool) {
	r.ControlInputEnabled = enabled
	if enabled {
		return
	}
	r.Stk.Input[0] = STICK_NONE
	r.Stk.Input[1] = STICK_NONE
	r.Stk.visible[0] = false
	r.Stk.visible[1] = false
	r.Stk.touchIDs[0] = -1
	r.Stk.touchIDs[1] = -1
}

func (r *Renderer) SetJumpButtonVisible(visible bool) {
	r.JumpButtonVisible = visible
}

func (r *Renderer) SetPseudoShadowEnabled(enabled bool) {
	r.pseudoShadowConfig.Enabled = enabled && r.pseudoShadowShader != nil
}

func (r *Renderer) SetPseudoShadowConfig(cfg PseudoShadowConfig) {
	if cfg.GroundY == 0 {
		cfg.GroundY = float32(r.screenHeight) * 0.7
	}
	cfg.ShadowStrength = clamp01(cfg.ShadowStrength)
	cfg.VignetteStrength = clamp01(cfg.VignetteStrength)
	cfg.TintStrength = clamp01(cfg.TintStrength)
	cfg.Enabled = cfg.Enabled && r.pseudoShadowShader != nil
	r.pseudoShadowConfig = cfg
}

func (r *Renderer) GetPseudoShadowConfig() PseudoShadowConfig {
	return r.pseudoShadowConfig
}

func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func (r *Renderer) NewTextureSheet(src []*ebiten.Image) *ImageSrc {
	cols := int(r.screenWidth) / r.texSize
	if cols < 1 {
		cols = 1
	}
	sheetW := int(r.screenWidth)
	rows := (len(src) + cols - 1) / cols
	if rows < 1 {
		rows = 1
	}
	sheetH := int(r.screenHeight)
	if rows*r.texSize > sheetH {
		sheetH = rows * r.texSize
	}

	textureSheet := &ImageSrc{
		Src:    ebiten.NewImage(sheetW, sheetH),
		Offset: sheetW * (sheetH - 1),
	}
	for i, s := range src {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64((i%cols)*r.texSize), float64((i/cols)*r.texSize))

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

					r.aimPos.Z = r.heightAtBlocks(hitPos.X, hitPos.Y)
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
				r.aimPos.Z = r.heightAtBlocks(downRay.hitPosOnMap.X, downRay.hitPosOnMap.Y) * float64(r.texSize)
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

func (r *Renderer) heightAtBlocks(x, y float64) float64 {
	maxX := float64(r.canvasWidth) - 1e-6
	maxY := float64(r.canvasHeight) - 1e-6
	if x < 0 {
		x = 0
	} else if x > maxX {
		x = maxX
	}
	if y < 0 {
		y = 0
	} else if y > maxY {
		y = maxY
	}
	cellX := math.Floor(x)
	cellY := math.Floor(y)
	fracX := x - cellX
	fracY := y - cellY
	heightSample := func(sampleX, sampleY int) float64 {
		if sampleX < 0 || sampleY < 0 || sampleX >= r.canvasWidth || sampleY >= r.canvasHeight {
			return 0
		}
		return float64(r.Wld.HeightMap[sampleY*r.canvasWidth+sampleX])
	}

	x0 := int(cellX)
	y0 := int(cellY)
	h00 := heightSample(x0, y0)
	h10 := heightSample(x0+1, y0)
	h01 := heightSample(x0, y0+1)
	h11 := heightSample(x0+1, y0+1)
	top := h00 + (h10-h00)*fracX
	bottom := h01 + (h11-h01)*fracX
	return top + (bottom-top)*fracY
}

func (r *Renderer) Update() {
	// r.updateCamera()
	// r.UpdateCamRotationByMouse()
	// r.UpdateCamRotationByTouch()

	if r.ControlInputEnabled {
		if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
			r.UpdateCamRotationAroundSubjectByMouse()
		} else {
			r.UpdateCamRotationAroundSubjectByTouch(true)
		}
		r.UpdateCamPos(r.Cam.GetCollisionAnchorPos())
	}
	if r.VerticalMovementEnabled {
		r.UpdateCamPosZ()
	}
	// r.UpdateCameraPos()

	if r.CameraInterferenceEnabled {
		r.GetDistanceInterferenceFromSubject()
	}

	r.updateSpriteParameters()
	r.sortSpriteParameters()
	if r.spriteParametersChanged() {
		spriteeditor.WriteTexture(r.Textures[0].Src, r.SpriteParameters, r.Textures[0].Offset)
		r.syncLastSpriteParameters()
	}

	// r.CalculateAimPosition()
	if r.JumpButtonVisible {
		r.jumpButton.SimpleUpdate()
	}

	r.counter++
}

func (r *Renderer) spriteParametersChanged() bool {
	if len(r.SpriteParameters) != len(r.lastSpriteParameters) {
		return true
	}
	for i, v := range r.SpriteParameters {
		if v != r.lastSpriteParameters[i] {
			return true
		}
	}
	return false
}

func (r *Renderer) syncLastSpriteParameters() {
	if cap(r.lastSpriteParameters) < len(r.SpriteParameters) {
		r.lastSpriteParameters = make([]float32, len(r.SpriteParameters))
	} else {
		r.lastSpriteParameters = r.lastSpriteParameters[:len(r.SpriteParameters)]
	}
	copy(r.lastSpriteParameters, r.SpriteParameters)
}

func (r *Renderer) Draw(screen *ebiten.Image) {
	if r.worldBuffer == nil || r.worldBuffer.Bounds().Dx() != int(r.screenWidth) || r.worldBuffer.Bounds().Dy() != int(r.screenHeight) {
		r.worldBuffer = ebiten.NewImage(int(r.screenWidth), int(r.screenHeight))
	}
	r.worldBuffer.Clear()
	r.renderWall(r.worldBuffer)
	r.GetAimPositionFromScreen(r.worldBuffer)

	if r.pseudoShadowConfig.Enabled && r.pseudoShadowShader != nil {
		r.renderPseudoShadow(screen, r.worldBuffer)
	} else {
		screen.DrawImage(r.worldBuffer, nil)
	}

	// r.renderWithNoTextures(screen)

	// r.Wld.DrawTopView(screen)

	// ebitenutil.DrawRect(screen, r.Cam.pos.X/2-2, r.Cam.pos.Y/2-2, 4, 4, color.RGBA{255, 0, 0, 255})

	// ebitenutil.DrawLine(screen, r.Cam.pos.X/2, r.Cam.pos.Y/2, r.Cam.pos.X/2+r.Cam.dir.X*200, r.Cam.pos.Y/2+r.Cam.dir.Y*200, color.RGBA{255, 0, 0, 255})

	// s.fps.Draw(screen)
	// s.debug.Draw(screen)

	// r.Stk.Draw(screen)
	// screen.DrawImage(r.textures[0], nil)

	if r.JumpButtonVisible {
		r.jumpButton.Draw(screen)
	}
}

func (r *Renderer) DrawTopView(screen *ebiten.Image) {
	r.Wld.DrawTopView(screen)
}

func (r *Renderer) renderWall(screen *ebiten.Image) {
	occlusionTargetPos := r.Cam.GetCollisionAnchorPos()
	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = map[string]interface{}{
		"ScreenSize": []float32{float32(r.screenWidth), float32(r.screenHeight)},
		"Pos":        []float32{float32(r.Cam.pos.X / float64(r.texSize)), float32(r.Cam.pos.Y / float64(r.texSize))},
		"Dir":        []float32{float32(r.Cam.dir.X), float32(r.Cam.dir.Y)},
		"Plane":      []float32{float32(r.Cam.plane.X), float32(r.Cam.plane.Y)},
		"POVScale":   r.POVScale,

		"PosZ":               float32(r.Cam.pos.Z / float64(r.texSize)),
		"Pitch":              r.Cam.pitch,
		"SpriteNum":          len(r.Wld.Sprites),
		"SpriteParameterNum": r.SpriteParameterNum,

		"AimPos":                []float32{float32(r.aimPos.X), float32(r.aimPos.Y), float32(r.aimPos.Z)},
		"HandTextureID":         float32(r.HandTextureID),
		"TexSize":               float32(r.texSize),
		"WorldSize":             []float32{float32(r.Wld.canvasWidth), float32(r.Wld.canvasHeight)},
		"OcclusionTargetPos":    []float32{float32(occlusionTargetPos.X / float64(r.texSize)), float32(occlusionTargetPos.Y / float64(r.texSize)), float32(occlusionTargetPos.Z / float64(r.texSize))},
		"OccludingWallFade":     float32(0),
		"CeilingHeight":         r.CeilingHeight,
		"CeilingTextureID":      r.CeilingTextureID,
		"DefaultFloorTextureID": r.DefaultFloorTextureID,
		"DefaultFloorColor":     []float32{r.DefaultFloorColor[0], r.DefaultFloorColor[1], r.DefaultFloorColor[2]},
		"AimHighlightEnabled":   float32(0),
	}
	if r.AimHighlightEnabled {
		op.Uniforms["AimHighlightEnabled"] = float32(1)
	}
	if r.OccludingWallFadeEnabled {
		op.Uniforms["OccludingWallFade"] = float32(1)
	}

	op.Images[0] = r.Textures[0].Src //wall(texture), sprite(texture)
	op.Images[1] = r.Textures[1].Src //floor(texture)
	op.Images[2] = r.Textures[2].Src //sprite(data)
	op.Images[3] = r.Textures[3].Src //map(data)
	screen.DrawRectShader(int(r.screenWidth), int(r.screenHeight), r.shader, op)
}

func (r *Renderer) renderPseudoShadow(screen, src *ebiten.Image) {
	cfg := r.pseudoShadowConfig
	op := &ebiten.DrawRectShaderOptions{
		Uniforms: map[string]interface{}{
			"ScreenSize":       []float32{float32(r.screenWidth), float32(r.screenHeight)},
			"LightDir":         []float32{cfg.LightDir[0], cfg.LightDir[1]},
			"GroundY":          cfg.GroundY,
			"ShadowStrength":   cfg.ShadowStrength,
			"VignetteStrength": cfg.VignetteStrength,
			"TintStrength":     cfg.TintStrength,
		},
	}
	op.Images[0] = src
	screen.DrawRectShader(int(r.screenWidth), int(r.screenHeight), r.pseudoShadowShader, op)
}

func (r *Renderer) UpdateCamPos(playerPos vec3.Vec3) {
	requestedDelta := vec2.New(0, 0)
	if ebiten.IsKeyPressed(ebiten.KeyW) || r.Stk.Input[0] == STICK_UP {
		// if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.GamepadAxisValue(0, 1) < -0.1 || r.Stk.Input[0] == STICK_UP {
		requestedDelta = r.Cam.dir.Scale(r.Cam.Speed)

		// r.Cam.pos = r.GetValidPos(r.Cam.por.X + r.Cam.dir.X*v, r.Cam.por.Y + r.Cam.dir.Y*v)
	} else if ebiten.IsKeyPressed(ebiten.KeyS) || r.Stk.Input[0] == STICK_DOWN {
		// } else if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.GamepadAxisValue(0, 1) > 0.1 || r.Stk.Input[0] == STICK_DOWN {
		requestedDelta = r.Cam.dir.Scale(-r.Cam.Speed)
		// r.Cam.pos = r.Cam.pos.Add(delta)
		// r.Cam.subjectPos = r.Cam.subjectPos.Add(delta)

	} else if ebiten.IsKeyPressed(ebiten.KeyD) || r.Stk.Input[0] == STICK_RIGHT {
		// } else if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.GamepadAxisValue(0, 0) > 0.1 || r.Stk.Input[0] == STICK_RIGHT {
		strafeDir := vec2.New(-r.Cam.dir.Y, r.Cam.dir.X)
		requestedDelta = strafeDir.Scale(r.Cam.Speed)
		// r.Cam.pos = r.Cam.pos.Add(delta)
		// r.Cam.subjectPos = r.Cam.subjectPos.Add(delta)

	} else if ebiten.IsKeyPressed(ebiten.KeyA) || r.Stk.Input[0] == STICK_LEFT {
		// } else if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.GamepadAxisValue(0, 0) < -0.1 || r.Stk.Input[0] == STICK_LEFT {
		strafeDir := vec2.New(-r.Cam.dir.Y, r.Cam.dir.X)
		requestedDelta = strafeDir.Scale(-r.Cam.Speed)
		// r.Cam.pos = r.Cam.pos.Add(delta)
		// r.Cam.subjectPos = r.Cam.subjectPos.Add(delta)

	}

	r.moveCameraOnTerrain(playerPos, requestedDelta)
}

func (r *Renderer) moveCameraOnTerrain(playerPos vec3.Vec3, requestedDelta vec2.Vec2) {
	groundBeforeMove := r.GetGroundHeightUnderCollisionBox(vec2.New(r.Cam.subjectPos.X, r.Cam.subjectPos.Y)) + r.Cam.shooterHeight
	wasGrounded := math.Abs(r.Cam.subjectPos.Z-groundBeforeMove) < 0.5
	r.Cam.v = r.collisionCheckedDelta(playerPos, requestedDelta, r.Cam.collisionDistance)
	r.Cam.subjectPos = r.Cam.subjectPos.Add(r.Cam.v)
	if wasGrounded {
		groundAfterMove := r.GetGroundHeightUnderCollisionBox(vec2.New(r.Cam.subjectPos.X, r.Cam.subjectPos.Y)) + r.Cam.shooterHeight
		r.Cam.subjectPos.Z = groundAfterMove
		r.Cam.pos.Z = groundAfterMove
	}
}

func (r *Renderer) IsRunningOnGround() bool {
	groundHeight := r.GetGroundHeightUnderCollisionBox(vec2.New(r.Cam.subjectPos.X, r.Cam.subjectPos.Y))
	return r.Cam.v.SquaredLength() > 0 && r.Cam.subjectPos.Z-(groundHeight+r.Cam.shooterHeight) == 0
}

func (r *Renderer) UpdateCamPosZ() {
	// if inpututil.IsKeyJustReleased(ebiten.KeySpace) || flib.IsThereJustReleasedTouch(r.jumpButton.Spr.Pos, vec2.New(float64(r.jumpButton.Spr.Img.Bounds().Dx()), float64(r.jumpButton.Spr.Img.Bounds().Dy()))) {
	// if (inpututil.IsKeyJustReleased(ebiten.KeySpace) || flib.IsThereJustReleasedTouch(r.jumpButton.Spr.Pos, vec2.New(100, 100))) && r.Cam.subjectPos.Z - (r.GetGroundHeight(r.Cam.subjectPos)+r.Cam.shooterHeight) == 0 {
	touchJump := r.JumpButtonVisible && r.jumpButton.IsTouchJustReleased()
	if (inpututil.IsKeyJustReleased(ebiten.KeySpace) || touchJump) && r.jumpCounter < r.JumpCountMax {
		r.Cam.vZ = 6
		r.jumpCounter++
	}

	r.Cam.vZ += GRAVITY

	delta := r.collisionCheckedDeltaZ(r.Cam.subjectPos, r.Cam.vZ)
	r.Cam.subjectPos.Z += delta
	groundHeight := r.GetGroundHeightUnderCollisionBox(vec2.New(r.Cam.subjectPos.X, r.Cam.subjectPos.Y))
	if delta == 0 {
		r.Cam.vZ = 0
		r.Cam.subjectPos.Z = groundHeight + r.Cam.shooterHeight
	}

	if r.Cam.subjectPos.Z-(groundHeight+r.Cam.shooterHeight) == 0 {
		r.jumpCounter = 0
	}

	r.Cam.pos.Z = r.Cam.subjectPos.Z
}

func (r *Renderer) GetTexSize() int {
	return r.texSize
}
