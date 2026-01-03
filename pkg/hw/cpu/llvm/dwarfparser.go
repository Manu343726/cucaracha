package llvm

// DWARF Debug Information Parser
//
// This file implements parsing of DWARF debug information from ELF object files.
// DWARF (Debugging With Attributed Record Formats) is the standard debug info
// format used by compilers like GCC and Clang on Unix-like systems.
//
// When code is compiled with debug info (-g flag), the compiler generates several
// DWARF sections in the ELF file:
//
//   - .debug_info: Compilation units, functions, variables, types
//   - .debug_line: Line number program mapping addresses to source lines
//   - .debug_abbrev: Abbreviation tables for .debug_info encoding
//   - .debug_str: String table for debug info
//   - .debug_addr: Address table (DWARF 5)
//
// This parser extracts:
//
//   1. Line Information: Maps instruction addresses to source file/line/column
//   2. Function Information: Function names, address ranges, parameters
//   3. Variable Information: Local variables, their types and storage locations
//   4. Scope Information: Lexical scopes (blocks) within functions
//
// DWARF Location Expressions:
//
// Variables can be stored in registers, on the stack, or as constants. DWARF uses
// location expressions to describe where a variable's value can be found:
//
//   - DW_OP_reg0..DW_OP_reg31: Value is in register N
//   - DW_OP_breg0..DW_OP_breg31: Value is at [register N + offset]
//   - DW_OP_fbreg: Value is at [frame base + offset]
//   - DW_OP_addr: Value is at absolute address
//   - DW_OP_const*: Value is a constant
//
// The Cucaracha CPU uses registers r0-r9 (general purpose), sp (r13), lr (r14).
// The DWARF parser maps DWARF register numbers to Cucaracha register numbers.
//
// Note: Addresses in DWARF are relative to the ELF file's layout. When the code
// is loaded at a different address (e.g., 0x10000), the MemoryResolver remaps
// all debug info addresses accordingly.

import (
	"debug/dwarf"
	"debug/elf"
	"fmt"
	"io"
	"sort"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
)

// DWARFParser extracts debug information from DWARF sections in ELF files.
// It uses Go's debug/dwarf package to parse the standard DWARF format and
// converts the information into Cucaracha's mc.DebugInfo structure.
type DWARFParser struct {
	elfFile   *elf.File
	dwarfData *dwarf.Data
	debugInfo *mc.DebugInfo
}

// NewDWARFParser creates a new DWARF parser for the given ELF file
func NewDWARFParser(elfFile *elf.File) (*DWARFParser, error) {
	dwarfData, err := elfFile.DWARF()
	if err != nil {
		return nil, fmt.Errorf("no DWARF data: %w", err)
	}

	return &DWARFParser{
		elfFile:   elfFile,
		dwarfData: dwarfData,
		debugInfo: mc.NewDebugInfo(),
	}, nil
}

// Parse extracts all debug information from the DWARF data
func (p *DWARFParser) Parse() (*mc.DebugInfo, error) {
	// Parse line number information
	if err := p.parseLineInfo(); err != nil {
		// Line info is optional, continue even if it fails
		// fmt.Printf("Warning: failed to parse line info: %v\n", err)
	}

	// Parse compilation units and variable information
	if err := p.parseCompilationUnits(); err != nil {
		// Also optional
		// fmt.Printf("Warning: failed to parse compilation units: %v\n", err)
	}

	return p.debugInfo, nil
}

