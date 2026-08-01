package fray

import (
	"container/heap"
	"math"

	"github.com/ichibankunio/fvec/vec3"
)

type TerrainNavigationConfig struct {
	MaxSlope      float64
	MaxStep       float64
	SlopePenalty  float64
	AllowDiagonal bool
}

type TerrainNavigation struct {
	Width    int
	Height   int
	Walkable []bool
	Cost     []float64
	heights  []float64
	config   TerrainNavigationConfig
}

// BuildTerrainNavigation derives a navigation grid from the interpolated
// height surface. It requires no authored navigation data.
func (w *World) BuildTerrainNavigation(config TerrainNavigationConfig) *TerrainNavigation {
	if config.MaxSlope <= 0 {
		config.MaxSlope = .8
	}
	if config.MaxStep <= 0 {
		config.MaxStep = 1.25
	}
	if config.SlopePenalty < 0 {
		config.SlopePenalty = 0
	}
	navigation := &TerrainNavigation{
		Width:    w.canvasWidth,
		Height:   w.canvasHeight,
		Walkable: make([]bool, w.canvasWidth*w.canvasHeight),
		Cost:     make([]float64, w.canvasWidth*w.canvasHeight),
		heights:  make([]float64, w.canvasWidth*w.canvasHeight),
		config:   config,
	}
	for y := 0; y < w.canvasHeight; y++ {
		for x := 0; x < w.canvasWidth; x++ {
			index := y*w.canvasWidth + x
			height := w.SampleTerrainHeight(float64(x)+.5, float64(y)+.5)
			slope := w.SampleTerrainSlope(float64(x)+.5, float64(y)+.5)
			navigation.heights[index] = height
			navigation.Walkable[index] = height > 0 && slope <= config.MaxSlope
			navigation.Cost[index] = 1 + slope*config.SlopePenalty
		}
	}
	return navigation
}

// FindPath returns cell-center positions in grid blocks, including terrain Z.
func (n *TerrainNavigation) FindPath(startX, startY, goalX, goalY int) ([]vec3.Vec3, bool) {
	if !n.inBounds(startX, startY) || !n.inBounds(goalX, goalY) {
		return nil, false
	}
	start := startY*n.Width + startX
	goal := goalY*n.Width + goalX
	if !n.Walkable[start] || !n.Walkable[goal] {
		return nil, false
	}

	count := n.Width * n.Height
	costs := make([]float64, count)
	parents := make([]int, count)
	closed := make([]bool, count)
	for i := range costs {
		costs[i] = math.Inf(1)
		parents[i] = -1
	}
	costs[start] = 0
	queue := &terrainPathQueue{{index: start, priority: n.heuristic(startX, startY, goalX, goalY)}}
	heap.Init(queue)
	directions := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	if n.config.AllowDiagonal {
		directions = append(directions, [2]int{-1, -1}, [2]int{1, -1}, [2]int{-1, 1}, [2]int{1, 1})
	}

	for queue.Len() > 0 {
		current := heap.Pop(queue).(terrainPathNode).index
		if closed[current] {
			continue
		}
		if current == goal {
			return n.reconstructPath(parents, goal), true
		}
		closed[current] = true
		cx := current % n.Width
		cy := current / n.Width
		for _, direction := range directions {
			nx := cx + direction[0]
			ny := cy + direction[1]
			if !n.inBounds(nx, ny) {
				continue
			}
			next := ny*n.Width + nx
			if closed[next] || !n.Walkable[next] || math.Abs(n.heights[next]-n.heights[current]) > n.config.MaxStep {
				continue
			}
			moveCost := n.Cost[next]
			if direction[0] != 0 && direction[1] != 0 {
				moveCost *= math.Sqrt2
			}
			candidate := costs[current] + moveCost
			if candidate >= costs[next] {
				continue
			}
			costs[next] = candidate
			parents[next] = current
			priority := candidate + n.heuristic(nx, ny, goalX, goalY)
			heap.Push(queue, terrainPathNode{index: next, priority: priority})
		}
	}
	return nil, false
}

func (n *TerrainNavigation) inBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < n.Width && y < n.Height
}

func (n *TerrainNavigation) heuristic(x, y, goalX, goalY int) float64 {
	return math.Hypot(float64(goalX-x), float64(goalY-y))
}

func (n *TerrainNavigation) reconstructPath(parents []int, goal int) []vec3.Vec3 {
	indices := make([]int, 0)
	for current := goal; current >= 0; current = parents[current] {
		indices = append(indices, current)
	}
	path := make([]vec3.Vec3, len(indices))
	for i := range indices {
		reversed := indices[len(indices)-1-i]
		x := reversed % n.Width
		y := reversed / n.Width
		path[i] = vec3.New(float64(x)+.5, float64(y)+.5, n.heights[reversed])
	}
	return path
}

type terrainPathNode struct {
	index    int
	priority float64
}

type terrainPathQueue []terrainPathNode

func (q terrainPathQueue) Len() int           { return len(q) }
func (q terrainPathQueue) Less(i, j int) bool { return q[i].priority < q[j].priority }
func (q terrainPathQueue) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q *terrainPathQueue) Push(value any)    { *q = append(*q, value.(terrainPathNode)) }
func (q *terrainPathQueue) Pop() any {
	old := *q
	last := old[len(old)-1]
	*q = old[:len(old)-1]
	return last
}
