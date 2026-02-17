package llvm

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// pendingLabel holds a label name and the line number it appears on
type pendingLabel struct {
	Name string
	Line int
}

// AssemblyFileParser handles parsing of .cucaracha assembly files
type AssemblyFileParser struct {
	// Input
	path string

	// Compiled regex patterns
	funcHeader    *regexp.Regexp
	globalHeader  *regexp.Regexp
	fileHeader    *regexp.Regexp
	typeHeader    *regexp.Regexp
	sizeHeader    *regexp.Regexp
	labelHeader   *regexp.Regexp
	symbolRef     *regexp.Regexp
	dataDirective *regexp.Regexp

	// Parsing state
	functionNames  map[string]bool
	currentFunc    *FunctionBody
	sourceFile     string
	lineNum        int
	funcInstrStart int
	funcInstrCount int
	pendingLabels  []pendingLabel
	allLabels      []pendingLabel
	pendingGlobal  *GlobalSymbol

	// Output
	file *AssemblyFile
}

// NewAssemblyFileParser creates a new parser for the given file path
func NewAssemblyFileParser(path string) *AssemblyFileParser {
	return &AssemblyFileParser{
		path:          path,
		funcHeader:    regexp.MustCompile(`^([A-Za-z0-9_\.]+):`),
		globalHeader:  regexp.MustCompile(`^\.globl\s+(\S+)`),
		fileHeader:    regexp.MustCompile(`^\.file\s+"?([^\"]+)"?`),
		typeHeader:    regexp.MustCompile(`^\.type\s+(\S+),@(\w+)`),
		sizeHeader:    regexp.MustCompile(`^\.size\s+(\S+),\s*(\S+)`),
		labelHeader:   regexp.MustCompile(`^([A-Za-z0-9_\.]+):`),
		symbolRef:     regexp.MustCompile(`([A-Za-z_\.][A-Za-z0-9_\.]*(:?@[a-z]+)?)`),
		dataDirective: regexp.MustCompile(`^\.(long|byte|word|zero)\s+(.+)`),
		functionNames: map[string]bool{},
	}
}

// Parse executes the full parsing process and returns the AssemblyFile
func (p *AssemblyFileParser) Parse() (*AssemblyFile, error) {
	if err := p.collectFunctionNames(); err != nil {
		return nil, err
	}

	if err := p.parseFile(); err != nil {
		return nil, err
	}

	p.finalizeLastFunction()
	p.resolveLabels()
	p.finalizePendingGlobal()

	p.file.SourceFileValue = p.sourceFile
	return p.file, nil
}

