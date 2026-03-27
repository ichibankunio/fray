package fray

import (
	// "image/color"
	// "math"

	"math"
	"runtime"

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

	collisionDistance  float64
	collisionAnchorPos vec3.Vec3
	useCollisionAnchor bool
	collisionHalfWidth float64
	collisionHalfDepth float64
	collisionHeight    float64

	shooterHeight float64
	shooterRadius float64

	Speed float64
	v     vec3.Vec3
	vZ    float64
	// vecV vec2.Vec2
}

func (c *Camera) SetPos(pos vec3.Vec3) {
	c.SetSubjectPos(pos)
}

func (c *Camera) SetCameraPos(pos vec3.Vec3) {
	c.pos = pos
	c.syncSubjectPosFromCamera()
}

func (c *Camera) SetSubjectPos(pos vec3.Vec3) {
	c.subjectPos = pos
	c.syncCameraPosFromSubject()
}

// func (c *Camera) SetPos(pos vec2.Vec2) {
// 	c.pos = pos
// }

func (c *Camera) GetPos() vec3.Vec3 {
	return c.pos
}

func (c *Camera) GetSubjectPos() vec3.Vec3 {
	return c.subjectPos
}

func (c *Camera) SetCollisionAnchorPos(pos vec3.Vec3) {
	c.collisionAnchorPos = pos
	c.useCollisionAnchor = true
}

func (c *Camera) ClearCollisionAnchorPos() {
	c.useCollisionAnchor = false
}

func (c *Camera) GetCollisionAnchorPos() vec3.Vec3 {
	if c.useCollisionAnchor {
		return c.collisionAnchorPos
	}
	return c.subjectPos
}

// SetCollisionBoxSize configures the collision body dimensions in world grid
// blocks. For example, (1, 1, 1) means a 1x1x1 block solid body.
func (c *Camera) SetCollisionBoxSize(width, depth, height float64) {
	if width < 0 {
		width = 0
	}
	if depth < 0 {
		depth = 0
	}
	if height < 0 {
		height = 0
	}
	c.collisionHalfWidth = width / 2
	c.collisionHalfDepth = depth / 2
	c.collisionHeight = height
}

func (c *Camera) GetCollisionBoxSize() (width, depth, height float64) {
	return c.collisionHalfWidth * 2, c.collisionHalfDepth * 2, c.collisionHeight
}

func (c *Camera) GetPlane() vec2.Vec2 {
	return c.plane
}

func (c *Camera) GetDir() vec2.Vec2 {
	return c.dir
}

func (c *Camera) GetPitch() float32 {
	return c.pitch
}

func (c *Camera) SetPitch(pitch float32) {
	c.pitch = pitch
}

func (c *Camera) SetPose(pos vec3.Vec3, dir vec2.Vec2, plane vec2.Vec2, pitch float32) {
	c.SetSubjectPose(pos, dir, plane, pitch)
}

func (c *Camera) SetSubjectPose(pos vec3.Vec3, dir vec2.Vec2, plane vec2.Vec2, pitch float32) {
	c.dir = dir
	c.plane = plane
	c.pitch = pitch
	c.SetSubjectPos(pos)
}

func (c *Camera) SetCameraPose(pos vec3.Vec3, dir vec2.Vec2, plane vec2.Vec2, pitch float32) {
	c.dir = dir
	c.plane = plane
	c.pitch = pitch
	c.SetCameraPos(pos)
}

func (c *Camera) SetDafaultDistanceBetweenSubjectCamera(distance float64) {
	c.distanceBetweenSubjectCamera = distance
	c.syncCameraPosFromSubject()
}

// SetMinDistanceToWallWhileApproaching configures the minimum distance between
// camera(subject) and walls when movement is approaching the wall.
// Unit: world grid blocks (e.g. 0.25 == 16px when texSize is 64).
func (c *Camera) SetMinDistanceToWallWhileApproaching(distance float64) {
	if distance < 0 {
		distance = 0
	}
	c.collisionDistance = distance
}

func (c *Camera) GetMinDistanceToWallWhileApproaching() float64 {
	return c.collisionDistance
}

func (c *Camera) SetShooterHeight(height float64) {
	c.shooterHeight = height
}

func (c *Camera) GetShooterHeight() float64 {
	return c.shooterHeight
}

func (c *Camera) Init(screenWidth, screenHeight float64) {
	initialCameraPos := vec3.New(64*10*3/4, 64*10/2, 32)
	// c.pos = vec3.New(500, 650, 32)
	c.dir = vec2.New(-1, 0)
	// c.plane = vec2.New(0, screenWidth/screenHeight)
	// c.plane = vec2.New(0, 0.66*screenWidth/screenHeight*3/4)
	c.plane = vec2.New(0, 0.66*screenWidth/screenHeight*3/4)
	// c.plane = vec2.New(0, 1)
	// c.plane = vec2.New(0, 0.66 * SCREEN_WIDTH / 960 * 720 / SCREEN_HEIGHT)

	c.distanceBetweenSubjectCamera = 64

	c.SetCameraPos(initialCameraPos)

	c.collisionDistance = 0.25
	c.collisionAnchorPos = c.subjectPos
	c.useCollisionAnchor = false
	c.collisionHalfWidth = c.shooterRadius
	c.collisionHalfDepth = c.shooterRadius
	c.collisionHeight = 0

	c.zoomed = false

	c.Speed = 10.0
	c.v = vec3.New(0, 0, 0)
	c.vZ = -1.0
	// c.posZ = 0

	c.shooterHeight = 128
	c.shooterRadius = 0.25 //equivalent to 16px
	c.collisionHalfWidth = c.shooterRadius
	c.collisionHalfDepth = c.shooterRadius

	if runtime.GOOS != "js" {
		ebiten.SetCursorMode(ebiten.CursorModeCaptured)
	}
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
	c.syncSubjectPosFromCamera()

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

	c.syncCameraPosFromSubject()
}

func (c *Camera) syncSubjectPosFromCamera() {
	c.subjectPos = c.pos.Add(c.subjectOffset())
}

func (c *Camera) syncCameraPosFromSubject() {
	c.pos = c.subjectPos.Sub(c.subjectOffset())
}

func (c *Camera) subjectOffset() vec3.Vec3 {
	return vec3.New(c.dir.X, c.dir.Y, 0).Scale(c.distanceBetweenSubjectCamera)
}
