//===-- Cucaracha.td - Describe the Cucaracha Target Machine -------*- tablegen -*-===//
//
//                     The LLVM Compiler Infrastructure
//
// This file is distributed under the University of Illinois Open Source
// License. See LICENSE.TXT for details.
//
//===----------------------------------------------------------------------===//
//
// This is the top level entry point for the Cucaracha target.
//
//===----------------------------------------------------------------------===//

//===----------------------------------------------------------------------===//
// Cucaracha documentation
//===----------------------------------------------------------------------===//
/*
{{ .DocString }}
*/

//===----------------------------------------------------------------------===//
// Target-independent interfaces which we are implementing
//===----------------------------------------------------------------------===//

include "llvm/Target/Target.td"

//===----------------------------------------------------------------------===//
// Descriptions
//===----------------------------------------------------------------------===//

{{ template "CucarachaRegisterInfo.td" . }}
{{ template "CucarachaConditionCodes.td" . }}
{{ template "CucarachaInstrInfo.td" . }}
{{ template "CucarachaCallingConv.td" . }}

def CucarachaInstrInfo : InstrInfo;

//===----------------------------------------------------------------------===//
// Cucaracha processors supported.
//===----------------------------------------------------------------------===//

class Proc<string Name, list<SubtargetFeature> Features>
    : Processor<Name, NoItineraries, Features>;

def : Proc<"generic", []>;

//===----------------------------------------------------------------------===//
// Declare the target which we are implementing
//===----------------------------------------------------------------------===//


def CucarachaAsmWriter : AsmWriter {
  string AsmWriterClassName  = "InstPrinter";
  int PassSubtarget = 1;
  int Variant = 0;
}

def Cucaracha : Target {
  let InstructionSet = CucarachaInstrInfo;
  let AssemblyWriters = [CucarachaAsmWriter];
  let AllowRegisterRenaming = 1;
}
