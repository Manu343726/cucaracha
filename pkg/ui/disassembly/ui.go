package disassembly

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Manu343726/cucaracha/pkg/ui/debugger"
)

// DisasmApp holds the tview application and UI components
type DisasmApp struct {
	app     *tview.Application
	table   *tview.Table
	footer  *tview.TextView
	flex    *tview.Flex
	session *Session
}

// NewDisasmApp creates a new tview-based application
func NewDisasmApp(session *Session) (*DisasmApp, error) {
	app := tview.NewApplication()

	// Create table for disassembly
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(0, 0)

	// Create footer for status
	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true)
	footer.SetText("View: 0/0 | Addr: 0x0 | Press ? for help")

	// Create main layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true).
		AddItem(footer, 1, 1, false)

	da := &DisasmApp{
		app:     app,
		table:   table,
		footer:  footer,
		flex:    flex,
		session: session,
	}

	// Setup input handling
	da.setupInputHandlers()

	// Populate table with instructions
	da.updateTable()

	app.SetRoot(flex, true)

	return da, nil
}

// setupInputHandlers sets up vim-style key handlers
func (da *DisasmApp) setupInputHandlers() {
	da.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			ch := event.Rune()
			switch ch {
			case 'q':
				da.session.SetExit(true)
				da.app.Stop()
				return nil
			case 'j':
				row := da.table.GetRowCount()
				if row > 0 {
					idx, _ := da.table.GetSelection()
					da.table.Select(idx+3, 0)
				}
				return nil
			case 'k':
				idx, _ := da.table.GetSelection()
				if idx > 0 {
					da.table.Select(idx-3, 0)
				}
				return nil
			case 'g':
				da.table.Select(0, 0)
				return nil
			case 'G':
				da.table.Select(da.table.GetRowCount()-1, 0)
				return nil
			case '?':
				da.showHelp()
				return nil
			case ':':
				da.showCommandPrompt()
				return nil
			case '/':
				da.showSearchPrompt()
				return nil
			case 'd':
				da.showDependenciesModal()
				return nil
			}
		case tcell.KeyUp:
			idx, _ := da.table.GetSelection()
			if idx > 0 {
				da.table.Select(idx-3, 0)
			}
			return nil
		case tcell.KeyDown:
			idx, _ := da.table.GetSelection()
			if idx < da.table.GetRowCount()-1 {
				da.table.Select(idx+3, 0)
			}
			return nil
		}
		return event
	})
}

