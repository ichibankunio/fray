package fray

import (
	"fmt"
	"time"
)

type PerformanceBudget struct {
	MinimumAverageFPS       float64
	MinimumOnePercentLowFPS float64
	MaximumFrameTime        time.Duration
	MaximumDrawCPUTime      time.Duration
	MaximumConsecutiveSlow  int
}

type PerformanceViolation struct {
	Metric string
	Actual float64
	Limit  float64
	Unit   string
}

func (v PerformanceViolation) Error() string {
	return fmt.Sprintf("performance budget exceeded: %s %.3f %s (limit %.3f %s)", v.Metric, v.Actual, v.Unit, v.Limit, v.Unit)
}

// EvaluatePerformanceBudget reports every exceeded budget without changing quality settings.
func EvaluatePerformanceBudget(stats FrameStatsSnapshot, budget PerformanceBudget) []PerformanceViolation {
	violations := make([]PerformanceViolation, 0, 5)
	if budget.MinimumAverageFPS > 0 && stats.AverageFPS < budget.MinimumAverageFPS {
		violations = append(violations, PerformanceViolation{Metric: "average_fps", Actual: stats.AverageFPS, Limit: budget.MinimumAverageFPS, Unit: "fps"})
	}
	if budget.MinimumOnePercentLowFPS > 0 && stats.OnePercentLowFPS < budget.MinimumOnePercentLowFPS {
		violations = append(violations, PerformanceViolation{Metric: "one_percent_low_fps", Actual: stats.OnePercentLowFPS, Limit: budget.MinimumOnePercentLowFPS, Unit: "fps"})
	}
	if budget.MaximumFrameTime > 0 && stats.MaximumFrameTime > budget.MaximumFrameTime {
		violations = append(violations, durationViolation("maximum_frame_time", stats.MaximumFrameTime, budget.MaximumFrameTime))
	}
	if budget.MaximumDrawCPUTime > 0 && stats.MaximumDrawCPUTime > budget.MaximumDrawCPUTime {
		violations = append(violations, durationViolation("maximum_draw_cpu_time", stats.MaximumDrawCPUTime, budget.MaximumDrawCPUTime))
	}
	if budget.MaximumConsecutiveSlow > 0 && stats.ConsecutiveSlowFrames > budget.MaximumConsecutiveSlow {
		violations = append(violations, PerformanceViolation{Metric: "consecutive_slow_frames", Actual: float64(stats.ConsecutiveSlowFrames), Limit: float64(budget.MaximumConsecutiveSlow), Unit: "frames"})
	}
	return violations
}

func durationViolation(metric string, actual, limit time.Duration) PerformanceViolation {
	return PerformanceViolation{Metric: metric, Actual: float64(actual) / float64(time.Millisecond), Limit: float64(limit) / float64(time.Millisecond), Unit: "ms"}
}
