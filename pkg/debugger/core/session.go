package core

import (
	"fmt"

	"github.com/Manu343726/cucaracha/pkg/runtime"
	"github.com/Manu343726/cucaracha/pkg/runtime/program"
	"github.com/Manu343726/cucaracha/pkg/runtime/program/loader"
	"github.com/Manu343726/cucaracha/pkg/system"
)

// Manages a debugging session, allowing to set the system config, the program being debugged, and the runtime where it runs.
//
// Once all three are set, the session can create a debugger instance to control execution.
// Any change to the system, program, or runtime invalidates the debugger, and the session resets it.
type Session struct {
	eventCallback EventCallback
	system        *system.SystemDescriptor
	runtime       runtime.Runtime
	program       program.ProgramFile
	debugger      Debugger
}

// Sets the event callback to be used by the session debugger.
func (s *Session) SetEventCallback(callback EventCallback) {
	s.eventCallback = callback

	if s.debugger != nil {
		s.debugger.SetEventCallback(callback)
	}
}

// Returns the debugger. If the session is not ready, returns an error.
func (s *Session) Debugger() (Debugger, error) {
	if s.debugger == nil {
		return nil, fmt.Errorf("debugger not ready, load system, program, and runtime first")
	}

	return s.debugger, nil
}

// Returns the system descriptor of the configured system
func (s *Session) System() *system.SystemDescriptor {
	return s.system
}

// Returns the runtime where the program is executed
func (s *Session) Runtime() runtime.Runtime {
	return s.runtime
}

// Returns the program being debugged
func (s *Session) Program() program.ProgramFile {
	return s.program
}

// Returns whether the session is ready, so the debugger can be used
func (s *Session) IsReady() bool {
	return s.debugger != nil
}

func (s *Session) setup() error {
	if s.system == nil {
		return fmt.Errorf("system not configured")
	}

	if s.runtime == nil {
		return fmt.Errorf("runtime not loaded")
	}

	if s.program == nil {
		return fmt.Errorf("program not loaded")
	}

	if err := runtime.Load(s.runtime, s.system, s.program); err != nil {
		return fmt.Errorf("failed to load runtime: %w", err)
	}

	s.debugger = NewDebugger(s.runtime, s.program)
	s.debugger.SetEventCallback(s.eventCallback)

	return nil
}

// Loads the system configuration into the session
//
// If the session becomes ready after this call (all system, runtime, and program are set), the debugger is set up automatically.
// If there's no error during the debugger setup, the call will return no error and the session is ready to use.
func (s *Session) LoadSystem(systemDesc *system.SystemDescriptor) error {
	if s.system != systemDesc {
		s.system = systemDesc
		s.debugger = nil
	}

	if s.runtime != nil && s.program != nil {
		return s.setup()
	}

	return nil
}

// Loads the execution runtime into the session
//
// If the session becomes ready after this call (all system, runtime, and program are set), the debugger is set up automatically.
// If there's no error during the debugger setup, the call will return no error and the session is ready to use.
func (s *Session) LoadRuntime(runtime runtime.Runtime) error {
	if s.runtime != runtime {
		s.runtime = runtime
		s.debugger = nil
	}

	if s.program != nil && s.system != nil {
		return s.setup()
	}

	return nil
}

// Loads the program into the session
//
// If the session becomes ready after this call (all system, runtime, and program are set), the debugger is set up automatically.
// If there's no error during the debugger setup, the call will return no error and the session is ready to use.
func (s *Session) LoadProgram(program program.ProgramFile) error {
	if s.program != program {
		s.program = program
		s.debugger = nil
	}

	if s.runtime != nil && s.system != nil {
		return s.setup()
	}

	return nil
}

// Loads a program into the session from a file.
//
// Program loading from file supports binary, assembly, and source files, using the loader package.
// If any error occurs during loading (compilation error for example), an error is returned and the session is kept unchanged.
func (s *Session) LoadProgramFromFile(path string, options *loader.Options) (*loader.Result, error) {
	// Ensure options are initialized
	if options == nil {
		options = &loader.Options{}
	}

	// If we have a runtime, use its memory layout for program resolution
	if s.runtime != nil {
		runtimeLayout := s.runtime.MemoryLayout()
		options.MemoryLayout = &runtimeLayout
	}

	loadResult, err := loader.LoadFile(path, options)
	if err != nil {
		return nil, fmt.Errorf("failed to load program from file: %w", err)
	}

	return loadResult, s.LoadProgram(loadResult.Program)
}

// Loads system configuration from a system YAML config file into the session.
//
// This involves loading the system config file, settings up the system, and loading it into the session (i.e. loading the system into the session runtime).
// If any error occurs during config loading or system setup, an error is returned and the session is kept unchanged.
func (s *Session) LoadSystemFromFile(path string) error {
	systemConfig, err := system.LoadFile(path)
	if err != nil {
		return fmt.Errorf("failed to load system config from file: %w", err)
	}

	if system, err := systemConfig.Setup(); err != nil {
		return fmt.Errorf("failed to setup system from config: %w", err)
	} else {
		return s.LoadSystem(system)
	}
}
