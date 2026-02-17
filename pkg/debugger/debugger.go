package debugger

import (
	"fmt"
	"os"

	"github.com/Manu343726/cucaracha/pkg/debugger/core"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/interpreter"
	"github.com/Manu343726/cucaracha/pkg/runtime/program"
	"github.com/Manu343726/cucaracha/pkg/runtime/program/loader"
	"github.com/Manu343726/cucaracha/pkg/runtime/program/sourcecode"
	"github.com/Manu343726/cucaracha/pkg/ui"
	"github.com/Manu343726/cucaracha/pkg/utils"
	"gopkg.in/yaml.v3"
)

type debugger struct {
	session *core.Session
}

func NewDebugger() Debugger {
	return &debugger{
		session: &core.Session{},
	}
}

func (d *debugger) SetEventCallback(callback ui.DebuggerEventCallback) {
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

func (d *debugger) LoadRuntime(args *ui.LoadRuntimeArgs) *ui.LoadRuntimeResult {
	switch args.Runtime {
	case ui.RuntimeTypeInterpreter:
		if d.session.System() == nil {
			return &ui.LoadRuntimeResult{
				Error: fmt.Errorf("system must be configured before loading interpreter runtime"),
			}
		}

		r, err := interpreter.NewInterpreter(d.session.System())
		if err != nil {
			return &ui.LoadRuntimeResult{
				Error: fmt.Errorf("failed to create interpreter runtime: %w", err),
			}
		}

		return &ui.LoadRuntimeResult{
			Error: d.session.LoadRuntime(r),
		}
	default:
		return &ui.LoadRuntimeResult{
			Error: fmt.Errorf("unsupported runtime: %s", args.Runtime),
		}
	}
}

// extractSystemInfo builds a LoadSystemResult with system information
func (d *debugger) extractSystemInfo(err error) *ui.LoadSystemResult {
	result := &ui.LoadSystemResult{
		Error: err,
	}

	// Only populate system info if no error and system was loaded successfully
	if err == nil && d.session.System() != nil {
		sys := d.session.System()
		layout := sys.MemoryLayout

		result.TotalMemory = layout.TotalSize
		result.CodeSize = layout.CodeSize
		result.DataSize = layout.DataSize
		result.StackSize = layout.StackSize
		result.HeapSize = layout.HeapSize
		result.PeripheralSize = layout.PeripheralSize
		result.NumberOfVectors = sys.VectorTable.NumberOfVectors
		result.VectorEntrySize = sys.VectorTable.VectorEntrySize
		result.NumPeripherals = len(sys.Peripherals)

		// Extract peripheral information
		result.Peripherals = make([]ui.PeripheralInfo, len(sys.Peripherals))
		for i, p := range sys.Peripherals {
			metadata := p.Metadata()
			typeStr := ""
			displayName := ""
			if metadata.Descriptor != nil {
				typeStr = metadata.Descriptor.Type.String()
				displayName = metadata.Descriptor.DisplayName
			}
			result.Peripherals[i] = ui.PeripheralInfo{
				Name:            metadata.Name,
				Type:            typeStr,
				DisplayName:     displayName,
				Description:     metadata.Description,
				BaseAddress:     metadata.BaseAddress,
				Size:            metadata.Size,
				InterruptVector: metadata.InterruptVector,
			}
		}
	}

	return result
}

func (d *debugger) LoadSystemFromFile(args *ui.LoadSystemArgs) *ui.LoadSystemResult {
	if args.FilePath == "default" {
		return d.LoadSystemFromEmbedded()
	}

	err := d.session.LoadSystemFromFile(args.FilePath)
	return d.extractSystemInfo(err)
}

func (d *debugger) LoadSystemFromEmbedded() *ui.LoadSystemResult {
	config, err := DefaultSystemConfig()
	if err != nil {
		return &ui.LoadSystemResult{
			Error: fmt.Errorf("failed to load embedded system config: %w", err),
		}
	}

	system, err := config.Setup()
	if err != nil {
		return &ui.LoadSystemResult{
			Error: fmt.Errorf("failed to setup system from embedded config: %w", err),
		}
	}

	err = d.session.LoadSystem(system)
	return d.extractSystemInfo(err)
}

func (d *debugger) LoadProgramFromFile(args *ui.LoadProgramArgs) *ui.LoadProgramResult {
	if d.session.System() == nil {
		return &ui.LoadProgramResult{
			Error: fmt.Errorf("system must be configured before loading a program"),
		}
	}

	options := &loader.Options{
		Verbose:        false,
		MemoryLayout:   &d.session.System().MemoryLayout,
		OutputFormat:   "object",
		AutoBuildClang: true,
	}

	loadedProgram, err := d.session.LoadProgramFromFile(args.FilePath, options)
	if err != nil {
		return &ui.LoadProgramResult{
			Error: err,
		}
	}

	return &ui.LoadProgramResult{
		Warnings:   loadedProgram.Warnings,
		SourceFile: &loadedProgram.OriginalPath,
		ObjectFile: &loadedProgram.CompiledPath,
	}
}

type allFile struct {
	Runtime     ui.RuntimeType `json:"runtime"`
	ProgramFile string         `json:"programFile"`
}

func (d *debugger) loadFromSingleFile(fullDescriptorPath string) *ui.LoadResult {
	raw, err := os.ReadFile(fullDescriptorPath)
	if err != nil {
		return &ui.LoadResult{
			Error: fmt.Errorf("failed to read file '%s': %w", fullDescriptorPath, err),
		}
	}

	var all allFile
	if err := yaml.Unmarshal(raw, &all); err != nil {
		return &ui.LoadResult{
			Error: fmt.Errorf("failed to parse YAML file '%s': %w", fullDescriptorPath, err),
		}
	}

	return d.Load(&ui.LoadArgs{
		Runtime:          utils.Ptr(all.Runtime),
		SystemConfigPath: utils.Ptr(all.ProgramFile),
		ProgramPath:      utils.Ptr(all.ProgramFile),
	})
}

func (d *debugger) Load(args *ui.LoadArgs) *ui.LoadResult {
	if args.FullDescriptorPath != nil {
		return d.loadFromSingleFile(*args.FullDescriptorPath)
	}

	if args.SystemConfigPath == nil {
		args.SystemConfigPath = utils.Ptr("default")
	}

	loadSysResult := d.LoadSystemFromFile(&ui.LoadSystemArgs{
		FilePath: *args.SystemConfigPath,
	})
	if loadSysResult.Error != nil {
		return &ui.LoadResult{
			Error: fmt.Errorf("failed to load system from file '%s': %w", *args.SystemConfigPath, loadSysResult.Error),
		}
	}

	if args.Runtime == nil {
		args.Runtime = utils.Ptr(ui.RuntimeTypeInterpreter)
	}

	loadRuntimeResult := d.LoadRuntime(&ui.LoadRuntimeArgs{
		Runtime: *args.Runtime,
	})
	if loadRuntimeResult.Error != nil {
		return &ui.LoadResult{
			Error: fmt.Errorf("failed to load runtime '%s': %w", *args.Runtime, loadRuntimeResult.Error),
		}
	}

	loadProgResult := d.LoadProgramFromFile(&ui.LoadProgramArgs{
		FilePath: *args.ProgramPath,
	})
	if loadProgResult.Error != nil {
		return &ui.LoadResult{
			Error: fmt.Errorf("failed to load program from file '%s': %w", *args.ProgramPath, loadProgResult.Error),
		}
	}

	return &ui.LoadResult{
		System:  loadSysResult,
		Program: loadProgResult,
		Runtime: loadRuntimeResult,
	}
}

func (d *debugger) Continue() *ui.ExecutionResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &ui.ExecutionResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	return ExecutionResultToUI(debugger.Continue())
}

