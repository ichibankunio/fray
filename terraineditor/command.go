package terraineditor

import (
	"fmt"
	"image"
	"math"
)

type Command struct {
	Operation  string     `json:"operation"`
	Parameters Parameters `json:"parameters"`
}

type Parameters struct {
	X       int     `json:"x,omitempty"`
	Y       int     `json:"y,omitempty"`
	Width   int     `json:"width,omitempty"`
	Rows    int     `json:"rows,omitempty"`
	Height  int     `json:"height,omitempty"`
	Radius  int     `json:"radius,omitempty"`
	Amount  int     `json:"amount,omitempty"`
	Falloff string  `json:"falloff,omitempty"`
	Blend   float64 `json:"blend,omitempty"`
}

type ColumnChange struct {
	X      int     `json:"x"`
	Y      int     `json:"y"`
	Before []uint8 `json:"before"`
	After  []uint8 `json:"after"`
}

type ChangeSet struct {
	Region  image.Rectangle `json:"region"`
	Changes []ColumnChange  `json:"changes"`
}

func (d *Document) Apply(command Command) (ChangeSet, error) {
	if err := d.Validate(); err != nil {
		return ChangeSet{}, err
	}
	region, err := d.commandRegion(command)
	if err != nil {
		return ChangeSet{}, err
	}
	var beforeHeights []uint8
	if command.Operation == "smooth" {
		beforeHeights = d.heightSnapshot()
	}
	targets := make(map[int]int, region.Dx()*region.Dy())
	for y := region.Min.Y; y < region.Max.Y; y++ {
		for x := region.Min.X; x < region.Max.X; x++ {
			if !commandIncludesCell(command, x, y) {
				continue
			}
			index := y*d.CanvasWidth + x
			current := d.HeightAt(x, y)
			if beforeHeights != nil {
				current = int(beforeHeights[index])
			}
			target := current
			switch command.Operation {
			case "set-height", "flatten":
				target = command.Parameters.Height
			case "raise", "lower":
				amount := max(1, command.Parameters.Amount)
				if command.Operation == "lower" {
					amount = -amount
				}
				target = current + brushAmount(command, x, y, amount)
			case "smooth":
				average := d.neighborAverage(beforeHeights, x, y)
				blend := command.Parameters.Blend
				if blend <= 0 || blend > 1 {
					blend = .5
				}
				target = int(math.Round(float64(current) + (average-float64(current))*blend))
			default:
				return ChangeSet{}, fmt.Errorf("unknown terrain operation %q", command.Operation)
			}
			targets[index] = max(0, min(d.CanvasDepth, target))
		}
	}

	changeSet := ChangeSet{}
	for y := region.Min.Y; y < region.Max.Y; y++ {
		for x := region.Min.X; x < region.Max.X; x++ {
			index := y*d.CanvasWidth + x
			target, ok := targets[index]
			if !ok || target == d.HeightAt(x, y) {
				continue
			}
			before := d.column(x, y)
			d.setColumnHeight(x, y, target)
			after := d.column(x, y)
			changeSet.Changes = append(changeSet.Changes, ColumnChange{X: x, Y: y, Before: before, After: after})
			cell := image.Rect(x, y, x+1, y+1)
			if changeSet.Region.Empty() {
				changeSet.Region = cell
			} else {
				changeSet.Region = changeSet.Region.Union(cell)
			}
		}
	}
	return changeSet, nil
}

func (d *Document) ApplyChangeSet(changeSet ChangeSet, forward bool) error {
	for _, change := range changeSet.Changes {
		if change.X < 0 || change.Y < 0 || change.X >= d.CanvasWidth || change.Y >= d.CanvasHeight {
			return fmt.Errorf("change coordinate (%d,%d) is outside terrain", change.X, change.Y)
		}
		column := change.Before
		if forward {
			column = change.After
		}
		if len(column) != d.CanvasDepth {
			return fmt.Errorf("change column at (%d,%d) has depth %d, expected %d", change.X, change.Y, len(column), d.CanvasDepth)
		}
		d.setColumn(change.X, change.Y, column)
	}
	return nil
}

func (d *Document) commandRegion(command Command) (image.Rectangle, error) {
	p := command.Parameters
	region := image.Rect(p.X, p.Y, p.X+1, p.Y+1)
	if p.Width > 0 && p.Rows > 0 {
		region = image.Rect(p.X, p.Y, p.X+p.Width, p.Y+p.Rows)
	} else if p.Radius > 0 {
		region = image.Rect(p.X-p.Radius, p.Y-p.Radius, p.X+p.Radius+1, p.Y+p.Radius+1)
	}
	region = region.Intersect(image.Rect(0, 0, d.CanvasWidth, d.CanvasHeight))
	if region.Empty() {
		return image.Rectangle{}, fmt.Errorf("terrain command is outside document bounds")
	}
	return region, nil
}

func commandIncludesCell(command Command, x, y int) bool {
	p := command.Parameters
	if p.Radius <= 0 || p.Width > 0 || p.Rows > 0 {
		return true
	}
	dx, dy := x-p.X, y-p.Y
	return dx*dx+dy*dy <= p.Radius*p.Radius
}

func brushAmount(command Command, x, y, amount int) int {
	p := command.Parameters
	if p.Falloff != "smooth" || p.Radius <= 0 {
		return amount
	}
	distance := math.Hypot(float64(x-p.X), float64(y-p.Y))
	weight := max(0.0, 1-distance/float64(p.Radius+1))
	return int(math.Round(float64(amount) * weight))
}

func (d *Document) neighborAverage(heights []uint8, x, y int) float64 {
	total, count := 0, 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx, ny := x+dx, y+dy
			if nx >= 0 && ny >= 0 && nx < d.CanvasWidth && ny < d.CanvasHeight {
				total += int(heights[ny*d.CanvasWidth+nx])
				count++
			}
		}
	}
	return float64(total) / float64(count)
}

func (d *Document) heightSnapshot() []uint8 {
	heights := make([]uint8, d.CanvasWidth*d.CanvasHeight)
	for y := 0; y < d.CanvasHeight; y++ {
		for x := 0; x < d.CanvasWidth; x++ {
			heights[y*d.CanvasWidth+x] = uint8(d.HeightAt(x, y))
		}
	}
	return heights
}

func (d *Document) ensureLayers(depth int) {
	cells := d.CanvasWidth * d.CanvasHeight
	for len(d.Layers) < depth {
		d.Layers = append(d.Layers, make([]uint8, cells))
	}
}

func (d *Document) setColumnHeight(x, y, height int) {
	d.ensureLayers(height)
	index := y*d.CanvasWidth + x
	material := uint8(1)
	for z := min(len(d.Layers), d.CanvasDepth) - 1; z >= 0; z-- {
		if d.Layers[z][index] != 0 {
			material = d.Layers[z][index]
			break
		}
	}
	for z := 0; z < len(d.Layers); z++ {
		if z < height {
			if d.Layers[z][index] == 0 {
				d.Layers[z][index] = material
			}
		} else {
			d.Layers[z][index] = 0
		}
	}
}

func (d *Document) column(x, y int) []uint8 {
	column := make([]uint8, d.CanvasDepth)
	index := y*d.CanvasWidth + x
	for z := 0; z < len(d.Layers); z++ {
		column[z] = d.Layers[z][index]
	}
	return column
}

func (d *Document) setColumn(x, y int, column []uint8) {
	d.ensureLayers(d.CanvasDepth)
	index := y*d.CanvasWidth + x
	for z := range column {
		d.Layers[z][index] = column[z]
	}
}
