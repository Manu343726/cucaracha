package debugger

import (
	"fmt"
	"os"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/debugger/core"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/interpreter"
	"github.com/Manu343726/cucaracha/pkg/runtime/program"
	"github.com/Manu343726/cucaracha/pkg/runtime/program/loader"
	"github.com/Manu343726/cucaracha/pkg/runtime/program/sourcecode"
	uiDebugger "github.com/Manu343726/cucaracha/pkg/ui/debugger"
	"github.com/Manu343726/cucaracha/pkg/utils"
	"gopkg.in/yaml.v3"
)

type debugger struct {
	session     *core.Session
	runtimeType uiDebugger.RuntimeType // Store the loaded runtime type for later retrieval
}

// Returns a debugger instance
func NewDebugger() uiDebugger.Debugger {
	return &debugger{
		session: &core.Session{},
	}
}

func (d *debugger) SetEventCallback(callback uiDebugger.DebuggerEventCallback) {
	// Convert UI callback to core callback
	d.session.SetEventCallback(func(event *core.Event) bool {
		if callback != nil {
			// For now, we'll call with nil
			// Full implementation would require mapping core event types to UI
			callback(EventToUI(event))
		}
		return true // continue processing
	})
}

func (d *debugger) LoadRuntime(args *uiDebugger.LoadRuntimeArgs) *uiDebugger.LoadRuntimeResult {
	switch args.Runtime {
	case uiDebugger.RuntimeTypeInterpreter:
		if d.session.System() == nil {
			return &uiDebugger.LoadRuntimeResult{
				Error: fmt.Errorf("system must be configured before loading interpreter runtime"),
			}
		}

		r, err := interpreter.NewInterpreter(d.session.System())
		if err != nil {
			return &uiDebugger.LoadRuntimeResult{
				Error: fmt.Errorf("failed to create interpreter runtime: %w", err),
			}
		}

		err = d.session.LoadRuntime(r)
		if err != nil {
			return &uiDebugger.LoadRuntimeResult{
				Error: err,
			}
		}

		// Store the runtime type for later retrieval
		d.runtimeType = args.Runtime

		return &uiDebugger.LoadRuntimeResult{
			Runtime: &uiDebugger.RuntimeInfo{
				Runtime: args.Runtime,
			},
		}
	default:
		return &uiDebugger.LoadRuntimeResult{
			Error: fmt.Errorf("unsupported runtime: %s", args.Runtime),
		}
	}
}

// extractSystemInfo builds a SystemInfo with system information
func (d *debugger) extractSystemInfo() *uiDebugger.SystemInfo {
	if d.session.System() == nil {
		return nil
	}

	sys := d.session.System()
	layout := sys.MemoryLayout

	info := &uiDebugger.SystemInfo{
		TotalMemory:     layout.TotalSize,
		CodeSize:        layout.CodeSize,
		DataSize:        layout.DataSize,
		StackSize:       layout.StackSize,
		HeapSize:        layout.HeapSize,
		PeripheralSize:  layout.PeripheralSize,
		NumberOfVectors: sys.VectorTable.NumberOfVectors,
		VectorEntrySize: sys.VectorTable.VectorEntrySize,
		NumPeripherals:  len(sys.Peripherals),
	}

	// Extract peripheral information
	info.Peripherals = make([]uiDebugger.PeripheralInfo, len(sys.Peripherals))
	for i, p := range sys.Peripherals {
		metadata := p.Metadata()
		typeStr := ""
		displayName := ""
		if metadata.Descriptor != nil {
			typeStr = metadata.Descriptor.Type.String()
			displayName = metadata.Descriptor.DisplayName
		}
		info.Peripherals[i] = uiDebugger.PeripheralInfo{
			Name:            metadata.Name,
			Type:            typeStr,
			DisplayName:     displayName,
			Description:     metadata.Description,
			BaseAddress:     metadata.BaseAddress,
			Size:            metadata.Size,
			InterruptVector: metadata.InterruptVector,
		}
	}

	return info
}

func (d *debugger) LoadSystemFromFile(args *uiDebugger.LoadSystemFromFileArgs) *uiDebugger.LoadSystemFromFileResult {
	if args.FilePath == "default" {
		result := d.LoadSystemFromEmbedded()
		return &uiDebugger.LoadSystemFromFileResult{
			Error:  result.Error,
			System: result.System,
		}
	}

	err := d.session.LoadSystemFromFile(args.FilePath)
	if err != nil {
		return &uiDebugger.LoadSystemFromFileResult{
			Error: err,
		}
	}

	return &uiDebugger.LoadSystemFromFileResult{
		System: d.extractSystemInfo(),
	}
}

