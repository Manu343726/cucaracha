package mc

import (
	"fmt"
	"io"
	"sort"
)

// WriteProgramFile writes a text (assembly) representation of a ProgramFile to the given writer.
// The output format is compatible with LLVM's .cucaracha output files and can be parsed
// back by the llvm package's assembly file parser.
func WriteProgramFile(w io.Writer, pf ProgramFile) error {
	pw := &programWriter{w: w, pf: pf}
	return pw.write()
}

type programWriter struct {
	w  io.Writer
	pf ProgramFile
}

func (pw *programWriter) write() error {
	if err := pw.writeHeader(); err != nil {
		return err
	}
	if err := pw.writeFunctions(); err != nil {
		return err
	}
	if err := pw.writeGlobals(); err != nil {
		return err
	}
	return nil
}

func (pw *programWriter) writeHeader() error {
	// Write .text section
	if _, err := fmt.Fprintln(pw.w, "\t.text"); err != nil {
		return err
	}

	// Write .file directive if source file is known
	if sf := pw.pf.SourceFile(); sf != "" {
		if _, err := fmt.Fprintf(pw.w, "\t.file\t\"%s\"\n", sf); err != nil {
			return err
		}
	}

	return nil
}

func (pw *programWriter) writeFunctions() error {
	functions := pw.pf.Functions()
	instructions := pw.pf.Instructions()
	labels := pw.pf.Labels()

	// Build label map: instruction index -> labels at that index
	labelMap := make(map[int][]Label)
	for _, lbl := range labels {
		if lbl.InstructionIndex >= 0 {
			labelMap[lbl.InstructionIndex] = append(labelMap[lbl.InstructionIndex], lbl)
		}
	}

	// Sort functions by their first instruction index for deterministic output
	funcNames := make([]string, 0, len(functions))
	for name := range functions {
		funcNames = append(funcNames, name)
	}
	sort.Slice(funcNames, func(i, j int) bool {
		fi, fj := functions[funcNames[i]], functions[funcNames[j]]
		if len(fi.InstructionRanges) == 0 {
			return true
		}
		if len(fj.InstructionRanges) == 0 {
			return false
		}
		return fi.InstructionRanges[0].Start < fj.InstructionRanges[0].Start
	})

	// Track which instructions have been written (for functions with multiple ranges)
	writtenInstructions := make(map[int]bool)

	for _, name := range funcNames {
		fn := functions[name]
		if err := pw.writeFunction(fn, instructions, labelMap, writtenInstructions); err != nil {
			return err
		}
	}

	return nil
}

func (pw *programWriter) writeFunction(fn Function, instructions []Instruction, labelMap map[int][]Label, writtenInstructions map[int]bool) error {
	// Build set of function names to avoid duplicate labels
	funcNames := make(map[string]bool)
	for name := range pw.pf.Functions() {
		funcNames[name] = true
	}

	// Write function header
	if _, err := fmt.Fprintf(pw.w, "\t.globl\t%s\n", fn.Name); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(pw.w, "\t.type\t%s,@function\n", fn.Name); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(pw.w, "%s:\n", fn.Name); err != nil {
		return err
	}

	// Write instructions for each range
	for _, r := range fn.InstructionRanges {
		for i := r.Start; i < r.Start+r.Count && i < len(instructions); i++ {
			if writtenInstructions[i] {
				continue
			}

			// Write any labels at this instruction (excluding function names)
			if lbls, ok := labelMap[i]; ok {
				for _, lbl := range lbls {
					if !funcNames[lbl.Name] {
						if _, err := fmt.Fprintf(pw.w, "%s:\n", lbl.Name); err != nil {
							return err
						}
					}
				}
			}

			// Write the instruction
			inst := instructions[i]
			if _, err := fmt.Fprintf(pw.w, "\t%s\n", inst.Text); err != nil {
				return err
			}
			writtenInstructions[i] = true
		}
	}

	// Write function end marker and size
	funcEndLabel := fmt.Sprintf(".Lfunc_end%s", fn.Name)
	if _, err := fmt.Fprintf(pw.w, "%s:\n", funcEndLabel); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(pw.w, "\t.size\t%s, %s-%s\n", fn.Name, funcEndLabel, fn.Name); err != nil {
		return err
	}

	return nil
}

func (pw *programWriter) writeGlobals() error {
	globals := pw.pf.Globals()
	if len(globals) == 0 {
		return nil
	}

	for _, g := range globals {
		if g.Type != GlobalObject {
			continue
		}

		// Write type directive
		if _, err := fmt.Fprintf(pw.w, "\t.type\t%s,@object\n", g.Name); err != nil {
			return err
		}

		// Write section directive for read-only data
		if _, err := fmt.Fprintln(pw.w, "\t.section\t.rodata,\"a\",@progbits"); err != nil {
			return err
		}

		// Write alignment (default to 4-byte alignment)
		if _, err := fmt.Fprintln(pw.w, "\t.p2align\t2, 0x0"); err != nil {
			return err
		}

		// Write label
		if _, err := fmt.Fprintf(pw.w, "%s:\n", g.Name); err != nil {
			return err
		}

		// Write data
		if err := pw.writeGlobalData(g); err != nil {
			return err
		}

		// Write size
		if _, err := fmt.Fprintf(pw.w, "\t.size\t%s, %d\n", g.Name, g.Size); err != nil {
			return err
		}

		// Add blank line between globals
		if _, err := fmt.Fprintln(pw.w); err != nil {
			return err
		}
	}

	return nil
}

func (pw *programWriter) writeGlobalData(g Global) error {
	data := g.InitialData

	// Try to write as .long directives (4 bytes each) when possible
	i := 0
	for i+4 <= len(data) {
		val := int32(data[i]) | int32(data[i+1])<<8 | int32(data[i+2])<<16 | int32(data[i+3])<<24
		if _, err := fmt.Fprintf(pw.w, "\t.long\t%d\n", val); err != nil {
			return err
		}
		i += 4
	}

	// Write remaining bytes individually
	for ; i < len(data); i++ {
		if _, err := fmt.Fprintf(pw.w, "\t.byte\t%d\n", data[i]); err != nil {
			return err
		}
	}

	// If no initial data but size is known, use .zero
	if len(data) == 0 && g.Size > 0 {
		if _, err := fmt.Fprintf(pw.w, "\t.zero\t%d\n", g.Size); err != nil {
			return err
		}
	}

	return nil
}
