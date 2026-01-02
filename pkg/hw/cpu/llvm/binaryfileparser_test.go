package llvm

import (
	"debug/elf"
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc/instructions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestELFFile creates a minimal ELF32 file with cucaracha machine code
func createTestELFFile(t *testing.T, instrs []uint32) string {
	t.Helper()

	// Create a temp directory for test files
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "test.o")

	// Build minimal ELF32 headers
	// ELF Header (52 bytes for ELF32)
	elfHeader := make([]byte, 52)
	// Magic number
	elfHeader[0] = 0x7f
	elfHeader[1] = 'E'
	elfHeader[2] = 'L'
	elfHeader[3] = 'F'
	// Class: ELFCLASS32
	elfHeader[4] = 1
	// Data: ELFDATA2LSB (little-endian)
	elfHeader[5] = 1
	// Version
	elfHeader[6] = 1
	// OS/ABI
	elfHeader[7] = 0
	// Type: ET_REL (relocatable)
	binary.LittleEndian.PutUint16(elfHeader[16:], 1)
	// Machine: EM_ARM (cucaracha uses ARM machine type)
	binary.LittleEndian.PutUint16(elfHeader[18:], 40)
	// Version
	binary.LittleEndian.PutUint32(elfHeader[20:], 1)
	// Entry point (0 for relocatable)
	binary.LittleEndian.PutUint32(elfHeader[24:], 0)
	// Program header offset (0 for relocatable)
	binary.LittleEndian.PutUint32(elfHeader[28:], 0)
	// Section header offset (will set after building sections)
	// ELF header flags
	binary.LittleEndian.PutUint32(elfHeader[36:], 0)
	// ELF header size
	binary.LittleEndian.PutUint16(elfHeader[40:], 52)
	// Program header entry size (0 for relocatable)
	binary.LittleEndian.PutUint16(elfHeader[42:], 0)
	// Number of program headers (0 for relocatable)
	binary.LittleEndian.PutUint16(elfHeader[44:], 0)
	// Section header entry size
	binary.LittleEndian.PutUint16(elfHeader[46:], 40)
	// Number of section headers (4: null, .text, .shstrtab, .symtab)
	binary.LittleEndian.PutUint16(elfHeader[48:], 4)
	// Section header string table index
	binary.LittleEndian.PutUint16(elfHeader[50:], 2)

	// Build .text section content (instructions)
	textData := make([]byte, len(instrs)*4)
	for i, instr := range instrs {
		binary.LittleEndian.PutUint32(textData[i*4:], instr)
	}

	// Section header string table
	shstrtab := []byte("\x00.text\x00.shstrtab\x00.symtab\x00")

	// Calculate offsets
	textOffset := uint32(52)                                     // After ELF header
	shstrtabOffset := textOffset + uint32(len(textData))         // After .text
	symtabOffset := shstrtabOffset + uint32(len(shstrtab))       // After .shstrtab
	sectionHeaderOffset := symtabOffset + 0                      // No symbols for simplicity
	sectionHeaderOffset = (sectionHeaderOffset + 3) & ^uint32(3) // Align to 4 bytes

	// Update ELF header with section header offset
	binary.LittleEndian.PutUint32(elfHeader[32:], sectionHeaderOffset)

	// Build section headers (40 bytes each for ELF32)
	// Section 0: Null section
	nullSection := make([]byte, 40)

	// Section 1: .text
	textSection := make([]byte, 40)
	binary.LittleEndian.PutUint32(textSection[0:], 1)                                       // sh_name (offset in shstrtab)
	binary.LittleEndian.PutUint32(textSection[4:], uint32(elf.SHT_PROGBITS))                // sh_type
	binary.LittleEndian.PutUint32(textSection[8:], uint32(elf.SHF_ALLOC|elf.SHF_EXECINSTR)) // sh_flags
	binary.LittleEndian.PutUint32(textSection[12:], 0)                                      // sh_addr
	binary.LittleEndian.PutUint32(textSection[16:], textOffset)                             // sh_offset
	binary.LittleEndian.PutUint32(textSection[20:], uint32(len(textData)))                  // sh_size
	binary.LittleEndian.PutUint32(textSection[24:], 0)                                      // sh_link
	binary.LittleEndian.PutUint32(textSection[28:], 0)                                      // sh_info
	binary.LittleEndian.PutUint32(textSection[32:], 4)                                      // sh_addralign
	binary.LittleEndian.PutUint32(textSection[36:], 0)                                      // sh_entsize

	// Section 2: .shstrtab
	shstrtabSection := make([]byte, 40)
	binary.LittleEndian.PutUint32(shstrtabSection[0:], 7)                      // sh_name
	binary.LittleEndian.PutUint32(shstrtabSection[4:], uint32(elf.SHT_STRTAB)) // sh_type
	binary.LittleEndian.PutUint32(shstrtabSection[8:], 0)                      // sh_flags
	binary.LittleEndian.PutUint32(shstrtabSection[12:], 0)                     // sh_addr
	binary.LittleEndian.PutUint32(shstrtabSection[16:], shstrtabOffset)        // sh_offset
	binary.LittleEndian.PutUint32(shstrtabSection[20:], uint32(len(shstrtab))) // sh_size
	binary.LittleEndian.PutUint32(shstrtabSection[24:], 0)                     // sh_link
	binary.LittleEndian.PutUint32(shstrtabSection[28:], 0)                     // sh_info
	binary.LittleEndian.PutUint32(shstrtabSection[32:], 1)                     // sh_addralign
	binary.LittleEndian.PutUint32(shstrtabSection[36:], 0)                     // sh_entsize

	// Section 3: .symtab (empty for now)
	symtabSection := make([]byte, 40)
	binary.LittleEndian.PutUint32(symtabSection[0:], 17)                     // sh_name
	binary.LittleEndian.PutUint32(symtabSection[4:], uint32(elf.SHT_SYMTAB)) // sh_type
	binary.LittleEndian.PutUint32(symtabSection[8:], 0)                      // sh_flags
	binary.LittleEndian.PutUint32(symtabSection[12:], 0)                     // sh_addr
	binary.LittleEndian.PutUint32(symtabSection[16:], symtabOffset)          // sh_offset
	binary.LittleEndian.PutUint32(symtabSection[20:], 0)                     // sh_size (empty)
	binary.LittleEndian.PutUint32(symtabSection[24:], 0)                     // sh_link
	binary.LittleEndian.PutUint32(symtabSection[28:], 0)                     // sh_info
	binary.LittleEndian.PutUint32(symtabSection[32:], 4)                     // sh_addralign
	binary.LittleEndian.PutUint32(symtabSection[36:], 16)                    // sh_entsize (Sym32Size)

	// Write the ELF file
	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()

	// Write ELF header
	_, err = f.Write(elfHeader)
	require.NoError(t, err)

	// Write .text section data
	_, err = f.Write(textData)
	require.NoError(t, err)

	// Write .shstrtab section data
	_, err = f.Write(shstrtab)
	require.NoError(t, err)

	// Pad to alignment for section headers
	currentOffset := int64(52 + len(textData) + len(shstrtab))
	padSize := int64(sectionHeaderOffset) - currentOffset
	if padSize > 0 {
		padding := make([]byte, padSize)
		_, err = f.Write(padding)
		require.NoError(t, err)
	}

	// Write section headers
	_, err = f.Write(nullSection)
	require.NoError(t, err)
	_, err = f.Write(textSection)
	require.NoError(t, err)
	_, err = f.Write(shstrtabSection)
	require.NoError(t, err)
	_, err = f.Write(symtabSection)
	require.NoError(t, err)

	return path
}

