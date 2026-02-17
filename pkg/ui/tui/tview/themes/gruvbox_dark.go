package themes

import (
	"github.com/gdamore/tcell/v2"
	tvlib "github.com/rivo/tview"
)

// GruvboxDark is the gruvbox-dark theme instance
var GruvboxDark = &Theme{
	Name:        "Gruvbox Dark",
	Description: "Retro groove color scheme",
	UserIO: &UserIOTheme{
		CommandPrompt: tcell.GetColor("#D3869B"),
		Info:          tcell.GetColor("#83A598"),
		Error:         tcell.GetColor("#FB4934"),
		Success:       tcell.GetColor("#B8BB26"),
		Warning:       tcell.GetColor("#FABD2F"),
	},
	SourceSnippet: &SourceSnippetTheme{
		LineNumber:       tcell.GetColor("#D79921"),
		CurrentLine:      tcell.GetColor("#3C3836"),
		BreakpointMarker: tcell.GetColor("#FB4934"),
		C: &CSyntaxTheme{
			Keyword:      tcell.GetColor("#FB4934"),
			Type:         tcell.GetColor("#83A598"),
			Function:     tcell.GetColor("#B8BB26"),
			String:       tcell.GetColor("#B8BB26"),
			Comment:      tcell.GetColor("#928374"),
			Preprocessor: tcell.GetColor("#D3869B"),
			Number:       tcell.GetColor("#D3869B"),
			Operator:     tcell.GetColor("#FB4934"),
		},
	},
	MemoryDump: &MemoryDumpTheme{
		Address:          tcell.GetColor("#FABD2F"),
		HexDump:          tcell.GetColor("#83A598"),
		ASCII:            tcell.GetColor("#B8BB26"),
		HighlightBG:      tcell.GetColor("#3C3836"),
		WatchpointMarker: tcell.GetColor("#FB4934"),
	},
	Disassembly: &DisassemblyTheme{
		Address:                     tcell.GetColor("#FABD2F"),
		RegisterOperand:             tcell.GetColor("#B8BB26"),
		ImmediateOperand:            tcell.GetColor("#D3869B"),
		Mnemonic:                    tcell.GetColor("#83A598"),
		HighlightBG:                 tcell.GetColor("#3C3836"),
		BreakpointMarker:            tcell.GetColor("#FB4934"),
		CallGraphColors:             []tcell.Color{tcell.GetColor("#D3869B"), tcell.GetColor("#FE8019"), tcell.GetColor("#83A598")},
		DataDependenciesGraphColors: []tcell.Color{tcell.GetColor("#FB4934"), tcell.GetColor("#FABD2F"), tcell.GetColor("#B8BB26")},
	},
	Registers: &RegistersTheme{
		RegisterName:              tcell.GetColor("#B8BB26"),
		RegisterValue_Decimal:     tcell.GetColor("#EBDBB2"),
		RegisterValue_Hexadecimal: tcell.GetColor("#D3869B"),
		RegisterValue_Binary:      tcell.GetColor("#83A598"),
	}, Events: &EventsTheme{
		Timestamp:             tcell.GetColor("#8ec07c"),
		ProgramLoaded:         tcell.GetColor("#b8bb26"),
		Stepped:               tcell.GetColor("#fabd2f"),
		BreakpointHit:         tcell.GetColor("#fb4934"),
		WatchpointHit:         tcell.GetColor("#fb4934"),
		ProgramTerminated:     tcell.GetColor("#b8bb26"),
		ProgramHalted:         tcell.GetColor("#fb4934"),
		Error:                 tcell.GetColor("#fb4934"),
		SourceLocationChanged: tcell.GetColor("#8ec07c"),
		Interrupted:           tcell.GetColor("#fe8019"),
		Lagging:               tcell.GetColor("#fe8019"),
		EventDetail:           tcell.GetColor("#fabd2f"),
	}, Theme: &tvlib.Theme{
		PrimitiveBackgroundColor:    tcell.GetColor("#282828"),
		ContrastBackgroundColor:     tcell.GetColor("#282828"),
		MoreContrastBackgroundColor: tcell.GetColor("#1D2021"),
		BorderColor:                 tcell.GetColor("#3C3836"),
		TitleColor:                  tcell.GetColor("#EBDBB2"),
		GraphicsColor:               tcell.GetColor("#3C3836"),
		PrimaryTextColor:            tcell.GetColor("#EBDBB2"),
		SecondaryTextColor:          tcell.GetColor("#FABD2F"),
		TertiaryTextColor:           tcell.GetColor("#B8BB26"),
		InverseTextColor:            tcell.GetColor("#282828"),
		ContrastSecondaryTextColor:  tcell.GetColor("#FB4934"),
	},
}