func (d *debugger) Interrupt() *ui.ExecutionResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &ui.ExecutionResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	return ExecutionResultToUI(debugger.Interrupt())
}

func (d *debugger) Step(args *ui.StepArgs) *ui.ExecutionResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &ui.ExecutionResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	switch args.StepMode {
	case ui.StepModeInto:
		switch args.CountMode {
		case ui.StepCountInstructions:
			return ExecutionResultToUI(debugger.Step())
		case ui.StepCountSourceLines:
			return ExecutionResultToUI(debugger.StepIntoSource())
		default:
			return &ui.ExecutionResult{
				Error: fmt.Errorf("unsupported single step count mode '%s'", args.CountMode),
			}
		}
	case ui.StepModeOver:
		switch args.CountMode {
		case ui.StepCountInstructions:
			return ExecutionResultToUI(debugger.StepOver())
		case ui.StepCountSourceLines:
			return ExecutionResultToUI(debugger.StepOverSource())
		default:
			return &ui.ExecutionResult{
				Error: fmt.Errorf("unsupported step over count mode '%s'", args.CountMode),
			}
		}
	case ui.StepModeOut:
		return ExecutionResultToUI(debugger.StepOut())
	default:
		return &ui.ExecutionResult{
			Error: fmt.Errorf("unsupported step mode: %d", args.StepMode),
		}
	}
}

