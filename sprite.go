package fray

import (
	"math"

	"github.com/ichibankunio/fvec/vec2"
	"github.com/ichibankunio/fvec/vec3"
)

type Sprite struct {
	Pos   vec3.Vec3
	ID    int
	TexID int

	PosOnScreen      vec2.Vec2
	Size             vec2.Vec2
	DistanceToCamera float32

	TexIDBegin       int
	CurrentAnimation int
	IsVisible bool
}

func (r *Renderer) NewSprite(pos vec2.Vec2, texID int, texIDBegin int, currentAnimation int, isVisible bool) {
	s := &Sprite{
		TexID:            texID,
		Pos:              vec3.New(pos.X, pos.Y, r.GetGroundHeight(vec3.New(pos.X, pos.Y, 0))+float64(r.texSize)),
		TexIDBegin:       texIDBegin,
		CurrentAnimation: currentAnimation,
		IsVisible: isVisible,
	}

	r.Wld.Sprites = append(r.Wld.Sprites, s)
}

func (r *Renderer) updateSpriteParameters() {
	if len(r.Wld.Sprites) == 0 {
		return
	}
	data := make([]float32, len(r.Wld.Sprites)*r.SpriteParameterNum)

	// fmt.Printf("%f, %f, %f\n", r.Cam.GetPos(), r.Cam.subjectPos, r.Wld.Sprites[0].Pos.Z)

	r.Wld.Sprites[0].Pos = r.Cam.subjectPos
	// r.Wld.Sprites[0].CurrentAnimation =
	if r.Cam.v.SquaredLength() > 0 {
		if r.counter%2 == 0 {
			r.Wld.Sprites[0].CurrentAnimation = (r.Wld.Sprites[0].CurrentAnimation + 1) % 16
		}
	} else {
		r.Wld.Sprites[0].CurrentAnimation = 16
	}

	// fmt.Printf("res: %d, %d\n", int(r.Wld.Sprites[0].Pos.X)%r.texSize, int(r.Wld.Sprites[0].Pos.Y)%r.texSize)

	invDet := 1.0 / (r.Cam.plane.X*r.Cam.dir.Y - r.Cam.dir.X*r.Cam.plane.Y) // 1/(ad-bc)
	for i, spr := range r.Wld.Sprites {
		relPos := spr.Pos.Sub(r.Cam.pos).Scale(1.0 / float64(r.texSize))
		transPos := vec2.New(r.Cam.dir.Y*relPos.X-r.Cam.dir.X*relPos.Y, -r.Cam.plane.Y*relPos.X+r.Cam.plane.X*relPos.Y).Scale(invDet)
		screenX := (r.screenWidth / 2) * (1.0 - transPos.X/transPos.Y)

		// println("vmove:", vMoveScreen)

		// vMoveScreen := (r.Cam.pos.Z - r.GetGroundHeight(r.Cam.pos) - r.Cam.shooterHeight)
		// println(int(r.Cam.pos.Z - r.GetGroundHeight(r.Cam.pos) - r.Cam.shooterHeight))
		// offsetByPlayerZ := math.Abs(r.screenHeight/transPos.Y) * (r.Cam.subjectPos.Z - r.GetGroundHeight(r.Cam.subjectPos) - r.Cam.shooterHeight) / float64(r.texSize)
		offsetByCamZ := math.Abs(r.screenHeight/transPos.Y) * (r.Cam.pos.Z + r.Cam.subjectPos.Z - r.Cam.pos.Z) / float64(r.texSize)
		offsetBySpriteZ := math.Abs(r.screenHeight/transPos.Y) * (spr.Pos.Z - float64(r.texSize)/2) / float64(r.texSize)
		// 2023/4/1 r.texSize/2を引いてspriteがPosよりも半ブロック分上に描画される問題を仮修正

		//calculate height of the sprite on screen
		spriteSize := vec2.New(math.Abs(r.screenHeight/transPos.Y), math.Abs(r.screenHeight/transPos.Y))
		// spriteHeight := math.Abs(SCREEN_HEIGHT / transPos.Y) //using 'transformY' instead of the real distance prevents fisheye
		// spriteWidth := math.Abs(SCREEN_HEIGHT / transPos.Y)

		// fmt.Printf("spriteSize: x: %f, y: %f\n", spriteSize.X, spriteSize.Y)

		//calculate lowest and highest pixel to fill in current stripe
		drawStart := vec2.New(-spriteSize.X/2+screenX, -spriteSize.Y/2+r.screenHeight/2+float64(r.Cam.pitch)+offsetByCamZ-offsetBySpriteZ)
		// drawEnd := vec2.New(spriteWidth/2+screenX, spriteHeight/2+SCREEN_HEIGHT/2)

		spr.Size = spriteSize.Mul(vec2.New(1/(r.screenWidth*100), 1/(r.screenHeight*100))) //0-1に正規化

		// fmt.Printf("decoded: %f\n", spr.Size.Mul(vec2.New(r.screenWidth, r.screenHeight)))

		// spr.Size = spriteSize.Scale(1.0/1000.0) //0-1に正規化
		spr.DistanceToCamera = float32(relPos.SquaredLength()) / 200.0

		if spr.PosOnScreen.Y > -0.004 && spr.PosOnScreen.Y < 0 { //-0.002など1/255=0.003921...より絶対値が小さいときencode/decodeが上手くいかないので-0.004に変えてチラつくのを回避
			spr.PosOnScreen.Y = -0.004
		}

		// println("transPos.Y" , transPos.Y)
		if transPos.Y < 0 {
			// spr.DistanceToCamera = 1.0
			spr.Size = vec2.New(0, 0)
		}
		spr.PosOnScreen = drawStart.Mul(vec2.New(1/r.screenWidth, 1/r.screenHeight)) //0-1に正規化
		// spr.PosOnScreen = drawStart.Scale(1.0/1000.0) //0-1に正規化

		if math.Abs(spr.PosOnScreen.X) > 1.0 {
			spr.Size = vec2.New(0, 0)
		}

		if math.Abs(spr.PosOnScreen.Y) > 1.0 {
			spr.Size = vec2.New(0, 0)
		}

		if !spr.IsVisible {
			spr.Size = vec2.New(0, 0)
		}

		// if spr.Size.X < 0.00001 {
		// 	spr.Size = vec2.New(0, 0)
		// }

		if spr.PosOnScreen.X > -0.004 && spr.PosOnScreen.X < 0 { //-0.002など1/255=0.003921...より絶対値が小さいときencode/decodeが上手くいかないので-0.004に変えてチラつくのを回避
			spr.PosOnScreen.X = -0.004
		}

		if spr.PosOnScreen.Y > -0.004 && spr.PosOnScreen.Y < 0 { //-0.002など1/255=0.003921...より絶対値が小さいときencode/decodeが上手くいかないので-0.004に変えてチラつくのを回避
			spr.PosOnScreen.Y = -0.004
		}

		// println("distance to camera", spr.DistanceToCamera)

		signOfPos := vec2.New(0, 0)
		if spr.PosOnScreen.X < 0 { //これが-1.4 < -1.0の時うまくいっていない --> 一旦解決
			signOfPos.X = 1.0
		}
		if spr.PosOnScreen.Y < 0 {
			signOfPos.Y = 1.0
		}

		// fmt.Printf("posonscreen: %f\n", spr.PosOnScreen)
		// fmt.Printf("size: %f\n", spr.Size)
		// fmt.Printf("signofpos: %f\n", signOfPos)

		data[r.SpriteParameterNum*i] = float32(spr.PosOnScreen.X)
		data[r.SpriteParameterNum*i+1] = float32(spr.PosOnScreen.Y)
		data[r.SpriteParameterNum*i+2] = float32(spr.Size.X)
		data[r.SpriteParameterNum*i+3] = float32(spr.Size.Y)
		data[r.SpriteParameterNum*i+4] = float32(spr.DistanceToCamera)
		data[r.SpriteParameterNum*i+5] = float32(signOfPos.X)
		data[r.SpriteParameterNum*i+6] = float32(signOfPos.Y)
		data[r.SpriteParameterNum*i+7] = float32(spr.TexIDBegin) / 255
		data[r.SpriteParameterNum*i+8] = float32(spr.CurrentAnimation) / 255

		// println(data[7*i],data[7*i+1],data[7*i+2],data[7*i+3],data[7*i+4],data[7*i+5],data[7*i+6])
	}

	r.SpriteParameters = data

}

