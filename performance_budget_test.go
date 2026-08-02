package fray

import (
	"testing"
	"time"
)

func TestEvaluatePerformanceBudgetReportsAllViolations(t *testing.T) {
	stats := FrameStatsSnapshot{AverageFPS: 58, OnePercentLowFPS: 45, MaximumFrameTime: 30 * time.Millisecond, MaximumDrawCPUTime: 12 * time.Millisecond, ConsecutiveSlowFrames: 4}
	budget := PerformanceBudget{MinimumAverageFPS: 59, MinimumOnePercentLowFPS: 50, MaximumFrameTime: 20 * time.Millisecond, MaximumDrawCPUTime: 10 * time.Millisecond, MaximumConsecutiveSlow: 2}
	if violations := EvaluatePerformanceBudget(stats, budget); len(violations) != 5 {
		t.Fatalf("violations = %v", violations)
	}
}

func TestEvaluatePerformanceBudgetPasses(t *testing.T) {
	stats := FrameStatsSnapshot{AverageFPS: 60, OnePercentLowFPS: 59, MaximumFrameTime: 17 * time.Millisecond}
	budget := PerformanceBudget{MinimumAverageFPS: 59, MinimumOnePercentLowFPS: 58, MaximumFrameTime: 18 * time.Millisecond}
	if violations := EvaluatePerformanceBudget(stats, budget); len(violations) != 0 {
		t.Fatalf("violations = %v", violations)
	}
}
