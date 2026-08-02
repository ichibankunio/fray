package fray

import (
	"sort"
	"time"
)

const frameStatsCapacity = 600

// FrameStatsSnapshot summarizes recent frame pacing. P99FrameTime is the
// threshold used to derive OnePercentLowFPS.
type FrameStatsSnapshot struct {
	Samples               int
	AverageFPS            float64
	OnePercentLowFPS      float64
	AverageFrameTime      time.Duration
	P99FrameTime          time.Duration
	MaximumFrameTime      time.Duration
	AverageDrawCPUTime    time.Duration
	MaximumDrawCPUTime    time.Duration
	ConsecutiveSlowFrames int
}

type frameStatsRecorder struct {
	enabled         bool
	lastFrame       time.Time
	frameTimes      [frameStatsCapacity]time.Duration
	drawTimes       [frameStatsCapacity]time.Duration
	next            int
	count           int
	consecutiveSlow int
}

// SetFrameStatsEnabled enables low-overhead CPU frame pacing collection.
// Collection is opt-in so release builds pay no clock-query cost by default.
func (r *Renderer) SetFrameStatsEnabled(enabled bool) {
	r.frameStats = frameStatsRecorder{enabled: enabled}
}

func (r *Renderer) beginFrameStats() time.Time {
	if !r.frameStats.enabled {
		return time.Time{}
	}
	now := time.Now()
	if r.frameStats.lastFrame.IsZero() {
		r.frameStats.lastFrame = now
		return time.Time{}
	}
	frameTime := now.Sub(r.frameStats.lastFrame)
	index := r.frameStats.next
	r.frameStats.frameTimes[index] = frameTime
	if frameTime > time.Second/59 {
		r.frameStats.consecutiveSlow++
	} else {
		r.frameStats.consecutiveSlow = 0
	}
	r.frameStats.lastFrame = now
	return now
}

func (r *Renderer) endFrameStats(start time.Time) {
	if start.IsZero() {
		return
	}
	index := r.frameStats.next
	r.frameStats.drawTimes[index] = time.Since(start)
	r.frameStats.next = (index + 1) % frameStatsCapacity
	if r.frameStats.count < frameStatsCapacity {
		r.frameStats.count++
	}
}

// FrameStats returns a snapshot without altering the active collection window.
func (r *Renderer) FrameStats() FrameStatsSnapshot {
	count := r.frameStats.count
	result := FrameStatsSnapshot{Samples: count, ConsecutiveSlowFrames: r.frameStats.consecutiveSlow}
	if count == 0 {
		return result
	}
	ordered := make([]time.Duration, count)
	var frameTotal, drawTotal time.Duration
	for i := 0; i < count; i++ {
		frame := r.frameStats.frameTimes[i]
		draw := r.frameStats.drawTimes[i]
		ordered[i] = frame
		frameTotal += frame
		drawTotal += draw
		if frame > result.MaximumFrameTime {
			result.MaximumFrameTime = frame
		}
		if draw > result.MaximumDrawCPUTime {
			result.MaximumDrawCPUTime = draw
		}
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	result.AverageFrameTime = frameTotal / time.Duration(count)
	result.AverageDrawCPUTime = drawTotal / time.Duration(count)
	p99Index := max(0, int(float64(count)*.99+.999999)-1)
	result.P99FrameTime = ordered[p99Index]
	if result.AverageFrameTime > 0 {
		result.AverageFPS = float64(time.Second) / float64(result.AverageFrameTime)
	}
	slowCount := max(1, int(float64(count)*.01+.999999))
	var slowTotal time.Duration
	for _, frame := range ordered[count-slowCount:] {
		slowTotal += frame
	}
	if slowAverage := slowTotal / time.Duration(slowCount); slowAverage > 0 {
		result.OnePercentLowFPS = float64(time.Second) / float64(slowAverage)
	}
	return result
}
