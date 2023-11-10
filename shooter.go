package fray

import (
	"math"

	"github.com/ichibankunio/fvec/vec2"
	"github.com/ichibankunio/fvec/vec3"
)

func (r *Renderer) GetDistanceInterferenceFromSubject() {
	// ray := r.castRayMultiHeight(r.Cam.dir.Scale(-r.Cam.distanceBetweenSubjectCamera), r.Cam.plane, r.Cam.subjectPos)
	// d := ray.squaredEuclidean * 64
	// d := r.collisionCheckedDelta(r.Cam.subjectPos, r.Cam.dir.Scale(-1), r.Cam.collisionDistance)
	ray := r.GetNearestWallDistance(r.Cam.dir.Scale(-1), r.Cam.plane, r.Cam.subjectPos)
	// fmt.Printf("%f, %f\n", -r.Cam.dir.X, -r.Cam.dir.Y)
	// fmt.Printf("%f, d.length = %f\n", d, d.Length())
	// fmt.Printf("squaredEuclidean: %f\n", ray.squaredEuclidean)
	// if ray.squaredEuclidean < 1 && r.GetGroundHeight(r.Cam.pos) == r.GetGroundHeight(r.Cam.subjectPos) && r.Cam.vZ == 0 {
	delta := r.Cam.dir.Scale(-r.Cam.distanceBetweenSubjectCamera)

	// if ray.squaredEuclidean < 1 && r.GetGroundHeight(r.Cam.subjectPos.Add(vec3.New(delta.X, delta.Y, 0))) - r.GetGroundHeight(r.Cam.subjectPos) == 0{
	// println(int(r.Cam.pos.Z))
	// if ray.squaredEuclidean < 1 && ((r.GetGroundHeight(r.Cam.subjectPos.Add(vec3.New(delta.X, delta.Y, 0))) - r.GetGroundHeight(r.Cam.subjectPos) == 0) || r.Cam.vZ == 0) {
	if ray.squaredEuclidean < 1 && !(r.Cam.v.SquaredLength() > 0 && r.Cam.vZ < 0) {
		// println("needs zoom", int(r.Cam.pos.Z))

		delta = r.Cam.dir.Scale(-r.Cam.distanceBetweenSubjectCamera * (ray.squaredEuclidean))

		// hitPosOnMap := ray.hitPosOnMap.Scale(float64(r.texSize))
		// internalDivisionPoint := vec2.New(/(r.Cam.subjectPos.X + r.Cam.pos.X))
		// r.Cam.pos = vec3.New(hitPosOnMap.X, hitPosOnMap.Y, r.Cam.subjectPos.Z)

		// r.Cam.pos = r.Cam.subjectPos.Add(vec3.New(r.Cam.dir.Scale(-ray.squaredEuclidean))
	}

	r.Cam.pos = r.Cam.subjectPos.Add(vec3.New(delta.X, delta.Y, 0))

}

func (r *Renderer) GetNearestWallDistance(dir, plane vec2.Vec2, pos vec3.Vec3) *Ray {
	cameraX := 0.0 //x-coordinate in camera space
	rayDir := dir.Add(plane.Scale(cameraX))
	rayPos := vec2.New(pos.X/float64(r.texSize), pos.Y/float64(r.texSize))
	mapPos := vec2.New(math.Floor(rayPos.X), math.Floor(rayPos.Y))
	deltaDist := vec2.New(math.Abs(1.0/rayDir.X), math.Abs(1.0/rayDir.Y))
	perpWallDist := 0.0
	unit := vec2.New(1, 1)
	sideDist := vec2.New(0, 0)
	if rayDir.X < 0 {
		unit.X = -1
		sideDist.X = (rayPos.X - mapPos.X) * deltaDist.X
	} else {
		unit.X = 1
		sideDist.X = (mapPos.X + 1.0 - rayPos.X) * deltaDist.X
	}

	if rayDir.Y < 0 {
		unit.Y = -1
		sideDist.Y = (rayPos.Y - mapPos.Y) * deltaDist.Y
	} else {
		unit.Y = 1
		sideDist.Y = (mapPos.Y + 1.0 - rayPos.Y) * deltaDist.Y
	}

	rayHeight := r.GetGroundHeight(pos)

	side := -1.0
	for i := 0; i < 4; i++ {
		//jump to next map square, OR in x-direction, OR in y-direction
		if sideDist.X < sideDist.Y {
			sideDist.X += deltaDist.X
			mapPos.X += unit.X
			side = 0.0
		} else {
			sideDist.Y += deltaDist.Y
			mapPos.Y += unit.Y
			side = 1.0
		}

		// if r.Wld.levelUint8[0][int(mapPos.Y)*r.Wld.width+int(mapPos.X)] >= 1 {
		if r.GetGroundHeight(vec3.New(mapPos.X*float64(r.texSize), mapPos.Y*float64(r.texSize), 0)) > rayHeight {
			// hit = 1
			break
		}

		if mapPos.Sub(rayPos).SquaredLength() > 2 {
			return &Ray{
				perpWallDist:       100,
				squaredEuclidean:   100,
				detectedWallHeight: 100,
				hitPosOnMap:        vec2.New(100, 100),
			}
		}

		//Calculate distance of perpendicular ray (oblique distance will give fisheye effect!)
	}

	if side == 0 {
		perpWallDist = sideDist.X - deltaDist.X
	} else {
		perpWallDist = sideDist.Y - deltaDist.Y

	}
	// println((r.Cam.pos.Z / float64(r.Wld.gridSize)), r.Wld.levelUint8[1][int(mapPos.Y)*r.Wld.width+int(mapPos.X)])

	return &Ray{
		perpWallDist:       perpWallDist,
		squaredEuclidean:   perpWallDist * perpWallDist * (rayDir.X*rayDir.X + rayDir.Y*rayDir.Y),
		detectedWallHeight: r.Wld.HeightMap[int(mapPos.Y)*r.canvasWidth+int(mapPos.X)],
		hitPosOnMap:        mapPos,
	}
}
