package themes

import (
	"github.com/gdamore/tcell/v2"
	tvlib "github.com/rivo/tview"
)

// OneDark is the one-dark theme instance
var OneDark = &Theme{
	Name:        "One Dark",
	Description: "Atom One Dark theme",
	UserIO: &UserIOTheme{
		CommandPrompt: tcell.GetColor("#C678DD"),
		Info:          tcell.GetColor("#61AFEF"),
		Error:         tcell.GetColor("#E06C75"),
		Success:       tcell.GetColor("#98C379"),
		Warning:       tcell.GetColor("#E5C07B"),
	},
	SourceSnippet: &SourceSnippetTheme{
		LineNumber:       tcell.GetColor("#E5C07B"),
		CurrentLine:      tcell.GetColor("#3E4451"),
		BreakpointMarker: tcell.GetColor("#E06C75"),
		C: &CSyntaxTheme{
			Keyword:      tcell.GetColor("#C678DD"),
			Type:         tcell.GetColor("#61AFEF"),
			Function:     tcell.GetColor("#61AFEF"),
			String:       tcell.GetColor("#98C379"),
			Comment:      tcell.GetColor("#5C6370"),
			Preprocessor: tcell.GetColor("#C678DD"),
			Number:       tcell.GetColor("#D19A66"),
			Operator:     tcell.GetColor("#56B6C2"),
		},
	},
	MemoryDump: &MemoryDumpTheme{
		Address:          tcell.GetColor("#E5C07B"),
		HexDump:          tcell.GetColor("#56B6C2"),
		ASCII:            tcell.GetColor("#98C379"),
		HighlightBG:      tcell.GetColor("#3E4451"),
		WatchpointMarker: tcell.GetColor("#E06C75"),
	},
	Disassembly: &DisassemblyTheme{
		Address:                     tcell.GetColor("#E5C07B"),
		RegisterOperand:             tcell.GetColor("#98C379"),
		ImmediateOperand:            tcell.GetColor("#C678DD"),
		Mnemonic:                    tcell.GetColor("#61AFEF"),
		HighlightBG:                 tcell.GetColor("#3E4451"),
		BreakpointMarker:            tcell.GetColor("#E06C75"),
		CallGraphColors:             []tcell.Color{tcell.GetColor("#C678DD"), tcell.GetColor("#D19A66"), tcell.GetColor("#61AFEF")},
		DataDependenciesGraphColors: []tcell.Color{tcell.GetColor("#E06C75"), tcell.GetColor("#E5C07B"), tcell.GetColor("#98C379")},
	},
	Registers: &RegistersTheme{
		RegisterName:              tcell.GetColor("#98C379"),
		RegisterValue_Decimal:     tcell.GetColor("#ABB2BF"),
		RegisterValue_Hexadecimal: tcell.GetColor("#C678DD"),
		RegisterValue_Binary:      tcell.GetColor("#61AFEF"),
	}, Events: &EventsTheme{
		Timestamp:             tcell.GetColor("#56b6f2"),
		ProgramLoaded:         tcell.GetColor("#98c379"),
		Stepped:               tcell.GetColor("#e5c07b"),
		BreakpointHit:         tcell.GetColor("#e06c75"),
		WatchpointHit:         tcell.GetColor("#e06c75"),
		ProgramTerminated:     tcell.GetColor("#98c379"),
		ProgramHalted:         tcell.GetColor("#e06c75"),
		Error:                 tcell.GetColor("#e06c75"),
		SourceLocationChanged: tcell.GetColor("#56b6f2"),
		Interrupted:           tcell.GetColor("#d19a66"),
		Lagging:               tcell.GetColor("#d19a66"),
		EventDetail:           tcell.GetColor("#e5c07b"),
	}, Theme: &tvlib.Theme{
		PrimitiveBackgroundColor:    tcell.GetColor("#282C34"),
		ContrastBackgroundColor:     tcell.GetColor("#282C34"),
		MoreContrastBackgroundColor: tcell.GetColor("#21252B"),
		BorderColor:                 tcell.GetColor("#3E4451"),
		TitleColor:                  tcell.GetColor("#ABB2BF"),
		GraphicsColor:               tcell.GetColor("#3E4451"),
		PrimaryTextColor:            tcell.GetColor("#ABB2BF"),
		SecondaryTextColor:          tcell.GetColor("#E5C07B"),
		TertiaryTextColor:           tcell.GetColor("#98C379"),
		InverseTextColor:            tcell.GetColor("#282C34"),
		ContrastSecondaryTextColor:  tcell.GetColor("#E06C75"),
	},
}