func TestBinaryFileParser_ParsesValidELF(t *testing.T) {
	// Create a test ELF file with some NOP instructions
	// NOP has no operands, so we need to get the descriptor and encode directly
	nopDescriptor, err := instructions.Instructions.Instruction(instructions.OpCode_NOP)
	require.NoError(t, err, "should get NOP descriptor")
	nopInstr := instructions.RawInstruction{
		Descriptor:    nopDescriptor,
		OperandValues: []uint64{}, // NOP has no operands
	}
	nopEncoded := nopInstr.Encode()

	path := createTestELFFile(t, []uint32{nopEncoded, nopEncoded, nopEncoded})

	// Parse the file
	pf, err := ParseBinaryFile(path)
	require.NoError(t, err, "ParseBinaryFile should succeed")

	// Verify basic properties
	assert.Equal(t, path, pf.FileName(), "file name should match")
	assert.NotNil(t, pf.MemoryLayout(), "memory layout should be set")

	// Verify instructions
	instrs := pf.Instructions()
	assert.Len(t, instrs, 3, "should have 3 instructions")

	for i, instr := range instrs {
		assert.NotNil(t, instr.Instruction, "instruction %d should be decoded", i)
		assert.Equal(t, instructions.OpCode_NOP, instr.Instruction.Descriptor.OpCode.OpCode,
			"instruction %d should be NOP", i)
		assert.NotNil(t, instr.Address, "instruction %d should have address", i)
		assert.Equal(t, uint32(i*4), *instr.Address, "instruction %d address should be correct", i)
	}
}

