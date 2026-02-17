package themes

import (
	"github.com/gdamore/tcell/v2"
	tvlib "github.com/rivo/tview"
)

// Dracula is the dracula theme instance
var Dracula = &Theme{
	Name:        "Dracula",
	Description: "Dark dracula theme with muted colors",
	UserIO: &UserIOTheme{
		CommandPrompt: tcell.GetColor("#BD93F9"),
		Info:          tcell.GetColor("#8BE9FD"),
		Error:         tcell.GetColor("#FF5555"),
		Success:       tcell.GetColor("#50FA7B"),
		Warning:       tcell.GetColor("#FFB86C"),
	},
	SourceSnippet: &SourceSnippetTheme{
		LineNumber:       tcell.GetColor("#FFB86C"),
		CurrentLine:      tcell.GetColor("#44475A"),
		BreakpointMarker: tcell.GetColor("#FF5555"),
		C: &CSyntaxTheme{
			Keyword:      tcell.GetColor("#FF79C6"),
			Type:         tcell.GetColor("#8BE9FD"),
			Function:     tcell.GetColor("#50FA7B"),
			String:       tcell.GetColor("#F1FA8C"),
			Comment:      tcell.GetColor("#6272A4"),
			Preprocessor: tcell.GetColor("#BD93F9"),
			Number:       tcell.GetColor("#BD93F9"),
			Operator:     tcell.GetColor("#FF79C6"),
		},
	},
	MemoryDump: &MemoryDumpTheme{
		Address:          tcell.GetColor("#FFB86C"),
		HexDump:          tcell.GetColor("#8BE9FD"),
		ASCII:            tcell.GetColor("#50FA7B"),
		HighlightBG:      tcell.GetColor("#44475A"),
		WatchpointMarker: tcell.GetColor("#FF5555"),
	},
	Disassembly: &DisassemblyTheme{
		Address:                     tcell.GetColor("#FFB86C"),
		RegisterOperand:             tcell.GetColor("#50FA7B"),
		ImmediateOperand:            tcell.GetColor("#BD93F9"),
		Mnemonic:                    tcell.GetColor("#8BE9FD"),
		HighlightBG:                 tcell.GetColor("#44475A"),
		BreakpointMarker:            tcell.GetColor("#FF5555"),
		CallGraphColors:             []tcell.Color{tcell.GetColor("#BD93F9"), tcell.GetColor("#FF79C6"), tcell.GetColor("#8BE9FD")},
		DataDependenciesGraphColors: []tcell.Color{tcell.GetColor("#FF5555"), tcell.GetColor("#FFB86C"), tcell.GetColor("#50FA7B")},
	},
	Registers: &RegistersTheme{
		RegisterName:              tcell.GetColor("#50FA7B"),
		RegisterValue_Decimal:     tcell.GetColor("#F8F8F2"),
		RegisterValue_Hexadecimal: tcell.GetColor("#BD93F9"),
		RegisterValue_Binary:      tcell.GetColor("#8BE9FD"),
	},
	Events: &EventsTheme{
		Timestamp:             tcell.GetColor("#8BE9FD"),
		ProgramLoaded:         tcell.GetColor("#50FA7B"),
		Stepped:               tcell.GetColor("#F1FA8C"),
		BreakpointHit:         tcell.GetColor("#FF79C6"),
		WatchpointHit:         tcell.GetColor("#FF79C6"),
		ProgramTerminated:     tcell.GetColor("#50FA7B"),
		ProgramHalted:         tcell.GetColor("#FF79C6"),
		Error:                 tcell.GetColor("#FF79C6"),
		SourceLocationChanged: tcell.GetColor("#8BE9FD"),
		Interrupted:           tcell.GetColor("#FFB86C"),
		Lagging:               tcell.GetColor("#FFB86C"),
		EventDetail:           tcell.GetColor("#F1FA8C"),
	},
	Theme: &tvlib.Theme{
		PrimitiveBackgroundColor:    tcell.GetColor("#282A36"),
		ContrastBackgroundColor:     tcell.GetColor("#282A36"),
		MoreContrastBackgroundColor: tcell.GetColor("#21222C"),
		BorderColor:                 tcell.GetColor("#44475A"),
		TitleColor:                  tcell.GetColor("#F8F8F2"),
		GraphicsColor:               tcell.GetColor("#44475A"),
		PrimaryTextColor:            tcell.GetColor("#F8F8F2"),
		SecondaryTextColor:          tcell.GetColor("#FFB86C"),
		TertiaryTextColor:           tcell.GetColor("#50FA7B"),
		InverseTextColor:            tcell.GetColor("#282A36"),
		ContrastSecondaryTextColor:  tcell.GetColor("#FF5555"),
	},
}
