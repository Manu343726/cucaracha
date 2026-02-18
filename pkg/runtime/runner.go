package runtime

import (
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/registers"
	"github.com/Manu343726/cucaracha/pkg/hw/memory"
	"github.com/Manu343726/cucaracha/pkg/hw/peripheral"
	"github.com/Manu343726/cucaracha/pkg/utils/contract"
)

type RunnerCommandType int

const (
	RunnerCommandContinue RunnerCommandType = iota
	RunnerCommandInterrupt
	RunnerCommandStop
	RunnerCommandReset

	RunnerCommandAddBreakpoint
	RunnerCommandRemoveBreakpoint
	RunnerCommandAddWatchpoint
	RunnerCommandRemoveWatchpoint

	RunnerCommandReadMemory
	RunnerCommandWriteMemory

	RunnerCommandReadRegister
	RunnerCommandWriteRegister
)

func (t RunnerCommandType) String() string {
	switch t {
	case RunnerCommandContinue:
		return "Continue"
	case RunnerCommandInterrupt:
		return "Interrupt"
	case RunnerCommandStop:
		return "Stop"
	case RunnerCommandReset:
		return "Reset"
	case RunnerCommandAddBreakpoint:
		return "AddBreakpoint"
	case RunnerCommandRemoveBreakpoint:
		return "RemoveBreakpoint"
	case RunnerCommandAddWatchpoint:
		return "AddWatchpoint"
	case RunnerCommandRemoveWatchpoint:
		return "RemoveWatchpoint"
	case RunnerCommandReadMemory:
		return "ReadMemory"
	case RunnerCommandWriteMemory:
		return "WriteMemory"
	case RunnerCommandReadRegister:
		return "ReadRegister"
	case RunnerCommandWriteRegister:
		return "WriteRegister"
	default:
		return "Unknown"
	}
}

type AddBreakpointArgs struct {
	Address uint32
}

type RemoveBreakpointArgs struct {
	Address uint32
}

type AddWatchpointArgs struct {
	Range *memory.Range
}

type RemoveWatchpointArgs struct {
	Range *memory.Range
}

type ReadMemoryArgs struct {
	Range *memory.Range
}

type WriteMemoryArgs struct {
	Range *memory.Range
	Data  []byte
}

type ReadRegisterArgs struct {
	Register *registers.RegisterDescriptor
}

type WriteRegisterArgs struct {
	Register *registers.RegisterDescriptor
	Value    uint32
}

type RunnerCommand struct {
	Type RunnerCommandType
	Args any
}

type RunnerCommandResult struct {
	Error error
	Data  any
}

type RunnerRequest struct {
	Command RunnerCommand
	Result  chan RunnerCommandResult
}

func NewRunnerRequest(command RunnerCommand) RunnerRequest {
	return RunnerRequest{
		Command: command,
		Result:  make(chan RunnerCommandResult, 1),
	}
}

func (r RunnerRequest) Error(err error) {
	r.Result <- RunnerCommandResult{
		Error: err,
	}
}

func (r RunnerRequest) Success(data any) {
	r.Result <- RunnerCommandResult{
		Error: nil,
		Data:  data,
	}
}

type RunnerState = uint32

const (
	RunnerStateIdle RunnerState = iota
	RunnerStateRunning
	RunnerStateInterrupted
	RunnerStateStopped
)

type RunnerEventType int

const (
	RunnerEventBreakpointHit RunnerEventType = iota
	RunnerEventWatchpointHit
	RunnerEventStepCompleted
	RunnerEventRuntimeError
	RunnerEventInterrupted
	RunnerEventStopped
)

type RunnerEvent struct {
	Type RunnerEventType
	Data any
}

type Runner struct {
	contract.Base

	r Runtime

	state RunnerState

	// Channel for receiving commands from the outside (e.g. debugger)
	commands chan RunnerRequest
	// Channel for sending events to the outside (e.g. debugger)
	events chan RunnerEvent
}

func NewRunner(r Runtime) *Runner {
	runner := &Runner{
		Base:     contract.NewBase(log().Child("runner")),
		r:        r,
		state:    RunnerStateIdle,
		commands: make(chan RunnerRequest, 100),
		events:   make(chan RunnerEvent, 100),
	}
	go runner.commandLoop()

	return runner
}

func (runner *Runner) interrupt() {
	// Interrupt the runner by changing its state to Interrupted
	atomic.StoreUint32(&runner.state, RunnerStateInterrupted)

	runner.events <- RunnerEvent{
		Type: RunnerEventInterrupted,
	}
}