func (r *Renderer) sortSpriteParameters() {
	if len(r.Wld.Sprites) < 2 {
		return
	}

	for i := 0; i < len(r.Wld.Sprites)-1; i++ {
		for j := 0; j < len(r.Wld.Sprites)-i-1; j++ {
			if r.SpriteParameters[r.SpriteParameterNum*j+4] > r.SpriteParameters[r.SpriteParameterNum*(j+1)+4] {
				for k := 0; k < r.SpriteParameterNum; k++ {
					// fmt.Printf("%d, %d\n", 6*j+k, 6*(j+1)+k)

					// tmp := w.SpriteRenderParam[6*j+k]
					r.SpriteParameters[r.SpriteParameterNum*j+k], r.SpriteParameters[r.SpriteParameterNum*(j+1)+k] = r.SpriteParameters[r.SpriteParameterNum*(j+1)+k], r.SpriteParameters[r.SpriteParameterNum*j+k]
					// r.SpriteParameters[6*j+k] = r.SpriteParameters[6*(j+1)+k]
					// r.SpriteParameters[6*(j+1)+k] = tmp
				}
				// r.Wld.Sprites[j], r.Wld.Sprites[j+1] = r.Wld.Sprites[j+1], r.Wld.Sprites[j]
			}
		}
	}
}