// updateTable populates the table with disassembly data
func (da *DisasmApp) updateTable() {
	instructions := da.session.GetInstructions()
	da.table.Clear()

	// Set header
	headers := []string{"CFG", "Deps", "Index", "Address", "Instruction", "Operands", "Source"}
	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetExpansion(0).
			SetMaxWidth(0)
		da.table.SetCell(0, col, cell)
	}

	// Build CFG and dependency visualization data
	// Only show arrows for code within the visible window (current selection ± buffer)
	currentRow, _ := da.table.GetSelection()
	visibleStart := 0
	visibleEnd := len(instructions) - 1

	// If we have a selection, use it to determine visible range
	if currentRow > 0 {
		// Show arrows for instructions around the current view
		// Adjust buffer size as needed (20 = scroll context)
		buffer := 20
		visibleStart = currentRow - buffer
		visibleEnd = currentRow + da.table.GetRowCount() + buffer

		if visibleStart < 0 {
			visibleStart = 0
		}
		if visibleEnd >= len(instructions) {
			visibleEnd = len(instructions) - 1
		}
	}

	cfgArrows := da.buildCFGArrows(instructions, visibleStart, visibleEnd)
	depArrows := da.buildDependencyArrows(instructions, visibleStart, visibleEnd)

	// Color palette for rotating colors across different edges
	colorPalette := []tcell.Color{
		tcell.ColorRed,
		tcell.ColorGreen,
		tcell.ColorYellow,
		tcell.ColorBlue,
		tcell.ColorWhite,
		tcell.ColorLime,
	}

	// Track previous source location for de-duplication
	var prevSourceFile string
	var prevSourceLine int

	// Add instruction rows
	for i, instr := range instructions {
		// Get CFG arrow for this instruction
		cfgArrow := cfgArrows[i]
		cfgColor := tcell.ColorDefault
		if cfgArrow != "" {
			// Rotate color based on the arrow content
			// Use a simple hash to assign consistent colors
			colorIdx := 0
			for _, ch := range cfgArrow {
				colorIdx = (colorIdx*31 + int(ch)) % len(colorPalette)
			}
			cfgColor = colorPalette[colorIdx]
		}

		// Get dependency arrow for this instruction
		depArrow := depArrows[i]
		depColor := tcell.ColorDefault
		if depArrow != "" {
			// Rotate color based on the arrow content
			colorIdx := 0
			for _, ch := range depArrow {
				colorIdx = (colorIdx*31 + int(ch)) % len(colorPalette)
			}
			depColor = colorPalette[colorIdx]
		}

		// Format operands
		operandsStr := make([]string, len(instr.Operands))
		for j, op := range instr.Operands {
			operandsStr[j] = formatOperand(op)
		}
		operands := strings.Join(operandsStr, ", ")

		// Format source code (only show if location changed)
		sourceText := ""
		if instr.SourceLine != nil && instr.SourceLine.Location != nil {
			curSourceFile := instr.SourceLine.Location.File
			curSourceLine := instr.SourceLine.Location.Line

			// Only show source if file or line changed
			if curSourceFile != prevSourceFile || curSourceLine != prevSourceLine {
				sourceText = fmt.Sprintf("%s:%d %s",
					curSourceFile,
					curSourceLine,
					instr.SourceLine.Text)
				prevSourceFile = curSourceFile
				prevSourceLine = curSourceLine
			}
		}

		row := i + 1
		cols := []string{
			cfgArrow,
			depArrow,
			fmt.Sprintf("[%d]", i),
			fmt.Sprintf("0x%x", instr.Address),
			instr.Mnemonic,
			operands,
			sourceText,
		}

		for col, text := range cols {
			cell := tview.NewTableCell(text).
				SetExpansion(0).
				SetMaxWidth(0)
			if col == 0 {
				// Set color for CFG column
				cell.SetTextColor(cfgColor)
			} else if col == 1 {
				// Set color for Deps column
				cell.SetTextColor(depColor)
			}
			da.table.SetCell(row, col, cell)
		}
	}
}

