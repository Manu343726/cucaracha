package themes

import (
	"github.com/gdamore/tcell/v2"
	tvlib "github.com/rivo/tview"
)

// Nord is the nord theme instance
var Nord = &Theme{
	Name:        "Nord",
	Description: "Arctic, north-bluish theme",
	UserIO: &UserIOTheme{
		CommandPrompt: tcell.GetColor("#B48EAD"),
		Info:          tcell.GetColor("#88C0D0"),
		Error:         tcell.GetColor("#BF616A"),
		Success:       tcell.GetColor("#A3BE8C"),
		Warning:       tcell.GetColor("#EBCB8B"),
	},
	SourceSnippet: &SourceSnippetTheme{
		LineNumber:       tcell.GetColor("#EBCB8B"),
		CurrentLine:      tcell.GetColor("#3B4252"),
		BreakpointMarker: tcell.GetColor("#BF616A"),
		C: &CSyntaxTheme{
			Keyword:      tcell.GetColor("#81A1C1"),
			Type:         tcell.GetColor("#88C0D0"),
			Function:     tcell.GetColor("#A3BE8C"),
			String:       tcell.GetColor("#A3BE8C"),
			Comment:      tcell.GetColor("#616E88"),
			Preprocessor: tcell.GetColor("#B48EAD"),
			Number:       tcell.GetColor("#B48EAD"),
			Operator:     tcell.GetColor("#81A1C1"),
		},
	},
	MemoryDump: &MemoryDumpTheme{
		Address:          tcell.GetColor("#EBCB8B"),
		HexDump:          tcell.GetColor("#88C0D0"),
		ASCII:            tcell.GetColor("#A3BE8C"),
		HighlightBG:      tcell.GetColor("#3B4252"),
		WatchpointMarker: tcell.GetColor("#BF616A"),
	},
	Disassembly: &DisassemblyTheme{
		Address:                     tcell.GetColor("#EBCB8B"),
		RegisterOperand:             tcell.GetColor("#A3BE8C"),
		ImmediateOperand:            tcell.GetColor("#B48EAD"),
		Mnemonic:                    tcell.GetColor("#88C0D0"),
		HighlightBG:                 tcell.GetColor("#3B4252"),
		BreakpointMarker:            tcell.GetColor("#BF616A"),
		CallGraphColors:             []tcell.Color{tcell.GetColor("#B48EAD"), tcell.GetColor("#D08770"), tcell.GetColor("#88C0D0")},
		DataDependenciesGraphColors: []tcell.Color{tcell.GetColor("#BF616A"), tcell.GetColor("#EBCB8B"), tcell.GetColor("#A3BE8C")},
	},
	Registers: &RegistersTheme{
		RegisterName:              tcell.GetColor("#A3BE8C"),
		RegisterValue_Decimal:     tcell.GetColor("#ECEFF4"),
		RegisterValue_Hexadecimal: tcell.GetColor("#B48EAD"),
		RegisterValue_Binary:      tcell.GetColor("#88C0D0"),
	}, Events: &EventsTheme{
		Timestamp:             tcell.GetColor("#88c0d0"),
		ProgramLoaded:         tcell.GetColor("#a3be8c"),
		Stepped:               tcell.GetColor("#ebcb8b"),
		BreakpointHit:         tcell.GetColor("#bf616a"),
		WatchpointHit:         tcell.GetColor("#bf616a"),
		ProgramTerminated:     tcell.GetColor("#a3be8c"),
		ProgramHalted:         tcell.GetColor("#bf616a"),
		Error:                 tcell.GetColor("#bf616a"),
		SourceLocationChanged: tcell.GetColor("#88c0d0"),
		Interrupted:           tcell.GetColor("#d08770"),
		Lagging:               tcell.GetColor("#d08770"),
		EventDetail:           tcell.GetColor("#ebcb8b"),
	}, Theme: &tvlib.Theme{
		PrimitiveBackgroundColor:    tcell.GetColor("#2E3440"),
		ContrastBackgroundColor:     tcell.GetColor("#2E3440"),
		MoreContrastBackgroundColor: tcell.GetColor("#272C36"),
		BorderColor:                 tcell.GetColor("#3B4252"),
		TitleColor:                  tcell.GetColor("#ECEFF4"),
		GraphicsColor:               tcell.GetColor("#3B4252"),
		PrimaryTextColor:            tcell.GetColor("#ECEFF4"),
		SecondaryTextColor:          tcell.GetColor("#EBCB8B"),
		TertiaryTextColor:           tcell.GetColor("#A3BE8C"),
		InverseTextColor:            tcell.GetColor("#2E3440"),
		ContrastSecondaryTextColor:  tcell.GetColor("#BF616A"),
	},
}
