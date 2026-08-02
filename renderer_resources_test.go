package fray

import "testing"

func TestInitWithErrorRejectsInvalidDimensions(t *testing.T) {
	r := &Renderer{}
	if err := r.InitWithError(0, 480, 16, 16, 8, 16); err == nil {
		t.Fatal("InitWithError accepted zero screen width")
	}
}

func TestReleaseGPUResourcesIsIdempotent(t *testing.T) {
	r := &Renderer{shaderSource: []byte("source"), terrainUploadBuffer: make([]byte, 16)}
	r.ReleaseGPUResources()
	r.ReleaseGPUResources()
	if r.shaderSource == nil {
		t.Fatal("ReleaseGPUResources discarded source needed for restoration")
	}
	if r.terrainUploadBuffer != nil {
		t.Fatal("ReleaseGPUResources retained upload buffer")
	}
}

func TestRestoreGPUResourcesRejectsUninitializedRenderer(t *testing.T) {
	if err := (&Renderer{}).RestoreGPUResources(); err == nil {
		t.Fatal("RestoreGPUResources accepted uninitialized renderer")
	}
}
