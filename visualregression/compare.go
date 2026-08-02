// Package visualregression provides renderer-independent screenshot comparison.
package visualregression

import "image"

type Options struct {
	IgnoreTopRows    int
	ChannelTolerance uint32
	MaxChangedRatio  float64
	MaxMeanError     float64
}

type Result struct {
	Pixels, Changed int
	ChangedRatio    float64
	MeanError       float64
	Passed          bool
}

func DefaultOptions() Options {
	return Options{ChannelTolerance: 8, MaxChangedRatio: .005, MaxMeanError: 1}
}

// Compare checks equally sized images while tolerating minor GPU rounding differences.
func Compare(baseline, actual image.Image, options Options) Result {
	if baseline == nil || actual == nil || baseline.Bounds() != actual.Bounds() {
		return Result{}
	}
	bounds := baseline.Bounds()
	changed, pixels := 0, 0
	var totalError uint64
	for y := max(bounds.Min.Y, options.IgnoreTopRows); y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			br, bg, bb, _ := baseline.At(x, y).RGBA()
			ar, ag, ab, _ := actual.At(x, y).RGBA()
			dr, dg, db := absDiff(br>>8, ar>>8), absDiff(bg>>8, ag>>8), absDiff(bb>>8, ab>>8)
			totalError += uint64(dr + dg + db)
			if dr > options.ChannelTolerance || dg > options.ChannelTolerance || db > options.ChannelTolerance {
				changed++
			}
			pixels++
		}
	}
	if pixels == 0 {
		return Result{}
	}
	ratio := float64(changed) / float64(pixels)
	mean := float64(totalError) / float64(pixels*3)
	return Result{Pixels: pixels, Changed: changed, ChangedRatio: ratio, MeanError: mean, Passed: ratio <= options.MaxChangedRatio && mean <= options.MaxMeanError}
}

func absDiff(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}
