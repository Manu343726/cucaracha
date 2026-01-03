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
│   │   ├── llvm.go        # LLVM TableGen generation
│   │   ├── exec.go        # Execute programs
│   │   ├── debug.go       # Interactive debugger CLI
│   │   └── compile.go     # Clang driver command
│   └── tools/
├── pkg/
│   ├── hw/cpu/
│   │   ├── interpreter/   # CPU emulator/interpreter
│   │   │   └── debugger.go # Debugger API
│   │   ├── llvm/          # LLVM integration & assembly parser
│   │   │   ├── clang.go   # Clang toolchain discovery & compilation
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

The `cucaracha` CLI includes a built-in clang driver that automatically discovers the LLVM toolchain:

```bash
# Compile C source to assembly (auto-discovers clang)
./cucaracha cpu clang hello.c

# Specify output format (assembly, object, llvm-ir)
./cucaracha cpu clang --format llvm-ir hello.c

# Specify optimization level
./cucaracha cpu clang -O2 hello.c

# Verbose output
./cucaracha cpu clang -v hello.c

# Specify output path
./cucaracha cpu clang -o output.cucaracha hello.c
```

You can also use clang directly:

```bash
# Using the built clang
path/to/llvm-project/build_vs2022/Release/bin/clang.exe \
  --target=cucaracha -O0 -S -o output.cucaracha input.c
```

The `cpu clang` command supports:
- **Auto-discovery**: Finds clang in project build directories or system PATH
- **Multiple output formats**: Assembly (`.cucaracha`), Object (`.o`), LLVM IR (`.ll`)
- **Optimization levels**: `-O0`, `-O1`, `-O2`, `-O3`, `-Os`, `-Oz`
- **Include paths**: `-I/path/to/include`
- **Preprocessor defines**: `-DDEBUG`
- **Extra flags**: `-X<flag>` to pass flags directly to clang
- **Verbose mode**: `-v` streams all clang output (warnings, errors) in real-time
- **Build from source**: `--build-clang` builds clang if not found

#### Clang Auto-Discovery

The toolchain searches for clang in the following order:
1. Explicit path via `--clang-path` flag
2. Project build directories:
   - `llvm-project/build_vs2022/{Release,Debug}/bin/clang.exe` (Windows)
   - `llvm-project/build/{Release,Debug}/bin/clang` (Linux)
   - `llvm-project/build_docker_linux_gcc/bin/clang` (Docker)
3. System PATH (validates `--target=cucaracha` support)

#### Clang Command Examples

```bash
# Basic compilation
cucaracha cpu clang hello.c                    # Output: hello.cucaracha
cucaracha cpu clang -o out.s hello.c           # Custom output path

# Output formats
cucaracha cpu clang -f assembly hello.c        # Assembly (default)
cucaracha cpu clang -f object hello.c          # ELF object file
cucaracha cpu clang -f llvm-ir hello.c         # LLVM IR

# Optimization
cucaracha cpu clang -O2 hello.c                # Standard optimization
cucaracha cpu clang -Os hello.c                # Size optimization

# Verbose (shows all clang output including warnings)
cucaracha cpu clang -v -X -Wall hello.c

# Include paths and defines
cucaracha cpu clang -I./include -DDEBUG hello.c

# Pass extra flags to clang
cucaracha cpu clang -X -fno-builtin hello.c
```

### Executing Cucaracha Programs

```bash
# Execute assembly file
./cucaracha cpu exec program.cucaracha

# Execute with tracing
./cucaracha cpu exec -t program.cucaracha

# Execute ELF binary
./cucaracha cpu exec program.o

# Execute C source file directly (compiles first)
./cucaracha cpu exec hello.c
```

### Interactive Debugger

The `debug` command provides a GDB-style interactive debugger with source-level debugging support:

```bash
# Start debugging a program
./cucaracha cpu debug program.cucaracha
./cucaracha cpu debug program.o

# Debug a C source file directly (compiles first with DWARF debug info)
./cucaracha cpu debug hello.c
```

#### Debugger Commands

