package fray

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/ichibankunio/fvec/vec2"
)

func (r *Renderer) UpdateCamRotationByMouse() {
	x, y := ebiten.CursorPosition()

	current := vec2.New(float64(x), float64(y))
	rel := current.Sub(r.Stk.mousePosLastFrame)

	r.Cam.RotateHorizontal(rel.X * 0.004)
	r.Cam.RotateVertical(rel.Y)

	r.Stk.mousePosLastFrame = current
}

func (r *Renderer) UpdateCamRotationAroundSubjectByTouch() {
	if len(inpututil.AppendJustPressedTouchIDs(nil)) > 0 {
		for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
			x, y := ebiten.TouchPosition(id)
			if r.Stk.touchIDs[0] < 0 && x < int(r.Stk.screenWidth/2) {
				// r.Stk.pos[0] = vec2.New(float64(x-r.Stk.img.Bounds().Dx()/2), float64(y-r.Stk.img.Bounds().Dy()/2))
				r.Stk.pos[0] = vec2.New(float64(x), float64(y))
				r.Stk.visible[0] = true
				r.Stk.touchIDs[0] = id
				r.Stk.isMobile = true
				continue
			}
			if r.Stk.touchIDs[1] < 0 && x >= int(r.Stk.screenWidth/2) {
				// r.Stk.pos[1] = vec2.New(float64(x-r.Stk.img.Bounds().Dx()/2), float64(y-r.Stk.img.Bounds().Dy()/2))
				r.Stk.pos[1] = vec2.New(float64(x), float64(y))
				r.Stk.visible[1] = true
				r.Stk.touchIDs[1] = id
				r.Stk.isMobile = true
				continue
			}
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if x < int(r.Stk.screenWidth/2) {
			// r.Stk.pos[0] = vec2.New(float64(x-r.Stk.img.Bounds().Dx()/2), float64(y-r.Stk.img.Bounds().Dy()/2))
			r.Stk.pos[0] = vec2.New(float64(x), float64(y))
			r.Stk.visible[0] = true
		}
		if x >= int(r.Stk.screenWidth/2) {
			// r.Stk.pos[1] = vec2.New(float64(x-r.Stk.img.Bounds().Dx()/2), float64(y-r.Stk.img.Bounds().Dy()/2))
			r.Stk.pos[1] = vec2.New(float64(x), float64(y))
			r.Stk.visible[1] = true
		}
	}

	if r.Stk.visible[0] {
		x, y := ebiten.CursorPosition()
		if r.Stk.isMobile {
			x, y = ebiten.TouchPosition(r.Stk.touchIDs[0])
		}
		current := vec2.New(float64(x), float64(y))
		rel := current.Sub(r.Stk.pos[0])
		if rel.X > 0 && math.Abs(rel.Y/rel.X) < 0.8 {
			r.Stk.Input[0] = STICK_RIGHT
		} else if rel.X < 0 && math.Abs(rel.Y/rel.X) < 0.8 {
			r.Stk.Input[0] = STICK_LEFT
		} else if rel.Y > 0 && math.Abs(rel.X/rel.Y) < 0.8 {
			r.Stk.Input[0] = STICK_DOWN
		} else if rel.Y < 0 && math.Abs(rel.X/rel.Y) < 0.8 {
			r.Stk.Input[0] = STICK_UP
		} else {
			r.Stk.Input[0] = STICK_NONE
		}

		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			r.Stk.visible[0] = false
			r.Stk.Input[0] = STICK_NONE
		}
		if inpututil.IsTouchJustReleased(r.Stk.touchIDs[0]) {
			r.Stk.visible[0] = false
			r.Stk.Input[0] = STICK_NONE
			r.Stk.touchIDs[0] = -1
		}
	}

	if r.Stk.visible[1] {
		x, y := ebiten.CursorPosition()
		if r.Stk.isMobile {
			x, y = ebiten.TouchPosition(r.Stk.touchIDs[1])
		}


		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			r.Stk.visible[1] = false
			r.Stk.Input[1] = STICK_NONE
			return
		}

		if inpututil.IsTouchJustReleased(r.Stk.touchIDs[1]) {
			r.Stk.visible[1] = false
			r.Stk.Input[1] = STICK_NONE
			r.Stk.touchIDs[1] = -1
			return
		}


		current := vec2.New(float64(x), float64(y))
		rel := current.Sub(r.Stk.pos[1])

		// fmt.Printf("%f, %f\n", r.Stk.pos[1], current)

		r.Cam.RotateHorizontalAroundSubject(rel.X * 0.0001)
		r.Cam.RotateVertical(rel.Y * 0.05)

		r.Stk.mousePosLastFrame = current

		//subjectとcameraの間に壁があるときにカメラを壁の前に出す(subjectに近づく)-----------
		// d := r.collisionCheckedDelta(r.Cam.subjectPos, r.Cam.dir.Scale(-64))
		// // fmt.Printf("%f, %f\n", d, d.Length())

		// if d.Length() < 63.9 {
		// 	r.Cam.zoomed = true
		// 	// println("zoomed")
		// 	r.Cam.pos = r.Cam.subjectPos.Add(d.Scale(1.0))
		// } else {
		// 	r.Cam.zoomed = false
		// }

		// println("distance: ", r.Cam.subjectPos.Mul(vec3.New(1, 1, 0)).Sub(r.Cam.pos.Mul(vec3.New(1, 1, 0))).Length())

		//------------------------------

	}
}

