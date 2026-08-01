package fray

import (
	"image/color"
	"math"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/ichibankunio/fvec/vec2"
	"github.com/ichibankunio/fvec/vec3"
)

// BillboardVertex describes camera-facing geometry in height-relative units.
// X is horizontal from the center and Y is height above the instance origin.
type BillboardVertex struct {
	X, Y       float32
	SrcX, SrcY float32
	Color      color.RGBA
}

type BillboardMesh struct {
	Vertices []BillboardVertex
	Indices  []uint16
}

type BillboardInstance struct {
	Position vec3.Vec3
	Height   float64
	Tint     color.RGBA
}

type BillboardConfig struct {
	NearDistance          float64
	FarDistance           float64
	ProjectionScale       float64
	ScreenMargin          float32
	MinPixelHeight        float32
	MaxPixelHeight        float32
	TerrainOcclusion      bool
	Occlusion             TerrainOcclusionConfig
	OcclusionWidthSamples int
	DebugBounds           bool
	DebugColor            color.RGBA
}

func DefaultBillboardConfig() BillboardConfig {
	return BillboardConfig{
		NearDistance: 1, FarDistance: 14, ProjectionScale: .48,
		ScreenMargin: 80, MinPixelHeight: .5, MaxPixelHeight: 4096,
		TerrainOcclusion: true, Occlusion: DefaultTerrainOcclusionConfig(), OcclusionWidthSamples: 3,
	}
}

type projectedBillboard struct {
	instance BillboardInstance
	depth    float64
	x, baseY float32
	clipY    float32
}

// DrawBillboardInstances projects, culls, terrain-clips, depth-sorts, and
// batches camera-facing instances into as few DrawTriangles calls as possible.
func (r *Renderer) DrawBillboardInstances(dst, texture *ebiten.Image, mesh BillboardMesh, instances []BillboardInstance, config BillboardConfig) {
	if dst == nil || texture == nil || len(mesh.Vertices) == 0 || len(mesh.Indices) == 0 || len(instances) == 0 {
		return
	}
	config = normalizedBillboardConfig(config)
	position := r.Cam.GetSubjectPos().Scale(1 / float64(r.texSize))
	direction := r.Cam.GetDir()
	right := vec2.New(-direction.Y, direction.X)
	projected := make([]projectedBillboard, 0, len(instances))
	meshHalfWidth := billboardMeshHalfWidth(mesh)
	for _, instance := range instances {
		if instance.Height <= 0 {
			continue
		}
		relX := instance.Position.X - position.X
		relY := instance.Position.Y - position.Y
		depth := relX*direction.X + relY*direction.Y
		if depth <= config.NearDistance || depth > config.FarDistance {
			continue
		}
		lateral := relX*right.X + relY*right.Y
		x := float32(r.screenWidth/2 + lateral/(depth*config.ProjectionScale)*(r.screenWidth/2))
		if x < -config.ScreenMargin || x > float32(r.screenWidth)+config.ScreenMargin {
			continue
		}
		baseY := float32(r.screenHeight/2) + r.Cam.GetPitch() + float32((position.Z-instance.Position.Z)/depth*r.screenHeight)
		if baseY < -config.ScreenMargin || baseY > float32(r.screenHeight)+config.ScreenMargin {
			continue
		}
		clipY := baseY
		if config.TerrainOcclusion {
			visibleHeight := r.Wld.TerrainVisibleHeight(position.X, position.Y, position.Z, instance.Position.X, instance.Position.Y, config.Occlusion)
			// Most objects are fully visible and need only the center ray. Sample
			// both edges only near a terrain silhouette where partial clipping is
			// actually possible.
			if visibleHeight > instance.Position.Z && config.OcclusionWidthSamples > 1 {
				halfWidth := meshHalfWidth * instance.Height
				for _, offset := range [2]float64{-halfWidth, halfWidth} {
					targetX := instance.Position.X + right.X*offset
					targetY := instance.Position.Y + right.Y*offset
					visibleHeight = max(visibleHeight, r.Wld.TerrainVisibleHeight(position.X, position.Y, position.Z, targetX, targetY, config.Occlusion))
				}
			}
			if visibleHeight >= instance.Position.Z+instance.Height {
				continue
			}
			if visibleHeight > instance.Position.Z {
				clipY = float32(r.screenHeight/2) + r.Cam.GetPitch() + float32((position.Z-visibleHeight)/depth*r.screenHeight)
			}
		}
		projected = append(projected, projectedBillboard{instance: instance, depth: depth, x: x, baseY: baseY, clipY: clipY})
	}
	sort.Slice(projected, func(i, j int) bool { return projected[i].depth > projected[j].depth })
	r.drawProjectedBillboards(dst, texture, mesh, projected, config)
	if config.DebugBounds {
		r.drawBillboardDebugBounds(dst, mesh, projected, config)
	}
}

func normalizedBillboardConfig(config BillboardConfig) BillboardConfig {
	defaults := DefaultBillboardConfig()
	if config.NearDistance < 0 {
		config.NearDistance = 0
	}
	if config.FarDistance <= config.NearDistance {
		config.FarDistance = defaults.FarDistance
	}
	if config.ProjectionScale <= 0 {
		config.ProjectionScale = defaults.ProjectionScale
	}
	if config.MinPixelHeight <= 0 {
		config.MinPixelHeight = defaults.MinPixelHeight
	}
	if config.MaxPixelHeight < config.MinPixelHeight {
		config.MaxPixelHeight = defaults.MaxPixelHeight
	}
	if config.Occlusion.SampleSpacing <= 0 {
		config.Occlusion = defaults.Occlusion
	}
	config.OcclusionWidthSamples = max(1, min(5, config.OcclusionWidthSamples))
	return config
}

