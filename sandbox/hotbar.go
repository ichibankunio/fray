package sandbox

type SandboxHotbar struct {
	Selected int
}

func (hb *SandboxHotbar) Init() {
	hb.Selected = 0
}
