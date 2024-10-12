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

//===----------------------------------------------------------------------===//
// Operand Definitions.
//===----------------------------------------------------------------------===//

def bl_target : Operand<i32>;

def b_target : Operand<OtherVT>;

def cc_val : Operand<i32> {
  let PrintMethod = "printCondCode";
}

def memsrc : Operand<i32> {
  let MIOperandInfo = (ops GRRegs, i32imm);
  let PrintMetbod = "printAddrModeMemSrc";
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