func (d *debugger) CurrentSource(args *ui.CurrentSourceArgs) *ui.SourceResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &ui.SourceResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	pc, err := cpu.ReadPC(debugger.Runtime().CPU().Registers())
	if err != nil {
		return &ui.SourceResult{
			Error: fmt.Errorf("failed to read PC register: %w", err),
		}
	}

	currentSourceLoc, err := program.SourceLocationAtInstructionAddress(d.session.Program(), pc)
	if err != nil {
		return &ui.SourceResult{
			Error: fmt.Errorf("failed to get current source location: %w", err),
		}
	}

	return d.Source(&ui.SourceArgs{
		File:         currentSourceLoc.File.Path(),
		Line:         currentSourceLoc.Line,
		ContextLines: args.ContextLines,
		ContextMode:  args.ContextMode,
	})
}

func (d *debugger) Source(args *ui.SourceArgs) *ui.SourceResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &ui.SourceResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	var sourceRange sourcecode.Range
	sourceRange.File = sourcecode.FileNamed(args.File)
	sourceRange.LineCount = args.ContextLines

	switch args.ContextMode {
	case ui.SourceContextTop:
		sourceRange.StartLine = args.Line
	case ui.SourceContextCentered:
		sourceRange.StartLine = args.Line - args.ContextLines/2
		if sourceRange.StartLine < 1 {
			sourceRange.StartLine = 1
		}
	case ui.SourceContextBottom:
		sourceRange.StartLine = args.Line - args.ContextLines + 1
		if sourceRange.StartLine < 1 {
			sourceRange.StartLine = 1
		}
	default:
		return &ui.SourceResult{
			Error: fmt.Errorf("unsupported context mode: %d", args.ContextMode),
		}
	}

	snippet, err := sourcecode.ReadSnippet(debugger.Program().DebugInfo().SourceLibrary, &sourceRange)
	return &ui.SourceResult{
		Error:   err,
		Snippet: SourceCodeSnippetToUI(snippet),
	}
}

func (d *debugger) Break(args *ui.BreakArgs) *ui.BreakResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &ui.BreakResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	if args.Address != nil {
		if bp, err := debugger.AddBreakpoint(*args.Address); err != nil {
			return &ui.BreakResult{
				Error: err,
			}
		} else {
			return &ui.BreakResult{
				Breakpoint: BreakpointToUI(bp),
			}
		}
	}

	if args.SourceLocation != nil {
		address, err := program.InstructionAddressAtSourceLocation(d.session.Program(), SourceLocationFromUI(args.SourceLocation))
		if err != nil {
			return &ui.BreakResult{
				Error: fmt.Errorf("failed to resolve source location to instruction address: %w", err),
			}
		}

		if bp, err := debugger.AddBreakpoint(address); err != nil {
			return &ui.BreakResult{
				Error: fmt.Errorf("failed to add breakpoint at address 0x%X: %w", address, err),
			}
		} else {
			return &ui.BreakResult{
				Breakpoint: BreakpointToUI(bp),
			}
		}
	}

	return &ui.BreakResult{
		Error: fmt.Errorf("either address or source location must be provided to set a breakpoint"),
	}
}

func (d *debugger) Watch(args *ui.WatchArgs) *ui.WatchResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &ui.WatchResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	watchpoint, err := debugger.AddWatchpoint(MemoryRangeFromUI(args.Range), WatchpointTypeFromUI(args.Type))
	return &ui.WatchResult{
		Error:      err,
		Watchpoint: WatchpointToUI(watchpoint),
	}
}

func (d *debugger) RemoveBreakpoint(args *ui.RemoveBreakpointArgs) *ui.RemoveBreakpointResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &ui.RemoveBreakpointResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	bp, err := debugger.RemoveBreakpoint(args.ID)
	return &ui.RemoveBreakpointResult{
		Error:      err,
		Breakpoint: BreakpointToUI(bp),
	}
}

func (d *debugger) RemoveWatchpoint(args *ui.RemoveWatchpointArgs) *ui.RemoveWatchpointResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &ui.RemoveWatchpointResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	wp, err := debugger.RemoveWatchpoint(args.ID)
	return &ui.RemoveWatchpointResult{
		Error:      err,
		Watchpoint: WatchpointToUI(wp),
	}
}