// parseLineInfo extracts source line number information from .debug_line
// DWARF line info only records entries at statement boundaries. This function
// propagates each entry to cover all instruction addresses (every 4 bytes)
// until the next entry.
func (p *DWARFParser) parseLineInfo() error {
	reader := p.dwarfData.Reader()

	// Collect all line entries first, then propagate
	type lineEntryData struct {
		addr   uint32
		file   string
		line   int
		column int
	}
	var entries []lineEntryData

	for {
		entry, err := reader.Next()
		if err != nil {
			return err
		}
		if entry == nil {
			break
		}

		// Look for compilation units
		if entry.Tag == dwarf.TagCompileUnit {
			// Get the line reader for this compilation unit
			lineReader, err := p.dwarfData.LineReader(entry)
			if err != nil {
				continue
			}
			if lineReader == nil {
				continue
			}

			// Read all line entries
			var lineEntry dwarf.LineEntry
			for {
				err := lineReader.Next(&lineEntry)
				if err == io.EOF {
					break
				}
				if err != nil {
					break
				}

				// Skip entries without valid addresses
				if lineEntry.Address == 0 {
					continue
				}

				entries = append(entries, lineEntryData{
					addr:   uint32(lineEntry.Address),
					file:   lineEntry.File.Name,
					line:   lineEntry.Line,
					column: lineEntry.Column,
				})
			}
		}
	}

	// Sort entries by address
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].addr < entries[j].addr
	})

	// Propagate each entry to cover all instruction addresses until the next entry
	// Cucaracha instructions are 4 bytes each
	const instrSize = 4

	for i, entry := range entries {
		loc := &mc.SourceLocation{
			File:   entry.file,
			Line:   entry.line,
			Column: entry.column,
		}

		// Determine the end address for this entry
		var endAddr uint32
		if i+1 < len(entries) {
			endAddr = entries[i+1].addr
		} else {
			// Last entry - just record this address
			endAddr = entry.addr + instrSize
		}

		// Fill in all instruction addresses from this entry to the next
		for addr := entry.addr; addr < endAddr; addr += instrSize {
			p.debugInfo.InstructionLocations[addr] = loc
		}
	}

	return nil
}

// parseCompilationUnits extracts function and variable information
func (p *DWARFParser) parseCompilationUnits() error {
	reader := p.dwarfData.Reader()

	var currentFunc *mc.FunctionDebugInfo
	var scopeStack []*mc.ScopeInfo

	for {
		entry, err := reader.Next()
		if err != nil {
			return err
		}
		if entry == nil {
			break
		}

		switch entry.Tag {
		case dwarf.TagCompileUnit:
			// Get compilation unit info
			if name, ok := entry.Val(dwarf.AttrName).(string); ok {
				p.debugInfo.CompilationUnit = name
			}
			if producer, ok := entry.Val(dwarf.AttrProducer).(string); ok {
				p.debugInfo.Producer = producer
			}

		case dwarf.TagSubprogram:
			// Function entry
			funcInfo := &mc.FunctionDebugInfo{}

			if name, ok := entry.Val(dwarf.AttrName).(string); ok {
				funcInfo.Name = name
			}

			if lowPC, ok := entry.Val(dwarf.AttrLowpc).(uint64); ok {
				funcInfo.StartAddress = uint32(lowPC)
			}

			if highPC, ok := entry.Val(dwarf.AttrHighpc).(uint64); ok {
				funcInfo.EndAddress = uint32(highPC)
			} else if highPCOff, ok := entry.Val(dwarf.AttrHighpc).(int64); ok {
				// High PC can be an offset from low PC
				funcInfo.EndAddress = funcInfo.StartAddress + uint32(highPCOff)
			}

			if declFile, ok := entry.Val(dwarf.AttrDeclFile).(int64); ok {
				funcInfo.SourceFile = p.getFileName(int(declFile))
			}

			if declLine, ok := entry.Val(dwarf.AttrDeclLine).(int64); ok {
				funcInfo.StartLine = int(declLine)
			}

			if funcInfo.Name != "" {
				p.debugInfo.Functions[funcInfo.Name] = funcInfo
				currentFunc = funcInfo
			}

			if !entry.Children {
				currentFunc = nil
			}

		case dwarf.TagFormalParameter:
			// Function parameter
			if currentFunc != nil {
				param := p.parseVariable(entry)
				if param != nil {
					param.IsParameter = true
					currentFunc.Parameters = append(currentFunc.Parameters, *param)
				}
			}

		case dwarf.TagVariable:
			// Local variable
			if currentFunc != nil {
				variable := p.parseVariable(entry)
				if variable != nil {
					if len(scopeStack) > 0 {
						// Add to innermost scope
						scope := scopeStack[len(scopeStack)-1]
						scope.Variables = append(scope.Variables, *variable)
					} else {
						currentFunc.LocalVariables = append(currentFunc.LocalVariables, *variable)
					}
				}
			}

		case 0x0b: // DW_TAG_lexical_block - Lexical scope (e.g., a block or loop body)
			// Lexical scope (e.g., a block or loop body)
			if currentFunc != nil {
				scope := &mc.ScopeInfo{}

				if lowPC, ok := entry.Val(dwarf.AttrLowpc).(uint64); ok {
					scope.StartAddress = uint32(lowPC)
				}

				if highPC, ok := entry.Val(dwarf.AttrHighpc).(uint64); ok {
					scope.EndAddress = uint32(highPC)
				} else if highPCOff, ok := entry.Val(dwarf.AttrHighpc).(int64); ok {
					scope.EndAddress = scope.StartAddress + uint32(highPCOff)
				}

				if entry.Children {
					scopeStack = append(scopeStack, scope)
				} else {
					currentFunc.Scopes = append(currentFunc.Scopes, *scope)
				}
			}

		case 0:
			// End of children marker
			if len(scopeStack) > 0 {
				// Pop scope and add to function
				scope := scopeStack[len(scopeStack)-1]
				scopeStack = scopeStack[:len(scopeStack)-1]
				if currentFunc != nil {
					currentFunc.Scopes = append(currentFunc.Scopes, *scope)
				}
			} else if currentFunc != nil {
				// End of function
				currentFunc = nil
			}
		}
	}

	// Build instruction variables map from function info
	p.buildInstructionVariables()

	return nil
}