// buildCFGArrows creates control flow visualization arrows for each instruction
// Only includes arrows where both source and target are within the visible range [visibleStart, visibleEnd]
// Uses a proper lane-based system where arrows never cross
func (da *DisasmApp) buildCFGArrows(instructions []*debugger.Instruction, visibleStart, visibleEnd int) map[int]string {
	arrows := make(map[int]string)

	// Initialize all arrows as empty
	for i := range instructions {
		arrows[i] = ""
	}

	// Build address to index map
	addrToIdx := make(map[uint32]int)
	for i, instr := range instructions {
		addrToIdx[instr.Address] = i
	}

	// Get jump graph
	jumpGraph := da.session.GetJumpGraph()
	if jumpGraph == nil {
		return arrows
	}

	// Collect all edges with their spans
	type EdgeInfo struct {
		from, to  int
		isForward bool
		laneIdx   int
		targetSym string // Symbol name at the target
	}
	var edges []EdgeInfo

	for source, target := range jumpGraph.Edges {
		fromIdx, fromOk := addrToIdx[source]
		toIdx, toOk := addrToIdx[target]
		if !fromOk || !toOk {
			continue
		}

		// Only include edges where both endpoints are within visible range
		if fromIdx < visibleStart || fromIdx > visibleEnd || toIdx < visibleStart || toIdx > visibleEnd {
			continue
		}

		// Skip fallthrough edges (sequential instruction flow)
		// Only show explicit jumps/branches, not implicit fallthrough to next instruction
		if toIdx == fromIdx+1 {
			continue
		}

		isForward := fromIdx < toIdx

		// Get the symbol name from the source instruction (the one doing the branch)
		targetSym := ""
		if fromIdx < len(instructions) && instructions[fromIdx].BranchTargetSym != nil {
			targetSym = *instructions[fromIdx].BranchTargetSym
		}

		edges = append(edges, EdgeInfo{from: fromIdx, to: toIdx, isForward: isForward, laneIdx: -1, targetSym: targetSym})
	}

	// Assign lanes to edges, with longer edges on the left (lower lane numbers)
	// This prevents edges from crossing

	// First, calculate span length for each edge and sort
	type EdgeWithSpan struct {
		idx  int
		span int
	}
	edgesWithSpan := make([]EdgeWithSpan, len(edges))
	for i := range edges {
		minIdx := edges[i].from
		maxIdx := edges[i].to
		if minIdx > maxIdx {
			minIdx, maxIdx = maxIdx, minIdx
		}
		span := maxIdx - minIdx
		edgesWithSpan[i] = EdgeWithSpan{idx: i, span: span}
	}

	// Sort by span length descending (longest first)
	for i := 0; i < len(edgesWithSpan)-1; i++ {
		for j := i + 1; j < len(edgesWithSpan); j++ {
			if edgesWithSpan[j].span > edgesWithSpan[i].span {
				edgesWithSpan[i], edgesWithSpan[j] = edgesWithSpan[j], edgesWithSpan[i]
			}
		}
	}

	// Assign lanes in order (longer edges get lower lane numbers = left side)
	for _, ewSpan := range edgesWithSpan {
		ei := ewSpan.idx
		lane := 0
	assignLane:
		for {
			// Check if this lane conflicts with other edges
			conflicts := false
			for j := 0; j < ei; j++ {
				if edges[j].laneIdx == lane {
					// Check if these edges' ranges overlap
					minI := edges[ei].from
					maxI := edges[ei].to
					if minI > maxI {
						minI, maxI = maxI, minI
					}

					minJ := edges[j].from
					maxJ := edges[j].to
					if minJ > maxJ {
						minJ, maxJ = maxJ, minJ
					}

					// If ranges overlap, conflict
					if !(maxI < minJ || maxJ < minI) {
						conflicts = true
						break
					}
				}
			}
			if !conflicts {
				edges[ei].laneIdx = lane
				break assignLane
			}
			lane++
		}
	}

	// Build the arrow strings for each row
	for rowIdx := range instructions {
		var activeEdges []int

		// Find all edges active at this row
		for ei, edge := range edges {
			minIdx := edge.from
			maxIdx := edge.to
			if minIdx > maxIdx {
				minIdx, maxIdx = maxIdx, minIdx
			}

			if rowIdx >= minIdx && rowIdx <= maxIdx {
				activeEdges = append(activeEdges, ei)
			}
		}

		if len(activeEdges) == 0 {
			continue
		}

		// Build visual representation for this row
		maxLane := -1
		for _, ei := range activeEdges {
			if edges[ei].laneIdx > maxLane {
				maxLane = edges[ei].laneIdx
			}
		}

		// Create a lane visualization with actual visual lines connecting source to target
		visualStr := ""
		symbolStr := ""
		for lane := 0; lane <= maxLane; lane++ {
			hasEdge := false
			isSource := false
			isTarget := false
			isForward := false
			targetSym := ""

			for _, ei := range activeEdges {
				if edges[ei].laneIdx == lane {
					hasEdge = true
					isForward = edges[ei].isForward
					if edges[ei].isForward && edges[ei].from == rowIdx {
						isSource = true
					} else if !edges[ei].isForward && edges[ei].from == rowIdx {
						isSource = true
					}

					if edges[ei].to == rowIdx {
						isTarget = true
						targetSym = edges[ei].targetSym
					}
					break
				}
			}

			if hasEdge {
				if isSource {
					// Source: use heavy corners to show clear start point
					if isForward {
						visualStr += "┏" // Heavy top-left corner for forward jump
					} else {
						visualStr += "┗" // Heavy bottom-left corner for backward jump
					}
				} else if isTarget {
					// Target: heavy corners with arrow to show clear endpoint
					if isForward {
						visualStr += "┛→" // Heavy top-right corner + arrow for forward target
					} else {
						visualStr += "┓→" // Heavy bottom-right corner + arrow for backward target
					}
					// Append symbol name if this is the target
					if targetSym != "" && symbolStr == "" {
						symbolStr = " " + targetSym
					}
				} else {
					// Middle: heavy vertical line
					visualStr += "┃"
				}
			} else {
				visualStr += " " // Empty lane
			}
		}

		arrows[rowIdx] = visualStr + symbolStr
	}

	return arrows
}

