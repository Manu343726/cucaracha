//===-- CucarachaRegisterInfo.td - Cucaracha Register defs ---------*- tablegen -*-===//
//
//                     The LLVM Compiler Infrastructure
//
// This file is distributed under the University of Illinois Open Source
// License. See LICENSE.TXT for details.
//
//===----------------------------------------------------------------------===//

//===----------------------------------------------------------------------===//
//  Declarations that describe the Cucaracha register file 
//===----------------------------------------------------------------------===//

class CucarachaReg<bits<16> Enc, string n> : Register<n> {
  let HWEncoding = Enc;
  let Namespace = "Cucaracha";
}

// CPU registers

{{ range .RegisterClasses.AllClasses }}
//
// {{ .Description }}
//
{{ range .AllRegisters }}
// {{ .Description }}
def {{ .Name | ToUpper }}  : CucarachaReg<{{ .Encode }}, "{{ .Name }}">;
{{ end }}
{{ end }}

// Register classes.
{{ range .RegisterMetaClasses }}
def {{ .Name }} : RegisterClass<"Cucaracha",
  [{{ .ValueType | LLVMType }}],
  {{ .ValueType.Bits }},
  (add {{ .AllRegisters | MapMember "Name" | MapStrings "ToUpper" | Join ", " }})>;
{{ end }}
