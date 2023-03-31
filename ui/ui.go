package ui

import (
	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

type UIManager struct {
	ScreenWidth  int
	ScreenHeight int

	DialogFont font.Face
}

//go:embed shaders/dialogBox.kage
var shaderByteDialog []byte

var shader *ebiten.Shader

func init() {
	var err error
	shader, err = ebiten.NewShader(shaderByteDialog)
	if err != nil {
		panic(err)
	}
}

func NewUIManager(width, height int, dialogFont font.Face) *UIManager {
	return &UIManager{
		ScreenWidth: width,
		ScreenHeight: height,
		DialogFont: dialogFont,
	}
}

func (uim UIManager) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawRectShaderOptions{}

	op.Uniforms = map[string]interface{}{
		"Resolution": []float32{float32(uim.ScreenWidth), float32(uim.ScreenHeight)},
	}

	screen.DrawRectShader(uim.ScreenWidth, uim.ScreenHeight, shader, op)

	uim.drawTextTest(screen)
}

func (uim UIManager) drawTextTest(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(200, 300)
	text.DrawWithOptions(screen, "Hello", uim.DialogFont, op)
}
