package mc

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// DumpProgramFile writes a detailed debugging representation of a ProgramFile to the given writer.
// This output is intended for debugging and inspection, not for parsing.
func DumpProgramFile(w io.Writer, pf ProgramFile) error {
	d := &programDumper{w: w, pf: pf}
	return d.dump()
}

type programDumper struct {
	w  io.Writer
	pf ProgramFile
}

func (d *programDumper) dump() error {
	if err := d.dumpHeader(); err != nil {
		return err
	}
	if err := d.dumpMemoryLayout(); err != nil {
		return err
	}
	if err := d.dumpFunctions(); err != nil {
		return err
	}
	if err := d.dumpLabels(); err != nil {
		return err
	}
	if err := d.dumpGlobals(); err != nil {
		return err
	}
	if err := d.dumpInstructions(); err != nil {
		return err
	}
	return nil
}

func (d *programDumper) dumpHeader() error {
	fmt.Fprintln(d.w, "=== Program File ===")
	fmt.Fprintf(d.w, "File: %s\n", d.pf.FileName())
	fmt.Fprintf(d.w, "Source: %s\n", d.pf.SourceFile())
	fmt.Fprintln(d.w)
	return nil
}

func (d *programDumper) dumpMemoryLayout() error {
	layout := d.pf.MemoryLayout()
	fmt.Fprintln(d.w, "=== Memory Layout ===")
	if layout == nil {
		fmt.Fprintln(d.w, "(not resolved)")
	} else {
		fmt.Fprintf(d.w, "Base Address: 0x%08X\n", layout.BaseAddress)
		fmt.Fprintf(d.w, "Total Size:   %d bytes\n", layout.TotalSize)
		fmt.Fprintf(d.w, "Code Section: 0x%08X - 0x%08X (%d bytes)\n",
			layout.CodeStart, layout.CodeStart+layout.CodeSize, layout.CodeSize)
		fmt.Fprintf(d.w, "Data Section: 0x%08X - 0x%08X (%d bytes)\n",
			layout.DataStart, layout.DataStart+layout.DataSize, layout.DataSize)
	}
	fmt.Fprintln(d.w)
	return nil
}

func (d *programDumper) dumpFunctions() error {
	functions := d.pf.Functions()
	fmt.Fprintf(d.w, "=== Functions (%d) ===\n", len(functions))

	if len(functions) == 0 {
		fmt.Fprintln(d.w, "(none)")
		fmt.Fprintln(d.w)
		return nil
	}

	// Sort by name for deterministic output
	names := make([]string, 0, len(functions))
	for name := range functions {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fn := functions[name]
		fmt.Fprintf(d.w, "  %s:\n", name)
		if fn.SourceFile != "" {
			fmt.Fprintf(d.w, "    Source: %s", fn.SourceFile)
			if fn.StartLine > 0 {
				fmt.Fprintf(d.w, ":%d", fn.StartLine)
				if fn.EndLine > fn.StartLine {
					fmt.Fprintf(d.w, "-%d", fn.EndLine)
				}
			}
			fmt.Fprintln(d.w)
		}
		if len(fn.InstructionRanges) > 0 {
			fmt.Fprintf(d.w, "    Instruction Ranges: ")
			for i, r := range fn.InstructionRanges {
				if i > 0 {
					fmt.Fprint(d.w, ", ")
				}
				fmt.Fprintf(d.w, "[%d..%d]", r.Start, r.Start+r.Count-1)
			}
			fmt.Fprintln(d.w)
		}
	}
	fmt.Fprintln(d.w)
	return nil
}

func (d *programDumper) dumpLabels() error {
	labels := d.pf.Labels()
	fmt.Fprintf(d.w, "=== Labels (%d) ===\n", len(labels))

	if len(labels) == 0 {
		fmt.Fprintln(d.w, "(none)")
		fmt.Fprintln(d.w)
		return nil
	}

	for _, lbl := range labels {
		if lbl.InstructionIndex >= 0 {
			fmt.Fprintf(d.w, "  %s -> instruction #%d\n", lbl.Name, lbl.InstructionIndex)
		} else {
			fmt.Fprintf(d.w, "  %s -> (unresolved)\n", lbl.Name)
		}
	}
	fmt.Fprintln(d.w)
	return nil
}

