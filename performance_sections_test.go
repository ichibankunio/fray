package fray

import (
	"testing"
	"time"
)

func TestPerformanceSectionsAreOptIn(t *testing.T) {
	r := &Renderer{}
	token := r.BeginPerformanceSection("terrain")
	token.End()
	if sections := r.PerformanceSections(); len(sections) != 0 {
		t.Fatalf("disabled sections = %v", sections)
	}
}

func TestPerformanceSectionRecordsAndResets(t *testing.T) {
	r := &Renderer{}
	r.SetPerformanceSectionsEnabled(true)
	token := r.BeginPerformanceSection("terrain")
	time.Sleep(time.Millisecond)
	token.End()
	token.End()
	stats := r.PerformanceSections()["terrain"]
	if stats.Samples != 1 || stats.Total <= 0 || stats.Average != stats.Total {
		t.Fatalf("stats = %+v", stats)
	}
	r.ResetPerformanceSections()
	if sections := r.PerformanceSections(); len(sections) != 0 {
		t.Fatalf("sections after reset = %v", sections)
	}
}
