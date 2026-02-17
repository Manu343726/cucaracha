//===-- CucarachaOperators.td - Cucaracha-specific operators ------*- tblgen-*-===//
//
//                     The LLVM Compiler Infrastructure
//
// This file is distributed under the University of Illinois Open Source
// License. See LICENSE.TXT for details.
//
//===----------------------------------------------------------------------===//

//===----------------------------------------------------------------------===//
// Type profiles
//===----------------------------------------------------------------------===//

def MoveImm32Ty : SDTypeProfile<1, 1, [
  SDTCisSameAs<0, 1>, SDTCisInt<0>
]>;

def SDT_CucarachaCallSeqStart : SDCallSeqStart<[ SDTCisVT<0, i32>, SDTCisVT<1, i32> ]>;

def SDT_CucarachaCallSeqEnd   : SDCallSeqEnd<[
  SDTCisVT<0, i32>, SDTCisVT<1, i32>
]>;

def CucarachaRetFlag    : SDNode<"CucarachaISD::RET_FLAG", SDTNone,
                           [SDNPHasChain, SDNPOptInGlue, SDNPVariadic]>;
def callseq_start : SDNode<"ISD::CALLSEQ_START", SDT_CucarachaCallSeqStart,
                           [SDNPHasChain, SDNPOutGlue]>;
def callseq_end   : SDNode<"ISD::CALLSEQ_END",   SDT_CucarachaCallSeqEnd,
                           [SDNPHasChain, SDNPOptInGlue, SDNPOutGlue]>;

def SDT_CucarachaCall    : SDTypeProfile<0, -1, [SDTCisPtrTy<0>]>;

//===----------------------------------------------------------------------===//
// Custom SDNodes.
//===----------------------------------------------------------------------===//

def load_sym : SDNode<"CucarachaISD::LOAD_SYM", SDTIntUnaryOp>;

def movei32 : SDNode<"CucarachaISD::MOVEi32", MoveImm32Ty>;

def leg_call
    : SDNode<"CucarachaISD::CALL", SDT_CucarachaCall,
             [ SDNPHasChain, SDNPOptInGlue, SDNPOutGlue, SDNPVariadic ]>;

// SELECT_CC node type profile
def SDT_CucarachaSelectCC : SDTypeProfile<1, 5, [
  SDTCisVT<0, i32>,
  SDTCisSameAs<0, 1>,
  SDTCisSameAs<1, 2>,
  SDTCisSameAs<0, 3>,
  SDTCisSameAs<0, 4>,
  SDTCisVT<5, i32>
]>;

def Cucaracha_select_cc : SDNode<"CucarachaISD::SELECT_CC", SDT_CucarachaSelectCC>;

// CMP node type profile - compares two values and produces result in CPSR
// Result: i32 (CPSR value)
// Operands: LHS, RHS (both i32 registers)
def SDT_CucarachaCmp : SDTypeProfile<1, 2, [
  SDTCisVT<0, i32>,    // Result (CPSR)
  SDTCisVT<1, i32>,    // LHS
  SDTCisVT<2, i32>     // RHS
]>;

def cucaracha_cmp : SDNode<"CucarachaISD::CMP", SDT_CucarachaCmp, []>;

// BR_COND node type profile - takes CPSR value explicitly
def SDT_CucarachaBrCond : SDTypeProfile<0, 3, [
  SDTCisVT<0, i32>,    // cpsr register value
  SDTCisVT<1, i32>,    // mask immediate
  SDTCisVT<2, OtherVT> // target basic block
]>;

def cucaracha_brcond : SDNode<"CucarachaISD::BR_COND", SDT_CucarachaBrCond,
                              [SDNPHasChain]>;

//===----------------------------------------------------------------------===//
// Operand Definitions.
//===----------------------------------------------------------------------===//

def bl_target : Operand<i32>;

def b_target : Operand<OtherVT>;

// Branch target operand for direct branches to basic blocks
def brtarget : Operand<OtherVT> {
  let EncoderMethod = "getBranchTargetOpValue";
}

def cc_val : Operand<i32> {
  let PrintMethod = "printCondCode";
}

def memsrc : Operand<i32> {
  let MIOperandInfo = (ops IntegerRegisters, i32imm);
  let PrintMethod = "printAddrModeMemSrc";
  let EncoderMethod = "getMemSrcValue";
}

def i32imm_lo : Operand<i32>, ImmLeaf<i32, [{
  return Imm >= 0 && Imm < 65536;
}]>;

def Cucarachaimm8 : Operand<i32>, ImmLeaf<i32, [{
  return Imm >= 0 && Imm < 256;
}]>;

//===----------------------------------------------------------------------===//
// Complex Pattern Definitions.
//===----------------------------------------------------------------------===//

def addr : ComplexPattern<iPTR, 2, "SelectAddr", [frameindex], []>;