func (d *debugger) LoadSystemFromEmbedded() *uiDebugger.LoadSystemFromEmbeddedResult {
	config, err := DefaultSystemConfig()
	if err != nil {
		return &uiDebugger.LoadSystemFromEmbeddedResult{
			Error: fmt.Errorf("failed to load embedded system config: %w", err),
		}
	}

	system, err := config.Setup()
	if err != nil {
		return &uiDebugger.LoadSystemFromEmbeddedResult{
			Error: fmt.Errorf("failed to setup system from embedded config: %w", err),
		}
	}

	err = d.session.LoadSystem(system)
	if err != nil {
		return &uiDebugger.LoadSystemFromEmbeddedResult{
			Error: err,
		}
	}

	return &uiDebugger.LoadSystemFromEmbeddedResult{
		System: d.extractSystemInfo(),
	}
}

func (d *debugger) LoadProgramFromFile(args *uiDebugger.LoadProgramFromFileArgs) *uiDebugger.LoadProgramFromFileResult {
	if d.session.System() == nil {
		return &uiDebugger.LoadProgramFromFileResult{
			Error: fmt.Errorf("system must be configured before loading a program"),
		}
	}

	// Determine auto_build_clang value: use provided value or default to true
	autoBuildClang := true
	if args.AutoBuildClang != nil {
		autoBuildClang = *args.AutoBuildClang
	}

	// Determine force_rebuild_clang value: use provided value or default to false
	forceRebuildClang := false
	if args.ForceRebuildClang != nil {
		forceRebuildClang = *args.ForceRebuildClang
	}

	options := &loader.Options{
		Verbose:           false,
		MemoryLayout:      &d.session.System().MemoryLayout,
		OutputFormat:      "object",
		AutoBuildClang:    autoBuildClang,
		ForceRebuildClang: forceRebuildClang,
	}

	loadedProgram, err := d.session.LoadProgramFromFile(args.FilePath, options)
	if err != nil {
		return &uiDebugger.LoadProgramFromFileResult{
			Error: err,
		}
	}

	// Get the loaded program from the session
	prog := d.session.Program()

	// Get the entry point
	entryPoint, err := program.ProgramEntryPoint(prog)
	if err != nil {
		// If we can't get the entry point, continue anyway (entry point will be 0)
		_ = err // Ignore error, entry point will be 0 which is fine
	}

	return &uiDebugger.LoadProgramFromFileResult{
		Program: &uiDebugger.ProgramInfo{
			Warnings:   loadedProgram.Warnings,
			SourceFile: &loadedProgram.OriginalPath,
			ObjectFile: &loadedProgram.CompiledPath,
			EntryPoint: entryPoint,
		},
	}
}

type allFile struct {
	Runtime     uiDebugger.RuntimeType `json:"runtime"`
	ProgramFile string                 `json:"programFile"`
}

func (d *debugger) loadFromSingleFile(fullDescriptorPath string) *uiDebugger.LoadResult {
	raw, err := os.ReadFile(fullDescriptorPath)
	if err != nil {
		return &uiDebugger.LoadResult{
			Error: fmt.Errorf("failed to read file '%s': %w", fullDescriptorPath, err),
		}
	}

	var all allFile
	if err := yaml.Unmarshal(raw, &all); err != nil {
		return &uiDebugger.LoadResult{
			Error: fmt.Errorf("failed to parse YAML file '%s': %w", fullDescriptorPath, err),
		}
	}

	return d.Load(&uiDebugger.LoadArgs{
		Runtime:          utils.Ptr(all.Runtime),
		SystemConfigPath: utils.Ptr(all.ProgramFile),
		ProgramPath:      utils.Ptr(all.ProgramFile),
	})
}

func (d *debugger) Load(args *uiDebugger.LoadArgs) *uiDebugger.LoadResult {
	if args.FullDescriptorPath != nil {
		return d.loadFromSingleFile(*args.FullDescriptorPath)
	}

	if args.SystemConfigPath == nil {
		args.SystemConfigPath = utils.Ptr("default")
	}

	loadSysResult := d.LoadSystemFromFile(&uiDebugger.LoadSystemFromFileArgs{
		FilePath: *args.SystemConfigPath,
	})
	if loadSysResult.Error != nil {
		return &uiDebugger.LoadResult{
			Error: fmt.Errorf("failed to load system from file '%s': %w", *args.SystemConfigPath, loadSysResult.Error),
		}
	}

	if args.Runtime == nil {
		args.Runtime = utils.Ptr(uiDebugger.RuntimeTypeInterpreter)
	}

	loadRuntimeResult := d.LoadRuntime(&uiDebugger.LoadRuntimeArgs{
		Runtime: *args.Runtime,
	})
	if loadRuntimeResult.Error != nil {
		return &uiDebugger.LoadResult{
			Error: fmt.Errorf("failed to load runtime '%s': %w", *args.Runtime, loadRuntimeResult.Error),
		}
	}

	loadProgResult := d.LoadProgramFromFile(&uiDebugger.LoadProgramFromFileArgs{
		FilePath: *args.ProgramPath,
	})
	if loadProgResult.Error != nil {
		return &uiDebugger.LoadResult{
			Error: fmt.Errorf("failed to load program from file '%s': %w", *args.ProgramPath, loadProgResult.Error),
		}
	}

	return &uiDebugger.LoadResult{
		System:  loadSysResult,
		Program: loadProgResult,
		Runtime: loadRuntimeResult,
	}
}