func (d *debugger) CurrentInstruction() *ui.CurrentInstructionResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &ui.CurrentInstructionResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	pc, err := cpu.ReadPC(debugger.Runtime().CPU().Registers())
	if err != nil {
		return &ui.CurrentInstructionResult{
			Error: fmt.Errorf("failed to read PC register: %w", err),
		}
	}

	instr, err := program.InstructionAtAddress(debugger.Program(), pc)
	return &ui.CurrentInstructionResult{
		Error:       err,
		Instruction: d.uiInstruction(debugger, instr),
	}
}

func (d *debugger) uiBreakpoint(debugger core.Debugger, bp *core.Breakpoint) *ui.Breakpoint {
	if bp == nil {
		return nil
	}

	result := BreakpointToUI(bp)

	srcLoc, _ := program.SourceLocationAtInstructionAddress(debugger.Program(), bp.Address)
	result.Location = SourceLocationToUI(srcLoc)

	return result
}

func (d *debugger) uiInstruction(debugger core.Debugger, instr *program.Instruction) *ui.Instruction {
	if instr == nil {
		return nil
	}

	if instr.Address != nil {
		panic("program instruction has no resolved address????")
	}

	pc, err := cpu.ReadPC(debugger.Runtime().CPU().Registers())
	if err != nil {
		panic("failed to read PC register: " + err.Error())
	}

	srcLoc, _ := program.SourceLocationAtInstructionAddress(debugger.Program(), pc)

	branchTarget, targetSymbol, _ := program.BranchTargetAtInstruction(debugger.Program(), *instr.Address)
	uiInstr := InstructionToUI(instr)
	uiInstr.IsCurrentPC = (*instr.Address == pc)
	uiInstr.BranchTarget = branchTarget
	if targetSymbol != nil {
		uiInstr.BranchTargetSym = &targetSymbol.Name
	}
	uiInstr.Breakpoints = []*ui.Breakpoint{BreakpointToUI(debugger.GetBreakpointAt(pc))}
	uiInstr.SourceLocation = SourceLocationToUI(srcLoc)

	return uiInstr
}

func (d *debugger) Disasm(args *ui.DisasmArgs) *ui.DisassemblyResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &ui.DisassemblyResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	instructions, err := program.InstructionsAtAddress(debugger.Program(), args.Address, args.Count)
	return &ui.DisassemblyResult{
		Error: err,
		Instructions: utils.Map(instructions, func(instr *program.Instruction) *ui.Instruction {
			return d.uiInstruction(debugger, instr)
		}),
	}
}

func (d *debugger) Eval(args *ui.EvalArgs) *ui.EvalResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &ui.EvalResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	value, err := core.Eval(debugger.Runtime(), debugger.Program(), args.Expression)
	return &ui.EvalResult{
		Error: err,
		Value: value,
	}
}

func (d *debugger) Info() *ui.InfoResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		var status ui.DebuggerStatus
		if d.session.System() == nil {
			status = ui.DebuggerStatusNotReady_MissingSystemConfig
		} else if d.session.Runtime() == nil {
			status = ui.DebuggerStatusNotReady_MissingRuntime
		} else if d.session.Program() == nil {
			status = ui.DebuggerStatusNotReady_MissingProgram
		}

		return &ui.InfoResult{
			DebuggerState: &ui.DebuggerState{
				Status: status,
			},
		}
	}

	var status ui.DebuggerStatus

	if debugger.IsInterrupted() {
		status = ui.DebuggerStatusPaused
	} else {
		status = ui.DebuggerStatusRunning
	}

	registers := core.NewRegisters(debugger.Runtime())
	cpsr := registers.ReadCPSR()

	return &ui.InfoResult{
		DebuggerState: &ui.DebuggerState{
			Status:    status,
			Registers: RegistersValuesToUI(registers.ReadAllRegisters()),
			Flags: &ui.FlagState{
				N: instructions.TestCPSRFlag(cpsr, instructions.FLAG_N),
				Z: instructions.TestCPSRFlag(cpsr, instructions.FLAG_Z),
				C: instructions.TestCPSRFlag(cpsr, instructions.FLAG_C),
				V: instructions.TestCPSRFlag(cpsr, instructions.FLAG_V),
			},
		},
	}
}

