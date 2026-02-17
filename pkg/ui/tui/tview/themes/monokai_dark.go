package themes

import (
	"github.com/gdamore/tcell/v2"
	tvlib "github.com/rivo/tview"
)

// MonokaiDark is the monokai-dark theme instance
var MonokaiDark = &Theme{
	Name:        "Monokai Dark",
	Description: "Dark monokai theme with vibrant colors",
	UserIO: &UserIOTheme{
		CommandPrompt: tcell.GetColor("#AE81FF"),
		Info:          tcell.GetColor("#66D9EF"),
		Error:         tcell.GetColor("#F92672"),
		Success:       tcell.GetColor("#A6E22E"),
		Warning:       tcell.GetColor("#E6DB74"),
	},
	SourceSnippet: &SourceSnippetTheme{
		LineNumber:       tcell.GetColor("#E6DB74"),
		CurrentLine:      tcell.GetColor("#49483E"),
		BreakpointMarker: tcell.GetColor("#F92672"),
		C: &CSyntaxTheme{
			Keyword:      tcell.GetColor("#F92672"),
			Type:         tcell.GetColor("#66D9EF"),
			Function:     tcell.GetColor("#A6E22E"),
			String:       tcell.GetColor("#E6DB74"),
			Comment:      tcell.GetColor("#75715E"),
			Preprocessor: tcell.GetColor("#AE81FF"),
			Number:       tcell.GetColor("#AE81FF"),
			Operator:     tcell.GetColor("#F92672"),
		},
	},
	MemoryDump: &MemoryDumpTheme{
		Address:          tcell.GetColor("#E6DB74"),
		HexDump:          tcell.GetColor("#A1EFE4"),
		ASCII:            tcell.GetColor("#A6E22E"),
		HighlightBG:      tcell.GetColor("#49483E"),
		WatchpointMarker: tcell.GetColor("#F92672"),
	},
	Disassembly: &DisassemblyTheme{
		Address:                     tcell.GetColor("#E6DB74"),
		RegisterOperand:             tcell.GetColor("#A6E22E"),
		ImmediateOperand:            tcell.GetColor("#AE81FF"),
		Mnemonic:                    tcell.GetColor("#66D9EF"),
		HighlightBG:                 tcell.GetColor("#49483E"),
		BreakpointMarker:            tcell.GetColor("#F92672"),
		CallGraphColors:             []tcell.Color{tcell.GetColor("#AE81FF"), tcell.GetColor("#FD971F"), tcell.GetColor("#A1EFE4")},
		DataDependenciesGraphColors: []tcell.Color{tcell.GetColor("#F92672"), tcell.GetColor("#E6DB74"), tcell.GetColor("#A6E22E")},
	},
	Registers: &RegistersTheme{
		RegisterName:              tcell.GetColor("#A6E22E"),
		RegisterValue_Decimal:     tcell.GetColor("#F8F8F2"),
		RegisterValue_Hexadecimal: tcell.GetColor("#AE81FF"),
		RegisterValue_Binary:      tcell.GetColor("#66D9EF"),
	},
	Events: &EventsTheme{
		Timestamp:             tcell.GetColor("#A1EFE4"),
		ProgramLoaded:         tcell.GetColor("#A6E22E"),
		Stepped:               tcell.GetColor("#E6DB74"),
		BreakpointHit:         tcell.GetColor("#F92672"),
		WatchpointHit:         tcell.GetColor("#F92672"),
		ProgramTerminated:     tcell.GetColor("#A6E22E"),
		ProgramHalted:         tcell.GetColor("#F92672"),
		Error:                 tcell.GetColor("#F92672"),
		SourceLocationChanged: tcell.GetColor("#A1EFE4"),
		Interrupted:           tcell.GetColor("#FD971F"),
		Lagging:               tcell.GetColor("#FD971F"),
		EventDetail:           tcell.GetColor("#E6DB74"),
	},
	Theme: &tvlib.Theme{
		PrimitiveBackgroundColor:    tcell.GetColor("#272822"),
		ContrastBackgroundColor:     tcell.GetColor("#272822"),
		MoreContrastBackgroundColor: tcell.GetColor("#1E1F1C"),
		BorderColor:                 tcell.GetColor("#75715E"),
		TitleColor:                  tcell.GetColor("#F8F8F2"),
		GraphicsColor:               tcell.GetColor("#75715E"),
		PrimaryTextColor:            tcell.GetColor("#F8F8F2"),
		SecondaryTextColor:          tcell.GetColor("#E6DB74"),
		TertiaryTextColor:           tcell.GetColor("#A6E22E"),
		InverseTextColor:            tcell.GetColor("#272822"),
		ContrastSecondaryTextColor:  tcell.GetColor("#F92672"),
	},
}