func (d *debugger) Continue() *uiDebugger.ExecutionResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.ExecutionResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	return ExecutionResultToUI(debugger.Continue())
}

func (d *debugger) Interrupt() *uiDebugger.ExecutionResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.ExecutionResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	return ExecutionResultToUI(debugger.Interrupt())
}

func (d *debugger) Step(args *uiDebugger.StepArgs) *uiDebugger.ExecutionResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.ExecutionResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	switch args.StepMode {
	case uiDebugger.StepModeInto:
		switch args.CountMode {
		case uiDebugger.StepCountInstructions:
			return ExecutionResultToUI(debugger.Step())
		case uiDebugger.StepCountSourceLines:
			return ExecutionResultToUI(debugger.StepIntoSource())
		default:
			return &uiDebugger.ExecutionResult{
				Error: fmt.Errorf("unsupported single step count mode '%s'", args.CountMode),
			}
		}
	case uiDebugger.StepModeOver:
		switch args.CountMode {
		case uiDebugger.StepCountInstructions:
			return ExecutionResultToUI(debugger.StepOver())
		case uiDebugger.StepCountSourceLines:
			return ExecutionResultToUI(debugger.StepOverSource())
		default:
			return &uiDebugger.ExecutionResult{
				Error: fmt.Errorf("unsupported step over count mode '%s'", args.CountMode),
			}
		}
	case uiDebugger.StepModeOut:
		return ExecutionResultToUI(debugger.StepOut())
	default:
		return &uiDebugger.ExecutionResult{
			Error: fmt.Errorf("unsupported step mode: %d", args.StepMode),
		}
	}
}

func (d *debugger) CurrentSource(args *uiDebugger.CurrentSourceArgs) *uiDebugger.SourceResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.SourceResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	pc, err := cpu.ReadPC(debugger.Runtime().CPU().Registers())
	if err != nil {
		return &uiDebugger.SourceResult{
			Error: fmt.Errorf("failed to read PC register: %w", err),
		}
	}

	currentSourceLoc, err := program.SourceLocationAtInstructionAddress(d.session.Program(), pc)
	if err != nil {
		return &uiDebugger.SourceResult{
			Error: fmt.Errorf("failed to get current source location: %w", err),
		}
	}

	return d.Source(&uiDebugger.SourceArgs{
		Location: &uiDebugger.SourceLocation{
			File: currentSourceLoc.File.Path(),
			Line: currentSourceLoc.Line,
		},
		ContextLines: args.ContextLines,
		ContextMode:  args.ContextMode,
	})
}

func (d *debugger) Source(args *uiDebugger.SourceArgs) *uiDebugger.SourceResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.SourceResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	// Get current PC to mark the current line
	pc, err := cpu.ReadPC(debugger.Runtime().CPU().Registers())
	if err != nil {
		return &uiDebugger.SourceResult{
			Error: fmt.Errorf("failed to read PC register: %w", err),
		}
	}

	var sourceRange sourcecode.Range
	sourceRange.File = sourcecode.FileNamed(args.Location.File)
	sourceRange.LineCount = args.ContextLines

	switch args.ContextMode {
	case uiDebugger.SourceContextTop:
		sourceRange.StartLine = args.Location.Line
	case uiDebugger.SourceContextCentered:
		sourceRange.StartLine = args.Location.Line - args.ContextLines/2
		if sourceRange.StartLine < 1 {
			sourceRange.StartLine = 1
		}
	case uiDebugger.SourceContextBottom:
		sourceRange.StartLine = args.Location.Line - args.ContextLines + 1
		if sourceRange.StartLine < 1 {
			sourceRange.StartLine = 1
		}
	default:
		return &uiDebugger.SourceResult{
			Error: fmt.Errorf("unsupported context mode: %d", args.ContextMode),
		}
	}

	snippet, err := sourcecode.ReadSnippet(debugger.Program().DebugInfo().SourceLibrary, &sourceRange)
	if err != nil {
		return &uiDebugger.SourceResult{
			Error: err,
		}
	}

	uiSnippet := SourceCodeSnippetToUI(snippet)

	// Mark lines that contain the current PC
	if uiSnippet != nil && uiSnippet.Lines != nil {
		for _, line := range uiSnippet.Lines {
			if line.Location != nil {
				// Check if this source line contains the current PC
				if srcLoc, err := program.SourceLocationAtInstructionAddress(debugger.Program(), pc); err == nil && srcLoc != nil {
					if srcLoc.File.Path() == line.Location.File && srcLoc.Line == line.Location.Line {
						line.IsCurrent = true
					}
				}
			}
		}
	}

	return &uiDebugger.SourceResult{
		Error:   err,
		Snippet: uiSnippet,
	}
}