func (r *Renderer) UpdateCamRotationAroundSubjectByMouse() {
	x, y := ebiten.CursorPosition()

	current := vec2.New(float64(x), float64(y))
	rel := current.Sub(r.Stk.mousePosLastFrame)

	r.Cam.RotateHorizontalAroundSubject(rel.X * 0.004)
	r.Cam.RotateVertical(rel.Y)

	r.Stk.mousePosLastFrame = current

	//subjectとcameraの間に壁があるときにカメラを壁の前に出す(subjectに近づく)-----------
	// d := r.collisionCheckedDelta(r.Cam.subjectPos, r.Cam.dir.Scale(-64))
	// fmt.Printf("%f, d.length = %f\n", d, d.Length())

	// if d.Length() < 63.9999 {
	// 	r.Cam.zoomed = true
	// 	// println("zoomed")
	// 	r.Cam.pos = r.Cam.subjectPos.Add(d.Scale(1.0))
	// } else {
	// 	// println("zoom end")

	// 	r.Cam.zoomed = false
	// }

	// println(r.Cam.subjectPos.Mul(vec3.New(1, 1, 0)).Sub(r.Cam.pos.Mul(vec3.New(1, 1, 0))).Length())

	// println("distance: ", r.Cam.subjectPos.Mul(vec3.New(1, 1, 0)).Sub(r.Cam.pos.Mul(vec3.New(1, 1, 0))).Length())

	//------------------------------
}

func (r *Renderer) UpdateCamRotationByTouch() {
	if len(inpututil.AppendJustPressedTouchIDs(nil)) > 0 {
		for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
			x, y := ebiten.TouchPosition(id)
			if r.Stk.touchIDs[0] < 0 && x < int(r.Stk.screenWidth/2) {
				// r.Stk.pos[0] = vec2.New(float64(x-r.Stk.img.Bounds().Dx()/2), float64(y-r.Stk.img.Bounds().Dy()/2))
				r.Stk.pos[0] = vec2.New(float64(x), float64(y))
				r.Stk.visible[0] = true
				r.Stk.touchIDs[0] = id
				r.Stk.isMobile = true
				continue
			}
			if r.Stk.touchIDs[1] < 0 && x >= int(r.Stk.screenWidth/2) {
				// r.Stk.pos[1] = vec2.New(float64(x-r.Stk.img.Bounds().Dx()/2), float64(y-r.Stk.img.Bounds().Dy()/2))
				r.Stk.pos[1] = vec2.New(float64(x), float64(y))
				r.Stk.visible[1] = true
				r.Stk.touchIDs[1] = id
				r.Stk.isMobile = true
				continue
			}
		}
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if x < int(r.Stk.screenWidth/2) {
			// r.Stk.pos[0] = vec2.New(float64(x-r.Stk.img.Bounds().Dx()/2), float64(y-r.Stk.img.Bounds().Dy()/2))
			r.Stk.pos[0] = vec2.New(float64(x), float64(y))
			r.Stk.visible[0] = true
		}
		if x >= int(r.Stk.screenWidth/2) {
			// r.Stk.pos[1] = vec2.New(float64(x-r.Stk.img.Bounds().Dx()/2), float64(y-r.Stk.img.Bounds().Dy()/2))
			r.Stk.pos[1] = vec2.New(float64(x), float64(y))
			r.Stk.visible[1] = true
		}
	}

	if r.Stk.visible[0] {
		x, y := ebiten.CursorPosition()
		if r.Stk.isMobile {
			x, y = ebiten.TouchPosition(r.Stk.touchIDs[0])
		}
		current := vec2.New(float64(x), float64(y))
		rel := current.Sub(r.Stk.pos[0])
		if rel.X > 0 && math.Abs(rel.Y/rel.X) < 0.8 {
			r.Stk.Input[0] = STICK_RIGHT
		} else if rel.X < 0 && math.Abs(rel.Y/rel.X) < 0.8 {
			r.Stk.Input[0] = STICK_LEFT
		} else if rel.Y > 0 && math.Abs(rel.X/rel.Y) < 0.8 {
			r.Stk.Input[0] = STICK_DOWN
		} else if rel.Y < 0 && math.Abs(rel.X/rel.Y) < 0.8 {
			r.Stk.Input[0] = STICK_UP
		} else {
			r.Stk.Input[0] = STICK_NONE
		}

		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			r.Stk.visible[0] = false
			r.Stk.Input[0] = STICK_NONE
		}
		if inpututil.IsTouchJustReleased(r.Stk.touchIDs[0]) {
			r.Stk.visible[0] = false
			r.Stk.Input[0] = STICK_NONE
			r.Stk.touchIDs[0] = -1
		}
	}

	if r.Stk.visible[1] {
		x, y := ebiten.CursorPosition()
		if r.Stk.isMobile {
			x, y = ebiten.TouchPosition(r.Stk.touchIDs[1])
		}
		current := vec2.New(float64(x), float64(y))
		rel := current.Sub(r.Stk.pos[1])

		// fmt.Printf("%f, %f\n", r.Stk.pos[1], current)

		r.Cam.RotateHorizontal(rel.X * 0.0001)
		r.Cam.RotateVertical(rel.Y * 0.1)

		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			r.Stk.visible[1] = false
			r.Stk.Input[1] = STICK_NONE

		}
		if inpututil.IsTouchJustReleased(r.Stk.touchIDs[1]) {
			r.Stk.visible[1] = false
			r.Stk.Input[1] = STICK_NONE
			r.Stk.touchIDs[1] = -1
		}
	}
}
