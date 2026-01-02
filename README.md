# Cucaracha

A custom 32-bit CPU architecture with a complete toolchain, including an emulator/interpreter written in Go and an LLVM backend for compiling C/C++ code.

## Overview

Cucaracha is an educational project that implements a complete custom CPU architecture from scratch:

- **Custom ISA**: A simple 32-bit instruction set with ARM-inspired design
- **Go Emulator**: An interpreter that can execute Cucaracha assembly and ELF binaries
- **LLVM Backend**: A full LLVM target that compiles C/C++ to Cucaracha machine code
- **Toolchain Integration**: Works with Clang for seamless C/C++ compilation

## Architecture Features

### Registers
- **16 General Purpose Registers**: `r0`-`r15`
- **State Registers**: `PC` (Program Counter), `SP` (Stack Pointer), `LR` (Link Register), `CPSR` (Status Flags)

### Instruction Set
| Opcode | Mnemonic | Description |
|--------|----------|-------------|
| 0 | `NOP` | No operation |
| 1 | `MOVIMM16H` | Load high 16 bits of immediate |
| 2 | `MOVIMM16L` | Load low 16 bits of immediate |
| 3 | `MOV` | Register-to-register copy |
| 4 | `LD` | Load from memory |
| 5 | `ST` | Store to memory |
| 6 | `ADD` | Integer addition |
| 7 | `SUB` | Integer subtraction |
| 8 | `MUL` | Integer multiplication |
| 9 | `DIV` | Signed integer division |
| 10 | `MOD` | Integer modulo |
| 11 | `CMP` | Compare (sets CPSR flags) |
| 12 | `LSL` | Logical shift left |
| 13 | `LSR` | Logical shift right |
| 14 | `ASR` | Arithmetic shift right |
| 15 | `JMP` | Unconditional jump |
| 16 | `CJMP` | Conditional jump |

### Condition Codes (ARM-style CPSR)
- `EQ` (0): Equal (Z=1)
- `NE` (1): Not Equal (Z=0)
- `CS` (2): Carry Set (C=1)
- `CC` (3): Carry Clear (C=0)
- `MI` (4): Minus/Negative (N=1)
- `PL` (5): Plus/Positive (N=0)
- `GT` (12): Signed Greater Than
- `LE` (13): Signed Less or Equal
- `AL` (14): Always

## Project Structure

```
cucaracha/
├── main.go                 # CLI entry point
├── cmd/                    # CLI commands
│   ├── root.go
│   ├── cpu/               # CPU-related commands
│   │   └── llvm.go        # LLVM TableGen generation
│   └── tools/
├── pkg/
│   ├── hw/cpu/
│   │   ├── interpreter/   # CPU emulator/interpreter
│   │   ├── llvm/          # LLVM integration & assembly parser
│   │   │   └── templates/ # TableGen templates
│   │   └── mc/            # Machine code definitions
│   │       ├── instructions/
│   │       └── registers/
│   └── utils/
└── go.mod
```

## Dependencies

### LLVM Fork

This project requires a custom LLVM fork with the Cucaracha backend:

- **Repository**: [Manu343726/llvm-project](https://github.com/Manu343726/llvm-project)
- **Branch**: `cucaracha-backend`

The LLVM fork adds:
- Full Cucaracha target backend in `llvm/lib/Target/Cucaracha/`
- Custom instruction selection, register allocation, and code generation
- Clang support for the `--target=cucaracha` triple

### Go Dependencies
- Go 1.22+
- Cobra (CLI framework)
- Viper (configuration)
- Testify (testing)

## Setup

### 1. Clone the Repositories

```bash
# Clone the main cucaracha project
git clone https://github.com/Manu343726/cucaracha.git
cd cucaracha

# The llvm-project fork should be cloned as a sibling directory
cd ..
git clone -b cucaracha-backend https://github.com/Manu343726/llvm-project.git
```

Expected directory structure:
```
parent-directory/
├── cucaracha/          # This Go project
│   └── cucaracha/      # Go module root
└── llvm-project/       # LLVM fork with Cucaracha backend
    └── llvm/lib/Target/Cucaracha/
```

### 2. Build LLVM with Cucaracha Backend

#### Windows (Visual Studio 2022)

```bash
cd llvm-project
mkdir build_vs2022 && cd build_vs2022

cmake -G "Visual Studio 17 2022" -A x64 \
  -DLLVM_ENABLE_PROJECTS="clang" \
  -DLLVM_TARGETS_TO_BUILD="X86" \
  -DLLVM_EXPERIMENTAL_TARGETS_TO_BUILD="Cucaracha" \
  -DCMAKE_BUILD_TYPE=Release \
  ../llvm

cmake --build . --target clang --config Release -j 16
```

#### Linux (GCC)

```bash
cd llvm-project
mkdir build && cd build

cmake -G Ninja \
  -DLLVM_ENABLE_PROJECTS="clang" \
  -DLLVM_TARGETS_TO_BUILD="X86" \
  -DLLVM_EXPERIMENTAL_TARGETS_TO_BUILD="Cucaracha" \
  -DCMAKE_BUILD_TYPE=Release \
  ../llvm

ninja clang
```

### 3. Build the Go Project

```bash
cd cucaracha/cucaracha
go build -o cucaracha.exe .
```

### 4. Run Tests

```bash
# Go unit tests
go test ./...

# LLVM Cucaracha tests (from llvm-project/build_vs2022)
ctest -C Release -R cucaracha --output-on-failure
```

## Usage

### Compiling C to Cucaracha Assembly

```bash
# Using the built clang
path/to/llvm-project/build_vs2022/Release/bin/clang.exe \
  --target=cucaracha -O0 -S -o output.cucaracha input.c
```

### Executing Cucaracha Programs

```bash
# Execute assembly file
./cucaracha cpu exec program.cucaracha

# Execute with tracing
./cucaracha cpu exec -t program.cucaracha

# Execute ELF binary
./cucaracha cpu exec program.o
```

### Generating LLVM TableGen Files

The LLVM backend's TableGen files can be regenerated from Go templates:

```bash
./cucaracha cpu generateLlvmTablegen --output ../llvm-project/llvm/lib/Target/Cucaracha/
```

## Current Status

✅ **Fully Working** (January 2026)

| Component | Status |
|-----------|--------|
| Go Emulator | ✅ 100% tests passing |
| LLVM Backend | ✅ 100% tests passing (48/48) |
| Assembly Execution | ✅ Working |
| Binary Execution | ✅ Working |

### Test Programs
All test programs compile and execute correctly:
- `hello_world` - Basic program structure
- `arithmetic` - Math operations (+, -, *, /, %)
- `fibonacci` - Recursive function calls
- `factorial` - Iterative computation
- `loops` - For/while loops
- `conditionals` - If/else branches
- `arrays` - Array initialization and access
- `functions` - Function calls with arguments

## Development

### Regenerating LLVM Backend

When modifying the instruction set or registers in Go:

1. Update descriptors in `pkg/hw/cpu/mc/instructions/descriptor.go`
2. Regenerate TableGen: `./cucaracha cpu generateLlvmTablegen`
3. Rebuild LLVM: `cmake --build . --target clang --config Release`
4. Run tests: `ctest -C Release -R cucaracha`

### Adding New Instructions

1. Add opcode in `pkg/hw/cpu/mc/instructions/opcodes.go`
2. Add descriptor in `pkg/hw/cpu/mc/instructions/descriptor.go`
3. Update interpreter in `pkg/hw/cpu/interpreter/interpreter.go`
4. Regenerate TableGen files
5. Add patterns in LLVM backend if needed

## License

MIT License - See [LICENSE](LICENSE) file

## Author

Manuel Sánchez (Manu343726)

## AI Assistance Disclaimer

Yeah, this statement may look redundant considering the very particular odor coming from certain documentation files, code comments, and scripts of the project. But since this topic is fairly hot at the moment, I will give a clear disclaimer here:

**This project is developed with the assistance of GitHub Copilot (Claude).** 

That said, the agent doesn't do anything I wouldn't be capable of doing myself—it's simply a productivity tool that helps save time, much like the difference between coding with and without autocompletion. I'm not in favor of what's known as "vibe coding" (i.e. dumping verbal diarrhea into a prompt and expecting a fully working application as if software development was an act of sorcery). But in my personal opinion disregarding tools such as AI agents is as simple minded as sticking with ed in the dawn of full screen editors just because *"Real programmers use paper and write code line by line"*.