// buildDependencyArrows creates dependency visualization arrows
// Only includes arrows where both source and target are within the visible range [visibleStart, visibleEnd]
func (da *DisasmApp) buildDependencyArrows(instructions []*debugger.Instruction, visibleStart, visibleEnd int) map[int]string {
	arrows := make(map[int]string)

	// Initialize all arrows
	for i := range instructions {
		arrows[i] = ""
	}

	depGraph := da.session.GetDependencyGraph()
	if depGraph == nil {
		return arrows
	}

	// Build address to index map
	addrToIdx := make(map[uint32]int)
	for i, instr := range instructions {
		addrToIdx[instr.Address] = i
	}

	// Get instructions by address for easy lookup
	instrByAddr := make(map[uint32]*debugger.Instruction)
	for _, instr := range instructions {
		instrByAddr[instr.Address] = instr
	}

	// Collect dependency edges with register information
	type DepEdgeInfo struct {
		from, to  int
		laneIdx   int
		regName   string // Register that creates the dependency
		isForward bool
	}
	var depEdges []DepEdgeInfo

	for _, instr := range instructions {
		deps := depGraph.GetDependencies(instr.Address)
		toIdx, ok := addrToIdx[instr.Address]
		if !ok {
			continue
		}

		for _, depAddr := range deps {
			fromIdx, ok := addrToIdx[depAddr]
			if !ok {
				continue
			}

			// Only include edges where both endpoints are within visible range
			if fromIdx < visibleStart || fromIdx > visibleEnd || toIdx < visibleStart || toIdx > visibleEnd {
				continue
			}

			// Find which register creates this dependency
			regName := findDependencyRegister(instrByAddr[depAddr], instr)

			isForward := fromIdx < toIdx
			depEdges = append(depEdges, DepEdgeInfo{
				from:      fromIdx,
				to:        toIdx,
				laneIdx:   -1,
				regName:   regName,
				isForward: isForward,
			})
		}
	}

	// Assign lanes to dependency edges, with longer edges on the left (lower lane numbers)
	// This prevents edges from crossing

	// First, calculate span length for each edge and sort
	type DepEdgeWithSpan struct {
		idx  int
		span int
	}
	depEdgesWithSpan := make([]DepEdgeWithSpan, len(depEdges))
	for i := range depEdges {
		minIdx := depEdges[i].from
		maxIdx := depEdges[i].to
		if minIdx > maxIdx {
			minIdx, maxIdx = maxIdx, minIdx
		}
		span := maxIdx - minIdx
		depEdgesWithSpan[i] = DepEdgeWithSpan{idx: i, span: span}
	}

	// Sort by span length descending (longest first)
	for i := 0; i < len(depEdgesWithSpan)-1; i++ {
		for j := i + 1; j < len(depEdgesWithSpan); j++ {
			if depEdgesWithSpan[j].span > depEdgesWithSpan[i].span {
				depEdgesWithSpan[i], depEdgesWithSpan[j] = depEdgesWithSpan[j], depEdgesWithSpan[i]
			}
		}
	}

	// Assign lanes in order (longer edges get lower lane numbers = left side)
	for _, desSpan := range depEdgesWithSpan {
		ei := desSpan.idx
		lane := 0
	assignLane:
		for {
			conflicts := false
			for j := 0; j < ei; j++ {
				if depEdges[j].laneIdx == lane {
					minI := depEdges[ei].from
					maxI := depEdges[ei].to
					if minI > maxI {
						minI, maxI = maxI, minI
					}

					minJ := depEdges[j].from
					maxJ := depEdges[j].to
					if minJ > maxJ {
						minJ, maxJ = maxJ, minJ
					}

					if !(maxI < minJ || maxJ < minI) {
						conflicts = true
						break
					}
				}
			}
			if !conflicts {
				depEdges[ei].laneIdx = lane
				break assignLane
			}
			lane++
		}
	}

	// Build the dependency strings for each row
	for rowIdx := range instructions {
		var activeEdges []int

		// Find all dependency edges active at this row
		for ei, edge := range depEdges {
			minIdx := edge.from
			maxIdx := edge.to
			if minIdx > maxIdx {
				minIdx, maxIdx = maxIdx, minIdx
			}

			if rowIdx >= minIdx && rowIdx <= maxIdx {
				activeEdges = append(activeEdges, ei)
			}
		}

		if len(activeEdges) == 0 {
			continue
		}

		// Build visual representation
		maxLane := -1
		for _, ei := range activeEdges {
			if depEdges[ei].laneIdx > maxLane {
				maxLane = depEdges[ei].laneIdx
			}
		}

		visualStr := ""
		regStr := ""
		for lane := 0; lane <= maxLane; lane++ {
			hasEdge := false
			isSource := false
			isTarget := false
			regName := ""
			isForward := false

			for _, ei := range activeEdges {
				if depEdges[ei].laneIdx == lane {
					hasEdge = true
					isForward = depEdges[ei].isForward
					if depEdges[ei].from == rowIdx {
						isSource = true
						regName = depEdges[ei].regName
					}
					if depEdges[ei].to == rowIdx {
						isTarget = true
					}
					break
				}
			}

			if hasEdge {
				if isSource {
					// Source: use heavy corners to show clear start point
					if isForward {
						visualStr += "┏" // Heavy top-left corner for forward dependency
					} else {
						visualStr += "┗" // Heavy bottom-left corner for backward dependency
					}
					// Append register name at source
					if regName != "" && regStr == "" {
						regStr = " " + regName
					}
				} else if isTarget {
					// Target: heavy corners with arrow to show clear endpoint
					if isForward {
						visualStr += "┛→" // Heavy top-right corner + arrow for forward target
					} else {
						visualStr += "┓→" // Heavy bottom-right corner + arrow for backward target
					}
				} else {
					// Middle: heavy vertical line
					visualStr += "┃"
				}
			} else {
				visualStr += " "
			}
		}

		arrows[rowIdx] = visualStr + regStr
	}

	return arrows
}

