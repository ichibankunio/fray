package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/ichibankunio/fray/terraineditor"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		if hasArgument(os.Args[1:], "--json") {
			_ = json.NewEncoder(os.Stderr).Encode(map[string]any{"ok": false, "error": map[string]string{"code": "terrain_command_failed", "message": err.Error()}})
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "fray-terrain:", err)
		os.Exit(1)
	}
}

func hasArgument(arguments []string, wanted string) bool {
	for _, argument := range arguments {
		if argument == wanted {
			return true
		}
	}
	return false
}

func run(arguments []string) error {
	if len(arguments) == 0 {
		return fmt.Errorf("usage: fray-terrain <inspect|validate|analyze|apply|set-height|raise|lower|flatten|smooth|copy-region|move-region|flip-region> ...")
	}
	switch arguments[0] {
	case "inspect":
		return inspect(arguments[1:])
	case "validate":
		return validate(arguments[1:])
	case "analyze":
		return analyze(arguments[1:])
	case "apply":
		return applyBatch(arguments[1:])
	case "set-height", "raise", "lower", "flatten", "smooth", "copy-region", "move-region", "flip-region":
		return applyOne(arguments[0], arguments[1:])
	default:
		return fmt.Errorf("unknown command %q", arguments[0])
	}
}

func analyze(arguments []string) error {
	flags := flag.NewFlagSet("analyze", flag.ContinueOnError)
	maximumSlope := flags.Float64("max-slope", 4, "maximum adjacent height difference")
	spikeHeight := flags.Float64("spike-height", 3, "isolated peak or pit threshold")
	suggestions := flags.String("suggestions", "", "write suggested command list")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		return fmt.Errorf("analyze requires one terrain JSON path")
	}
	document, err := terraineditor.Load(flags.Arg(0))
	if err != nil {
		return err
	}
	report := terraineditor.Analyze(document, terraineditor.AnalysisOptions{MaximumSlope: *maximumSlope, SpikeHeight: *spikeHeight})
	if *suggestions != "" {
		commands := make([]terraineditor.Command, 0, len(report.Issues))
		for _, issue := range report.Issues {
			if issue.Suggestion != nil {
				commands = append(commands, *issue.Suggestion)
			}
		}
		data, _ := json.MarshalIndent(commands, "", "  ")
		if err := os.WriteFile(*suggestions, append(data, '\n'), 0o644); err != nil {
			return err
		}
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"ok": true, "result": report})
}

type inspection struct {
	Width         int      `json:"width"`
	Height        int      `json:"height"`
	Depth         int      `json:"depth"`
	MinimumHeight int      `json:"minimumHeight"`
	MaximumHeight int      `json:"maximumHeight"`
	AverageHeight float64  `json:"averageHeight"`
	MaximumSlope  float64  `json:"maximumSlope"`
	WaterLevel    *float64 `json:"waterLevel,omitempty"`
	Interpolation string   `json:"interpolation"`
}

type editResult struct {
	OK       bool                        `json:"ok"`
	DryRun   bool                        `json:"dryRun"`
	Output   string                      `json:"output,omitempty"`
	Commands int                         `json:"commands"`
	Summary  terraineditor.ChangeSummary `json:"summary"`
}

type constraintFlags struct {
	minHeight       int
	maxHeight       int
	maxSlope        float64
	protectBoundary bool
}

func addConstraintFlags(flags *flag.FlagSet, values *constraintFlags) {
	flags.IntVar(&values.minHeight, "min-height", -1, "reject resulting heights below this value")
	flags.IntVar(&values.maxHeight, "max-height", -1, "reject resulting heights above this value")
	flags.Float64Var(&values.maxSlope, "max-slope", -1, "reject resulting adjacent height differences above this value")
	flags.BoolVar(&values.protectBoundary, "protect-boundary", false, "reject edits touching the map boundary")
}

