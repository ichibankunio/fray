package terraineditor

import "math"

type Issue struct {
	Code       string   `json:"code"`
	Message    string   `json:"message"`
	X          int      `json:"x"`
	Y          int      `json:"y"`
	Value      float64  `json:"value"`
	Suggestion *Command `json:"suggestion,omitempty"`
}

type AnalysisOptions struct {
	MaximumSlope float64 `json:"maximumSlope"`
	SpikeHeight  float64 `json:"spikeHeight"`
}

type AnalysisReport struct {
	Width  int     `json:"width"`
	Height int     `json:"height"`
	Issues []Issue `json:"issues"`
}

func Analyze(document *Document, options AnalysisOptions) AnalysisReport {
	if options.MaximumSlope <= 0 {
		options.MaximumSlope = 4
	}
	if options.SpikeHeight <= 0 {
		options.SpikeHeight = 3
	}
	report := AnalysisReport{Width: document.CanvasWidth, Height: document.CanvasHeight, Issues: []Issue{}}
	for y := 0; y < document.CanvasHeight; y++ {
		for x := 0; x < document.CanvasWidth; x++ {
			h := document.HeightAt(x, y)
			if x+1 < document.CanvasWidth {
				if delta := math.Abs(float64(h - document.HeightAt(x+1, y))); delta > options.MaximumSlope {
					c := Command{Operation: "smooth", Parameters: Parameters{X: x, Y: y, Radius: 1, Blend: .5}}
					report.Issues = append(report.Issues, Issue{Code: "steep_edge", Message: "adjacent height difference exceeds limit", X: x, Y: y, Value: delta, Suggestion: &c})
				}
			}
			if y+1 < document.CanvasHeight {
				if delta := math.Abs(float64(h - document.HeightAt(x, y+1))); delta > options.MaximumSlope {
					c := Command{Operation: "smooth", Parameters: Parameters{X: x, Y: y, Radius: 1, Blend: .5}}
					report.Issues = append(report.Issues, Issue{Code: "steep_edge", Message: "adjacent height difference exceeds limit", X: x, Y: y, Value: delta, Suggestion: &c})
				}
			}
			if x > 0 && y > 0 && x+1 < document.CanvasWidth && y+1 < document.CanvasHeight {
				average := float64(document.HeightAt(x-1, y)+document.HeightAt(x+1, y)+document.HeightAt(x, y-1)+document.HeightAt(x, y+1)) / 4
				delta := float64(h) - average
				if math.Abs(delta) >= options.SpikeHeight {
					code := "isolated_peak"
					if delta < 0 {
						code = "isolated_pit"
					}
					c := Command{Operation: "smooth", Parameters: Parameters{X: x, Y: y, Radius: 1, Blend: 1}}
					report.Issues = append(report.Issues, Issue{Code: code, Message: "cell differs sharply from surrounding terrain", X: x, Y: y, Value: delta, Suggestion: &c})
				}
			}
		}
	}
	return report
}
