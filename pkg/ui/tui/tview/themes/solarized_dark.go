package themes

import (
	"github.com/gdamore/tcell/v2"
	tvlib "github.com/rivo/tview"
)

// SolarizedDark is the solarized-dark theme instance
var SolarizedDark = &Theme{
	Name:        "Solarized Dark",
	Description: "Precision colors for machines and people",
	UserIO: &UserIOTheme{
		CommandPrompt: tcell.GetColor("#6C71C4"),
		Info:          tcell.GetColor("#268BD2"),
		Error:         tcell.GetColor("#DC322F"),
		Success:       tcell.GetColor("#859900"),
		Warning:       tcell.GetColor("#B58900"),
	},
	SourceSnippet: &SourceSnippetTheme{
		LineNumber:       tcell.GetColor("#B58900"),
		CurrentLine:      tcell.GetColor("#073642"),
		BreakpointMarker: tcell.GetColor("#DC322F"),
		C: &CSyntaxTheme{
			Keyword:      tcell.GetColor("#859900"),
			Type:         tcell.GetColor("#268BD2"),
			Function:     tcell.GetColor("#859900"),
			String:       tcell.GetColor("#2AA198"),
			Comment:      tcell.GetColor("#586E75"),
			Preprocessor: tcell.GetColor("#6C71C4"),
			Number:       tcell.GetColor("#6C71C4"),
			Operator:     tcell.GetColor("#859900"),
		},
	},
	MemoryDump: &MemoryDumpTheme{
		Address:          tcell.GetColor("#B58900"),
		HexDump:          tcell.GetColor("#2AA198"),
		ASCII:            tcell.GetColor("#859900"),
		HighlightBG:      tcell.GetColor("#073642"),
		WatchpointMarker: tcell.GetColor("#DC322F"),
	},
	Disassembly: &DisassemblyTheme{
		Address:                     tcell.GetColor("#B58900"),
		RegisterOperand:             tcell.GetColor("#859900"),
		ImmediateOperand:            tcell.GetColor("#6C71C4"),
		Mnemonic:                    tcell.GetColor("#268BD2"),
		HighlightBG:                 tcell.GetColor("#073642"),
		BreakpointMarker:            tcell.GetColor("#DC322F"),
		CallGraphColors:             []tcell.Color{tcell.GetColor("#6C71C4"), tcell.GetColor("#D33682"), tcell.GetColor("#268BD2")},
		DataDependenciesGraphColors: []tcell.Color{tcell.GetColor("#DC322F"), tcell.GetColor("#B58900"), tcell.GetColor("#859900")},
	},
	Registers: &RegistersTheme{
		RegisterName:              tcell.GetColor("#859900"),
		RegisterValue_Decimal:     tcell.GetColor("#839496"),
		RegisterValue_Hexadecimal: tcell.GetColor("#6C71C4"),
		RegisterValue_Binary:      tcell.GetColor("#268BD2"),
	}, Events: &EventsTheme{
		Timestamp:             tcell.GetColor("#2aa198"),
		ProgramLoaded:         tcell.GetColor("#859900"),
		Stepped:               tcell.GetColor("#b58900"),
		BreakpointHit:         tcell.GetColor("#dc322f"),
		WatchpointHit:         tcell.GetColor("#dc322f"),
		ProgramTerminated:     tcell.GetColor("#859900"),
		ProgramHalted:         tcell.GetColor("#dc322f"),
		Error:                 tcell.GetColor("#dc322f"),
		SourceLocationChanged: tcell.GetColor("#2aa198"),
		Interrupted:           tcell.GetColor("#cb4b16"),
		Lagging:               tcell.GetColor("#cb4b16"),
		EventDetail:           tcell.GetColor("#b58900"),
	}, Theme: &tvlib.Theme{
		PrimitiveBackgroundColor:    tcell.GetColor("#002B36"),
		ContrastBackgroundColor:     tcell.GetColor("#002B36"),
		MoreContrastBackgroundColor: tcell.GetColor("#001B26"),
		BorderColor:                 tcell.GetColor("#073642"),
		TitleColor:                  tcell.GetColor("#839496"),
		GraphicsColor:               tcell.GetColor("#073642"),
		PrimaryTextColor:            tcell.GetColor("#839496"),
		SecondaryTextColor:          tcell.GetColor("#B58900"),
		TertiaryTextColor:           tcell.GetColor("#859900"),
		InverseTextColor:            tcell.GetColor("#002B36"),
		ContrastSecondaryTextColor:  tcell.GetColor("#DC322F"),
	},
}
