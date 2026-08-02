# Terrain JSON Compatibility

`TerrainJSONVersionCurrent` is the newest supported schema. fray accepts versions
`TerrainJSONVersionLegacy` through the current version and rejects future versions
instead of guessing their meaning.

| Version | Compatibility |
| --- | --- |
| 0 | Legacy documents without an explicit version; missing interpolation means `linear`. |
| 1-2 | Existing layered world documents; interpreted with legacy `linear` defaults. |
| 3 | Current schema; supports `monotonic` interpolation and `waterLevel`. |

Terrain shape remains authoritative in `layers`; height, slopes, normals, collision,
and visibility are inferred from it. New optional fields must preserve old defaults.
Removing or changing a field requires a new version and an explicit migration path.

Use `ValidateTerrainJSON` before loading user-authored data. `LoadTerrainJSON` uses
strict field checking, rejects multiple concatenated documents, and never accepts a
schema newer than the library understands.