// findDependencyRegister finds which register causes a dependency between source and target instructions
func findDependencyRegister(source *debugger.Instruction, target *debugger.Instruction) string {
	// Get registers written by source
	writtenRegs := make(map[string]bool)
	for _, op := range source.Operands {
		if op.Kind == debugger.OperandKindRegister && op.Register != nil {
			// Simplified: assume all operands are both read and written
			// A real implementation would need to know the instruction semantics
			writtenRegs[op.Register.Name] = true
		}
	}

	// Get registers read by target
	for _, op := range target.Operands {
		if op.Kind == debugger.OperandKindRegister && op.Register != nil {
			if writtenRegs[op.Register.Name] {
				return op.Register.Name
			}
		}
	}

	return ""
}

// showHelp displays help information
func (da *DisasmApp) showHelp() {
	helpText := `Vim-style Navigation:
  j/k        - Scroll down/up
  g/G        - Jump to top/bottom
  ↑↓         - Navigate by 3 lines

Commands:
  q          - Quit
  ?          - Show this help
  d          - Show dependencies for current instruction
  :cmd       - Execute command
  /pattern   - Search

Colon Commands:
  :info      - Show program info
  :jumps     - Show jump graph
  :deps addr - Show dependencies for address
  :help      - Show help`

	modal := tview.NewModal().
		SetText(helpText).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			da.app.SetRoot(da.flex, true)
		})

	da.app.SetRoot(modal, true)
}