func (values constraintFlags) constraints() terraineditor.Constraints {
	result := terraineditor.Constraints{ProtectBoundary: values.protectBoundary}
	if values.minHeight >= 0 {
		result.MinHeight = &values.minHeight
	}
	if values.maxHeight >= 0 {
		result.MaxHeight = &values.maxHeight
	}
	if values.maxSlope >= 0 {
		result.MaxSlope = &values.maxSlope
	}
	return result
}

func inspect(arguments []string) error {
	flags := flag.NewFlagSet("inspect", flag.ContinueOnError)
	jsonOutput := flags.Bool("json", false, "emit machine-readable JSON")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		return fmt.Errorf("inspect requires one terrain JSON path")
	}
	document, err := terraineditor.Load(flags.Arg(0))
	if err != nil {
		return err
	}
	result := inspectDocument(document)
	if *jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(result)
	}
	fmt.Printf("size: %dx%dx%d\nheight: %d..%d (average %.2f)\nmaximum slope: %.3f\ninterpolation: %s\n", result.Width, result.Height, result.Depth, result.MinimumHeight, result.MaximumHeight, result.AverageHeight, result.MaximumSlope, result.Interpolation)
	if result.WaterLevel != nil {
		fmt.Printf("water level: %.2f\n", *result.WaterLevel)
	}
	return nil
}

func inspectDocument(document *terraineditor.Document) inspection {
	minimum, maximum, total, maximumSlope := document.CanvasDepth, 0, 0, 0.0
	for y := 0; y < document.CanvasHeight; y++ {
		for x := 0; x < document.CanvasWidth; x++ {
			height := document.HeightAt(x, y)
			minimum, maximum, total = min(minimum, height), max(maximum, height), total+height
			if x+1 < document.CanvasWidth {
				maximumSlope = max(maximumSlope, math.Abs(float64(height-document.HeightAt(x+1, y))))
			}
			if y+1 < document.CanvasHeight {
				maximumSlope = max(maximumSlope, math.Abs(float64(height-document.HeightAt(x, y+1))))
			}
		}
	}
	return inspection{Width: document.CanvasWidth, Height: document.CanvasHeight, Depth: document.CanvasDepth, MinimumHeight: minimum, MaximumHeight: maximum, AverageHeight: float64(total) / float64(document.CanvasWidth*document.CanvasHeight), MaximumSlope: maximumSlope, WaterLevel: document.Terrain.WaterLevel, Interpolation: document.Terrain.Interpolation}
}

func validate(arguments []string) error {
	if len(arguments) != 1 {
		return fmt.Errorf("validate requires one terrain JSON path")
	}
	document, err := terraineditor.Load(arguments[0])
	if err != nil {
		return err
	}
	if err := document.Validate(); err != nil {
		return err
	}
	fmt.Println("valid")
	return nil
}

func applyBatch(arguments []string) error {
	flags := flag.NewFlagSet("apply", flag.ContinueOnError)
	output := flags.String("output", "", "output path; defaults to replacing input")
	dryRun := flags.Bool("dry-run", false, "calculate changes without writing")
	jsonOutput := flags.Bool("json", false, "emit machine-readable JSON")
	var limits constraintFlags
	addConstraintFlags(flags, &limits)
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if flags.NArg() != 2 {
		return fmt.Errorf("apply requires terrain JSON and command JSON paths")
	}
	document, err := terraineditor.Load(flags.Arg(0))
	if err != nil {
		return err
	}
	data, err := os.ReadFile(flags.Arg(1))
	if err != nil {
		return err
	}
	var commands []terraineditor.Command
	if err := json.Unmarshal(data, &commands); err != nil {
		var command terraineditor.Command
		if singleErr := json.Unmarshal(data, &command); singleErr != nil {
			return fmt.Errorf("decode command JSON: %w", err)
		}
		commands = []terraineditor.Command{command}
	}
	combined := terraineditor.ChangeSet{}
	for index, command := range commands {
		changeSet, err := document.ApplyConstrained(command, limits.constraints())
		if err != nil {
			return fmt.Errorf("apply command %d: %w", index, err)
		}
		combined.Changes = append(combined.Changes, changeSet.Changes...)
		if combined.Region.Empty() {
			combined.Region = changeSet.Region
		} else if !changeSet.Region.Empty() {
			combined.Region = combined.Region.Union(changeSet.Region)
		}
	}
	result := editResult{OK: true, DryRun: *dryRun, Commands: len(commands), Summary: terraineditor.Summarize(combined)}
	if !*dryRun {
		if err := saveDocument(document, flags.Arg(0), *output); err != nil {
			return err
		}
		result.Output = *output
		if result.Output == "" {
			result.Output = flags.Arg(0)
		}
	}
	if *jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(result)
	}
	fmt.Printf("changed %d cells in %v%s\n", result.Summary.ChangedCells, result.Summary.Region, map[bool]string{true: " (dry run)"}[*dryRun])
	return nil
}

