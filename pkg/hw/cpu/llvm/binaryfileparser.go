package llvm

import (
	"debug/elf"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
)

// Relocation types used by the Cucaracha backend (ARM-style)
const (
	// R_ARM_MOVW_PREL_NC - Low 16-bit relocation for MOVIMM16L
	R_ARM_MOVW_PREL_NC elf.R_ARM = 45
	// R_ARM_MOVT_PREL - High 16-bit relocation for MOVIMM16H
	R_ARM_MOVT_PREL elf.R_ARM = 46
)

// relocationEntry represents a parsed relocation from the ELF file
type relocationEntry struct {
	Offset     uint32                  // Offset in .text section where relocation applies
	SymbolName string                  // Name of the symbol being referenced
	Usage      mc.SymbolReferenceUsage // Lo or Hi part of the address
}

// BinaryFile represents a parsed .o ELF file containing cucaracha machine code
type BinaryFile struct {
	fileNameValue   string
	sourceFileValue string
	functionsMap    map[string]mc.Function
	instructionsAll []mc.Instruction
	globalsValue    []mc.Global
	labelsValue     []mc.Label
	memoryLayout    *mc.MemoryLayout
}

// ParseBinaryFile parses a .o ELF file and returns a ProgramFile
func ParseBinaryFile(path string) (mc.ProgramFile, error) {
	parser := NewBinaryFileParser(path)
	return parser.Parse()
}

// BinaryFileParser handles parsing of .o ELF object files
type BinaryFileParser struct {
	path string
}

// NewBinaryFileParser creates a new parser for the given .o file path
func NewBinaryFileParser(path string) *BinaryFileParser {
	return &BinaryFileParser{
		path: path,
	}
}