func (runner *Runner) stop() {
	// Stop the runner by changing its state to Stopped
	atomic.StoreUint32(&runner.state, RunnerStateStopped)

	runner.events <- RunnerEvent{
		Type: RunnerEventStopped,
	}
}

func (runner *Runner) step() (*cpu.StepInfo, error) {
	stepInfo, err := runner.r.Step()
	if err != nil {
		runner.events <- RunnerEvent{
			Type: RunnerEventRuntimeError,
			Data: err,
		}

		// Stop the runner on runtime error
		runner.stop()
		return nil, err
	}

	interrupt := false

	if stepInfo.BreakpointHit != nil {
		runner.events <- RunnerEvent{
			Type: RunnerEventBreakpointHit,
			Data: *stepInfo.BreakpointHit,
		}

		// Interrupt the runner on breakpoint hit
		interrupt = true
	}

	if stepInfo.WatchpointHit != nil {
		runner.events <- RunnerEvent{
			Type: RunnerEventWatchpointHit,
			Data: *stepInfo.WatchpointHit,
		}

		// Interrupt the runner on watchpoint hit
		interrupt = true
	}

	runner.events <- RunnerEvent{
		Type: RunnerEventStepCompleted,
		Data: stepInfo,
	}

	if interrupt {
		runner.interrupt()
	}

	return stepInfo, nil
}

