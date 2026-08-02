package fray

import "time"

type PerformanceSectionStats struct {
	Samples int
	Total   time.Duration
	Average time.Duration
	Maximum time.Duration
}

type performanceSectionRecord struct {
	samples int
	total   time.Duration
	maximum time.Duration
}

type PerformanceSectionToken struct {
	recorder *performanceSectionRecorder
	name     string
	start    time.Time
}

type performanceSectionRecorder struct {
	enabled bool
	records map[string]*performanceSectionRecord
}

// SetPerformanceSectionsEnabled enables opt-in named CPU section timing.
func (r *Renderer) SetPerformanceSectionsEnabled(enabled bool) {
	r.performanceSections.enabled = enabled
	if enabled && r.performanceSections.records == nil {
		r.performanceSections.records = make(map[string]*performanceSectionRecord)
	}
}

// BeginPerformanceSection starts a named interval. Call End on the returned token.
func (r *Renderer) BeginPerformanceSection(name string) PerformanceSectionToken {
	if !r.performanceSections.enabled || name == "" {
		return PerformanceSectionToken{}
	}
	return PerformanceSectionToken{recorder: &r.performanceSections, name: name, start: time.Now()}
}

// End records the interval. Repeated calls are ignored.
func (token *PerformanceSectionToken) End() {
	if token == nil || token.recorder == nil {
		return
	}
	duration := time.Since(token.start)
	record := token.recorder.records[token.name]
	if record == nil {
		record = &performanceSectionRecord{}
		token.recorder.records[token.name] = record
	}
	record.samples++
	record.total += duration
	if duration > record.maximum {
		record.maximum = duration
	}
	token.recorder = nil
}

// PerformanceSections returns a copy of all named timing summaries.
func (r *Renderer) PerformanceSections() map[string]PerformanceSectionStats {
	result := make(map[string]PerformanceSectionStats, len(r.performanceSections.records))
	for name, record := range r.performanceSections.records {
		result[name] = PerformanceSectionStats{Samples: record.samples, Total: record.total, Average: record.total / time.Duration(record.samples), Maximum: record.maximum}
	}
	return result
}

func (r *Renderer) ResetPerformanceSections() {
	clear(r.performanceSections.records)
}
