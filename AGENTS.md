# AGENTS.md

## Terrain Direction

These constraints are project invariants for every terrain-related change in this repository:

1. **Infer terrain from a heightmap.** The heightmap is the authoritative geometric input. Slopes, corners, ridges, valleys, normals, collision, and other surface properties must be inferred consistently from neighboring height samples. Do not require authored 3D models or explicit per-slope geometry.
2. **Keep terrain definitions JSON-based.** Games must be able to define terrain and its configuration with JSON plus referenced raster assets. Do not introduce a required external 3D modeling, mesh-authoring, or proprietary editor workflow.
3. **Maintain 60 FPS with VSync on.** fray targets smartphones. New rendering features must sustain the project's physical-60-Hz performance gate (59.0 measured FPS or better) in representative scenes. If a feature misses the gate, optimize it; if that is not sufficient, make it optional or remove it rather than silently lowering the established visual quality, draw distance, interpolation quality, or logical resolution.

## Implementation Rules

- Keep CPU collision/navigation and GPU rendering on the same terrain interpolation model. The current standard is monotonic cubic Hermite interpolation unless a deliberate repository-wide migration is approved.
- Prefer implementation and data-layout optimizations that preserve rendered output over quality reductions or magic-number tuning.
- Treat explicit JSON slope-angle, ridge, valley, or corner geometry as out of scope when those properties can be inferred from the heightmap.
- Add or update representative visual and performance checks for terrain rendering changes. Measure with VSync enabled and include demanding views, not only the easiest camera position.
- Stop and request a product decision before accepting a performance regression or weakening any invariant above.
