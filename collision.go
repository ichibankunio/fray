package fray

import (
	"math"

	"github.com/ichibankunio/fvec/vec2"
	"github.com/ichibankunio/fvec/vec3"
)

type Ray struct {
	perpWallDist       float64
	squaredEuclidean   float64
	detectedWallHeight uint8
	hitPosOnMap        vec2.Vec2
	rayHitType         RayHitType
	hitPos             vec3.Vec3
}

type RayHitType int

const (
	RAY_HIT_UP RayHitType = iota
	RAY_HIT_DOWN
)

func (r *Renderer) castRayMultiHeight(dir, plane vec2.Vec2, pos vec3.Vec3, hitType RayHitType) *Ray {
	// heightAtRayPosition := r.GetGroundHeight(pos)
	cameraX := 0.0 //x-coordinate in camera space
	rayDir := dir.Add(plane.Scale(cameraX))
	initRayPos := vec2.New(pos.X/float64(r.texSize), pos.Y/float64(r.texSize))
	rayPos := vec3.New(pos.X, pos.Y, pos.Z)
	mapPos := vec2.New(math.Floor(initRayPos.X), math.Floor(initRayPos.Y))
	deltaDist := vec2.New(math.Abs(1.0/rayDir.X), math.Abs(1.0/rayDir.Y))
	perpWallDist := 0.0
	unit := vec2.New(1, 1)
	sideDist := vec2.New(0, 0)
	if rayDir.X < 0 {
		unit.X = -1
		sideDist.X = (initRayPos.X - mapPos.X) * deltaDist.X
	} else {
		unit.X = 1
		sideDist.X = (mapPos.X + 1.0 - initRayPos.X) * deltaDist.X
	}

	if rayDir.Y < 0 {
		unit.Y = -1
		sideDist.Y = (initRayPos.Y - mapPos.Y) * deltaDist.Y
	} else {
		unit.Y = 1
		sideDist.Y = (mapPos.Y + 1.0 - initRayPos.Y) * deltaDist.Y
	}
	side := -1.0
	for i := 0; i < 10; i++ {
		//jump to next map square, OR in x-direction, OR in y-direction
		if sideDist.X < sideDist.Y {
			sideDist.X += deltaDist.X
			mapPos.X += unit.X
			side = 0.0
			rayPos.X += float64(r.texSize)
		} else {
			sideDist.Y += deltaDist.Y
			mapPos.Y += unit.Y
			side = 1.0
			rayPos.Y += float64(r.texSize)
		}

		//世界の端に衝突
		if mapPos.X < 0 || mapPos.Y < 0 || mapPos.X > float64(r.canvasWidth-1) || mapPos.Y > float64(r.canvasHeight-1) {
			if side == 0 {
				perpWallDist = sideDist.X - deltaDist.X
			} else {
				perpWallDist = sideDist.Y - deltaDist.Y
			}
			return &Ray{
				perpWallDist:       perpWallDist,
				squaredEuclidean:   perpWallDist * perpWallDist * (rayDir.X*rayDir.X + rayDir.Y*rayDir.Y),
				detectedWallHeight: 255,
				hitPosOnMap:        mapPos,
				rayHitType:         RAY_HIT_UP,
				hitPos:             rayPos,
			}
		}

		// if r.Wld.WorldMap[int(pos.Z/float64(r.texSize))][int(mapPos.Y)*r.canvasWidth+int(mapPos.X)] >= 1 {

		rayHeight := r.Wld.HeightMap[int(mapPos.Y)*r.canvasWidth+int(mapPos.X)]
		if hitType == RAY_HIT_UP {
			if rayHeight > uint8(pos.Z/float64(r.texSize)) {
				break
			}
		} else if hitType == RAY_HIT_DOWN {
			if rayHeight < uint8(pos.Z/float64(r.texSize)) {
				break
			}
		}
		//Calculate distance of perpendicular ray (oblique distance will give fisheye effect!)
	}

	if side == 0 {
		perpWallDist = sideDist.X - deltaDist.X
	} else {
		perpWallDist = sideDist.Y - deltaDist.Y
	}

	detectedWallHeight := r.Wld.HeightMap[int(mapPos.Y)*r.canvasWidth+int(mapPos.X)]
	rayPos.Z = float64(int(detectedWallHeight) * r.texSize)
	return &Ray{
		perpWallDist:       perpWallDist,
		squaredEuclidean:   perpWallDist * perpWallDist * (rayDir.X*rayDir.X + rayDir.Y*rayDir.Y),
		detectedWallHeight: detectedWallHeight,
		hitPosOnMap:        mapPos,
		rayHitType:         hitType,
		hitPos:             rayPos,
	}

	// return perpWallDist, r.Wld.levelUint8[1][int(mapPos.Y)*r.Wld.width+int(mapPos.X)]//(当たった壁までの距離, その壁の高さ)
}