func (d *debugger) Break(args *uiDebugger.BreakArgs) *uiDebugger.BreakResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.BreakResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	if args.Address != nil {
		val, err := core.Eval(debugger.Runtime(), debugger.Program(), *args.Address)
		if err != nil {
			return &uiDebugger.BreakResult{
				Error: fmt.Errorf("failed to evaluate address expression '%s': %w", args.Address, err),
			}
		}

		if bp, err := debugger.AddBreakpoint(val); err != nil {
			return &uiDebugger.BreakResult{
				Error: err,
			}
		} else {
			return &uiDebugger.BreakResult{
				Breakpoint: BreakpointToUI(bp),
			}
		}
	}

	if args.SourceLocation != nil {
		address, err := program.InstructionAddressAtSourceLocation(d.session.Program(), SourceLocationFromUI(args.SourceLocation))
		if err != nil {
			return &uiDebugger.BreakResult{
				Error: fmt.Errorf("failed to resolve source location to instruction address: %w", err),
			}
		}

		if bp, err := debugger.AddBreakpoint(address); err != nil {
			return &uiDebugger.BreakResult{
				Error: fmt.Errorf("failed to add breakpoint at address 0x%X: %w", address, err),
			}
		} else {
			return &uiDebugger.BreakResult{
				Breakpoint: BreakpointToUI(bp),
			}
		}
	}

	return &uiDebugger.BreakResult{
		Error: fmt.Errorf("either address or source location must be provided to set a breakpoint"),
	}
}

func (d *debugger) Watch(args *uiDebugger.WatchArgs) *uiDebugger.WatchResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.WatchResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	startAddress, err := core.Eval(debugger.Runtime(), debugger.Program(), args.StartAddress)
	if err != nil {
		return &uiDebugger.WatchResult{
			Error: fmt.Errorf("failed to evaluate address expression '%s': %w", args.StartAddress, err),
		}
	}

	var size uint32 = 4 // Default watch size is 4 bytes

	if args.EndAddress != nil {
		if args.Size != nil {
			return &uiDebugger.WatchResult{
				Error: fmt.Errorf("cannot specify both end address and size for watchpoint"),
			}
		}

		addr, err := core.Eval(debugger.Runtime(), debugger.Program(), *args.EndAddress)
		if err != nil {
			return &uiDebugger.WatchResult{
				Error: fmt.Errorf("failed to evaluate end address expression '%s': %w", *args.EndAddress, err),
			}
		}

		size = addr - startAddress
	}

	if args.Type == nil {
		// Default to read/write watchpoint if type is not specified
		args.Type = utils.Ptr(uiDebugger.WatchpointTypeReadWrite)
	}

	watchpoint, err := debugger.AddWatchpoint(&memory.Range{Start: startAddress, Size: size}, WatchpointTypeFromUI(*args.Type))
	return &uiDebugger.WatchResult{
		Error:      err,
		Watchpoint: WatchpointToUI(watchpoint),
	}
}

func (d *debugger) RemoveBreakpoint(args *uiDebugger.RemoveBreakpointArgs) *uiDebugger.RemoveBreakpointResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.RemoveBreakpointResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	bp, err := debugger.RemoveBreakpoint(args.ID)
	return &uiDebugger.RemoveBreakpointResult{
		Error:      err,
		Breakpoint: BreakpointToUI(bp),
	}
}

func (d *debugger) RemoveWatchpoint(args *uiDebugger.RemoveWatchpointArgs) *uiDebugger.RemoveWatchpointResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.RemoveWatchpointResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	wp, err := debugger.RemoveWatchpoint(args.ID)
	return &uiDebugger.RemoveWatchpointResult{
		Error:      err,
		Watchpoint: WatchpointToUI(wp),
	}
}

func (d *debugger) CurrentInstruction() *uiDebugger.CurrentInstructionResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.CurrentInstructionResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	pc, err := cpu.ReadPC(debugger.Runtime().CPU().Registers())
	if err != nil {
		return &uiDebugger.CurrentInstructionResult{
			Error: fmt.Errorf("failed to read PC register: %w", err),
		}
	}

	instr, err := program.InstructionAtAddress(debugger.Program(), pc)
	return &uiDebugger.CurrentInstructionResult{
		Error:       err,
		Instruction: d.uiInstruction(debugger, instr),
	}
}

