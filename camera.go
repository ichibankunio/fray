package fray

import (
	// "image/color"
	// "math"

	"math"

	"github.com/hajimehoshi/ebiten/v2"
	// "github.com/ichibankunio/flib"
	"github.com/ichibankunio/fvec/vec2"
	"github.com/ichibankunio/fvec/vec3"
)

type Camera struct {

	//--camera position, init to start position--//
	// pos vec2.Vec2

	// vertical camera strafing up/down, for jumping/crouching
	// posZ float32

	pos        vec3.Vec3
	subjectPos vec3.Vec3

	zoomed bool

	distanceBetweenSubjectCamera float64

	pitch float32

	//--current facing direction, init to values coresponding to FOV--//
	dir vec2.Vec2

	//--the 2d raycaster version of camera plane, adjust y component to change FOV (ratio between this and dir x resizes FOV)--//
	plane vec2.Vec2

	collisionDistance float64

	shooterHeight float64
	shooterRadius float64

	speed float64
	v     vec3.Vec3
	vZ    float64
	// vecV vec2.Vec2
}

func (c *Camera) SetPos(pos vec3.Vec3) {
	c.pos = pos
	c.subjectPos = pos.Add(vec3.New(c.dir.X, c.dir.Y, 0).Scale(c.distanceBetweenSubjectCamera))
}

// func (c *Camera) SetPos(pos vec2.Vec2) {
// 	c.pos = pos
// }

func (c *Camera) GetPos() vec3.Vec3 {
	return c.pos
}

func (c *Camera) GetPlane() vec2.Vec2 {
	return c.plane
}

func (c *Camera) GetPitch() float32 {
	return c.pitch
}

func (c *Camera) Init(screenWidth, screenHeight float64) {
	c.pos = vec3.New(64*10*3/4, 64*10/2, 32)
	// c.pos = vec3.New(500, 650, 32)
	c.dir = vec2.New(-1, 0)
	// c.plane = vec2.New(0, screenWidth/screenHeight)
	// c.plane = vec2.New(0, 0.66*screenWidth/screenHeight*3/4)
	c.plane = vec2.New(0, 0.66*screenWidth/screenHeight*3/4)
	// c.plane = vec2.New(0, 1)
	// c.plane = vec2.New(0, 0.66 * SCREEN_WIDTH / 960 * 720 / SCREEN_HEIGHT)

	c.distanceBetweenSubjectCamera = 64

	// c.subjectPos = c.pos.Add(vec3.New(c.dir.X, c.dir.Y, 0).Scale(c.distanceBetweenSubjectCamera))

	c.collisionDistance = 0.25

	c.zoomed = false

	c.speed = 2.0
	c.v = vec3.New(0, 0, 0)
	c.vZ = -1.0
	// c.posZ = 0

	c.shooterHeight = 64
	c.shooterRadius = 0.25 //equivalent to 16px

	ebiten.SetCursorMode(ebiten.CursorModeCaptured)

}

// func (c *Camera) Update(g *flib.Game) {

// }

func (c *Camera) Draw(screen *ebiten.Image) {

}

func (c *Camera) RotateHorizontal(v float64) {
	// rotateV := 0.02
	//right
	c.dir = vec2.New(math.Cos(v)*c.dir.X-math.Sin(v)*c.dir.Y, math.Sin(v)*c.dir.X+math.Cos(v)*c.dir.Y)

	c.plane = vec2.New(math.Cos(v)*c.plane.X-math.Sin(v)*c.plane.Y, math.Sin(v)*c.plane.X+math.Cos(v)*c.plane.Y)

	// //left
	// c.dir = vec2.New(math.Cos(-rotateV)*c.dir.X-math.Sin(-rotateV)*c.dir.Y, math.Sin(-rotateV)*c.dir.X+math.Cos(-rotateV)*c.dir.Y)

	// c.plane = vec2.New(math.Cos(-rotateV)*c.plane.X-math.Sin(-rotateV)*c.plane.Y, math.Sin(-rotateV)*c.plane.X+math.Cos(-rotateV)*c.plane.Y)
}

func (c *Camera) RotateVertical(speed float64) {
	c.pitch += -float32(speed)
	// if c.pitch < -300 {
	// 	c.pitch += 	1.0
	// }else if c.pitch > 200 {
	// 	c.pitch -= 1.0
	// }
}

func (c *Camera) RotateHorizontalAroundSubject(v float64) {
	c.dir = vec2.New(math.Cos(v)*c.dir.X-math.Sin(v)*c.dir.Y, math.Sin(v)*c.dir.X+math.Cos(v)*c.dir.Y)

	c.plane = vec2.New(math.Cos(v)*c.plane.X-math.Sin(v)*c.plane.Y, math.Sin(v)*c.plane.X+math.Cos(v)*c.plane.Y)

	c.pos = vec3.New(math.Cos(v)*(c.pos.X-c.subjectPos.X)-math.Sin(v)*(c.pos.Y-c.subjectPos.Y)+c.subjectPos.X, math.Sin(v)*(c.pos.X-c.subjectPos.X)+math.Cos(v)*(c.pos.Y-c.subjectPos.Y)+c.subjectPos.Y, c.pos.Z)

}
