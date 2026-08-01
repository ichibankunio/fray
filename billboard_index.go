package fray

import "math"

type billboardCell struct{ X, Y int }

// BillboardSpatialIndex partitions static instances into square world cells.
// QueryInto reuses caller-owned memory and returns only nearby candidates.
type BillboardSpatialIndex struct {
	cellSize  float64
	instances []BillboardInstance
	cells     map[billboardCell][]int
}

func NewBillboardSpatialIndex(instances []BillboardInstance, cellSize float64) *BillboardSpatialIndex {
	if cellSize <= 0 {
		cellSize = 8
	}
	index := &BillboardSpatialIndex{
		cellSize: cellSize, instances: append([]BillboardInstance(nil), instances...),
		cells: make(map[billboardCell][]int),
	}
	for instanceIndex, instance := range index.instances {
		cell := index.cellAt(instance.Position.X, instance.Position.Y)
		index.cells[cell] = append(index.cells[cell], instanceIndex)
	}
	return index
}

func (index *BillboardSpatialIndex) Len() int {
	if index == nil {
		return 0
	}
	return len(index.instances)
}

func (index *BillboardSpatialIndex) QueryInto(dst []BillboardInstance, x, y, radius float64) []BillboardInstance {
	dst = dst[:0]
	if index == nil || radius < 0 {
		return dst
	}
	minCell := index.cellAt(x-radius, y-radius)
	maxCell := index.cellAt(x+radius, y+radius)
	radiusSquared := radius * radius
	for cellY := minCell.Y; cellY <= maxCell.Y; cellY++ {
		for cellX := minCell.X; cellX <= maxCell.X; cellX++ {
			for _, instanceIndex := range index.cells[billboardCell{cellX, cellY}] {
				instance := index.instances[instanceIndex]
				dx, dy := instance.Position.X-x, instance.Position.Y-y
				if dx*dx+dy*dy <= radiusSquared {
					dst = append(dst, instance)
				}
			}
		}
	}
	return dst
}

func (index *BillboardSpatialIndex) cellAt(x, y float64) billboardCell {
	return billboardCell{int(math.Floor(x / index.cellSize)), int(math.Floor(y / index.cellSize))}
}