func (r *Renderer) collisionCheckedDelta(pos vec3.Vec3, delta vec2.Vec2, collisionBuffer float64) vec3.Vec3 { //deltaは絶対値が大きすぎるとうまくいかない（？）
	climbable := 0.0
	// distanceToSubject := 0.0

	if delta.X > 0 {
		// ray := r.castRayMultiHeight(vec2.New(1, 0), r.Cam.plane, pos.Add(vec3.New(r.Cam.dir.X, r.Cam.dir.Y, 0).Scale(distanceToSubject)))
		ray := r.castRayMultiHeight(vec2.New(1, 0), r.Cam.plane, pos.Add(vec3.New(0, 0, -r.Cam.shooterHeight)), RAY_HIT_UP)
		if ray.squaredEuclidean <= collisionBuffer*collisionBuffer && float64(ray.detectedWallHeight)-((pos.Z-r.Cam.shooterHeight)/float64(r.texSize)) > climbable {
			// delta.X = dist - collisionBuffer
			delta.X = 0
		}
	}

	if delta.X < 0 {
		// ray := r.castRayMultiHeight(vec2.New(-1, 0), r.Cam.plane, pos.Add(vec3.New(r.Cam.dir.X, r.Cam.dir.Y, 0).Scale(distanceToSubject)))
		ray := r.castRayMultiHeight(vec2.New(-1, 0), r.Cam.plane, pos.Add(vec3.New(0, 0, -r.Cam.shooterHeight)), RAY_HIT_UP)

		if ray.squaredEuclidean <= collisionBuffer*collisionBuffer && float64(ray.detectedWallHeight)-((pos.Z-r.Cam.shooterHeight)/float64(r.texSize)) > climbable {

			// delta.X = collisionBuffer - dist
			delta.X = 0

		}
	}

	if delta.Y > 0 {
		// ray := r.castRayMultiHeight(vec2.New(0, 1), r.Cam.plane, pos.Add(vec3.New(r.Cam.dir.X, r.Cam.dir.Y, 0).Scale(distanceToSubject)))
		ray := r.castRayMultiHeight(vec2.New(0, 1), r.Cam.plane, pos.Add(vec3.New(0, 0, -r.Cam.shooterHeight)), RAY_HIT_UP)

		if ray.squaredEuclidean <= collisionBuffer*collisionBuffer && float64(ray.detectedWallHeight)-((pos.Z-r.Cam.shooterHeight)/float64(r.texSize)) > climbable {

			// delta.Y = dist - collisionBuffer
			delta.Y = 0
		}
	}

	if delta.Y < 0 {

		ray := r.castRayMultiHeight(vec2.New(0, -1), r.Cam.plane, pos.Add(vec3.New(0, 0, -r.Cam.shooterHeight)), RAY_HIT_UP)

		if ray.squaredEuclidean <= collisionBuffer*collisionBuffer && float64(ray.detectedWallHeight)-((pos.Z-r.Cam.shooterHeight)/float64(r.texSize)) > climbable {

			// delta.Y = collisionBuffer - dist
			delta.Y = 0
		}
	}

	return vec3.New(delta.X, delta.Y, 0)
}

func (r *Renderer) collisionCheckedDeltaZ(pos vec3.Vec3, delta float64) float64 {
	if delta < 0 {
		dist := pos.Z - r.GetGroundHeight(pos) //今のz座標と地面の高さの差
		if dist <= r.Cam.shooterHeight {
			delta = 0
		}
	}

	return delta
}

func (r *Renderer) GetGroundHeight(pos vec3.Vec3) float64 {
	if pos.Y/float64(r.texSize) < 0 {
		pos.Y = 0
	}
	if pos.X/float64(r.texSize) < 0 {
		pos.X = 0
	}

	return float64(r.Wld.HeightMap[int(pos.Y/float64(r.texSize))*r.canvasWidth+int(pos.X/float64(r.texSize))]) * float64(r.texSize)
}