func (runner *Runner) commandLoop() {
	runner.Log().Info("runner command loop started")
	defer runner.Log().Info("runner command loop stopped")

	for {
		select {
		case cmd := <-runner.commands:
			runner.handleCommand(cmd)
		default:
			if runner.State() == RunnerStateRunning {
				runner.step()
			} else {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}

func (runner *Runner) commandReset(cmd RunnerRequest) {
	switch runner.State() {
	case RunnerStateRunning:
		cmd.Error(runner.Log().Errorf("cannot reset runtime while running"))
	case RunnerStateInterrupted:
		cmd.Error(runner.Log().Errorf("cannot reset runtime while interrupted"))
	case RunnerStateIdle:
		fallthrough
	case RunnerStateStopped:
		if err := runner.r.Reset(); err != nil {
			cmd.Error(runner.Log().Errorf("failed to reset runtime: %w", err))
			return
		}
		atomic.StoreUint32(&runner.state, RunnerStateIdle)
		cmd.Success(nil)
	}
}

func (runner *Runner) handleCommand(cmd RunnerRequest) {
	runner.Log().Debug("received command", slog.String("command", cmd.Command.Type.String()), slog.Any("args", cmd.Command.Args))

	switch cmd.Command.Type {
	case RunnerCommandContinue:
		runner.commandContinue(cmd)
	case RunnerCommandInterrupt:
		runner.commandInterrupt(cmd)
	case RunnerCommandStop:
		runner.commandStop(cmd)
	case RunnerCommandReset:
		runner.commandReset(cmd)
	case RunnerCommandAddBreakpoint:
		runner.commandAddBreakpoint(cmd)
	case RunnerCommandRemoveBreakpoint:
		runner.commandRemoveBreakpoint(cmd)
	case RunnerCommandAddWatchpoint:
		runner.commandAddWatchpoint(cmd)
	case RunnerCommandRemoveWatchpoint:
		runner.commandRemoveWatchpoint(cmd)
	case RunnerCommandReadMemory:
		runner.commandReadMemory(cmd)
	case RunnerCommandWriteMemory:
		runner.commandWriteMemory(cmd)
	case RunnerCommandReadRegister:
		runner.commandReadRegister(cmd)
	case RunnerCommandWriteRegister:
		runner.commandWriteRegister(cmd)
	default:
		cmd.Error(runner.Log().Errorf("unknown command type: %v", cmd.Command.Type))
	}
}

func (runner *Runner) Reset() error {
	switch runner.State() {
	case RunnerStateRunning:
		return runner.Log().Errorf("cannot reset runtime while running")
	case RunnerStateInterrupted:
		return runner.Log().Errorf("cannot reset runtime while interrupted")
	case RunnerStateStopped:
		if err := runner.r.Reset(); err != nil {
			return runner.Log().Errorf("failed to reset runtime: %w", err)
		}

		atomic.StoreUint32(&runner.state, RunnerStateIdle)
	}

	return nil
}

func (runner *Runner) commandContinue(cmd RunnerRequest) {
	switch runner.State() {
	case RunnerStateRunning:
		cmd.Error(runner.Log().Errorf("runtime is already running"))
		return
	case RunnerStateStopped:
		cmd.Error(runner.Log().Errorf("cannot continue a stopped runtime, please reset it first"))
		return
	}

	atomic.StoreUint32(&runner.state, RunnerStateRunning)

	cmd.Success(nil)
}

func (runner *Runner) commandInterrupt(cmd RunnerRequest) {
	switch runner.State() {
	case RunnerStateIdle:
		cmd.Error(runner.Log().Errorf("cannot interrupt, runtime is not running"))
		return
	case RunnerStateStopped:
		cmd.Error(runner.Log().Errorf("cannot interrupt, runtime is stopped"))
		return
	case RunnerStateInterrupted:
		cmd.Error(runner.Log().Errorf("runtime is already interrupted"))
		return
	}

	atomic.StoreUint32(&runner.state, RunnerStateInterrupted)

	cmd.Success(nil)
}

func (runner *Runner) commandStop(cmd RunnerRequest) {
	switch runner.State() {
	case RunnerStateStopped:
		cmd.Error(runner.Log().Errorf("runtime is already stopped"))
		return
	case RunnerStateIdle:
		cmd.Error(runner.Log().Errorf("runtime is not running"))
		return
	}

	atomic.StoreUint32(&runner.state, RunnerStateStopped)

	cmd.Success(nil)
}

func (runner *Runner) commandAddBreakpoint(cmd RunnerRequest) {
	if _, ok := cmd.Command.Args.(AddBreakpointArgs); ok {
		// TODO: implement adding a breakpoint at args.Address
		cmd.Success(nil)
	} else {
		cmd.Error(runner.Log().Errorf("invalid arguments for AddBreakpoint command. Expected AddBreakpointArgs, got %T", cmd.Command.Args))
	}
}

func (runner *Runner) commandRemoveBreakpoint(cmd RunnerRequest) {
	if _, ok := cmd.Command.Args.(RemoveBreakpointArgs); ok {
		// TODO: implement removing a breakpoint at args.Address
		cmd.Success(nil)
	} else {
		cmd.Error(runner.Log().Errorf("invalid arguments for RemoveBreakpoint command. Expected RemoveBreakpointArgs, got %T", cmd.Command.Args))
	}
}

func (runner *Runner) commandAddWatchpoint(cmd RunnerRequest) {
	if _, ok := cmd.Command.Args.(AddWatchpointArgs); ok {
		// TODO: implement adding a watchpoint for args.Range
		cmd.Success(nil)
	} else {
		cmd.Error(runner.Log().Errorf("invalid arguments for AddWatchpoint command. Expected AddWatchpointArgs, got %T", cmd.Command.Args))
	}
}

func (runner *Runner) commandRemoveWatchpoint(cmd RunnerRequest) {
	if _, ok := cmd.Command.Args.(RemoveWatchpointArgs); ok {
		// TODO: implement removing a watchpoint for args.Range
		cmd.Success(nil)
	} else {
		cmd.Error(runner.Log().Errorf("invalid arguments for RemoveWatchpoint command. Expected RemoveWatchpointArgs, got %T", cmd.Command.Args))
	}
}

func (runner *Runner) commandReadMemory(cmd RunnerRequest) {
	if args, ok := cmd.Command.Args.(ReadMemoryArgs); ok {
		view := memory.NewSlice(runner.r.Memory(), args.Range)
		data, err := view.ReadAll()
		if err != nil {
			cmd.Error(runner.Log().Errorf("failed to read memory: %w", err))
			return
		}

		cmd.Success(data)
	} else {
		cmd.Error(runner.Log().Errorf("invalid arguments for ReadMemory command. Expected ReadMemoryArgs, got %T", cmd.Command.Args))
	}
}

func (runner *Runner) commandWriteMemory(cmd RunnerRequest) {
	if args, ok := cmd.Command.Args.(WriteMemoryArgs); ok {
		view := memory.NewSlice(runner.r.Memory(), args.Range)
		if err := view.Write(args.Data); err != nil {
			cmd.Error(runner.Log().Errorf("failed to write memory: %w", err))
			return
		}

		cmd.Success(nil)
	} else {
		cmd.Error(runner.Log().Errorf("invalid arguments for WriteMemory command. Expected WriteMemoryArgs, got %T", cmd.Command.Args))
	}
}

func (runner *Runner) commandReadRegister(cmd RunnerRequest) {
	if args, ok := cmd.Command.Args.(ReadRegisterArgs); ok {
		value, err := runner.r.CPU().Registers().ReadByDescriptor(args.Register)
		if err != nil {
			cmd.Error(runner.Log().Errorf("failed to read register: %w", err))
			return
		}

		cmd.Success(value)
	} else {
		cmd.Error(runner.Log().Errorf("invalid arguments for ReadRegister command. Expected ReadRegisterArgs, got %T", cmd.Command.Args))
	}
}

func (runner *Runner) commandWriteRegister(cmd RunnerRequest) {
	if args, ok := cmd.Command.Args.(WriteRegisterArgs); ok {
		if err := runner.r.CPU().Registers().WriteByDescriptor(args.Register, args.Value); err != nil {
			cmd.Error(runner.Log().Errorf("failed to write register: %w", err))
			return
		}

		cmd.Success(nil)
	} else {
		cmd.Error(runner.Log().Errorf("invalid arguments for WriteRegister command. Expected WriteRegisterArgs, got %T", cmd.Command.Args))
	}
}

func (runner *Runner) State() RunnerState {
	return RunnerState(atomic.LoadUint32(&runner.state))
}

func (runner *Runner) SendCommand(cmd RunnerCommand) chan RunnerCommandResult {
	req := NewRunnerRequest(cmd)
	runner.commands <- req
	return req.Result
}

func (runner *Runner) Events() <-chan RunnerEvent {
	return runner.events
}

func (runner *Runner) Runtime() Runtime {
	return &runnerRuntime{r: runner}
}

// Implements the Runtime interface by delegating it to a runner, allowing async access to the runtime
type runnerRuntime struct {
	r *Runner
}

func (rr *runnerRuntime) CPU() cpu.CPU {
	return &runnerCpu{r: rr.r}
}

func (rr *runnerRuntime) Memory() memory.Memory {
	return &runnerMemory{r: rr.r}
}

func (rr *runnerRuntime) MemoryLayout() memory.MemoryLayout {
	return rr.r.r.MemoryLayout()
}

func (rr *runnerRuntime) Peripherals() map[string]peripheral.Peripheral {
	return nil
}

func (rr *runnerRuntime) Reset() error {
	result := <-rr.r.SendCommand(RunnerCommand{
		Type: RunnerCommandReset,
	})

	return result.Error
}

func (rr *runnerRuntime) Step() (*cpu.StepInfo, error) {
	result := <-rr.r.SendCommand(RunnerCommand{
		Type: RunnerCommandContinue,
	})

	if result.Error != nil {
		return nil, result.Error
	}

	return result.Data.(*cpu.StepInfo), nil
}

func (rr *runnerRuntime) SetBreakpoint(addr uint32) error {
	result := <-rr.r.SendCommand(RunnerCommand{
		Type: RunnerCommandAddBreakpoint,
		Args: AddBreakpointArgs{
			Address: addr,
		},
	})

	return result.Error
}

func (rr *runnerRuntime) ClearBreakpoint(addr uint32) error {
	result := <-rr.r.SendCommand(RunnerCommand{
		Type: RunnerCommandRemoveBreakpoint,
		Args: RemoveBreakpointArgs{
			Address: addr,
		},
	})

	return result.Error
}

func (rr *runnerRuntime) SetWatchpoint(r memory.Range) error {
	result := <-rr.r.SendCommand(RunnerCommand{
		Type: RunnerCommandAddWatchpoint,
		Args: AddWatchpointArgs{
			Range: &r,
		},
	})

	return result.Error
}

func (rr *runnerRuntime) ClearWatchpoint(r memory.Range) error {
	result := <-rr.r.SendCommand(RunnerCommand{
		Type: RunnerCommandRemoveWatchpoint,
		Args: RemoveWatchpointArgs{
			Range: &r,
		},
	})

	return result.Error
}

type runnerMemory struct {
	r *Runner
}

func (m *runnerMemory) ReadByte(addr uint32) (byte, error) {
	result := <-m.r.SendCommand(RunnerCommand{
		Type: RunnerCommandReadMemory,
		Args: ReadMemoryArgs{
			Range: &memory.Range{
				Start: addr,
				Size:  1,
			},
		},
	})

	if result.Error != nil {
		return 0, result.Error
	}

	data := result.Data.([]byte)
	if len(data) != 1 {
		return 0, fmt.Errorf("expected to read 1 byte, but got %d bytes", len(data))
	}

	return data[0], nil
}

func (m *runnerMemory) WriteByte(addr uint32, value byte) error {
	result := <-m.r.SendCommand(RunnerCommand{
		Type: RunnerCommandWriteMemory,
		Args: WriteMemoryArgs{
			Range: &memory.Range{
				Start: addr,
				Size:  1,
			},
			Data: []byte{value},
		},
	})

	return result.Error
}

func (m *runnerMemory) Size() int {
	return m.r.r.Memory().Size()
}

func (m *runnerMemory) Reset() error {
	return fmt.Errorf("cannot reset memory through runner, access the runtime directly to reset the memory")
}

func (m *runnerMemory) Ranges() []memory.Range {
	return m.r.r.Memory().Ranges()
}

type runnerCpu struct {
	r *Runner
}

func (c *runnerCpu) Registers() cpu.Registers {
	return &runnerRegisters{r: c.r}
}

func (c *runnerCpu) Interrupts() cpu.Interrupts {
	return &runnerInterrupts{r: c.r}
}

func (c *runnerCpu) Step() (*cpu.StepInfo, error) {
	result := <-c.r.SendCommand(RunnerCommand{
		Type: RunnerCommandContinue,
	})

	if result.Error != nil {
		return nil, result.Error
	}

	return result.Data.(*cpu.StepInfo), nil
}

func (c *runnerCpu) IsHalted() bool {
	return c.r.State() == RunnerStateStopped
}

func (c *runnerCpu) Halt() error {
	result := <-c.r.SendCommand(RunnerCommand{
		Type: RunnerCommandStop,
	})

	return result.Error
}

func (c *runnerCpu) Reset() error {
	return fmt.Errorf("cannot reset CPU through runner, access the runtime directly to reset the CPU")
}

type runnerRegisters struct {
	r *Runner
}

func (r *runnerRegisters) Read(idx uint32) (uint32, error) {
	desc, err := cpu.LookupRegister(idx)
	if err != nil {
		return 0, fmt.Errorf("failed to lookup register by index: %w", err)
	}

	return r.ReadByDescriptor(desc)
}

func (r *runnerRegisters) Write(idx uint32, value uint32) error {
	desc, err := cpu.LookupRegister(idx)
	if err != nil {
		return fmt.Errorf("failed to lookup register by index: %w", err)
	}

	return r.WriteByDescriptor(desc, value)
}

func (r *runnerRegisters) ReadByDescriptor(regDesc *registers.RegisterDescriptor) (uint32, error) {
	result := <-r.r.SendCommand(RunnerCommand{
		Type: RunnerCommandReadRegister,
		Args: ReadRegisterArgs{
			Register: regDesc,
		},
	})

	if result.Error != nil {
		return 0, result.Error
	}

	return result.Data.(uint32), nil
}

func (r *runnerRegisters) WriteByDescriptor(regDesc *registers.RegisterDescriptor, value uint32) error {
	result := <-r.r.SendCommand(RunnerCommand{
		Type: RunnerCommandWriteRegister,
		Args: WriteRegisterArgs{
			Register: regDesc,
			Value:    value,
		},
	})

	return result.Error
}

func (r *runnerRegisters) Reset() error {
	return fmt.Errorf("cannot resets registers through runner, access the runtime directly to access the CPU registers and reset them")
}

type runnerInterrupts struct {
	r *Runner
}

func (i *runnerInterrupts) Enable() error {
	return fmt.Errorf("cannot enable interrupts through runner, access the runtime directly to access the CPU interrupts and enable them")
}

func (i *runnerInterrupts) Disable() error {
	return fmt.Errorf("cannot disable interrupts through runner, access the runtime directly to access the CPU interrupts and disable them")
}

func (i *runnerInterrupts) Enabled() bool {
	return false
}

func (i *runnerInterrupts) Interrupt(interruptID uint8) error {
	return fmt.Errorf("cannot trigger interrupts through runner, access the runtime directly to access the CPU interrupts and trigger them")
}

func (i *runnerInterrupts) Servicing() bool {
	return false
}

func (i *runnerInterrupts) CurrentInterrupt() int {
	return 0
}

func (i *runnerInterrupts) PendingInterrupts() []uint8 {
	return nil
}