// collectFunctionNames performs a first pass to identify all function names
func (p *AssemblyFileParser) collectFunctionNames() error {
	f, err := os.Open(p.path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if p.typeHeader.MatchString(line) {
			matches := p.typeHeader.FindStringSubmatch(line)
			name := matches[1]
			typ := matches[2]
			if typ == "function" {
				p.functionNames[name] = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}
	return nil
}

// parseFile performs the main parsing pass
func (p *AssemblyFileParser) parseFile() error {
	f, err := os.Open(p.path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	p.file = &AssemblyFile{
		FileNameValue:   p.path,
		GlobalsValue:    []GlobalSymbol{},
		FunctionsMap:    map[string]*FunctionBody{},
		LabelsValue:     []LabelSymbol{},
		InstructionsAll: []Instruction{},
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		p.lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if p.processLine(line) {
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}
	return nil
}

// processLine handles a single line of assembly, returns true if the line was fully processed
func (p *AssemblyFileParser) processLine(line string) bool {
	if p.handleFileDirective(line) {
		return true
	}
	if p.handleTypeDirective(line) {
		return true
	}
	if p.handleFunctionHeader(line) {
		return true
	}
	if p.handleSizeDirective(line) {
		return true
	}
	if p.handleGlobalLabel(line) {
		return true
	}
	if p.handleDataDirective(line) {
		return true
	}
	if p.handleLabel(line) {
		return true
	}
	if p.handleInstruction(line) {
		return true
	}
	p.maybeFinalizePendingGlobal(line)
	return false
}

// handleFileDirective processes .file directives
func (p *AssemblyFileParser) handleFileDirective(line string) bool {
	if !p.fileHeader.MatchString(line) {
		return false
	}
	p.sourceFile = p.fileHeader.FindStringSubmatch(line)[1]
	p.file.SourceFileValue = p.sourceFile
	return true
}

// handleTypeDirective processes .type directives for objects
func (p *AssemblyFileParser) handleTypeDirective(line string) bool {
	if !p.typeHeader.MatchString(line) {
		return false
	}
	matches := p.typeHeader.FindStringSubmatch(line)
	name := matches[1]
	typ := matches[2]

	if typ == "object" {
		p.finalizePendingGlobalIfValid()
		p.pendingGlobal = &GlobalSymbol{
			Name: name,
			Type: GlobalObject,
		}
		return true
	}
	return false
}

// handleFunctionHeader processes function header labels
func (p *AssemblyFileParser) handleFunctionHeader(line string) bool {
	if !p.funcHeader.MatchString(line) {
		return false
	}
	label := p.funcHeader.FindStringSubmatch(line)[1]

	if !p.functionNames[label] {
		return false
	}

	// Close current function if open
	if p.currentFunc != nil {
		p.closeCurrentFunction()
	}

	// Start new function
	p.currentFunc = &FunctionBody{
		Name:       label,
		SourceFile: p.sourceFile,
		StartLine:  p.lineNum,
	}
	p.funcInstrStart = 0
	p.funcInstrCount = 0

	// Note: We do NOT add function labels to pendingLabels
	// Functions are tracked separately from labels
	return true
}

// handleSizeDirective processes .size directives
func (p *AssemblyFileParser) handleSizeDirective(line string) bool {
	if !p.sizeHeader.MatchString(line) {
		return false
	}
	matches := p.sizeHeader.FindStringSubmatch(line)
	name := matches[1]
	sz := matches[2]

	var sizeInt int
	fmt.Sscanf(sz, "%d", &sizeInt)

	if p.pendingGlobal != nil && p.pendingGlobal.Name == name {
		p.pendingGlobal.Size = sizeInt
	} else {
		for i := range p.file.GlobalsValue {
			if p.file.GlobalsValue[i].Name == name {
				p.file.GlobalsValue[i].Size = sizeInt
			}
		}
	}
	return true
}

// handleGlobalLabel processes labels that are part of global definitions
func (p *AssemblyFileParser) handleGlobalLabel(line string) bool {
	if !p.labelHeader.MatchString(line) {
		return false
	}
	label := p.labelHeader.FindStringSubmatch(line)[1]

	// Check if this label belongs to the pending global
	if p.pendingGlobal != nil && label == p.pendingGlobal.Name {
		return true
	}

	// Handle .L__const.* labels as new globals
	if strings.HasPrefix(label, ".L__const.") {
		p.finalizePendingGlobalIfValid()
		p.pendingGlobal = &GlobalSymbol{
			Name: label,
			Type: GlobalObject,
		}
		return true
	}

	return false
}

// handleDataDirective processes data directives (.long, .byte, etc.)
func (p *AssemblyFileParser) handleDataDirective(line string) bool {
	if !p.dataDirective.MatchString(line) {
		return false
	}
	matches := p.dataDirective.FindStringSubmatch(line)
	kind := matches[1]
	value := matches[2]

	if p.pendingGlobal != nil {
		p.appendDataToGlobal(p.pendingGlobal, kind, value)
		return true
	}

	if len(p.file.GlobalsValue) > 0 {
		g := &p.file.GlobalsValue[len(p.file.GlobalsValue)-1]
		p.appendDataToGlobal(g, kind, value)
		return true
	}

	return true
}

// appendDataToGlobal appends data to a global symbol based on directive type
func (p *AssemblyFileParser) appendDataToGlobal(g *GlobalSymbol, kind, value string) {
	switch kind {
	case "long":
		var v int32
		fmt.Sscanf(value, "%d", &v)
		g.InitialData = append(g.InitialData, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
	case "byte":
		var v int
		fmt.Sscanf(value, "%d", &v)
		g.InitialData = append(g.InitialData, byte(v))
	}
}

// handleLabel processes non-function, non-global labels
func (p *AssemblyFileParser) handleLabel(line string) bool {
	if !p.labelHeader.MatchString(line) {
		return false
	}

	// Skip if already handled as function header
	label := p.labelHeader.FindStringSubmatch(line)[1]
	if p.functionNames[label] {
		return false
	}

	p.pendingLabels = append(p.pendingLabels, pendingLabel{Name: label, Line: p.lineNum + 1})
	return true
}

// handleInstruction processes instruction lines inside functions
func (p *AssemblyFileParser) handleInstruction(line string) bool {
	if p.currentFunc == nil {
		return false
	}

	// Skip directives
	if strings.HasPrefix(line, ".") {
		return false
	}

	symbols := p.extractSymbolReferences(line)

	instrIdx := len(p.file.InstructionsAll)
	p.file.InstructionsAll = append(p.file.InstructionsAll, Instruction{
		LineNumber: p.lineNum,
		Text:       line,
		Symbols:    symbols,
	})

	if p.funcInstrCount == 0 {
		p.funcInstrStart = instrIdx
	}
	p.funcInstrCount++

	// Collect pending labels for this instruction
	for _, lbl := range p.pendingLabels {
		p.allLabels = append(p.allLabels, pendingLabel{Name: lbl.Name, Line: p.lineNum})
	}
	p.pendingLabels = nil

	return true
}

// extractSymbolReferences extracts symbol references from an instruction line
func (p *AssemblyFileParser) extractSymbolReferences(line string) []string {
	fields := strings.Fields(line)
	if len(fields) <= 1 {
		return []string{}
	}

	symbols := []string{}
	operands := strings.Join(fields[1:], " ")

	for _, m := range p.symbolRef.FindAllStringSubmatch(operands, -1) {
		name := m[1]
		if p.isRegisterOrSpecial(name) {
			continue
		}
		symbols = append(symbols, name)
	}

	return symbols
}

// isRegisterOrSpecial checks if a name is a register or special symbol
func (p *AssemblyFileParser) isRegisterOrSpecial(name string) bool {
	// Skip register references
	if strings.HasPrefix(name, "r") && len(name) <= 3 && len(name) > 1 && name[1] >= '0' && name[1] <= '9' {
		return true
	}
	if strings.HasPrefix(name, "#") {
		return true
	}
	if name == "sp" || name == "lr" {
		return true
	}
	return false
}

// maybeFinalizePendingGlobal checks if we should finalize a pending global
func (p *AssemblyFileParser) maybeFinalizePendingGlobal(line string) {
	if p.pendingGlobal == nil {
		return
	}
	if p.pendingGlobal.Type == GlobalFunction || p.pendingGlobal.Type == GlobalUnknown {
		return
	}

	// Don't finalize if we hit any directive (starts with .)
	// This allows .section, .p2align, etc. between .type and the label/data
	if strings.HasPrefix(line, ".") {
		return
	}

	// Finalize if we hit a non-global related line
	if !p.globalHeader.MatchString(line) &&
		!p.typeHeader.MatchString(line) &&
		!p.sizeHeader.MatchString(line) &&
		!p.labelHeader.MatchString(line) {
		p.file.GlobalsValue = append(p.file.GlobalsValue, *p.pendingGlobal)
		p.pendingGlobal = nil
	}
}

// closeCurrentFunction closes the current function and saves its instruction range
func (p *AssemblyFileParser) closeCurrentFunction() {
	if p.currentFunc == nil {
		return
	}
	p.currentFunc.EndLine = p.lineNum - 1
	if p.funcInstrCount > 0 {
		p.currentFunc.InstructionRanges = append(p.currentFunc.InstructionRanges, InstructionRange{
			Start: p.funcInstrStart,
			Count: p.funcInstrCount,
		})
	}
	p.file.FunctionsMap[p.currentFunc.Name] = p.currentFunc
}

// finalizeLastFunction closes the last function if still open
func (p *AssemblyFileParser) finalizeLastFunction() {
	if p.currentFunc == nil || p.funcInstrCount == 0 {
		return
	}
	p.currentFunc.EndLine = p.lineNum
	p.currentFunc.InstructionRanges = append(p.currentFunc.InstructionRanges, InstructionRange{
		Start: p.funcInstrStart,
		Count: p.funcInstrCount,
	})
	p.file.FunctionsMap[p.currentFunc.Name] = p.currentFunc

	// Add remaining pending labels
	for _, lbl := range p.pendingLabels {
		p.allLabels = append(p.allLabels, lbl)
	}
}

// finalizePendingGlobalIfValid finalizes the pending global if it's valid
func (p *AssemblyFileParser) finalizePendingGlobalIfValid() {
	if p.pendingGlobal != nil && p.pendingGlobal.Type != GlobalFunction && p.pendingGlobal.Type != GlobalUnknown {
		p.file.GlobalsValue = append(p.file.GlobalsValue, *p.pendingGlobal)
		p.pendingGlobal = nil
	}
}

// finalizePendingGlobal ensures any pending global is saved at EOF
func (p *AssemblyFileParser) finalizePendingGlobal() {
	if p.pendingGlobal != nil && p.pendingGlobal.Type != GlobalFunction && p.pendingGlobal.Type != GlobalUnknown {
		p.file.GlobalsValue = append(p.file.GlobalsValue, *p.pendingGlobal)
		p.pendingGlobal = nil
	}
}

// resolveLabels resolves all collected labels to instruction pointers
func (p *AssemblyFileParser) resolveLabels() {
	// Build sets for exclusion
	functionNamesSet := make(map[string]struct{})
	for name := range p.file.FunctionsMap {
		functionNamesSet[name] = struct{}{}
	}

	globalNamesSet := make(map[string]struct{})
	for _, g := range p.file.GlobalsValue {
		globalNamesSet[g.Name] = struct{}{}
	}

	// Resolve each label
	for _, lbl := range p.allLabels {
		if _, isFunc := functionNamesSet[lbl.Name]; isFunc {
			continue
		}
		if _, isGlob := globalNamesSet[lbl.Name]; isGlob {
			continue
		}

		instrPtr := p.findInstructionAtLine(lbl.Line)
		p.file.LabelsValue = append(p.file.LabelsValue, LabelSymbol{
			Name:        lbl.Name,
			Instruction: instrPtr,
		})
	}
}

// findInstructionAtLine finds the instruction at a given line number
func (p *AssemblyFileParser) findInstructionAtLine(line int) *Instruction {
	for _, fn := range p.file.FunctionsMap {
		for _, r := range fn.InstructionRanges {
			for i := r.Start; i < r.Start+r.Count; i++ {
				if p.file.InstructionsAll[i].LineNumber == line {
					return &p.file.InstructionsAll[i]
				}
			}
		}
	}
	return nil
}