// parseVariable extracts variable information from a DWARF entry
func (p *DWARFParser) parseVariable(entry *dwarf.Entry) *mc.VariableInfo {
	varInfo := &mc.VariableInfo{}

	if name, ok := entry.Val(dwarf.AttrName).(string); ok {
		varInfo.Name = name
	}

	if varInfo.Name == "" {
		return nil
	}

	// Get type information
	if typeOff, ok := entry.Val(dwarf.AttrType).(dwarf.Offset); ok {
		varInfo.TypeName, varInfo.Size = p.getTypeInfo(typeOff)
	}

	// Get location information
	varInfo.Location = p.parseLocation(entry)

	return varInfo
}

// parseLocation extracts the location of a variable from a DWARF entry
func (p *DWARFParser) parseLocation(entry *dwarf.Entry) mc.VariableLocation {
	// Try to get location expression
	locAttr := entry.Val(dwarf.AttrLocation)
	if locAttr == nil {
		return nil
	}

	// Location can be a []byte (location expression) or int64 (constant)
	switch loc := locAttr.(type) {
	case []byte:
		return p.decodeLocationExpr(loc)
	case int64:
		return mc.ConstantLocation{Value: loc}
	}

	return nil
}

// decodeLocationExpr decodes a DWARF location expression
// This is a simplified decoder that handles common cases
func (p *DWARFParser) decodeLocationExpr(expr []byte) mc.VariableLocation {
	if len(expr) == 0 {
		return nil
	}

	// DWARF location expression opcodes
	const (
		DW_OP_addr        = 0x03
		DW_OP_plus_uconst = 0x23
		DW_OP_reg0        = 0x50
		DW_OP_reg31       = 0x6f
		DW_OP_breg0       = 0x70
		DW_OP_breg31      = 0x8f
		DW_OP_regx        = 0x90
		DW_OP_fbreg       = 0x91
		DW_OP_piece       = 0x93
		DW_OP_stack_val   = 0x9f
		DW_OP_call_frame  = 0x9c
	)

	op := expr[0]

	// Register location (DW_OP_reg0 to DW_OP_reg31)
	if op >= DW_OP_reg0 && op <= DW_OP_reg31 {
		reg := uint32(op - DW_OP_reg0)
		return mc.RegisterLocation{Register: p.mapDWARFRegister(reg)}
	}

	// Base register + offset (DW_OP_breg0 to DW_OP_breg31)
	if op >= DW_OP_breg0 && op <= DW_OP_breg31 {
		reg := uint32(op - DW_OP_breg0)
		offset := int32(0)
		if len(expr) > 1 {
			offset = decodeSLEB128(expr[1:])
		}
		return mc.MemoryLocation{
			BaseRegister: p.mapDWARFRegister(reg),
			Offset:       offset,
		}
	}

	// Frame base relative (DW_OP_fbreg)
	if op == DW_OP_fbreg && len(expr) > 1 {
		offset := decodeSLEB128(expr[1:])
		// Frame base is typically SP or FP
		// For Cucaracha, we use SP (register 13)
		return mc.MemoryLocation{
			BaseRegister: 13, // SP
			Offset:       offset,
		}
	}

	// DW_OP_plus_uconst - adds unsigned constant
	// When this appears alone, it's typically relative to the frame base (SP)
	// This is commonly generated for local variables on the stack
	if op == DW_OP_plus_uconst && len(expr) > 1 {
		offset := int32(decodeULEB128(expr[1:]))
		return mc.MemoryLocation{
			BaseRegister: 13, // SP
			Offset:       offset,
		}
	}

	return nil
}

