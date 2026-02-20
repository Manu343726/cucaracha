package widgets

import (
	"fmt"

	debuggerUI "github.com/Manu343726/cucaracha/pkg/ui/debugger"
	"github.com/Manu343726/cucaracha/pkg/ui/tui/tview/themes"
	tvlib "github.com/rivo/tview"
)

// Stack is a widget for displaying the call stack
type Stack struct {
	*tvlib.TextView
	result *debuggerUI.StackResult
}

// NewStack creates a new Stack widget
func NewStack(result *debuggerUI.StackResult) *Stack {
	tv := tvlib.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle("Call Stack")
	s := &Stack{
		TextView: tv,
		result:   result,
	}
	s.refresh()
	return s
}

// SetResult sets the result for this widget
func (s *Stack) SetResult(result *debuggerUI.StackResult) {
	s.result = result
	s.refresh()
}

// refresh refreshes the display
func (s *Stack) refresh() {
	if s.result == nil || s.result.Error != nil || len(s.result.StackFrames) == 0 {
		s.TextView.SetText("[#F92672]No stack available[#F8F8F2]")
		return
	}

	var text string
	for i, frame := range s.result.StackFrames {
		funcName := "unknown"
		if frame.Function != nil {
			funcName = *frame.Function
		}
		memAddr := uint32(0)
		if frame.Memory != nil {
			memAddr = frame.Memory.Start
		}
		// Monokai: cyan for frame number, yellow for address
		text += fmt.Sprintf("[#A1EFE4]#%d[#F8F8F2] %s [#E6DB74](0x%x)[#F8F8F2]\n", i, funcName, memAddr)
	}
	s.TextView.SetText(text)
}

// Registers is a widget for displaying CPU registers
type Registers struct {
	*tvlib.TextView
	result *debuggerUI.RegistersResult
}

// NewRegisters creates a new Registers widget
func NewRegisters(result *debuggerUI.RegistersResult) *Registers {
	tv := tvlib.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle("Registers")
	r := &Registers{
		TextView: tv,
		result:   result,
	}
	r.refresh()
	return r
}

// SetResult sets the result for this widget
func (r *Registers) SetResult(result *debuggerUI.RegistersResult) {
	r.result = result
	r.refresh()
}

// refresh refreshes the display
func (r *Registers) refresh() {
	if r.result == nil || r.result.Error != nil || len(r.result.Registers) == 0 {
		r.TextView.SetText("[#F92672]No registers available[#F8F8F2]")
		return
	}

	var text string
	for _, reg := range r.result.Registers {
		// Monokai: green for register name, yellow for value
		text += fmt.Sprintf("[#A6E22E]%-3s[#F8F8F2]: [#E6DB74]0x%08x[#F8F8F2]\n", reg.Name, reg.Value)
	}

	if r.result.Flags != nil {
		// Monokai: orange for flags section
		text += fmt.Sprintf("\n[#FD971F]Flags[#F8F8F2]: N=%v Z=%v C=%v V=%v", r.result.Flags.N, r.result.Flags.Z, r.result.Flags.C, r.result.Flags.V)
	}

	r.TextView.SetText(text)
}

// Memory is a widget for displaying memory contents
type Memory struct {
	*tvlib.TextView
	result *debuggerUI.MemoryResult
}

// NewMemory creates a new Memory widget
func NewMemory(result *debuggerUI.MemoryResult) *Memory {
	tv := tvlib.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle("Memory")
	m := &Memory{
		TextView: tv,
		result:   result,
	}
	m.refresh()
	return m
}

// SetResult sets the result for this widget
func (m *Memory) SetResult(result *debuggerUI.MemoryResult) {
	m.result = result
	m.refresh()
}

// refresh refreshes the display
func (m *Memory) refresh() {
	if m.result == nil || m.result.Error != nil || len(m.result.Data) == 0 {
		m.TextView.SetText("[#F92672]No memory available[#F8F8F2]")
		return
	}

	var text string
	bytesPerRow := 16

	for i := 0; i < len(m.result.Data); i += bytesPerRow {
		endIdx := i + bytesPerRow
		if endIdx > len(m.result.Data) {
			endIdx = len(m.result.Data)
		}

		bytesLine := m.result.Data[i:endIdx]
		hexStr := ""
		for _, b := range bytesLine {
			hexStr += fmt.Sprintf("%02x ", b)
		}

		// Monokai: yellow for address, cyan for hex dump
		text += fmt.Sprintf("[#E6DB74]0x%08x[#F8F8F2]: [#A1EFE4]%-48s[#F8F8F2]", m.result.Address+uint32(i), hexStr)

		asciiStr := ""
		for _, b := range bytesLine {
			if b >= 32 && b < 127 {
				asciiStr += string(b)
			} else {
				asciiStr += "."
			}
		}
		// Monokai: green for ASCII representation
		text += fmt.Sprintf("  [#A6E22E]%s[#F8F8F2]\n", asciiStr)
	}

	m.TextView.SetText(text)
}

// SetTheme applies the theme to the Stack widget
func (s *Stack) SetTheme(theme *themes.Theme) *Stack {
	s.TextView.SetBackgroundColor(theme.PrimitiveBackgroundColor)
	s.TextView.SetTextColor(theme.PrimaryTextColor)
	s.TextView.SetBorderColor(theme.BorderColor)
	return s
}

// SetTheme applies the theme to the Registers widget
func (r *Registers) SetTheme(theme *themes.Theme) *Registers {
	r.TextView.SetBackgroundColor(theme.PrimitiveBackgroundColor)
	r.TextView.SetTextColor(theme.PrimaryTextColor)
	r.TextView.SetBorderColor(theme.BorderColor)
	return r
}

// SetTheme applies the theme to the Memory widget
func (m *Memory) SetTheme(theme *themes.Theme) *Memory {
	m.TextView.SetBackgroundColor(theme.PrimitiveBackgroundColor)
	m.TextView.SetTextColor(theme.PrimaryTextColor)
	m.TextView.SetBorderColor(theme.BorderColor)
	return m
}
