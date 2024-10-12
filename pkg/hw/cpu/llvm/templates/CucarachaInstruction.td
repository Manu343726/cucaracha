
// {{ .String }}
// {{ .Description }}
def {{ .OpCode.OpCode | String | ToUpper }}  : InstCucaracha<(outs {{ .LLVM.Outs | Join ", " }} ),
                   (ins {{ .LLVM.Ins | Join ", " }}),
                   "{{ .OpCode.OpCode | String }} {{ .LLVM.Params | Join ", " }}",
                   [{{ .LLVM.Pattern }}]> {
  let Inst{ {{ .OpCode.MostSignificantBit }} - {{ .OpCode.LeastSignificantBit }} } = {{ .OpCode.BinaryRepresentation | Binary .OpCode.EncodingBits }} // opcode

{{- range .LLVM.Operands}}
  bits<{{ .Bits }}> {{ .Name }}
  let Inst{ {{ .MostSignificantBit }}-{{ .LeastSignificantBit }} } = {{ .Name }}
{{- end }}

{{- range $flag, $set := .LLVM.Flags }}
  {{- if $set }}
  let {{ $flag }} = {{ $set }}
  {{- end }}
{{- end }}
}