func (d *debugger) uiBreakpoint(debugger core.Debugger, bp *core.Breakpoint) *uiDebugger.Breakpoint {
	if bp == nil {
		return nil
	}

	result := BreakpointToUI(bp)

	srcLoc, _ := program.SourceLocationAtInstructionAddress(debugger.Program(), bp.Address)
	result.Location = SourceLocationToUI(srcLoc)

	return result
}

func (d *debugger) uiInstruction(debugger core.Debugger, instr *program.Instruction) *uiDebugger.Instruction {
	return d.uiInstructionWithPrevious(debugger, instr, nil)
}

func (d *debugger) uiSourceLine(debugger core.Debugger, srcLine *sourcecode.Line, instr *program.Instruction, pc uint32) *uiDebugger.SourceLine {
	if srcLine == nil {
		return nil
	}

	result := SourceLineToUI(srcLine)
	result.IsCurrent = (*instr.Address == pc)

	if bp := debugger.GetBreakpointAt(*instr.Address); bp != nil {
		result.Breakpoints = []*uiDebugger.Breakpoint{BreakpointToUI(bp)}
	}

	return result
}

func (d *debugger) uiInstructionWithPrevious(debugger core.Debugger, instr *program.Instruction, prevSourceLoc *sourcecode.Location) *uiDebugger.Instruction {
	if instr == nil {
		return nil
	}

	if instr.Address == nil {
		panic("program instruction has no resolved address????")
	}

	pc, err := cpu.ReadPC(debugger.Runtime().CPU().Registers())
	if err != nil {
		panic("failed to read PC register: " + err.Error())
	}

	// Get the source location for this specific instruction
	srcLoc, _ := program.SourceLocationAtInstructionAddress(debugger.Program(), *instr.Address)

	// Get the source line information if available
	// Skip expensive source extraction if location matches previous instruction
	var sourceLine *uiDebugger.SourceLine
	shouldExtractSource := srcLoc != nil && (prevSourceLoc == nil || !sourceLocationsEqual(srcLoc, prevSourceLoc))

	if shouldExtractSource {
		if srcLineObj, err := program.SourceLineAtInstructionAddress(debugger.Program(), *instr.Address); err == nil {
			sourceLine = d.uiSourceLine(debugger, srcLineObj, instr, pc)
		}
	}

	branchTarget, targetSymbol, _ := program.BranchTargetAtInstruction(debugger.Program(), *instr.Address)
	uiInstr := InstructionToUI(instr)
	uiInstr.IsCurrentPC = (*instr.Address == pc)
	uiInstr.BranchTarget = branchTarget
	if targetSymbol != nil {
		uiInstr.BranchTargetSym = &targetSymbol.Name
	}
	uiInstr.Breakpoints = []*uiDebugger.Breakpoint{BreakpointToUI(debugger.GetBreakpointAt(pc))}
	uiInstr.SourceLine = sourceLine

	return uiInstr
}

func sourceLocationsEqual(loc1, loc2 *sourcecode.Location) bool {
	if loc1 == nil || loc2 == nil {
		return loc1 == loc2
	}
	return loc1.File.Path() == loc2.File.Path() && loc1.Line == loc2.Line
}

func (d *debugger) Disasm(args *uiDebugger.DisasmArgs) *uiDebugger.DisasmResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.DisasmResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	// Evaluate the address expression
	var address uint32
	if args.Address != "" {
		val, err := core.Eval(debugger.Runtime(), debugger.Program(), args.Address)
		if err != nil {
			return &uiDebugger.DisasmResult{
				Error: fmt.Errorf("failed to evaluate address expression '%s': %w", args.Address, err),
			}
		}
		address = val
	}

	count := 10 // Default instruction count
	if args.CountExpr != nil {
		val, err := core.Eval(debugger.Runtime(), debugger.Program(), *args.CountExpr)
		if err != nil {
			return &uiDebugger.DisasmResult{
				Error: fmt.Errorf("failed to evaluate count expression '%s': %w", *args.CountExpr, err),
			}
		}
		count = int(val)
	}

	instructions, err := program.InstructionsAtAddress(debugger.Program(), address, count)
	if err != nil {
		return &uiDebugger.DisasmResult{
			Error: err,
		}
	}

	result := make([]*uiDebugger.Instruction, 0, len(instructions))
	var prevSourceLoc *sourcecode.Location

	for _, instr := range instructions {
		uiInstr := d.uiInstructionWithPrevious(debugger, instr, prevSourceLoc)
		result = append(result, uiInstr)

		// Update previous source location for next iteration
		// Get the actual source location to track it for next instruction
		srcLoc, _ := program.SourceLocationAtInstructionAddress(debugger.Program(), *instr.Address)
		prevSourceLoc = srcLoc
	}

	// Compute the control flow graph for this instruction range
	cfg := d.computeControlFlowGraphForInstructions(debugger, result)

	return &uiDebugger.DisasmResult{
		Error:            err,
		Instructions:     result,
		ControlFlowGraph: cfg,
	}
}

