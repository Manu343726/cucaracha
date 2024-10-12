//===----------------------------------------------------------------------===//
// Instruction format superclass
//===----------------------------------------------------------------------===//
class InstCucaracha<dag outs, dag ins, string asmstr, list<dag> pattern>
    : Instruction {
  field bits<{{ .Instructions.InstructionBits }}> Inst;
  let Namespace = "Cucaracha";
  dag OutOperandList = outs;
  dag InOperandList = ins;
  let AsmString   = asmstr;
  let Pattern = pattern;
  let Size = {{ .Instructions.InstructionBytes }};
}

// Cucaracha pseudo instructions format
class CucarachaPseudoInst<dag outs, dag ins, string asmstr, list<dag> pattern>
    : InstCucaracha<outs, ins, asmstr, pattern> {
  let isPseudo = 1;
}