func (d *debugger) List() *ui.ListResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &ui.ListResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	breakpoints := debugger.ListBreakpoints()
	watchpoints := debugger.ListWatchpoints()

	return &ui.ListResult{
		Error:       nil,
		Breakpoints: utils.Map(breakpoints, func(bp *core.Breakpoint) *ui.Breakpoint { return d.uiBreakpoint(debugger, bp) }),
		Watchpoints: utils.Map(watchpoints, WatchpointToUI),
	}
}

func (d *debugger) Run() *ui.ExecutionResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &ui.ExecutionResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	return ExecutionResultToUI(debugger.Continue())
}

func (d *debugger) Memory(args *ui.MemoryArgs) *ui.MemoryResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &ui.MemoryResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	// Evaluate the address expression
	var address uint32
	if args.AddressExpr != "" {
		val, err := core.Eval(debugger.Runtime(), debugger.Program(), args.AddressExpr)
		if err != nil {
			return &ui.MemoryResult{
				Error: fmt.Errorf("failed to evaluate address expression '%s': %w", args.AddressExpr, err),
			}
		}
		address = val
	}

	// Get memory contents using the memory package functions
	mem := debugger.Runtime().Memory()
	count := args.Count
	if count == 0 {
		count = 256 // Default to 256 bytes
	}

	// Read all memory at once for efficiency
	data, err := memory.Read(mem, address, count)
	if err != nil {
		return &ui.MemoryResult{
			Error: fmt.Errorf("failed to read memory at 0x%x: %w", address, err),
		}
	}

	// Read memory regions
	regions := make([]*ui.MemoryRegion, 0)
	if data != nil && len(data) > 0 {
		regions = append(regions, &ui.MemoryRegion{
			Start: address,
			Size:  uint32(len(data)),
		})
	}

	return &ui.MemoryResult{
		Error:   nil,
		Regions: regions,
	}
}

func (d *debugger) Registers() *ui.RegistersResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &ui.RegistersResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	registers := core.NewRegisters(debugger.Runtime())
	cpsr := registers.ReadCPSR()

	// Get register values as map
	registerMap := RegistersValuesToUI(registers.ReadAllRegisters())

	return &ui.RegistersResult{
		Error:     nil,
		Registers: registerMap,
		Flags: &ui.FlagState{
			N: instructions.TestCPSRFlag(cpsr, instructions.FLAG_N),
			Z: instructions.TestCPSRFlag(cpsr, instructions.FLAG_Z),
			C: instructions.TestCPSRFlag(cpsr, instructions.FLAG_C),
			V: instructions.TestCPSRFlag(cpsr, instructions.FLAG_V),
		},
	}
}

func (d *debugger) Stack() *ui.StackResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &ui.StackResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	// Get stack frames - for now return minimal information
	frames := make([]*ui.StackFrame, 0)

	// Try to get current frame information
	pc, err := cpu.ReadPC(debugger.Runtime().CPU().Registers())
	if err != nil {
		return &ui.StackResult{
			Error: fmt.Errorf("failed to read PC: %w", err),
		}
	}

	sp, err := cpu.ReadSP(debugger.Runtime().CPU().Registers())
	if err != nil {
		return &ui.StackResult{
			Error: fmt.Errorf("failed to read SP: %w", err),
		}
	}

	// Get source location at current PC
	srcLoc, _ := program.SourceLocationAtInstructionAddress(d.session.Program(), pc)
	frames = append(frames, &ui.StackFrame{
		SourceLocation: SourceLocationToUI(srcLoc),
		Memory: &ui.MemoryRegion{
			Start: sp,
			Size:  1,
		},
	})

	return &ui.StackResult{
		Error:       nil,
		StackFrames: frames,
	}
}

func (d *debugger) Vars() *ui.VarsResult {
	debugger, err := d.session.Debugger()
	if err != nil {
		return &ui.VarsResult{
			Error: fmt.Errorf("debugger not ready: %w", err),
		}
	}

	// Get current PC
	pc, err := cpu.ReadPC(debugger.Runtime().CPU().Registers())
	if err != nil {
		return &ui.VarsResult{
			Error: fmt.Errorf("failed to read PC: %w", err),
		}
	}

	// For now, return empty variable list
	// Full variable extraction would require debug info parsing
	_ = pc // use pc to avoid unused variable
	vars := make([]*ui.VariableValue, 0)

	return &ui.VarsResult{
		Error:     nil,
		Variables: vars,
	}
}