// computeControlFlowGraphForInstructions builds a CFG from the given instructions
func (d *debugger) computeControlFlowGraphForInstructions(debuggerIface core.Debugger, instructions []*uiDebugger.Instruction) *uiDebugger.ControlFlowGraph {
	if len(instructions) == 0 {
		return &uiDebugger.ControlFlowGraph{Edges: make(map[uint32]uint32)}
	}

	// Build a memory range from the instructions
	minAddr := instructions[0].Address
	maxAddr := minAddr
	for _, instr := range instructions {
		if instr.Address > maxAddr {
			maxAddr = instr.Address
		}
	}

	// Create a memory range covering all instructions (assuming 4-byte instructions)
	region := &memory.Range{
		Start: minAddr,
		Size:  maxAddr - minAddr + 4,
	}

	// Build the CFG from the debugger interface
	cfg, err := debuggerIface.BuildControlFlowGraph(region)
	if err != nil {
		// Return empty CFG on error
		return &uiDebugger.ControlFlowGraph{Edges: make(map[uint32]uint32)}
	}

	// Convert the core CFG to UI CFG
	edges := make(map[uint32]uint32)
	for sourceAddr := range cfg.AllSources() {
		target := cfg.Target(sourceAddr)
		if target != nil {
			edges[sourceAddr] = *target
		}
	}

	return &uiDebugger.ControlFlowGraph{Edges: edges}
}

func (d *debugger) Eval(args *uiDebugger.EvalArgs) *uiDebugger.EvalResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.EvalResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	value, err := core.Eval(debugger.Runtime(), debugger.Program(), args.Expression)
	return &uiDebugger.EvalResult{
		Error: err,
		Value: value,
	}
}

func (d *debugger) Symbols(args *uiDebugger.SymbolsArgs) *uiDebugger.SymbolsResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.SymbolsResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	prog := debugger.Program()
	filter := ""
	if args.SymbolName != nil {
		filter = *args.SymbolName
	}

	result := &uiDebugger.SymbolsResult{
		Functions: []*uiDebugger.FunctionSymbol{},
		Globals:   []*uiDebugger.GlobalSymbol{},
		Labels:    []*uiDebugger.LabelSymbol{},
	}

	// Collect functions that match the filter
	for name, fn := range prog.Functions() {
		if filter != "" && !matchesFilter(name, filter) {
			continue
		}

		// Calculate size as instruction range sum
		var size *uint32
		if len(fn.InstructionRanges) > 0 {
			totalSize := uint32(0)
			for _, ir := range fn.InstructionRanges {
				totalSize += uint32(ir.Count)
			}
			size = &totalSize
		}

		// Format instruction ranges for display
		rangeStrs := make([]string, len(fn.InstructionRanges))
		for i, ir := range fn.InstructionRanges {
			rangeStrs[i] = fmt.Sprintf("%d-%d", ir.Start, ir.Start+ir.Count-1)
		}

		// Get start address from first instruction if available
		var startAddr *uint32
		if len(fn.InstructionRanges) > 0 {
			instructions := prog.Instructions()
			if fn.InstructionRanges[0].Start < len(instructions) {
				startAddr = instructions[fn.InstructionRanges[0].Start].Address
			}
		}

		result.Functions = append(result.Functions, &uiDebugger.FunctionSymbol{
			Name:              name,
			Address:           startAddr,
			Size:              size,
			SourceFile:        fn.SourceFile,
			StartLine:         fn.StartLine,
			EndLine:           fn.EndLine,
			InstructionRanges: rangeStrs,
		})
	}

	// Collect globals that match the filter
	for _, global := range prog.Globals() {
		if filter != "" && !matchesFilter(global.Name, filter) {
			continue
		}

		hasInitData := len(global.InitialData) > 0
		result.Globals = append(result.Globals, &uiDebugger.GlobalSymbol{
			Name:        global.Name,
			Address:     global.Address,
			Size:        global.Size,
			SymbolType:  global.Type.String(),
			HasInitData: hasInitData,
			InitDataLen: len(global.InitialData),
		})
	}

	// Collect labels that match the filter
	for _, label := range prog.Labels() {
		if filter != "" && !matchesFilter(label.Name, filter) {
			continue
		}

		var address *uint32
		if label.InstructionIndex >= 0 {
			instructions := prog.Instructions()
			if label.InstructionIndex < len(instructions) {
				address = instructions[label.InstructionIndex].Address
			}
		}

		result.Labels = append(result.Labels, &uiDebugger.LabelSymbol{
			Name:             label.Name,
			InstructionIndex: label.InstructionIndex,
			Address:          address,
		})
	}

	// Calculate total count
	result.TotalCount = len(result.Functions) + len(result.Globals) + len(result.Labels)

	return result
}