| Command | Shortcut | Description |
|---------|----------|-------------|
| `step [n]` | `s` | Step n source lines (or instructions if no debug info) |
| `stepi [n]` | `si` | Step n instructions |
| `continue` | `c` | Continue execution until breakpoint |
| `run` | `r` | Run until termination or breakpoint |
| `break <addr>` | `b` | Set breakpoint at address |
| `watch <addr>` | `w` | Set watchpoint on memory address |
| `delete <id>` | `d` | Delete breakpoint/watchpoint by ID |
| `list` | `l` | List all breakpoints/watchpoints |
| `print <reg/@addr>` | `p` | Print register (r0-r9, sp, lr, pc, cpsr) or memory (@addr) |
| `set <reg> <value>` | - | Set register value |
| `disasm [addr] [n]` | `x` | Disassemble n instructions at addr |
| `disasm -i` | `x -i` | **Interactive disassembly view with CFG** |
| `memory -i [addr]` | `m -i` | **Interactive memory view** |
| `source [n]` | `src` | Show n lines of source around current location |
| `vars` | `v` | Show accessible variables at current location |
| `info` | `i` | Show CPU state (registers, flags) |
| `stack` | - | Show stack contents |
| `memory <addr> [n]` | `m` | Show n bytes of memory at addr |
| `help` | `h` | Show help |
| `quit` | `q` | Exit debugger |

#### Interactive Disassembly View (CFG Visualization)

The interactive disassembly view (`disasm -i` or `x -i`) provides a radare2-style visualization:

```
╭─ int x = 5;
│  0x10000 12345678 │ MOVIMM16L #5, r0
╰  0x10004 87654321 │ ST r0, [sp+4]
╭─ if (x > 0) {
│  0x10008 AABBCCDD │ LD [sp+4], r0
│  0x1000C 11223344 │ CMP r0, r1, cpsr
│  0x10010 55667788 ╭ CJMP #1, r2, lr         ↓ 0x10020
╰  0x10014 99AABBCC ╰─────────────────────────→
```

**Features:**
- **CFG arrows**: Visual representation of control flow (jumps, branches)
- **Color-coded branches**: Different colors for different branch edges
- **Source line grouping**: Instructions grouped by source line with `╭│╰` brackets
- **C syntax highlighting**: Keywords, types, strings, numbers colored
- **Branch hints**: Target address or symbol shown at end of line
- **Keyboard navigation**: Arrow keys, Page Up/Down, Home/End

**Interactive View Keys:**
| Key | Action |
|-----|--------|
| ↑/↓ | Move up/down one instruction |
| Shift+↑/↓ | Move up/down 10 instructions |
| Page Up/Down | Move by screen height |
| Home/End | Jump to start/end of code |
| P | Jump to current PC |
| ? | Show help |
| Q/ESC | Exit view |

#### Source-Level Debugging

When debugging C files with DWARF debug info, the debugger provides:

- **Source location tracking**: Shows file:line for each instruction
- **Source line stepping**: `step` command steps by source lines
- **Variable inspection**: `vars` shows accessible local variables
- **C syntax highlighting**: Source code displayed with colors

**Example session:**
```
$ ./cucaracha cpu debug program.c
Compiled program.c to temporary file
Loaded 17 instructions with DWARF debug info
Entry point: 0x00010000
Type 'help' for available commands.

  main.c:5  int x = 5;
=> 0x00010000 [000C0410]: MOVIMM16L #5, r0
(cucaracha) s
  main.c:6  int y = x + 3;
=> 0x00010008 [00028820]: LD [sp+4], r1
(cucaracha) vars
Variables at main.c:6:
  x (int): 5 [sp+4]
  y (int): <not yet initialized> [sp+8]
(cucaracha) x -i
[Interactive disassembly view opens]
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
| Clang Integration | ✅ Auto-discovery, compilation, verbose output |
| Interactive Debugger | ✅ GDB-style CLI with breakpoints/watchpoints |
| Source-Level Debugging | ✅ DWARF debug info, source stepping, variable inspection |
| CFG Visualization | ✅ Interactive disassembly with control flow arrows |
| C Syntax Highlighting | ✅ Keywords, types, strings, numbers colored |

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
