package widgets

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/ui"
	"github.com/Manu343726/cucaracha/pkg/ui/tui/tview/themes"
	tvlib "github.com/rivo/tview"
)

// SourceSnippet is a widget for displaying current source code snippet
type SourceSnippet struct {
	*tvlib.TextView
	result   *ui.SourceResult
	maxLines int
}

// NewSourceSnippet creates a new SourceSnippet widget
func NewSourceSnippet(result *ui.SourceResult, maxLines int) *SourceSnippet {
	if result == nil {
		result = &ui.SourceResult{}
	}
	if maxLines <= 0 {
		maxLines = 10
	}
	tv := tvlib.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle("Source Snippet")
	ss := &SourceSnippet{
		TextView: tv,
		result:   result,
		maxLines: maxLines,
	}
	ss.refresh()
	return ss
}

// SetResult sets the result for this widget
func (ss *SourceSnippet) SetResult(result *ui.SourceResult) {
	ss.result = result
	ss.refresh()
}

// refresh refreshes the display
func (ss *SourceSnippet) refresh() {
	if ss.result == nil || ss.result.Error != nil || ss.result.Snippet == nil || len(ss.result.Snippet.Lines) == 0 {
		ss.TextView.SetText("[#F92672]No source available[#F8F8F2]")
		return
	}

	var text string
	linesToShow := ss.maxLines
	if len(ss.result.Snippet.Lines) < linesToShow {
		linesToShow = len(ss.result.Snippet.Lines)
	}

	for i := 0; i < linesToShow; i++ {
		line := ss.result.Snippet.Lines[i]
		if line.IsCurrent {
			// Monokai red for current line marker
			text += fmt.Sprintf("[#F92672]>>>[#F8F8F2] %s\n", line.Text)
		} else {
			// Monokai default for regular lines
			text += fmt.Sprintf("    %s\n", line.Text)
		}
	}
	ss.TextView.SetText(text)
}

// SourceFile is a widget for displaying full source code file
type SourceFile struct {
	*tvlib.TextView
	result *ui.SourceResult
}

// NewSourceFile creates a new SourceFile widget
func NewSourceFile(result *ui.SourceResult) *SourceFile {
	tv := tvlib.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle("Source File")
	sf := &SourceFile{
		TextView: tv,
		result:   result,
	}
	sf.refresh()
	return sf
}

// SetResult sets the result for this widget
func (sf *SourceFile) SetResult(result *ui.SourceResult) {
	sf.result = result
	sf.refresh()
}

// refresh refreshes the display
func (sf *SourceFile) refresh() {
	if sf.result == nil || sf.result.Error != nil || sf.result.Snippet == nil || len(sf.result.Snippet.Lines) == 0 {
		sf.TextView.SetText("[#F92672]No source file available[#F8F8F2]")
		return
	}

	var text string
	// Show all lines from the snippet
	for i, line := range sf.result.Snippet.Lines {
		// Calculate line number from source range if available
		lineNum := i + 1
		if sf.result.Snippet.SourceRange != nil && sf.result.Snippet.SourceRange.Start != nil {
			lineNum = sf.result.Snippet.SourceRange.Start.Line + i
		}
		// Monokai: yellow for line numbers, default for source code
		text += fmt.Sprintf("[#E6DB74]%4d[#F8F8F2]: %s\n", lineNum, line.Text)
	}
	sf.TextView.SetText(text)
}

// SetTheme applies the theme to the SourceSnippet widget
func (ss *SourceSnippet) SetTheme(theme *themes.Theme) *SourceSnippet {
	ss.TextView.SetBackgroundColor(theme.PrimitiveBackgroundColor)
	ss.TextView.SetTextColor(theme.PrimaryTextColor)
	ss.TextView.SetBorderColor(theme.BorderColor)
	return ss
}

// SetTheme applies the theme to the SourceFile widget
func (sf *SourceFile) SetTheme(theme *themes.Theme) *SourceFile {
	sf.TextView.SetBackgroundColor(theme.PrimitiveBackgroundColor)
	sf.TextView.SetTextColor(theme.PrimaryTextColor)
	sf.TextView.SetBorderColor(theme.BorderColor)
	return sf
}