func (d *debugger) Info(args *uiDebugger.InfoArgs) *uiDebugger.InfoResult {
	switch args.Type {
	case uiDebugger.InfoTypeGeneral:
		return d.infoGeneral()
	case uiDebugger.InfoTypeRuntime:
		return d.infoRuntime()
	case uiDebugger.InfoTypeProgram:
		return d.infoProgram()
	default:
		return d.infoGeneral()
	}
}

func (d *debugger) infoGeneral() *uiDebugger.InfoResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		var status uiDebugger.DebuggerStatus
		if d.session.System() == nil {
			status = uiDebugger.DebuggerStatusNotReady_MissingSystemConfig
		} else if d.session.Runtime() == nil {
			status = uiDebugger.DebuggerStatusNotReady_MissingRuntime
		} else if d.session.Program() == nil {
			status = uiDebugger.DebuggerStatusNotReady_MissingProgram
		}

		return &uiDebugger.InfoResult{
			DebuggerState: &uiDebugger.DebuggerState{
				Status: status,
			},
		}
	}

	// Use detailed status determination based on runner state and last execution result
	status := DetermineDetailedDebuggerStatus(debugger)

	registers := core.NewRegisters(debugger.Runtime())
	cpsr := registers.ReadCPSR()

	return &uiDebugger.InfoResult{
		DebuggerState: &uiDebugger.DebuggerState{
			Status:    status,
			Registers: RegistersValuesToUI(registers.ReadAllRegisters()),
			Flags: &uiDebugger.FlagState{
				N: instructions.TestCPSRFlag(cpsr, instructions.FLAG_N),
				Z: instructions.TestCPSRFlag(cpsr, instructions.FLAG_Z),
				C: instructions.TestCPSRFlag(cpsr, instructions.FLAG_C),
				V: instructions.TestCPSRFlag(cpsr, instructions.FLAG_V),
			},
		},
	}
}

func (d *debugger) infoRuntime() *uiDebugger.InfoResult {
	if d.session.System() == nil {
		return &uiDebugger.InfoResult{
			Error: fmt.Errorf("system not configured"),
		}
	}

	if d.session.Runtime() == nil {
		return &uiDebugger.InfoResult{
			Error: fmt.Errorf("runtime not loaded"),
		}
	}

	return &uiDebugger.InfoResult{
		RuntimeInfo: &uiDebugger.RuntimeInfo{
			Runtime: d.runtimeType,
		},
	}
}

func (d *debugger) infoProgram() *uiDebugger.InfoResult {
	if d.session.Program() == nil {
		return &uiDebugger.InfoResult{
			Error: fmt.Errorf("program not loaded"),
		}
	}

	prog := d.session.Program()

	// Get entry point
	entryPoint, err := program.ProgramEntryPoint(prog)
	if err != nil {
		// If we can't get the entry point, continue anyway (entry point will be 0)
		_ = err // Ignore error, entry point will be 0 which is fine
	}

	// Check if debug info is available
	hasDebugInfo := prog.DebugInfo() != nil

	// Get source file name
	sourceFile := prog.SourceFile()

	return &uiDebugger.InfoResult{
		ProgramInfo: &uiDebugger.ProgramInfo{
			SourceFile:   &sourceFile,
			ObjectFile:   nil, // Object file info not easily available from ProgramFile interface
			EntryPoint:   entryPoint,
			HasDebugInfo: hasDebugInfo,
			Warnings:     nil,
		},
	}
}

func (d *debugger) List() *uiDebugger.ListResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.ListResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	breakpoints := debugger.ListBreakpoints()
	watchpoints := debugger.ListWatchpoints()

	return &uiDebugger.ListResult{
		Error:       nil,
		Breakpoints: utils.Map(breakpoints, func(bp *core.Breakpoint) *uiDebugger.Breakpoint { return d.uiBreakpoint(debugger, bp) }),
		Watchpoints: utils.Map(watchpoints, WatchpointToUI),
	}
}

func (d *debugger) Run() *uiDebugger.ExecutionResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.ExecutionResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	return ExecutionResultToUI(debugger.Continue())
}