// Parse executes the full parsing process and returns the BinaryFile
func (p *BinaryFileParser) Parse() (*BinaryFile, error) {
	// Open the ELF file
	f, err := os.Open(p.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	elfFile, err := elf.NewFile(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ELF file: %w", err)
	}

	// Verify it's a 32-bit ELF file (cucaracha uses 32-bit)
	if elfFile.Class != elf.ELFCLASS32 {
		return nil, fmt.Errorf("expected 32-bit ELF file, got %v", elfFile.Class)
	}

	// Verify it's little-endian (cucaracha is little-endian)
	if elfFile.Data != elf.ELFDATA2LSB {
		return nil, fmt.Errorf("expected little-endian ELF file, got %v", elfFile.Data)
	}

	result := &BinaryFile{
		fileNameValue:   p.path,
		functionsMap:    make(map[string]mc.Function),
		instructionsAll: []mc.Instruction{},
		globalsValue:    []mc.Global{},
		labelsValue:     []mc.Label{},
	}

	// Find the .text section for code
	textSection := elfFile.Section(".text")
	if textSection == nil {
		return nil, fmt.Errorf("no .text section found in ELF file")
	}

	// Read the code section
	codeData, err := textSection.Data()
	if err != nil {
		return nil, fmt.Errorf("failed to read .text section: %w", err)
	}

	// Parse symbols to get function and label information
	symbols, err := elfFile.Symbols()
	if err != nil && err != elf.ErrNoSymbols {
		// Check for empty symbol section error message (Go's ELF parser returns this as a string error)
		if err.Error() != "symbol section is empty" {
			return nil, fmt.Errorf("failed to read symbols: %w", err)
		}
		// Treat empty symbol section the same as no symbols
		symbols = nil
	}

	// Build symbol maps for functions and labels
	functionSymbols := make(map[uint32]string)       // addr -> name for functions
	labelSymbols := make(map[uint32]string)          // addr -> name for labels
	globalDataSymbols := make(map[string]elf.Symbol) // name -> symbol for data
	symbolsByIndex := make(map[int]elf.Symbol)       // symbol table index -> symbol

	for i, sym := range symbols {
		symbolsByIndex[i] = sym
		switch elf.ST_TYPE(sym.Info) {
		case elf.STT_FUNC:
			// Function symbol
			functionSymbols[uint32(sym.Value)] = sym.Name
		case elf.STT_NOTYPE:
			// Labels typically have no type
			if sym.Section == elf.SHN_UNDEF {
				continue
			}
			// Check if it's in the text section (local label)
			if sym.Section != elf.SHN_UNDEF && elfFile.Sections[sym.Section] == textSection {
				labelSymbols[uint32(sym.Value)] = sym.Name
			}
		case elf.STT_OBJECT:
			// Global data object
			globalDataSymbols[sym.Name] = sym
		case elf.STT_SECTION:
			// Section symbol - store for relocation resolution
			if int(sym.Section) < len(elfFile.Sections) {
				symbolsByIndex[i] = sym
			}
		}
	}

	// Parse relocations from .rel.text section and get generated labels and rodata globals
	relocations, generatedLabels, rodataGlobals := p.parseRelocations(elfFile, textSection, symbols, labelSymbols)

	// Add generated labels to the label map
	for addr, name := range generatedLabels {
		labelSymbols[addr] = name
	}

	// Decode instructions from the code section
	instructionSize := instructions.Instructions.InstructionBits() / 8 // 4 bytes for 32-bit instructions
	numInstructions := len(codeData) / instructionSize

	// Build a map from offset to relocations for quick lookup
	relocationsByOffset := make(map[uint32][]relocationEntry)
	for _, rel := range relocations {
		relocationsByOffset[rel.Offset] = append(relocationsByOffset[rel.Offset], rel)
	}

	// Calculate opcode bits mask for fixing corrupted instructions
	opcodeBits := 5 // Cucaracha uses 5 bits for opcode
	opcodeMask := uint32((1 << opcodeBits) - 1)

	for i := 0; i < numInstructions; i++ {
		offset := i * instructionSize
		if offset+instructionSize > len(codeData) {
			break
		}

		// Read the 32-bit instruction (little-endian)
		instrBytes := codeData[offset : offset+instructionSize]
		instrValue := binary.LittleEndian.Uint32(instrBytes)

		// Check if this instruction has relocations
		instrRelocs := relocationsByOffset[uint32(offset)]

		// If this instruction has a relocation, the opcode bits might be corrupted
		// because LLVM's relocation handling for ARM-style relocations modifies
		// bits that overlap with our opcode field.
		// We must fix the opcode based on the relocation type:
		// - Lo relocation (R_ARM_MOVW_PREL_NC) -> MOVIMM16L (opcode 2)
		// - Hi relocation (R_ARM_MOVT_PREL) -> MOVIMM16H (opcode 1)
		if len(instrRelocs) > 0 {
			var correctOpcode uint32
			if instrRelocs[0].Usage == mc.SymbolUsageLo {
				correctOpcode = uint32(instructions.OpCode_MOV_IMM16L)
			} else {
				correctOpcode = uint32(instructions.OpCode_MOV_IMM16H)
			}
			instrValue = (instrValue &^ opcodeMask) | correctOpcode
		}

		// Try to decode the instruction
		decoded, err := instructions.Instructions.Decode(instrValue)
		if err != nil {
			// If decoding fails, store as unknown instruction
			result.instructionsAll = append(result.instructionsAll, mc.Instruction{
				LineNumber: i + 1,
				Text:       fmt.Sprintf(".word 0x%08x ; unknown instruction", instrValue),
				Raw: &instructions.RawInstruction{
					Descriptor:    nil,
					OperandValues: []uint64{uint64(instrValue)},
				},
			})
			continue
		}

		// Build the raw instruction
		rawInstr := &instructions.RawInstruction{
			Descriptor:    decoded.Descriptor,
			OperandValues: make([]uint64, len(decoded.OperandValues)),
		}
		for j, opVal := range decoded.OperandValues {
			rawInstr.OperandValues[j] = opVal.Encode()
		}

		// Calculate the address for this instruction
		instrAddr := textSection.Addr + uint64(offset)
		addr := uint32(instrAddr)

		// Build symbol references from relocations
		var symbolRefs []mc.SymbolReference
		for _, rel := range instrRelocs {
			symbolRefs = append(symbolRefs, mc.SymbolReference{
				Name:  rel.SymbolName,
				Usage: rel.Usage,
			})
		}

		result.instructionsAll = append(result.instructionsAll, mc.Instruction{
			LineNumber:  i + 1,
			Address:     &addr,
			Text:        decoded.String(),
			Raw:         rawInstr,
			Instruction: decoded,
			Symbols:     symbolRefs,
		})
	}

	// Build function entries from function symbols
	for addr, name := range functionSymbols {
		// Find the instruction index for this function
		instrIdx := -1
		for i, instr := range result.instructionsAll {
			if instr.Address != nil && *instr.Address == addr {
				instrIdx = i
				break
			}
		}

		if instrIdx >= 0 {
			// Count instructions until next function or end
			endIdx := len(result.instructionsAll)
			nextFuncAddr := uint32(0xFFFFFFFF)
			for otherAddr := range functionSymbols {
				if otherAddr > addr && otherAddr < nextFuncAddr {
					nextFuncAddr = otherAddr
				}
			}
			if nextFuncAddr != 0xFFFFFFFF {
				for i, instr := range result.instructionsAll {
					if instr.Address != nil && *instr.Address >= nextFuncAddr {
						endIdx = i
						break
					}
				}
			}

			result.functionsMap[name] = mc.Function{
				Name: name,
				InstructionRanges: []mc.InstructionRange{
					{Start: instrIdx, Count: endIdx - instrIdx},
				},
			}
		}
	}

	// Build label entries
	for addr, name := range labelSymbols {
		// Skip if this is also a function
		if _, isFunc := functionSymbols[addr]; isFunc {
			continue
		}

		// Find the instruction index for this label
		instrIdx := -1
		for i, instr := range result.instructionsAll {
			if instr.Address != nil && *instr.Address == addr {
				instrIdx = i
				break
			}
		}

		result.labelsValue = append(result.labelsValue, mc.Label{
			Name:             name,
			InstructionIndex: instrIdx,
		})
	}

	// Add rodata globals from relocation analysis
	// These have already been parsed with correct memory addresses
	result.globalsValue = append(result.globalsValue, rodataGlobals...)

	// Parse other data sections for named globals (skipping .rodata if already handled)
	hasRodataGlobal := len(rodataGlobals) > 0
	for _, section := range elfFile.Sections {
		if section.Type != elf.SHT_PROGBITS {
			continue
		}
		if section.Name == ".text" {
			continue
		}
		// Skip .rodata if we already have it from relocations
		if section.Name == ".rodata" && hasRodataGlobal {
			continue
		}
		// Look for .data, .rodata, .bss sections
		if section.Name == ".data" || section.Name == ".rodata" || section.Name == ".bss" {
			data, err := section.Data()
			if err != nil {
				continue
			}

			// Find symbols in this section
			for name, sym := range globalDataSymbols {
				symSection := elfFile.Sections[sym.Section]
				if symSection != section {
					continue
				}

				// Extract the data for this symbol
				offset := sym.Value - section.Addr
				size := int(sym.Size)
				if size == 0 {
					size = 1 // minimum size
				}

				var initData []byte
				if section.Name != ".bss" && offset+uint64(size) <= uint64(len(data)) {
					initData = make([]byte, size)
					copy(initData, data[offset:offset+uint64(size)])
				}

				result.globalsValue = append(result.globalsValue, mc.Global{
					Name:        name,
					Size:        size,
					InitialData: initData,
					Type:        mc.GlobalObject,
				})
			}
		}
	}

	// Build memory layout
	codeSize := uint32(len(codeData))
	dataSize := uint32(0)
	for _, g := range result.globalsValue {
		dataSize += uint32(g.Size)
	}

	result.memoryLayout = &mc.MemoryLayout{
		BaseAddress: uint32(textSection.Addr),
		TotalSize:   codeSize + dataSize,
		CodeSize:    codeSize,
		DataSize:    dataSize,
		CodeStart:   uint32(textSection.Addr),
		DataStart:   uint32(textSection.Addr) + codeSize,
	}

	return result, nil
}

// FileName returns the path to the binary file
func (f *BinaryFile) FileName() string {
	return f.fileNameValue
}

// SourceFile returns the original source file name (empty for binary files)
func (f *BinaryFile) SourceFile() string {
	return f.sourceFileValue
}

// Functions returns all functions in the program
func (f *BinaryFile) Functions() map[string]mc.Function {
	return f.functionsMap
}

// Instructions returns all instructions in the program
func (f *BinaryFile) Instructions() []mc.Instruction {
	return f.instructionsAll
}

// Globals returns all global symbols in the program
func (f *BinaryFile) Globals() []mc.Global {
	return f.globalsValue
}

// Labels returns all labels in the program
func (f *BinaryFile) Labels() []mc.Label {
	return f.labelsValue
}

// MemoryLayout returns the memory layout information
func (f *BinaryFile) MemoryLayout() *mc.MemoryLayout {
	return f.memoryLayout
}

// sectionInfo holds information about an ELF section for relocation processing
type sectionInfo struct {
	section      *elf.Section
	memoryOffset uint32 // Offset from code base where this section will be placed in memory
	data         []byte // Section data (pre-loaded to avoid issues with ELF file access)
}

// parseRelocations parses the .rel.text section and returns relocation entries
// that map instruction offsets to symbol references.
// For PC-relative relocations, it computes the target address from the instruction
// immediates and creates synthetic label/global names.
// Returns both the relocations and a map of generated label addresses to names.
// Also returns a list of globals found in .rodata that need to be loaded.
func (p *BinaryFileParser) parseRelocations(elfFile *elf.File, textSection *elf.Section, symbols []elf.Symbol, labelSymbols map[uint32]string) ([]relocationEntry, map[uint32]string, []mc.Global) {
	var relocations []relocationEntry
	generatedLabels := make(map[uint32]string)
	var rodataGlobals []mc.Global

	// Find the .rel.text section
	relSection := elfFile.Section(".rel.text")
	if relSection == nil {
		return relocations, generatedLabels, rodataGlobals
	}

	// Read the relocation section data
	relData, err := relSection.Data()
	if err != nil {
		return relocations, generatedLabels, rodataGlobals
	}

	// Read code data to extract immediates
	codeData, err := textSection.Data()
	if err != nil {
		return relocations, generatedLabels, rodataGlobals
	}

	// Build a map of sections by their symbol table index
	// Note: Go's elf.Symbols() skips the null symbol at index 0, so we need to
	// use (i+1) as the key to match ELF symbol table indices in relocations
	sectionMap := make(map[int]sectionInfo)
	codeSize := uint32(len(codeData))

	// Also build a map of function symbols (STT_FUNC) by their index
	// These are named functions that we can look up by name
	functionSymbols := make(map[int]elf.Symbol)

	// Calculate where each section will be placed in memory
	// .text is at offset 0, .rodata follows .text
	currentOffset := codeSize

	for i, sym := range symbols {
		// ELF symbol index = Go slice index + 1 (because Go skips the null symbol)
		elfSymIndex := i + 1
		symType := elf.ST_TYPE(sym.Info)

		if symType == elf.STT_SECTION {
			if int(sym.Section) < len(elfFile.Sections) {
				sect := elfFile.Sections[sym.Section]
				if sect == textSection {
					sectionMap[elfSymIndex] = sectionInfo{section: sect, memoryOffset: 0, data: codeData}
				} else if sect.Name == ".rodata" {
					// Align to 4 bytes
					if currentOffset%4 != 0 {
						currentOffset = (currentOffset + 3) &^ 3
					}
					// Pre-load the section data
					sectData, err := sect.Data()
					if err != nil {
						sectData = nil
					}
					sectionMap[elfSymIndex] = sectionInfo{section: sect, memoryOffset: currentOffset, data: sectData}
					// Note: Individual globals will be created per-reference when we parse relocations
					currentOffset += uint32(sect.Size)
				}
			}
		} else if symType == elf.STT_FUNC || symType == elf.STT_NOTYPE {
			// Named function or label symbol - store for later lookup
			functionSymbols[elfSymIndex] = sym
		}
	}

	// Each Rel32 entry is 8 bytes: 4 bytes offset + 4 bytes info
	const relEntrySize = 8
	numRelocs := len(relData) / relEntrySize

	// First pass: collect all relocation pairs (Lo followed by Hi)
	type relocPair struct {
		loOffset uint32
		hiOffset uint32
		symIndex int
	}
	var pairs []relocPair

	for i := 0; i < numRelocs-1; i += 2 {
		loEntryOffset := i * relEntrySize
		hiEntryOffset := (i + 1) * relEntrySize

		loOffset := binary.LittleEndian.Uint32(relData[loEntryOffset : loEntryOffset+4])
		loInfo := binary.LittleEndian.Uint32(relData[loEntryOffset+4 : loEntryOffset+8])
		hiOffset := binary.LittleEndian.Uint32(relData[hiEntryOffset : hiEntryOffset+4])
		hiInfo := binary.LittleEndian.Uint32(relData[hiEntryOffset+4 : hiEntryOffset+8])

		loType := elf.R_ARM(loInfo & 0xFF)
		hiType := elf.R_ARM(hiInfo & 0xFF)
		loSymIndex := int(loInfo >> 8)

		// Verify this is a Lo/Hi pair
		if loType == R_ARM_MOVW_PREL_NC && hiType == R_ARM_MOVT_PREL && hiOffset == loOffset+4 {
			pairs = append(pairs, relocPair{
				loOffset: loOffset,
				hiOffset: hiOffset,
				symIndex: loSymIndex,
			})
		}
	}

	// Map from target address to symbol name (label for code, global for data)
	targetSymbols := make(map[uint32]string)
	labelCounter := 0

	// Track which globals we've created for data references
	rodataGlobalNames := make(map[uint32]string)

	// Second pass: compute target addresses and create symbols
	for _, pair := range pairs {
		// Read the immediate values from both instructions
		if int(pair.loOffset)+4 > len(codeData) || int(pair.hiOffset)+4 > len(codeData) {
			continue
		}

		loInstr := binary.LittleEndian.Uint32(codeData[pair.loOffset : pair.loOffset+4])
		hiInstr := binary.LittleEndian.Uint32(codeData[pair.hiOffset : pair.hiOffset+4])

		// Decode fixup values from both instructions using Cucaracha-compatible format.
		// The LLVM Cucaracha backend encodes fixup immediates in bits 5-20 of the instruction,
		// which aligns with Cucaracha's MOV immediate format and preserves the opcode (bits 0-4).
		loImm := DecodeFixupFromInstruction(loInstr)
		hiImm := DecodeFixupFromInstruction(hiInstr)

		// Combine to get the full 32-bit offset (this is the stored addend)
		storedAddend := CombineLoHiImmediate(loImm, hiImm)

		var symbolName string
		var targetAddr uint32

		// Check if this relocation targets a named function symbol
		if funcSym, ok := functionSymbols[pair.symIndex]; ok {
			// Named function symbol - use the function name directly
			symbolName = funcSym.Name
			// Target address is the function's value (offset within .text)
			targetAddr = uint32(funcSym.Value)
			if _, exists := targetSymbols[targetAddr]; !exists {
				targetSymbols[targetAddr] = symbolName
			}
		} else if sectInfo, hasSect := sectionMap[pair.symIndex]; hasSect {
			// Section symbol
			if sectInfo.section != textSection {
				// Data section reference (.rodata)
				// The immediate is garbage - use section offset as target
				targetAddr = sectInfo.memoryOffset

				if existingName, ok := targetSymbols[targetAddr]; ok {
					symbolName = existingName
				} else if existingName, ok := rodataGlobalNames[targetAddr]; ok {
					symbolName = existingName
				} else {
					symbolName = fmt.Sprintf(".L_data_%d", labelCounter)
					labelCounter++

					rodataData := sectInfo.data
					var initData []byte
					if rodataData != nil && len(rodataData) > 0 {
						initData = make([]byte, len(rodataData))
						copy(initData, rodataData)
					}

					rodataGlobals = append(rodataGlobals, mc.Global{
						Name:        symbolName,
						Size:        len(initData),
						InitialData: initData,
						Type:        mc.GlobalObject,
					})
					rodataGlobalNames[targetAddr] = symbolName
					targetSymbols[targetAddr] = symbolName
				}
			} else {
				// .text section symbol - internal branch target
				// For PC-relative relocations, LLVM stores the addend in ARM MOVW format.
				// The storedAddend is the absolute target offset within the section.
				// This is because for a section symbol with value 0, the linker would
				// compute: (S + A) - P = (0 + addend) - P = addend - P
				// So the stored addend IS the target address (relative to section start).
				targetAddr = storedAddend

				// Debug logging (disabled)
				// fmt.Printf("DEBUG: loOffset=0x%X, loImm=0x%X, hiImm=0x%X, storedAddend=0x%X (%d), targetAddr=0x%X, codeSize=0x%X\n",
				// 	pair.loOffset, loImm, hiImm, storedAddend, storedAddend, targetAddr, codeSize)

				// Validate target is within .text bounds
				if targetAddr >= codeSize {
					// Invalid target - skip this relocation
					continue
				}

				if existingName, ok := targetSymbols[targetAddr]; ok {
					symbolName = existingName
				} else if existingLabel, ok := labelSymbols[targetAddr]; ok {
					symbolName = existingLabel
					targetSymbols[targetAddr] = symbolName
				} else {
					symbolName = fmt.Sprintf(".L_auto_%d", labelCounter)
					labelCounter++
					generatedLabels[targetAddr] = symbolName
					targetSymbols[targetAddr] = symbolName
				}
			}
		} else {
			// Unknown symbol type - skip this relocation
			continue
		}

		// Add relocations for both Lo and Hi
		relocations = append(relocations, relocationEntry{
			Offset:     pair.loOffset,
			SymbolName: symbolName,
			Usage:      mc.SymbolUsageLo,
		})
		relocations = append(relocations, relocationEntry{
			Offset:     pair.hiOffset,
			SymbolName: symbolName,
			Usage:      mc.SymbolUsageHi,
		})
	}

	return relocations, generatedLabels, rodataGlobals
}
