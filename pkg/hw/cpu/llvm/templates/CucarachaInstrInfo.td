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

def MOVi32 : InstCucaracha<(outs GRRegs:$dst), (ins i32imm:$src), "",
                     [(set i32:$dst, (movei32 imm:$src))]> {
  let isPseudo = 1;
}

//===----------------------------------------------------------------------===//
// Comparison Instructions
//===----------------------------------------------------------------------===//

def CMP : InstCucaracha<(outs), (ins GRRegs:$lhs, GRRegs:$rhs),
                    "cmp $lhs, $rhs", []> {
  bits<4> lhs;
  bits<4> rhs;
  let Inst{27-20} = 0b00010101;
  let Inst{19-16} = lhs;
  let Inst{15-12}  = 0b0000;
  let Inst{11-7}  = 0b00000; // imm5
  let Inst{6-5} = 0b00; // type
  let Inst{4} = 0b0;
  let Inst{3-0} = rhs;
  
  let Defs = [CPSR];
}

//===----------------------------------------------------------------------===//
// Load/Store Instructions
//===----------------------------------------------------------------------===//

def LDR  : InstCucaracha<(outs GRRegs:$val), (ins memsrc:$addr),
                   "ldr $val, $addr",
                   [(set i32:$val, (load addr:$addr))]> {
  bits<4> val;
  bits<16> addr;
  let Inst{27-25} = 0b010;
  let Inst{24} = 1; // not indexed
  let Inst{23} = 1; // add ofset
  let Inst{22} = 0;
  let Inst{21} = 0; // no writeback
  let Inst{20} = 1; // load
  let Inst{19-16} = addr{15-12};
  let Inst{15-12} = val;
  let Inst{11-0} = addr{11-0};
}

def STR : InstCucaracha<(outs), (ins GRRegs:$val, memsrc:$addr),
                  "str $val, $addr",
                  [(store i32:$val, addr:$addr)]> {
  bits<16> addr;
  bits<4> val;
  let Inst{27-25} = 0b010;
  let Inst{24} = 1; // not indexed
  let Inst{23} = 1; // add ofset
  let Inst{22} = 0;
  let Inst{21} = 0; // no writeback
  let Inst{20} = 0; // store
  let Inst{19-16} = addr{15-12};
  let Inst{15-12} = val;
  let Inst{11-0} = addr{11-0};
}

//===----------------------------------------------------------------------===//
// Return Instructions
//===----------------------------------------------------------------------===//

let isTerminator = 1, isReturn = 1, isBarrier = 1, Uses = [LR] in {
  def RET : InstCucaracha<(outs), (ins variable_ops),
                    "bx lr",  [(CucarachaRetFlag)]> {
    let Inst{27-0}  = 0b0001001011111111111100011110;
  }
}


//===----------------------------------------------------------------------===//
// Call Instructions
//===----------------------------------------------------------------------===//

let isCall = 1, Defs = [LR], Uses = [SP] in {
  def BL : BranchInst<0b1011, (outs), (ins GRRegs:$addr),
                      "bl $addr",
                      [(leg_call i32:$addr)]> {
    bits<4> addr;
    let Inst{31-28} = 0b1110;
    let Inst{3-0} = addr;
  }
}

def : Pattern<(i32 (load_sym tglobaladdr:$addr)),  [(MOVi32 $addr)]>;

//===----------------------------------------------------------------------===//
// Branch Instructions
//===----------------------------------------------------------------------===//

let isTerminator = 1, isBranch = 1, isBarrier = 1 in {
  def B : InstCucaracha<(outs), (ins b_target:$dst),
                  "b $dst", [(br bb:$dst)]> {
    bits<24> dst;
    let Inst{31-28} = 0b0000;
    let Inst{27-24} = 0b1010;
    let Inst{23-0} = dst;
  }
}

let isTerminator = 1, isBranch = 1, Uses = [CPSR] in {
  def Bcc : InstCucaracha<(outs), (ins cc_val:$cc, b_target:$dst),
                    "b$cc $dst",  []> {
  }
}

//===----------------------------------------------------------------------===//
// Pseudo Instructions
//===----------------------------------------------------------------------===//

let Defs = [SP], Uses = [SP] in {
def ADJCALLSTACKDOWN : CucarachaPseudoInst<(outs), (ins i32imm:$amt1, i32imm:$amt2),
                                    "# ADJCALLSTACKDOWN $amt1",
                                    [(callseq_start timm:$amt1, timm:$amt2)]>;
def ADJCALLSTACKUP   : CucarachaPseudoInst<(outs), (ins i32imm:$amt1, i32imm:$amt2),
                                    "# ADJCALLSTACKUP $amt1",
                                    [(callseq_end timm:$amt1, timm:$amt2)]>;
}