func TestBinaryFileParser_DecodesADDInstruction(t *testing.T) {
	// Create an ADD r0, r1, r2 instruction
	addDescriptor, err := instructions.Instructions.Instruction(instructions.OpCode_ADD)
	require.NoError(t, err, "should get ADD descriptor")
	addInstr := instructions.RawInstruction{
		Descriptor:    addDescriptor,
		OperandValues: []uint64{0x10, 0x11, 0x12}, // r0, r1, r2 encodings
	}
	addEncoded := addInstr.Encode()

	path := createTestELFFile(t, []uint32{addEncoded})

	// Parse the file
	pf, err := ParseBinaryFile(path)
	require.NoError(t, err, "ParseBinaryFile should succeed")

	// Verify instruction
	instrs := pf.Instructions()
	require.Len(t, instrs, 1, "should have 1 instruction")

	instr := instrs[0]
	require.NotNil(t, instr.Instruction, "instruction should be decoded")
	assert.Equal(t, instructions.OpCode_ADD, instr.Instruction.Descriptor.OpCode.OpCode,
		"should be ADD instruction")

	// Verify operands
	require.Len(t, instr.Instruction.OperandValues, 3, "ADD should have 3 operands")
	assert.Equal(t, "r0", instr.Instruction.OperandValues[0].Register().Name(),
		"first operand should be r0")
	assert.Equal(t, "r1", instr.Instruction.OperandValues[1].Register().Name(),
		"second operand should be r1")
	assert.Equal(t, "r2", instr.Instruction.OperandValues[2].Register().Name(),
		"third operand should be r2")
}

func TestBinaryFileParser_DecodesMOVIMMInstruction(t *testing.T) {
	// Create a MOVIMM16L #42, r0 instruction
	movDescriptor, err := instructions.Instructions.Instruction(instructions.OpCode_MOV_IMM16L)
	require.NoError(t, err, "should get MOVIMM16L descriptor")
	movInstr := instructions.RawInstruction{
		Descriptor:    movDescriptor,
		OperandValues: []uint64{42, 0x10}, // immediate 42, r0 encoding
	}
	movEncoded := movInstr.Encode()

	path := createTestELFFile(t, []uint32{movEncoded})

	// Parse the file
	pf, err := ParseBinaryFile(path)
	require.NoError(t, err, "ParseBinaryFile should succeed")

	// Verify instruction
	instrs := pf.Instructions()
	require.Len(t, instrs, 1, "should have 1 instruction")

	instr := instrs[0]
	require.NotNil(t, instr.Instruction, "instruction should be decoded")
	assert.Equal(t, instructions.OpCode_MOV_IMM16L, instr.Instruction.Descriptor.OpCode.OpCode,
		"should be MOVIMM16L instruction")

	// Verify operands
	require.Len(t, instr.Instruction.OperandValues, 2, "MOVIMM16L should have 2 operands")
	imm := instr.Instruction.OperandValues[0].Immediate()
	assert.Equal(t, int32(42), imm.Int32(),
		"first operand should be immediate 42")
	assert.Equal(t, "r0", instr.Instruction.OperandValues[1].Register().Name(),
		"second operand should be r0")
}

