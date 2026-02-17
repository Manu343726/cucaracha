//===-- CucarachaConditionCodes.td - Cucaracha Condition Codes -*- tablegen -*-===//
//
//                     The LLVM Compiler Infrastructure
//
// This file is distributed under the University of Illinois Open Source
// License. See LICENSE.TXT for details.
//
//===----------------------------------------------------------------------===//
//
// This file defines condition codes for the Cucaracha target.
// AUTO-GENERATED - Do not edit manually
//
// Uses LLVM's SearchableTables to generate C++ enums and lookup functions.
//
//===----------------------------------------------------------------------===//

include "llvm/TableGen/SearchableTable.td"

//===----------------------------------------------------------------------===//
// CPSR Flags - Individual flags in the Current Program Status Register
//===----------------------------------------------------------------------===//

class CPSRFlagEntry<bits<4> val, string nameStr> {
  bits<4> Value = val;
  string Name = nameStr;
}

{{- range .ConditionCodes.Flags }}
def {{ .Name }} : CPSRFlagEntry<{{ printf "0x%X" .Value }}, "{{ .Name }}">;
{{- end }}

def CPSRFlagsList : GenericEnum {
  let FilterClass = "CPSRFlagEntry";
  let NameField = "Name";
  let ValueField = "Value";
}

//===----------------------------------------------------------------------===//
// Condition Code Definition
//===----------------------------------------------------------------------===//

class CucarachaCondCodeEntry<bits<4> codeVal, string nameStr,
                             string oppositeStr, string testStr> {
  bits<4> Code = codeVal;
  string Name = nameStr;
  string OppositeName = oppositeStr;
  string TestExpr = testStr;
}

{{- range $i, $cc := .ConditionCodes.CondCodes }}
def COND_{{ $cc.Name }} : CucarachaCondCodeEntry<{{ $cc.Code }}, "COND_{{ $cc.Name }}", "COND_{{ $cc.OppositeName }}", "{{ $cc.TestExpr }}">;
{{- end }}
def COND_INVALID : CucarachaCondCodeEntry<15, "COND_INVALID", "COND_INVALID", "false">;

def CucarachaCondCodesList : GenericEnum {
  let FilterClass = "CucarachaCondCodeEntry";
  let NameField = "Name";
  let ValueField = "Code";
}

def CucarachaCondCodeLookup : GenericTable {
  let FilterClass = "CucarachaCondCodeEntry";
  let CppTypeName = "CondCodeInfo";
  let Fields = ["Code", "Name", "OppositeName", "TestExpr"];
  let PrimaryKey = ["Code"];
  let PrimaryKeyName = "lookupCondCodeByCode";
}

def CucarachaCondCodeByName : SearchIndex {
  let Table = CucarachaCondCodeLookup;
  let Key = ["Name"];
}

//===----------------------------------------------------------------------===//
// Condition Code Operand for Assembly Parsing
//===----------------------------------------------------------------------===//

def CucarachaConditionCodeOperand : AsmOperandClass {
  let Name = "ConditionCode";
  let RenderMethod = "addCondCodeOperands";
  let PredicateMethod = "isCondCode";
  let DiagnosticType = "InvalidCondCode";
}

def cond_code : Operand<i32> {
  let PrintMethod = "printCondCode";
  let ParserMatchClass = CucarachaConditionCodeOperand;
  let DecoderMethod = "DecodeCondCode";
}

//===----------------------------------------------------------------------===//
// Convenience ImmLeaf definitions for DAG patterns
//===----------------------------------------------------------------------===//

// Flag value matchers
{{- range .ConditionCodes.Flags }}
def Is{{ .Name }} : ImmLeaf<i32, [{ return Imm == {{ printf "0x%X" .Value }}; }]>;
{{- end }}
