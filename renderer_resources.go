package fray

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// ReleaseGPUResources drops renderer-owned GPU references. Game-owned textures
// are retained and can be reused when restoring the renderer.
func (r *Renderer) ReleaseGPUResources() {
	r.shader = nil
	r.shader2 = nil
	r.pseudoShadowShader = nil
	r.worldBuffer = nil
	r.terrainUploadBuffer = nil
}

// RestoreGPUResources recompiles renderer-owned shaders and buffers after a
// context or lifecycle reset. It must run on the Ebitengine main thread.
func (r *Renderer) RestoreGPUResources() error {
	if r.screenWidth <= 0 || r.screenHeight <= 0 {
		return fmt.Errorf("restore renderer resources: renderer is not initialized")
	}
	if len(r.shaderSource) == 0 {
		return fmt.Errorf("restore renderer resources: custom shader source is unavailable")
	}
	shader, err := ebiten.NewShader(r.shaderSource)
	if err != nil {
		return fmt.Errorf("restore renderer shader: %w", err)
	}
	shader2, err := ebiten.NewShader(shaderByte2)
	if err != nil {
		return fmt.Errorf("restore textureless renderer shader: %w", err)
	}
	pseudoShadow, err := ebiten.NewShader(pseudoShadowShaderByte)
	if err != nil {
		return fmt.Errorf("restore pseudo-shadow shader: %w", err)
	}
	r.shader = shader
	r.shader2 = shader2
	r.pseudoShadowShader = pseudoShadow
	r.worldBuffer = ebiten.NewImage(int(r.screenWidth), int(r.screenHeight))
	return nil
}