func (d *debugger) Memory(args *uiDebugger.MemoryArgs) *uiDebugger.MemoryResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.MemoryResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	// Evaluate the address expression
	var address uint32
	if args.AddressExpr != "" {
		val, err := core.Eval(debugger.Runtime(), debugger.Program(), args.AddressExpr)
		if err != nil {
			return &uiDebugger.MemoryResult{
				Error: fmt.Errorf("failed to evaluate address expression '%s': %w", args.AddressExpr, err),
			}
		}
		address = val
	}

	// Get memory contents using the memory package functions
	mem := debugger.Runtime().Memory()
	count := 256 // Default to 256 bytes
	if args.CountExpr != nil {
		val, err := core.Eval(debugger.Runtime(), debugger.Program(), *args.CountExpr)
		if err != nil {
			return &uiDebugger.MemoryResult{
				Error: fmt.Errorf("failed to evaluate count expression '%s': %w", *args.CountExpr, err),
			}
		}
		count = int(val)
	}

	// Read all memory at once for efficiency
	data, err := memory.Read(mem, address, count)
	if err != nil {
		return &uiDebugger.MemoryResult{
			Error: fmt.Errorf("failed to read memory at 0x%x: %w", address, err),
		}
	}

	// Read memory regions
	regions := make([]*uiDebugger.MemoryRegion, 0)
	if data != nil && len(data) > 0 {
		regions = append(regions, &uiDebugger.MemoryRegion{
			Start: address,
			Size:  uint32(len(data)),
		})
	}

	return &uiDebugger.MemoryResult{
		Error:   nil,
		Regions: regions,
		Data:    data,
		Address: address,
	}
}

func (d *debugger) Registers() *uiDebugger.RegistersResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.RegistersResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	registers := core.NewRegisters(debugger.Runtime())
	cpsr := registers.ReadCPSR()

	// Get register values as map
	registerMap := RegistersValuesToUI(registers.ReadAllRegisters())

	return &uiDebugger.RegistersResult{
		Error:     nil,
		Registers: registerMap,
		Flags: &uiDebugger.FlagState{
			N: instructions.TestCPSRFlag(cpsr, instructions.FLAG_N),
			Z: instructions.TestCPSRFlag(cpsr, instructions.FLAG_Z),
			C: instructions.TestCPSRFlag(cpsr, instructions.FLAG_C),
			V: instructions.TestCPSRFlag(cpsr, instructions.FLAG_V),
		},
	}
}

func (d *debugger) Stack() *uiDebugger.StackResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.StackResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	// Get stack frames - for now return minimal information
	frames := make([]*uiDebugger.StackFrame, 0)

	// Try to get current frame information
	pc, err := cpu.ReadPC(debugger.Runtime().CPU().Registers())
	if err != nil {
		return &uiDebugger.StackResult{
			Error: fmt.Errorf("failed to read PC: %w", err),
		}
	}

	sp, err := cpu.ReadSP(debugger.Runtime().CPU().Registers())
	if err != nil {
		return &uiDebugger.StackResult{
			Error: fmt.Errorf("failed to read SP: %w", err),
		}
	}

	// Get source location at current PC
	srcLoc, _ := program.SourceLocationAtInstructionAddress(d.session.Program(), pc)
	frames = append(frames, &uiDebugger.StackFrame{
		SourceLocation: SourceLocationToUI(srcLoc),
		Memory: &uiDebugger.MemoryRegion{
			Start: sp,
			Size:  1,
		},
	})

	return &uiDebugger.StackResult{
		Error:       nil,
		StackFrames: frames,
	}
}

func (d *debugger) Vars() *uiDebugger.VarsResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.VarsResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	// Get current PC
	pc, err := cpu.ReadPC(debugger.Runtime().CPU().Registers())
	if err != nil {
		return &uiDebugger.VarsResult{
			Error: fmt.Errorf("failed to read PC: %w", err),
		}
	}

	// For now, return empty variable list
	// Full variable extraction would require debug info parsing
	_ = pc // use pc to avoid unused variable
	vars := make([]*uiDebugger.VariableValue, 0)

	return &uiDebugger.VarsResult{
		Error:     nil,
		Variables: vars,
	}
}

func (d *debugger) Reset() *uiDebugger.ExecutionResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.ExecutionResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	return ExecutionResultToUI(debugger.Reset())
}

func (d *debugger) Restart() *uiDebugger.ExecutionResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &uiDebugger.ExecutionResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	// Reset the debugger
	resetResult := debugger.Reset()
	if resetResult.Error != nil {
		return ExecutionResultToUI(resetResult)
	}

	// Then continue execution
	return ExecutionResultToUI(debugger.Continue())
}

// matchesFilter checks if a symbol name matches the filter pattern
// Filter matching is substring-based (case-sensitive)
func matchesFilter(name, filter string) bool {
	if filter == "" {
		return true
	}
	return strings.Contains(name, filter)
}
