package terraineditor

import "testing"

func TestAnalyzeReturnsDeterministicSuggestedFix(t *testing.T) {
	d := testDocument()
	_, _ = d.Apply(Command{Operation: "set-height", Parameters: Parameters{X: 1, Y: 1, Height: 8}})
	report := Analyze(d, AnalysisOptions{MaximumSlope: 2, SpikeHeight: 2})
	if len(report.Issues) == 0 {
		t.Fatal("expected terrain issues")
	}
	found := false
	for _, issue := range report.Issues {
		if issue.Code == "isolated_peak" && issue.Suggestion != nil {
			found = true
		}
	}
	if !found {
		t.Fatalf("report lacks peak suggestion: %+v", report)
	}
}