// mapDWARFRegister maps a DWARF register number to Cucaracha register
func (p *DWARFParser) mapDWARFRegister(dwarfReg uint32) uint32 {
	// For ARM-like targets, DWARF registers typically map directly
	// r0-r12 = 0-12
	// sp = 13
	// lr = 14
	// pc = 15
	// For Cucaracha's encoding: r0-r9 are at 16-25, sp=13, lr=14, pc=15
	if dwarfReg <= 9 {
		return dwarfReg + 16 // r0-r9
	}
	if dwarfReg == 13 {
		return 13 // sp
	}
	if dwarfReg == 14 {
		return 14 // lr
	}
	if dwarfReg == 15 {
		return 15 // pc
	}
	return dwarfReg
}

// getTypeInfo retrieves type name and size from a type DIE offset
func (p *DWARFParser) getTypeInfo(typeOff dwarf.Offset) (string, int) {
	typ, err := p.dwarfData.Type(typeOff)
	if err != nil {
		return "", 0
	}

	return typ.String(), int(typ.Size())
}

// getFileName returns the file name for a given file index
func (p *DWARFParser) getFileName(index int) string {
	// This would require tracking the file table from line info
	// For now, return empty string
	return ""
}

// buildInstructionVariables populates InstructionVariables map based on function scopes
func (p *DWARFParser) buildInstructionVariables() {
	for _, funcInfo := range p.debugInfo.Functions {
		// Collect all addresses in this function
		for addr := funcInfo.StartAddress; addr < funcInfo.EndAddress; addr += 4 {
			var vars []mc.VariableInfo

			// Add function parameters (always in scope)
			vars = append(vars, funcInfo.Parameters...)

			// Add local variables (always in scope within function)
			vars = append(vars, funcInfo.LocalVariables...)

			// Check each scope
			for _, scope := range funcInfo.Scopes {
				if addr >= scope.StartAddress && addr < scope.EndAddress {
					vars = append(vars, scope.Variables...)
				}
			}

			if len(vars) > 0 {
				p.debugInfo.InstructionVariables[addr] = vars
			}
		}
	}
}

// decodeSLEB128 decodes a signed LEB128 value
func decodeSLEB128(data []byte) int32 {
	var result int32
	var shift uint
	var b byte

	for i := 0; i < len(data); i++ {
		b = data[i]
		result |= int32(b&0x7f) << shift
		shift += 7
		if b&0x80 == 0 {
			// Sign extend if needed
			if shift < 32 && (b&0x40) != 0 {
				result |= -(1 << shift)
			}
			break
		}
	}

	return result
}

// decodeULEB128 decodes an unsigned LEB128 value
func decodeULEB128(data []byte) uint32 {
	var result uint32
	var shift uint

	for i := 0; i < len(data); i++ {
		b := data[i]
		result |= uint32(b&0x7f) << shift
		shift += 7
		if b&0x80 == 0 {
			break
		}
	}

	return result
}