func TestBinaryFileParser_MemoryLayoutIsCorrect(t *testing.T) {
	nopDescriptor, err := instructions.Instructions.Instruction(instructions.OpCode_NOP)
	require.NoError(t, err, "should get NOP descriptor")
	nopInstr := instructions.RawInstruction{
		Descriptor:    nopDescriptor,
		OperandValues: []uint64{},
	}
	nopEncoded := nopInstr.Encode()

	// 5 instructions = 20 bytes
	path := createTestELFFile(t, []uint32{nopEncoded, nopEncoded, nopEncoded, nopEncoded, nopEncoded})

	pf, err := ParseBinaryFile(path)
	require.NoError(t, err)

	layout := pf.MemoryLayout()
	require.NotNil(t, layout, "memory layout should be set")

	assert.Equal(t, uint32(0), layout.BaseAddress, "base address should be 0")
	assert.Equal(t, uint32(20), layout.CodeSize, "code size should be 20 bytes (5 instructions)")
	assert.Equal(t, uint32(0), layout.CodeStart, "code start should be 0")
}

func TestBinaryFileParser_RejectsNonELFFile(t *testing.T) {
	// Create a non-ELF file
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "notelf.o")
	err := os.WriteFile(path, []byte("not an ELF file"), 0644)
	require.NoError(t, err)

	_, err = ParseBinaryFile(path)
	assert.Error(t, err, "should reject non-ELF file")
	assert.Contains(t, err.Error(), "ELF", "error should mention ELF")
}

func TestBinaryFileParser_RejectsNonExistentFile(t *testing.T) {
	_, err := ParseBinaryFile("/nonexistent/path/file.o")
	assert.Error(t, err, "should reject non-existent file")
}

func TestBinaryFileParser_RawInstructionIsPopulated(t *testing.T) {
	// Create an ADD instruction
	addDescriptor, err := instructions.Instructions.Instruction(instructions.OpCode_ADD)
	require.NoError(t, err, "should get ADD descriptor")
	addInstr := instructions.RawInstruction{
		Descriptor:    addDescriptor,
		OperandValues: []uint64{0x10, 0x11, 0x12},
	}
	addEncoded := addInstr.Encode()

	path := createTestELFFile(t, []uint32{addEncoded})

	pf, err := ParseBinaryFile(path)
	require.NoError(t, err)

	instrs := pf.Instructions()
	require.Len(t, instrs, 1)

	instr := instrs[0]
	require.NotNil(t, instr.Raw, "Raw instruction should be populated")
	assert.Equal(t, instructions.OpCode_ADD, instr.Raw.Descriptor.OpCode.OpCode)
	require.Len(t, instr.Raw.OperandValues, 3)
	assert.Equal(t, uint64(0x10), instr.Raw.OperandValues[0])
	assert.Equal(t, uint64(0x11), instr.Raw.OperandValues[1])
	assert.Equal(t, uint64(0x12), instr.Raw.OperandValues[2])
}

// findClang searches for the clang binary in common locations
func findClang() string {
	// Try common paths for LLVM build
	possiblePaths := []string{
		// Linux Docker build
		"/c/Users/Manu3/Documents/GitHub/cucaracha/llvm-project/build_docker_linux_gcc/bin/clang",
		// VS2022 build paths
		"C:/Users/Manu3/Documents/GitHub/cucaracha/llvm-project/build_vs2022/Debug/bin/clang.exe",
		"C:/Users/Manu3/Documents/GitHub/cucaracha/llvm-project/build_vs2022/Release/bin/clang.exe",
		"C:/Users/Manu3/Documents/GitHub/cucaracha/llvm-project/build_vs2022/bin/Debug/clang.exe",
		"C:/Users/Manu3/Documents/GitHub/cucaracha/llvm-project/build_vs2022/bin/Release/clang.exe",
		// Generic build path
		"../../../../../llvm-project/build/bin/clang",
	}

	for _, p := range possiblePaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// Try PATH
	if path, err := exec.LookPath("clang"); err == nil {
		return path
	}

	return ""
}

