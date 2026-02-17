//===-- CucarachaInstrInfo.td - Target Description for Cucaracha ---*- tablegen -*-===//
//
//                     The LLVM Compiler Infrastructure
//
// This file is distributed under the University of Illinois Open Source
// License. See LICENSE.TXT for details.
//
//===----------------------------------------------------------------------===//
//
// This file describes the Cucaracha instructions in TableGen format.
//
//===----------------------------------------------------------------------===//

{{ template "CucarachaInstrFormats.td" . }}
{{ template "CucarachaOperators.td" . }}

//===----------------------------------------------------------------------===//
// Automatically defined instructions
//===----------------------------------------------------------------------===//

{{ range .Instructions.AllInstructions }}
{{ template "CucarachaInstruction.td" . }}
{{ end }}

//===----------------------------------------------------------------------===//
// Pseudo instructions
//===----------------------------------------------------------------------===//

def MOVi32 : InstCucaracha<(outs IntegerRegisters:$dst), (ins i32imm:$src), "",
                     [(set i32:$dst, (movei32 imm:$src))]> {
  let isPseudo = 1;
}

//===----------------------------------------------------------------------===//
// Return Instructions
//===----------------------------------------------------------------------===//

// Return is implemented as JMP LR, LR (unconditional jump to link register)
// The JMP instruction saves the current PC to the second operand before jumping,
// but since we're returning we don't care about saving the return address again.
// Using LR as both target and link output makes the instruction: JMP LR, LR
let isTerminator = 1, isReturn = 1, isBarrier = 1, Uses = [LR], Defs = [LR] in {
  def RET : InstCucaracha<(outs), (ins variable_ops),
                    "JMP lr, lr",  [(CucarachaRetFlag)]> {
    // JMP instruction encoding: opcode 0b01100 (bits 4-0)
    // target register (LR = 3) in bits 12-5
    // link register output (LR = 3) in bits 20-13
    let Inst{ 4 - 0 } = 0b01100;  // JMP opcode
    let Inst{ 12-5 } = 3;         // target = LR (register 3)
    let Inst{ 20-13 } = 3;        // link = LR (register 3)
    let Inst{ 31-21 } = 0;        // unused bits
  }
}


//===----------------------------------------------------------------------===//
// Call Instructions
//===----------------------------------------------------------------------===//

// Call is implemented as JMP $addr, LR (jump to address, store return address in LR)
// The JMP instruction saves PC+4 to the link register, making this perfect for function calls
let isCall = 1, Defs = [LR], Uses = [SP] in {
  def BL : InstCucaracha<(outs), (ins IntegerRegisters:$addr),
                      "JMP $addr, lr",
                      [(leg_call i32:$addr)]> {
    bits<8> addr;
    // JMP instruction encoding: opcode 0b01100 (bits 4-0)
    // target register in bits 12-5
    // link register output (LR = 3) in bits 20-13
    let Inst{ 4 - 0 } = 0b01100;  // JMP opcode
    let Inst{ 12-5 } = addr;      // target = address register
    let Inst{ 20-13 } = 3;        // link = LR (register 3)
    let Inst{ 31-21 } = 0;        // unused bits
  }
}

def : Pattern<(i32 (load_sym tglobaladdr:$addr)),  [(MOVi32 $addr)]>;

//===----------------------------------------------------------------------===//
// Pseudo Instructions
//===----------------------------------------------------------------------===//

// SELECT pseudo instruction - expands to conditional branches
// This selects between two values based on a condition
let usesCustomInserter = 1 in {
  def SELECT : CucarachaPseudoInst<(outs IntegerRegisters:$dst),
                                   (ins IntegerRegisters:$cond, IntegerRegisters:$T, IntegerRegisters:$F),
                                   "# SELECT $dst = $cond ? $T : $F",
                                   [(set i32:$dst, (select i32:$cond, i32:$T, i32:$F))]>;
                                   
  // SELECT_CC pseudo instruction - does comparison and selection in one
  def SELECT_CC : CucarachaPseudoInst<(outs IntegerRegisters:$dst),
                                      (ins IntegerRegisters:$lhs, IntegerRegisters:$rhs, IntegerRegisters:$T, IntegerRegisters:$F, i32imm:$cc),
                                      "# SELECT_CC $dst = ($lhs cmp $rhs) ? $T : $F",
                                      [(set i32:$dst, (Cucaracha_select_cc i32:$lhs, i32:$rhs, i32:$T, i32:$F, (i32 timm:$cc)))]>;
}

