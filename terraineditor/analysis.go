package terraineditor

import (
	"fmt"
	"image"
	"math"
)

type ChangeSummary struct {
	ChangedCells int             `json:"changedCells"`
	Region       image.Rectangle `json:"region"`
	MinimumDelta int             `json:"minimumDelta"`
	MaximumDelta int             `json:"maximumDelta"`
}

type Constraints struct {
	MinHeight       *int     `json:"minHeight,omitempty"`
	MaxHeight       *int     `json:"maxHeight,omitempty"`
	MaxSlope        *float64 `json:"maxSlope,omitempty"`
	ProtectBoundary bool     `json:"protectBoundary,omitempty"`
}

func ColumnHeight(column []uint8) int {
	for z := len(column) - 1; z >= 0; z-- {
		if column[z] != 0 {
			return z + 1
		}
	}
	return 0
}

func Summarize(changeSet ChangeSet) ChangeSummary {
	summary := ChangeSummary{ChangedCells: len(changeSet.Changes), Region: changeSet.Region}
	for i, change := range changeSet.Changes {
		delta := ColumnHeight(change.After) - ColumnHeight(change.Before)
		if i == 0 || delta < summary.MinimumDelta {
			summary.MinimumDelta = delta
		}
		if i == 0 || delta > summary.MaximumDelta {
			summary.MaximumDelta = delta
		}
	}
	return summary
}

func Diff(before, after *Document) (ChangeSet, error) {
	if before == nil || after == nil {
		return ChangeSet{}, fmt.Errorf("terrain diff requires two documents")
	}
	if before.CanvasWidth != after.CanvasWidth || before.CanvasHeight != after.CanvasHeight || before.CanvasDepth != after.CanvasDepth {
		return ChangeSet{}, fmt.Errorf("terrain diff dimensions do not match")
	}
	result := ChangeSet{}
	for y := 0; y < before.CanvasHeight; y++ {
		for x := 0; x < before.CanvasWidth; x++ {
			index := y*before.CanvasWidth + x
			equal := true
			for z := 0; z < before.CanvasDepth; z++ {
				var b, a uint8
				if z < len(before.Layers) {
					b = before.Layers[z][index]
				}
				if z < len(after.Layers) {
					a = after.Layers[z][index]
				}
				if b != a {
					equal = false
					break
				}
			}
			if equal {
				continue
			}
			cell := image.Rect(x, y, x+1, y+1)
			if result.Region.Empty() {
				result.Region = cell
			} else {
				result.Region = result.Region.Union(cell)
			}
			result.Changes = append(result.Changes, ColumnChange{X: x, Y: y, Before: before.column(x, y), After: after.column(x, y)})
		}
	}
	return result, nil
}

func (d *Document) Clone() (*Document, error) {
	data, err := d.Marshal()
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

// Preview calculates the exact change without leaving the document modified.
func (d *Document) Preview(command Command) (ChangeSet, error) {
	changeSet, err := d.Apply(command)
	if err != nil {
		return ChangeSet{}, err
	}
	if err := d.ApplyChangeSet(changeSet, false); err != nil {
		return ChangeSet{}, fmt.Errorf("restore terrain after preview: %w", err)
	}
	return changeSet, nil
}

// ApplyConstrained applies atomically: a failed constraint restores every cell.
func (d *Document) ApplyConstrained(command Command, constraints Constraints) (ChangeSet, error) {
	changeSet, err := d.Apply(command)
	if err != nil {
		return ChangeSet{}, err
	}
	if err := d.ValidateConstraints(changeSet.Region, constraints); err != nil {
		if restoreErr := d.ApplyChangeSet(changeSet, false); restoreErr != nil {
			return ChangeSet{}, fmt.Errorf("%v; restore terrain: %w", err, restoreErr)
		}
		return ChangeSet{}, err
	}
	return changeSet, nil
}

func (d *Document) ValidateConstraints(region image.Rectangle, constraints Constraints) error {
	region = region.Intersect(image.Rect(0, 0, d.CanvasWidth, d.CanvasHeight))
	if constraints.ProtectBoundary && (region.Min.X == 0 || region.Min.Y == 0 || region.Max.X == d.CanvasWidth || region.Max.Y == d.CanvasHeight) {
		return fmt.Errorf("constraint protect_boundary: edit touches terrain boundary")
	}
	check := region
	if constraints.MaxSlope != nil {
		check = image.Rect(region.Min.X-1, region.Min.Y-1, region.Max.X+1, region.Max.Y+1).Intersect(image.Rect(0, 0, d.CanvasWidth, d.CanvasHeight))
	}
	for y := check.Min.Y; y < check.Max.Y; y++ {
		for x := check.Min.X; x < check.Max.X; x++ {
			h := d.HeightAt(x, y)
			if constraints.MinHeight != nil && h < *constraints.MinHeight {
				return fmt.Errorf("constraint min_height: height %d at (%d,%d) is below %d", h, x, y, *constraints.MinHeight)
			}
			if constraints.MaxHeight != nil && h > *constraints.MaxHeight {
				return fmt.Errorf("constraint max_height: height %d at (%d,%d) exceeds %d", h, x, y, *constraints.MaxHeight)
			}
			if constraints.MaxSlope != nil {
				if x+1 < d.CanvasWidth && math.Abs(float64(h-d.HeightAt(x+1, y))) > *constraints.MaxSlope {
					return fmt.Errorf("constraint max_slope: edge at (%d,%d) exceeds %.3f", x, y, *constraints.MaxSlope)
				}
				if y+1 < d.CanvasHeight && math.Abs(float64(h-d.HeightAt(x, y+1))) > *constraints.MaxSlope {
					return fmt.Errorf("constraint max_slope: edge at (%d,%d) exceeds %.3f", x, y, *constraints.MaxSlope)
				}
			}
		}
	}
	return nil
}
