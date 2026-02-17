package themes

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func makeTestTheme() *Theme {
	return &Theme{
		UserIO: &UserIOTheme{
			Error:         tcell.NewHexColor(0xFF0000),
			Success:       tcell.NewHexColor(0x00FF00),
			Warning:       tcell.NewHexColor(0xFFFF00),
			Info:          tcell.NewHexColor(0x0000FF),
			CommandPrompt: tcell.NewHexColor(0xCCCCCC),
		},
	}
}

func TestFormatError(t *testing.T) {
	th := makeTestTheme()
	out := th.FormatError("Error: %s", "fail")
	assert.Equal(t, "[#ff0000]Error: fail[-]", out)
}

func TestFormatSuccess(t *testing.T) {
	th := makeTestTheme()
	out := th.FormatSuccess("Success: %d", 42)
	assert.Equal(t, "[#00ff00]Success: 42[-]", out)
}

func TestFormatWarning(t *testing.T) {
	th := makeTestTheme()
	out := th.FormatWarning("Warning: %v", true)
	assert.Equal(t, "[#ffff00]Warning: true[-]", out)
}

func TestFormatInfo(t *testing.T) {
	th := makeTestTheme()
	out := th.FormatInfo("Info: %.2f", 3.14)
	assert.Equal(t, "[#0000ff]Info: 3.14[-]", out)
}

func TestFormatPrompt(t *testing.T) {
	th := makeTestTheme()
	out := th.FormatPrompt("Prompt: %s", "cmd")
	assert.Equal(t, "[#cccccc]Prompt: cmd[-]", out)
}