func (d *programDumper) dumpGlobals() error {
	globals := d.pf.Globals()
	fmt.Fprintf(d.w, "=== Globals (%d) ===\n", len(globals))

	if len(globals) == 0 {
		fmt.Fprintln(d.w, "(none)")
		fmt.Fprintln(d.w)
		return nil
	}

	for _, g := range globals {
		typeStr := "unknown"
		switch g.Type {
		case GlobalFunction:
			typeStr = "function"
		case GlobalObject:
			typeStr = "object"
		}

		addrStr := "(unresolved)"
		if g.Address != nil {
			addrStr = fmt.Sprintf("0x%08X", *g.Address)
		}

		fmt.Fprintf(d.w, "  %s:\n", g.Name)
		fmt.Fprintf(d.w, "    Type: %s\n", typeStr)
		fmt.Fprintf(d.w, "    Address: %s\n", addrStr)
		fmt.Fprintf(d.w, "    Size: %d bytes\n", g.Size)
		if len(g.InitialData) > 0 {
			fmt.Fprintf(d.w, "    Data: %s\n", formatBytes(g.InitialData))
		}
	}
	fmt.Fprintln(d.w)
	return nil
}

func (d *programDumper) dumpInstructions() error {
	instructions := d.pf.Instructions()
	fmt.Fprintf(d.w, "=== Instructions (%d) ===\n", len(instructions))

	if len(instructions) == 0 {
		fmt.Fprintln(d.w, "(none)")
		return nil
	}

	// Build label map for display
	labelMap := make(map[int][]string)
	for _, lbl := range d.pf.Labels() {
		if lbl.InstructionIndex >= 0 {
			labelMap[lbl.InstructionIndex] = append(labelMap[lbl.InstructionIndex], lbl.Name)
		}
	}

	// Build function start map
	funcStartMap := make(map[int]string)
	for name, fn := range d.pf.Functions() {
		if len(fn.InstructionRanges) > 0 {
			funcStartMap[fn.InstructionRanges[0].Start] = name
		}
	}

	for i, instr := range instructions {
		// Show function start
		if funcName, ok := funcStartMap[i]; ok {
			fmt.Fprintf(d.w, "\n  ; function: %s\n", funcName)
		}

		// Show labels
		if labels, ok := labelMap[i]; ok {
			for _, lbl := range labels {
				fmt.Fprintf(d.w, "  %s:\n", lbl)
			}
		}

		// Instruction index and address
		addrStr := "--------"
		if instr.Address != nil {
			addrStr = fmt.Sprintf("%08X", *instr.Address)
		}
		fmt.Fprintf(d.w, "  [%4d] 0x%s  ", i, addrStr)

		// Raw instruction info
		if instr.Raw != nil {
			fmt.Fprintf(d.w, "%s  ", instr.Raw.String())
		} else {
			fmt.Fprint(d.w, "(no raw)  ")
		}

		// Decoded instruction info
		if instr.Instruction != nil {
			fmt.Fprint(d.w, "[decoded]  ")
		} else {
			fmt.Fprint(d.w, "(not decoded)  ")
		}

		// Assembly text
		fmt.Fprint(d.w, instr.Text)

		// Symbol references
		if len(instr.Symbols) > 0 {
			fmt.Fprint(d.w, "  ; refs: ")
			for j, sym := range instr.Symbols {
				if j > 0 {
					fmt.Fprint(d.w, ", ")
				}
				fmt.Fprint(d.w, formatSymbolRef(&sym))
			}
		}

		// Line number
		if instr.LineNumber > 0 {
			fmt.Fprintf(d.w, "  ; line %d", instr.LineNumber)
		}

		fmt.Fprintln(d.w)
	}
	return nil
}

func formatBytes(data []byte) string {
	if len(data) == 0 {
		return "(empty)"
	}

	const maxDisplay = 32
	var sb strings.Builder

	for i, b := range data {
		if i >= maxDisplay {
			sb.WriteString(fmt.Sprintf("... (%d more bytes)", len(data)-maxDisplay))
			break
		}
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(fmt.Sprintf("%02X", b))
	}

	return sb.String()
}

func formatSymbolRef(sym *SymbolReference) string {
	var sb strings.Builder
	sb.WriteString(sym.Name)

	switch sym.Usage {
	case SymbolUsageLo:
		sb.WriteString("@lo")
	case SymbolUsageHi:
		sb.WriteString("@hi")
	}

	sb.WriteString(" (")
	switch sym.Kind() {
	case SymbolKindFunction:
		sb.WriteString("func")
	case SymbolKindGlobal:
		sb.WriteString("global")
	case SymbolKindLabel:
		sb.WriteString("label")
	default:
		sb.WriteString("unresolved")
	}
	sb.WriteString(")")

	return sb.String()
}
