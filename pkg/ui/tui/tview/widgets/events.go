package widgets

import (
	"fmt"
	"time"

	"github.com/Manu343726/cucaracha/pkg/ui"
	"github.com/Manu343726/cucaracha/pkg/ui/tui/tview/themes"
	tvlib "github.com/rivo/tview"
)

// Events is a widget for displaying debugger events
type Events struct {
	*tvlib.TextView
	events []*eventEntry
	maxLen int
	theme  *themes.Theme
}

// eventEntry represents a single event in the display
type eventEntry struct {
	timestamp time.Time
	eventType ui.DebuggerEventType
	result    *ui.ExecutionResult
}

// NewEvents creates a new Events widget
func NewEvents() *Events {
	tv := tvlib.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle("Events")
	e := &Events{
		TextView: tv,
		events:   make([]*eventEntry, 0),
		maxLen:   100, // Keep last 100 events
		theme:    nil,
	}
	e.refresh()
	return e
}

// AddEvent adds an event to the display
func (e *Events) AddEvent(event *ui.DebuggerEvent) {
	if event == nil {
		return
	}

	entry := &eventEntry{
		timestamp: time.Now(),
		eventType: event.Type,
		result:    event.Result,
	}

	e.events = append(e.events, entry)

	// Keep only the last maxLen events
	if len(e.events) > e.maxLen {
		e.events = e.events[len(e.events)-e.maxLen:]
	}

	e.refresh()
	e.ScrollToEnd()
}

// Clear clears all events
func (e *Events) Clear() {
	e.events = make([]*eventEntry, 0)
	e.refresh()
}

// SetTheme applies the theme to the Events widget
func (e *Events) SetTheme(theme *themes.Theme) *Events {
	e.theme = theme
	e.TextView.SetBackgroundColor(theme.PrimitiveBackgroundColor)
	e.TextView.SetTextColor(theme.PrimaryTextColor)
	e.TextView.SetBorderColor(theme.BorderColor)
	// Refresh display to apply new colors
	e.refresh()
	return e
}

// refresh refreshes the display
func (e *Events) refresh() {
	if len(e.events) == 0 {
		var waitColor string
		if e.theme != nil && e.theme.Events != nil {
			waitColor = fmt.Sprintf("%06x", e.theme.Events.Timestamp.Hex())
		} else {
			waitColor = "A1EFE4" // Default Monokai cyan
		}
		e.TextView.SetText(fmt.Sprintf("[#%s]Waiting for events...[#-:-]", waitColor))
		return
	}

	var text string
	for _, entry := range e.events {
		text += e.formatEvent(entry) + "\n"
	}
	e.TextView.SetText(text)
}

// formatEvent formats a single event for display
func (e *Events) formatEvent(entry *eventEntry) string {
	// Use theme colors if available, otherwise use default formatting
	timestamp := entry.timestamp.Format("15:04:05.000")

	var timestampColor string
	var eventTypeColor string
	var detailColor string

	if e.theme != nil && e.theme.Events != nil {
		timestampColor = fmt.Sprintf("%06x", e.theme.Events.Timestamp.Hex())
		detailColor = fmt.Sprintf("%06x", e.theme.Events.EventDetail.Hex())
	} else {
		timestampColor = "A1EFE4" // Monokai cyan
		detailColor = "E6DB74"    // Monokai yellow
	}

	eventTypeStr := entry.eventType.String()

	var details string
	if entry.result != nil {
		details = e.formatEventDetails(entry, &eventTypeColor)
	}

	if details != "" {
		return fmt.Sprintf("[#%s]%s[#-:-] %s [#%s]%s[#-:-]", timestampColor, timestamp, details, detailColor, "")
	}
	return fmt.Sprintf("[#%s]%s[#-:-] [#%s]%s[#-:-]", timestampColor, timestamp, eventTypeColor, eventTypeStr)
}