let Defs = [SP], Uses = [SP] in {
def ADJCALLSTACKDOWN : CucarachaPseudoInst<(outs), (ins i32imm:$amt1, i32imm:$amt2),
                                    "# ADJCALLSTACKDOWN $amt1",
                                    [(callseq_start timm:$amt1, timm:$amt2)]>;
def ADJCALLSTACKUP   : CucarachaPseudoInst<(outs), (ins i32imm:$amt1, i32imm:$amt2),
                                    "# ADJCALLSTACKUP $amt1",
                                    [(callseq_end timm:$amt1, timm:$amt2)]>;
}

// Unconditional branch pseudo-instruction
// This matches LLVM's 'br' to basic blocks and will be expanded to load address + JMP
let isBranch = 1, isTerminator = 1, isBarrier = 1 in {
  def PseudoBR : CucarachaPseudoInst<(outs), (ins brtarget:$target),
                                     "JMP $target",
                                     [(br bb:$target)]>;
}

// Conditional branch pseudo-instruction
// This matches CucarachaISD::BR_COND and will be expanded to CJMP with the mask
// Takes CPSR value explicitly as input register
let isBranch = 1, isTerminator = 1 in {
  def PseudoBRCOND : CucarachaPseudoInst<(outs), (ins IntegerRegisters:$cpsr, i32imm:$mask, brtarget:$target),
                                         "CJMP $cpsr, $mask, $target",
                                         [(cucaracha_brcond i32:$cpsr, (i32 timm:$mask), bb:$target)]>;
}
// Pseudo-instructions for loads/stores with base+offset addressing
// The hardware LD/ST only take register operands, so we use pseudos to match
// LLVM's addressing modes and then expand them to compute effective addresses

// PseudoLD - Load with base register and immediate offset
// Will be expanded to: compute effective addr (base + offset) -> temp reg, then LD
def PseudoLD : CucarachaPseudoInst<(outs IntegerRegisters:$dst),
                                   (ins IntegerRegisters:$base, i32imm:$offset),
                                   "# PseudoLD $dst, [$base, #$offset]",
                                   [(set i32:$dst, (load (add i32:$base, (i32 imm:$offset))))]> {
  let mayLoad = 1;
}

// Pattern for load from a simple register address (no offset)
def : Pat<(load i32:$addr), (PseudoLD i32:$addr, 0)>;

// Pattern for load from frame index
def : Pat<(load (add frameindex:$fi, (i32 imm:$offset))),
          (PseudoLD (COPY_TO_REGCLASS $fi, IntegerRegisters), imm:$offset)>;
def : Pat<(load frameindex:$fi),
          (PseudoLD (COPY_TO_REGCLASS $fi, IntegerRegisters), 0)>;

// PseudoST - Store with base register and immediate offset
// Will be expanded to: compute effective addr (base + offset) -> temp reg, then ST
def PseudoST : CucarachaPseudoInst<(outs),
                                   (ins IntegerRegisters:$src, IntegerRegisters:$base, i32imm:$offset),
                                   "# PseudoST $src, [$base, #$offset]",
                                   [(store i32:$src, (add i32:$base, (i32 imm:$offset)))]> {
  let mayStore = 1;
}

// Pattern for store to a simple register address (no offset)
def : Pat<(store i32:$src, i32:$addr), (PseudoST i32:$src, i32:$addr, 0)>;

// Pattern for store to frame index
def : Pat<(store i32:$src, (add frameindex:$fi, (i32 imm:$offset))),
          (PseudoST i32:$src, (COPY_TO_REGCLASS $fi, IntegerRegisters), imm:$offset)>;
def : Pat<(store i32:$src, frameindex:$fi),
          (PseudoST i32:$src, (COPY_TO_REGCLASS $fi, IntegerRegisters), 0)>;