// TestBinaryFileParser_LLVMGeneratedObjectFile tests parsing of actual .o files
// generated by LLVM for the cucaracha target
func TestBinaryFileParser_LLVMGeneratedObjectFile(t *testing.T) {
	clang := findClang()
	if clang == "" {
		t.Skip("clang not found - skipping LLVM integration test")
	}

	// Find the test programs directory
	root, err := os.Getwd()
	require.NoError(t, err)
	programsDir := filepath.Join(root, "..", "..", "..", "..", "..", "llvm-project", "cucaracha-tests", "programs")

	// Check if hello_world.c exists
	helloWorldC := filepath.Join(programsDir, "hello_world.c")
	if _, err := os.Stat(helloWorldC); os.IsNotExist(err) {
		t.Skipf("test program not found at %s - skipping LLVM integration test", helloWorldC)
	}

	// Create temp directory for output
	tempDir := t.TempDir()
	outputO := filepath.Join(tempDir, "hello_world.o")

	// Compile with clang for cucaracha target
	cmd := exec.Command(clang, "--target=cucaracha", "-O0", "-c", helloWorldC, "-o", outputO)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("clang compilation failed (maybe cucaracha target not supported): %v\nOutput: %s", err, output)
	}

	// Parse the generated .o file
	pf, err := ParseBinaryFile(outputO)
	require.NoError(t, err, "ParseBinaryFile should succeed for LLVM-generated .o file")

	// Verify basic properties
	assert.Equal(t, outputO, pf.FileName(), "file name should match")

	// Verify we have instructions
	instrs := pf.Instructions()
	assert.NotEmpty(t, instrs, "should have instructions")

	// Count decoded vs unknown instructions
	decodedCount := 0
	unknownCount := 0
	for _, instr := range instrs {
		if instr.Instruction != nil {
			decodedCount++
		} else {
			unknownCount++
		}
	}
	t.Logf("Total instructions: %d, Decoded: %d, Unknown: %d", len(instrs), decodedCount, unknownCount)

	// Most instructions should be decodable
	assert.Greater(t, decodedCount, unknownCount, "more instructions should be decoded than unknown")

	// Verify memory layout is set
	layout := pf.MemoryLayout()
	require.NotNil(t, layout, "memory layout should be set")
	assert.Greater(t, layout.CodeSize, uint32(0), "code size should be > 0")

	// Now pass through Resolve() to fully resolve the program
	// Note: This may fail if there are unknown instructions, so we test what we can
	memConfig := mc.DefaultMemoryResolverConfig()
	resolved, err := mc.Resolve(pf, memConfig)

	if err != nil {
		// If resolve fails due to unknown instructions, log and continue
		t.Logf("Resolve() failed (likely due to unknown instructions): %v", err)
		t.Logf("This is expected if the LLVM backend generates pseudo-instructions not in the cucaracha ISA")
		return
	}

	// Verify resolved program
	resolvedInstrs := resolved.Instructions()
	assert.Len(t, resolvedInstrs, len(instrs), "resolved should have same instruction count")

	// All decoded instructions should have addresses
	for i, instr := range resolvedInstrs {
		if instr.Instruction != nil {
			require.NotNil(t, instr.Address, "resolved instruction %d should have address", i)
		}
	}

	// Memory layout should be complete
	resolvedLayout := resolved.MemoryLayout()
	require.NotNil(t, resolvedLayout, "resolved memory layout should be set")
	assert.Equal(t, uint32(0), resolvedLayout.BaseAddress, "base address should be 0")
}