// formatEventDetails formats event-specific details
func (e *Events) formatEventDetails(entry *eventEntry, eventTypeColorPtr *string) string {
	result := entry.result
	if result == nil {
		return ""
	}

	// Get color for event type
	var eventTypeColor string
	if e.theme != nil && e.theme.Events != nil {
		eventTypeColor = e.getEventTypeColor(entry.eventType)
	} else {
		eventTypeColor = "A6E22E" // Monokai green
	}
	if eventTypeColorPtr != nil {
		*eventTypeColorPtr = eventTypeColor
	}

	detailColor := ""
	if e.theme != nil && e.theme.Events != nil {
		detailColor = fmt.Sprintf("%06x", e.theme.Events.EventDetail.Hex())
	} else {
		detailColor = "E6DB74" // Monokai yellow
	}

	switch entry.eventType {
	case ui.DebuggerEventStepped:
		return fmt.Sprintf("[#%s]%s[#-:-] [#%s]PC: 0x%08x Steps: %d[#-:-]", eventTypeColor, "Stepped", detailColor, result.LastInstruction, result.Steps)

	case ui.DebuggerEventBreakpointHit:
		msg := "Breakpoint hit"
		if result.Breakpoint != nil {
			msg = fmt.Sprintf("Breakpoint %d hit", result.Breakpoint.ID)
		}
		return fmt.Sprintf("[#%s]%s[#-:-] [#%s]at 0x%08x[#-:-]", eventTypeColor, msg, detailColor, result.LastInstruction)

	case ui.DebuggerEventWatchpointHit:
		msg := "Watchpoint triggered"
		if result.Watchpoint != nil {
			msg = fmt.Sprintf("Watchpoint %d triggered", result.Watchpoint.ID)
		}
		return fmt.Sprintf("[#%s]%s[#-:-] [#%s]at 0x%08x[#-:-]", eventTypeColor, msg, detailColor, result.LastInstruction)

	case ui.DebuggerEventProgramTerminated:
		return fmt.Sprintf("[#%s]Normal termination[#-:-] [#%s]Steps: %d[#-:-]", eventTypeColor, detailColor, result.Steps)

	case ui.DebuggerEventProgramHalted:
		return fmt.Sprintf("[#%s]CPU halted[#-:-] [#%s]at 0x%08x[#-:-]", eventTypeColor, detailColor, result.LastInstruction)

	case ui.DebuggerEventError:
		msg := "Error"
		if result.Error != nil {
			msg = fmt.Sprintf("Error: %v", result.Error)
		}
		return fmt.Sprintf("[#%s]%s[#-:-]", eventTypeColor, msg)

	case ui.DebuggerEventSourceLocationChanged:
		return fmt.Sprintf("[#%s]Source location changed[#-:-] [#%s]at 0x%08x[#-:-]", eventTypeColor, detailColor, result.LastInstruction)

	case ui.DebuggerEventInterrupted:
		return fmt.Sprintf("[#%s]Interrupted[#-:-] [#%s]at 0x%08x[#-:-]", eventTypeColor, detailColor, result.LastInstruction)

	case ui.DebuggerEventLagging:
		lagMsg := "lagging"
		if result.LaggingCycles == 0 {
			lagMsg = "on-time"
		}
		return fmt.Sprintf("[#%s]%s[#-:-] [#%s]by %d cycles[#-:-]", eventTypeColor, lagMsg, detailColor, result.LaggingCycles)

	case ui.DebuggerEventProgramLoaded:
		return fmt.Sprintf("[#%s]Program loaded[#-:-]", eventTypeColor)

	default:
		return ""
	}
}

// getEventTypeColor returns the appropriate color for an event type
func (e *Events) getEventTypeColor(eventType ui.DebuggerEventType) string {
	if e.theme == nil || e.theme.Events == nil {
		return "A6E22E" // Default Monokai green
	}

	var color interface{}
	switch eventType {
	case ui.DebuggerEventProgramLoaded:
		color = e.theme.Events.ProgramLoaded
	case ui.DebuggerEventStepped:
		color = e.theme.Events.Stepped
	case ui.DebuggerEventBreakpointHit:
		color = e.theme.Events.BreakpointHit
	case ui.DebuggerEventWatchpointHit:
		color = e.theme.Events.WatchpointHit
	case ui.DebuggerEventProgramTerminated:
		color = e.theme.Events.ProgramTerminated
	case ui.DebuggerEventProgramHalted:
		color = e.theme.Events.ProgramHalted
	case ui.DebuggerEventError:
		color = e.theme.Events.Error
	case ui.DebuggerEventSourceLocationChanged:
		color = e.theme.Events.SourceLocationChanged
	case ui.DebuggerEventInterrupted:
		color = e.theme.Events.Interrupted
	case ui.DebuggerEventLagging:
		color = e.theme.Events.Lagging
	default:
		return "A6E22E"
	}

	if tcColor, ok := color.(interface{ Hex() int32 }); ok {
		return fmt.Sprintf("%06x", tcColor.Hex())
	}
	return "A6E22E"
}