func applyOne(operation string, arguments []string) error {
	flags := flag.NewFlagSet(operation, flag.ContinueOnError)
	x := flags.Int("x", 0, "x coordinate")
	y := flags.Int("y", 0, "y coordinate")
	width := flags.Int("width", 0, "rectangle width")
	rows := flags.Int("rows", 0, "rectangle height")
	radius := flags.Int("radius", 0, "brush radius")
	height := flags.Int("height", 0, "target height")
	amount := flags.Int("amount", 1, "height delta")
	falloff := flags.String("falloff", "", "brush falloff: smooth or empty")
	blend := flags.Float64("blend", .5, "smoothing blend 0..1")
	toX := flags.Int("to-x", 0, "destination x coordinate")
	toY := flags.Int("to-y", 0, "destination y coordinate")
	axis := flags.String("axis", "horizontal", "flip axis: horizontal or vertical")
	shape := flags.String("shape", "circle", "brush shape: circle, ellipse, square, diamond, or noise")
	radiusX := flags.Int("radius-x", 0, "horizontal brush radius")
	radiusY := flags.Int("radius-y", 0, "vertical brush radius")
	seed := flags.Int("seed", 0, "deterministic noise brush seed")
	output := flags.String("output", "", "output path; defaults to replacing input")
	dryRun := flags.Bool("dry-run", false, "calculate changes without writing")
	jsonOutput := flags.Bool("json", false, "emit machine-readable JSON")
	var limits constraintFlags
	addConstraintFlags(flags, &limits)
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		return fmt.Errorf("%s requires one terrain JSON path", operation)
	}
	document, err := terraineditor.Load(flags.Arg(0))
	if err != nil {
		return err
	}
	command := terraineditor.Command{Operation: operation, Parameters: terraineditor.Parameters{X: *x, Y: *y, Width: *width, Rows: *rows, Radius: *radius, Height: *height, Amount: *amount, Falloff: *falloff, Blend: *blend, ToX: *toX, ToY: *toY, Axis: *axis, Shape: *shape, RadiusX: *radiusX, RadiusY: *radiusY, Seed: *seed}}
	changeSet, err := document.ApplyConstrained(command, limits.constraints())
	if err != nil {
		return err
	}
	result := editResult{OK: true, DryRun: *dryRun, Commands: 1, Summary: terraineditor.Summarize(changeSet)}
	if !*dryRun {
		if err := saveDocument(document, flags.Arg(0), *output); err != nil {
			return err
		}
		result.Output = *output
		if result.Output == "" {
			result.Output = flags.Arg(0)
		}
	}
	if *jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(result)
	}
	fmt.Printf("changed %d cells in %v%s\n", len(changeSet.Changes), changeSet.Region, map[bool]string{true: " (dry run)"}[*dryRun])
	return nil
}

func saveDocument(document *terraineditor.Document, input, output string) error {
	if output == "" {
		output = input
	}
	return document.Save(output)
}
