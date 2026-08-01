package fray

import (
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
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
	NearDistance     float64
	FarDistance      float64
	ProjectionScale  float64
	ScreenMargin     float32
	MinPixelHeight   float32
	MaxPixelHeight   float32
	TerrainOcclusion bool
	Occlusion        TerrainOcclusionConfig
}

func DefaultBillboardConfig() BillboardConfig {
	return BillboardConfig{
		NearDistance: 1, FarDistance: 14, ProjectionScale: .48,
		ScreenMargin: 80, MinPixelHeight: 1, MaxPixelHeight: 4096,
		TerrainOcclusion: true, Occlusion: DefaultTerrainOcclusionConfig(),
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
	return config
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
		pixelHeight := float32(item.instance.Height / item.depth * r.screenHeight)
		pixelHeight = max(config.MinPixelHeight, min(config.MaxPixelHeight, pixelHeight))
		base := uint16(len(vertices))
		tint := item.instance.Tint
		if tint == (color.RGBA{}) {
			tint = color.RGBA{255, 255, 255, 255}
		}
		for _, vertex := range mesh.Vertices {
			clr := multiplyBillboardColor(vertex.Color, tint)
			vertices = append(vertices, ebiten.Vertex{
				DstX: item.x + vertex.X*pixelHeight, DstY: min(item.baseY-vertex.Y*pixelHeight, item.clipY),
				SrcX: vertex.SrcX, SrcY: vertex.SrcY,
				ColorR: float32(clr.R) / 255, ColorG: float32(clr.G) / 255, ColorB: float32(clr.B) / 255, ColorA: float32(clr.A) / 255,
			})
		}
		for _, index := range mesh.Indices {
			indices = append(indices, base+index)
		}
	}
	flush()
}

func multiplyBillboardColor(a, b color.RGBA) color.RGBA {
	return color.RGBA{uint8(uint16(a.R) * uint16(b.R) / 255), uint8(uint16(a.G) * uint16(b.G) / 255), uint8(uint16(a.B) * uint16(b.B) / 255), uint8(uint16(a.A) * uint16(b.A) / 255)}
}