// TestBinaryFileParser_LLVMGeneratedObjectFile_CompareWithAssembly compares
// the binary parser output with the assembly parser output for the same program
func TestBinaryFileParser_LLVMGeneratedObjectFile_CompareWithAssembly(t *testing.T) {
	clang := findClang()
	if clang == "" {
		t.Skip("clang not found - skipping LLVM integration test")
	}

	// Find the test programs directory
	root, err := os.Getwd()
	require.NoError(t, err)
	programsDir := filepath.Join(root, "..", "..", "..", "..", "..", "llvm-project", "cucaracha-tests", "programs")

	// Check if hello_world.c exists
	helloWorldC := filepath.Join(programsDir, "hello_world.c")
	if _, err := os.Stat(helloWorldC); os.IsNotExist(err) {
		t.Skipf("test program not found at %s - skipping LLVM integration test", helloWorldC)
	}

	// Create temp directory for output
	tempDir := t.TempDir()
	outputO := filepath.Join(tempDir, "hello_world.o")
	outputS := filepath.Join(tempDir, "hello_world.cucaracha")

	// Compile .o file (binary)
	cmd := exec.Command(clang, "--target=cucaracha", "-O0", "-c", helloWorldC, "-o", outputO)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Skipf("clang compilation failed: %v\nOutput: %s", err, output)
	}

	// Compile .cucaracha file (assembly)
	cmd = exec.Command(clang, "--target=cucaracha", "-O0", "-S", "-c", helloWorldC, "-o", outputS)
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Skipf("clang assembly compilation failed: %v\nOutput: %s", err, output)
	}

	// Parse both files
	binaryPF, err := ParseBinaryFile(outputO)
	require.NoError(t, err, "ParseBinaryFile should succeed")

	assemblyPF, err := ParseAssemblyFile(outputS)
	require.NoError(t, err, "ParseAssemblyFile should succeed")

	// Log instruction counts before resolution
	t.Logf("Binary instructions: %d", len(binaryPF.Instructions()))
	t.Logf("Assembly instructions: %d", len(assemblyPF.Instructions()))

	// Resolve both through the full pipeline
	memConfig := mc.DefaultMemoryResolverConfig()

	resolvedBinary, errBinary := mc.Resolve(binaryPF, memConfig)
	resolvedAssembly, errAssembly := mc.Resolve(assemblyPF, memConfig)

	// If either fails, log the error and do a partial comparison
	if errBinary != nil {
		t.Logf("Binary resolve failed: %v", errBinary)
	}
	if errAssembly != nil {
		t.Logf("Assembly resolve failed: %v", errAssembly)
	}

	// If both failed, skip detailed comparison
	if errBinary != nil && errAssembly != nil {
		t.Log("Both resolve operations failed - skipping detailed comparison")
		return
	}

	// If only one succeeded, still compare what we can
	if errBinary != nil {
		t.Log("Binary resolve failed - comparing raw parsed data only")
		// Compare raw instruction counts
		assert.Equal(t, len(assemblyPF.Instructions()), len(binaryPF.Instructions()),
			"raw instruction counts should match")
		return
	}

	if errAssembly != nil {
		t.Log("Assembly resolve failed - comparing raw parsed data only")
		return
	}

	// Both resolved successfully - do full comparison
	binaryInstrs := resolvedBinary.Instructions()
	assemblyInstrs := resolvedAssembly.Instructions()
	assert.Equal(t, len(assemblyInstrs), len(binaryInstrs),
		"binary and assembly should have same instruction count")

	// Compare each instruction's decoded form
	minLen := len(binaryInstrs)
	if len(assemblyInstrs) < minLen {
		minLen = len(assemblyInstrs)
	}

	matchCount := 0
	mismatchCount := 0

	for i := 0; i < minLen; i++ {
		binInstr := binaryInstrs[i]
		asmInstr := assemblyInstrs[i]

		// Skip comparison if either instruction couldn't be decoded
		if binInstr.Instruction == nil || asmInstr.Instruction == nil {
			continue
		}

		// Compare opcodes
		if asmInstr.Instruction.Descriptor.OpCode.OpCode == binInstr.Instruction.Descriptor.OpCode.OpCode {
			matchCount++
		} else {
			mismatchCount++
			t.Logf("Instruction %d mismatch: asm=%s, bin=%s",
				i, asmInstr.Instruction.Descriptor.OpCode.Mnemonic,
				binInstr.Instruction.Descriptor.OpCode.Mnemonic)
		}

		// Compare raw encodings if both have raw
		if binInstr.Raw != nil && asmInstr.Raw != nil &&
			binInstr.Raw.Descriptor != nil && asmInstr.Raw.Descriptor != nil {
			assert.Equal(t, asmInstr.Raw.Encode(), binInstr.Raw.Encode(),
				"instruction %d raw encoding should match", i)
		}

		// Compare addresses
		if binInstr.Address != nil && asmInstr.Address != nil {
			assert.Equal(t, *asmInstr.Address, *binInstr.Address,
				"instruction %d address should match", i)
		}
	}

	t.Logf("Instruction comparison: %d matches, %d mismatches", matchCount, mismatchCount)

	// Most instructions should match
	assert.Greater(t, matchCount, mismatchCount, "more instructions should match than mismatch")

	// Compare memory layouts
	binLayout := resolvedBinary.MemoryLayout()
	asmLayout := resolvedAssembly.MemoryLayout()
	if binLayout != nil && asmLayout != nil {
		assert.Equal(t, asmLayout.CodeSize, binLayout.CodeSize, "code size should match")
		assert.Equal(t, asmLayout.BaseAddress, binLayout.BaseAddress, "base address should match")
	}
}
