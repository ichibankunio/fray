package fray

import (
	"math"
	"testing"
	"time"
)

func TestFrameStatsSummarizesFramePacing(t *testing.T) {
	r := &Renderer{}
	r.frameStats.enabled = true
	r.frameStats.count = 100
	for i := 0; i < 100; i++ {
		r.frameStats.frameTimes[i] = 10 * time.Millisecond
		r.frameStats.drawTimes[i] = 2 * time.Millisecond
	}
	r.frameStats.frameTimes[99] = 20 * time.Millisecond
	r.frameStats.drawTimes[99] = 4 * time.Millisecond
	snapshot := r.FrameStats()
	if snapshot.Samples != 100 || snapshot.MaximumFrameTime != 20*time.Millisecond || snapshot.MaximumDrawCPUTime != 4*time.Millisecond {
		t.Fatalf("snapshot = %+v", snapshot)
	}
	if math.Abs(snapshot.OnePercentLowFPS-50) > 1e-9 {
		t.Fatalf("1%% low = %v, want 50", snapshot.OnePercentLowFPS)
	}
	if snapshot.P99FrameTime != 10*time.Millisecond {
		t.Fatalf("p99 = %v, want 10ms", snapshot.P99FrameTime)
	}
}

func TestFrameStatsDisabledByDefault(t *testing.T) {
	r := &Renderer{}
	if got := r.beginFrameStats(); !got.IsZero() {
		t.Fatalf("disabled recorder returned start time %v", got)
	}
	if snapshot := r.FrameStats(); snapshot.Samples != 0 {
		t.Fatalf("disabled snapshot = %+v", snapshot)
	}
}