// showCommandPrompt shows an input box for commands
func (da *DisasmApp) showCommandPrompt() {
	inputField := tview.NewInputField().
		SetLabel(":").
		SetFieldWidth(40)

	inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			cmd := inputField.GetText()
			da.session.ExecuteCommand(cmd)
			da.updateTable()
		}
		da.app.SetRoot(da.flex, true)
	})

	modal := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(da.flex, 0, 1, false).
		AddItem(inputField, 1, 0, true)

	da.app.SetRoot(modal, true)
}

// showSearchPrompt shows an input box for search
func (da *DisasmApp) showSearchPrompt() {
	inputField := tview.NewInputField().
		SetLabel("/").
		SetFieldWidth(40)

	inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			pattern := inputField.GetText()
			da.session.ExecuteCommand("/" + pattern)
			da.updateTable()
		}
		da.app.SetRoot(da.flex, true)
	})

	modal := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(da.flex, 0, 1, false).
		AddItem(inputField, 1, 0, true)

	da.app.SetRoot(modal, true)
}

// showDependenciesModal shows dependencies for the currently selected instruction
func (da *DisasmApp) showDependenciesModal() {
	instructions := da.session.GetInstructions()
	depGraph := da.session.GetDependencyGraph()

	if depGraph == nil || len(instructions) == 0 {
		da.app.SetRoot(da.flex, true)
		return
	}

	// Get the current selected row
	row, _ := da.table.GetSelection()
	// Adjust for header row (row 0 is header)
	instrIdx := row - 1
	if instrIdx < 0 || instrIdx >= len(instructions) {
		da.app.SetRoot(da.flex, true)
		return
	}

	instr := instructions[instrIdx]
	deps := depGraph.GetDependencies(instr.Address)
	dependents := depGraph.GetDependents(instr.Address)

	// Build dependency info text
	depText := fmt.Sprintf("Instruction at 0x%x (%s)\n\n", instr.Address, instr.Mnemonic)

	if len(deps) > 0 {
		depText += "Depends on:\n"
		for _, depAddr := range deps {
			depInstr := da.session.GetInstruction(depAddr)
			if depInstr != nil {
				depText += fmt.Sprintf("  0x%x: %s\n", depAddr, depInstr.Mnemonic)
			}
		}
	} else {
		depText += "No incoming dependencies\n"
	}

	depText += "\n"

	if len(dependents) > 0 {
		depText += "Has dependents:\n"
		for _, depAddr := range dependents {
			depInstr := da.session.GetInstruction(depAddr)
			if depInstr != nil {
				depText += fmt.Sprintf("  0x%x: %s\n", depAddr, depInstr.Mnemonic)
			}
		}
	} else {
		depText += "No outgoing dependencies\n"
	}

	modal := tview.NewModal().
		SetText(depText).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			da.app.SetRoot(da.flex, true)
		})

	da.app.SetRoot(modal, true)
}

// Run starts the tview application
func (da *DisasmApp) Run() error {
	return da.app.Run()
}

// Stop stops the application
func (da *DisasmApp) Stop() {
	da.app.Stop()
}

// formatOperand formats an instruction operand for display
func formatOperand(op *debugger.InstructionOperand) string {
	switch op.Kind {
	case debugger.OperandKindRegister:
		if op.Register != nil {
			return op.Register.Name
		}
		return "?reg"
	case debugger.OperandKindImmediate:
		if op.Immediate != nil {
			return fmt.Sprintf("0x%x", *op.Immediate)
		}
		return "?imm"
	default:
		return "?"
	}
}