func billboardMeshHalfWidth(mesh BillboardMesh) float64 {
	width := float64(0)
	for _, vertex := range mesh.Vertices {
		width = max(width, math.Abs(float64(vertex.X)))
	}
	return width
}

func (r *Renderer) drawProjectedBillboards(dst, texture *ebiten.Image, mesh BillboardMesh, projected []projectedBillboard, config BillboardConfig) {
	vertices := make([]ebiten.Vertex, 0, len(projected)*len(mesh.Vertices))
	indices := make([]uint16, 0, len(projected)*len(mesh.Indices))
	flush := func() {
		if len(indices) > 0 {
			dst.DrawTriangles(vertices, indices, texture, nil)
		}
		vertices, indices = vertices[:0], indices[:0]
	}
	for _, item := range projected {
		if len(vertices)+len(mesh.Vertices) > 65535 {
			flush()
		}
		pixelHeight := billboardPixelHeight(item.instance.Height, item.depth, r.screenHeight, config)
		tint := item.instance.Tint
		if tint == (color.RGBA{}) {
			tint = color.RGBA{255, 255, 255, 255}
		}
		instanceVertices := make([]ebiten.Vertex, len(mesh.Vertices))
		for index, vertex := range mesh.Vertices {
			clr := multiplyBillboardColor(vertex.Color, tint)
			instanceVertices[index] = ebiten.Vertex{
				DstX: item.x + vertex.X*pixelHeight, DstY: item.baseY - vertex.Y*pixelHeight,
				SrcX: vertex.SrcX, SrcY: vertex.SrcY,
				ColorR: float32(clr.R) / 255, ColorG: float32(clr.G) / 255, ColorB: float32(clr.B) / 255, ColorA: float32(clr.A) / 255,
			}
		}
		for triangle := 0; triangle+2 < len(mesh.Indices); triangle += 3 {
			a, b, c := mesh.Indices[triangle], mesh.Indices[triangle+1], mesh.Indices[triangle+2]
			if int(a) >= len(instanceVertices) || int(b) >= len(instanceVertices) || int(c) >= len(instanceVertices) {
				continue
			}
			polygon := clipBillboardTriangle([3]ebiten.Vertex{instanceVertices[a], instanceVertices[b], instanceVertices[c]}, item.clipY)
			if len(vertices)+len(polygon) > 65535 {
				flush()
			}
			base := uint16(len(vertices))
			vertices = append(vertices, polygon...)
			for index := 1; index+1 < len(polygon); index++ {
				indices = append(indices, base, base+uint16(index), base+uint16(index+1))
			}
		}
	}
	flush()
}

func billboardPixelHeight(height, depth, screenHeight float64, config BillboardConfig) float32 {
	projected := float32(height / max(.001, depth) * screenHeight)
	return max(config.MinPixelHeight, min(config.MaxPixelHeight, projected))
}

func (r *Renderer) drawBillboardDebugBounds(dst *ebiten.Image, mesh BillboardMesh, projected []projectedBillboard, config BillboardConfig) {
	debugColor := config.DebugColor
	if debugColor == (color.RGBA{}) {
		debugColor = color.RGBA{255, 196, 32, 255}
	}
	halfWidth := float32(billboardMeshHalfWidth(mesh))
	for _, item := range projected {
		height := billboardPixelHeight(item.instance.Height, item.depth, r.screenHeight, config)
		left := item.x - halfWidth*height
		top := item.baseY - height
		width := halfWidth * height * 2
		bottom := min(item.baseY, item.clipY)
		vector.StrokeRect(dst, left, top, width, max(0, bottom-top), 1, debugColor, false)
		if item.clipY < item.baseY {
			vector.StrokeLine(dst, left, item.clipY, left+width, item.clipY, 1, color.RGBA{255, 64, 64, 255}, false)
		}
	}
}

func clipBillboardTriangle(triangle [3]ebiten.Vertex, clipY float32) []ebiten.Vertex {
	result := make([]ebiten.Vertex, 0, 4)
	previous := triangle[2]
	previousInside := previous.DstY <= clipY
	for _, current := range triangle {
		currentInside := current.DstY <= clipY
		if currentInside != previousInside {
			t := (clipY - previous.DstY) / (current.DstY - previous.DstY)
			result = append(result, interpolateBillboardVertex(previous, current, t))
		}
		if currentInside {
			result = append(result, current)
		}
		previous, previousInside = current, currentInside
	}
	return result
}

func interpolateBillboardVertex(a, b ebiten.Vertex, t float32) ebiten.Vertex {
	return ebiten.Vertex{
		DstX: a.DstX + (b.DstX-a.DstX)*t, DstY: a.DstY + (b.DstY-a.DstY)*t,
		SrcX: a.SrcX + (b.SrcX-a.SrcX)*t, SrcY: a.SrcY + (b.SrcY-a.SrcY)*t,
		ColorR: a.ColorR + (b.ColorR-a.ColorR)*t, ColorG: a.ColorG + (b.ColorG-a.ColorG)*t,
		ColorB: a.ColorB + (b.ColorB-a.ColorB)*t, ColorA: a.ColorA + (b.ColorA-a.ColorA)*t,
	}
}

func multiplyBillboardColor(a, b color.RGBA) color.RGBA {
	return color.RGBA{uint8(uint16(a.R) * uint16(b.R) / 255), uint8(uint16(a.G) * uint16(b.G) / 255), uint8(uint16(a.B) * uint16(b.B) / 255), uint8(uint16(a.A) * uint16(b.A) / 255)}
}
