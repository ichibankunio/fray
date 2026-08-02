# Terrain Editor

fray provides one editing model for headless automation and an optional
Ebitengine panel. Both paths modify only layered heightmap source data; slopes,
normals, collision, visibility, and rendering remain inferred by fray.

## CLI

Build or run the CLI without starting Ebitengine:

```sh
go run ./cmd/fray-terrain inspect --json game/map/worldmap.json
go run ./cmd/fray-terrain validate game/map/worldmap.json
go run ./cmd/fray-terrain raise --x 32 --y 48 --radius 6 --amount 2 --falloff smooth game/map/worldmap.json
go run ./cmd/fray-terrain smooth --x 32 --y 48 --radius 8 --blend 0.5 game/map/worldmap.json
go run ./cmd/fray-terrain flatten --x 10 --y 10 --width 20 --rows 20 --height 4 game/map/worldmap.json
go run ./cmd/fray-terrain copy-region --x 10 --y 10 --width 8 --rows 8 --to-x 32 --to-y 24 game/map/worldmap.json
go run ./cmd/fray-terrain flip-region --x 10 --y 10 --width 8 --rows 8 --axis horizontal game/map/worldmap.json
go run ./cmd/fray-terrain raise --dry-run --json --x 32 --y 48 --radius 6 game/map/worldmap.json
go run ./cmd/fray-terrain analyze --max-slope 4 --suggestions /tmp/fixes.json game/map/worldmap.json
```

`--output path` writes a separate document. Without it, saving uses an atomic
replacement of the input file.

Use `--dry-run` to calculate the exact affected region and height deltas without
writing the file. `--json` returns a stable machine-readable result. Automated
edits can be guarded with `--min-height`, `--max-height`, `--max-slope`, and
`--protect-boundary`; rejected commands are rolled back before returning.
`analyze` reports steep edges and isolated peaks or pits with coordinates and
deterministic smoothing suggestions. Suggestions are never applied implicitly.

Codex can apply a deterministic command list:

```sh
go run ./cmd/fray-terrain apply --output /tmp/candidate.json game/map/worldmap.json commands.json
```

```json
[
  {
    "operation": "raise",
    "parameters": {
      "x": 32,
      "y": 48,
      "radius": 6,
      "amount": 2,
      "falloff": "smooth"
    }
  },
  {
    "operation": "smooth",
    "parameters": {
      "x": 32,
      "y": 48,
      "radius": 8,
      "blend": 0.5
    }
  }
]
```

Supported operations are `set-height`, `raise`, `lower`, `flatten`, and
`smooth`, `copy-region`, `move-region`, and `flip-region`. Region operations
preserve material columns and are safe when source and destination overlap.
Commands return a `ChangeSet`; GUI and automation therefore share the
same undo unit and affected rectangle.

## Live Ebitengine Panel

Launch the standalone game-preview editor:

```sh
go run ./cmd/fray-terrain-gui ../mobile3dtest-slope/game/map/worldmap.json
```

The left side is the live fray terrain renderer and the right side is the editor.
Use WASD and mouse look in the game area, edit in the panel, and press `Ctrl+S`
to atomically save the JSON. VSync remains enabled.
The top-left diagnostics separate edit/rebuild and GPU upload time. A sustained
drop below 59 FPS emits a performance warning after startup rather than silently
accepting a feature that violates fray's 60 FPS requirement.
Saving also writes `<terrain>.history.json`, a deterministic command list that
can be reviewed or replayed with `fray-terrain apply`.

The GUI packages are separate from the headless `terraineditor` package, so CLI
usage does not require a window.

```go
layout := terraineditorui.Split(screen.Bounds(), 288)
session, err := terraineditorruntime.New(document, renderer)
if err != nil {
    return err
}
panel := terraineditorui.NewPanel(session, layout.Editor)
_ = panel.Watch("game/map/worldmap.json") // optional CLI live reload
```

In `Update`, call `panel.Update()`. In `Draw`, render the game into
`screen.SubImage(layout.Game)` and then call `panel.Draw(screen)`. The panel
updates `WorldMap`, rebuilds only the affected terrain region, and uploads the
merged GPU dirty rectangle in the same frame.

The optional watcher hashes file contents rather than relying on timestamp
resolution. Valid external CLI edits reload automatically. If GUI undo history
contains unsaved edits, the panel reports a conflict instead of overwriting them.

Controls:

- `1`-`5`: set height, raise, lower, flatten, smooth
- Shift + drag: select a rectangular region
- Enter: apply the current tool to the selection
- Alt + arrow: move the selection by one cell
- Ctrl + arrow: copy the selection by one cell
- `H` / `V`: flip the selection horizontally or vertically
- `C`: clear the selection
- `R`: compare the last saved heightmap with current edits
- `B`: cycle circle, square, diamond, and deterministic noise brushes
- `X`: show a row cross-section with cells and the monotonic Hermite surface
- Left mouse: apply selected tool
- Right mouse: lower
- `[` / `]`: brush radius
- `-` / `=`: amount
- Up / Down: target height
- `Z` / `Y`: undo / redo
- `Esc`: exit the GUI editor

Moving the pointer over the heightmap previews the exact command footprint in
green (raise) or red (lower) before applying it. The equivalent compact command
JSON is displayed below the map so the same operation can be reproduced by CLI.

The game decides whether to include these optional editor packages. Release
builds that do not import them carry no editor update or rendering cost.
