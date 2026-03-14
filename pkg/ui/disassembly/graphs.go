package disassembly

import (
	"strings"

	"github.com/Manu343726/cucaracha/pkg/ui/debugger"
)

// InstructionDependencyGraph represents data and control dependencies between instructions
type InstructionDependencyGraph struct {
	// Map from instruction address to addresses it depends on
	dependencies map[uint32][]uint32
	// Map from instruction address to addresses that depend on it
	dependents map[uint32][]uint32
	// Register usage: which instruction last wrote to each register before this one
	registerLastWriter map[string]uint32
	// Memory usage tracking
	memoryAliases map[uint32][]uint32
	edgeCount     int
}

// NewInstructionDependencyGraph builds a dependency graph from disassembled instructions
func NewInstructionDependencyGraph(result *debugger.DisasmResult) *InstructionDependencyGraph {
	graph := &InstructionDependencyGraph{
		dependencies:       make(map[uint32][]uint32),
		dependents:         make(map[uint32][]uint32),
		registerLastWriter: make(map[string]uint32),
		memoryAliases:      make(map[uint32][]uint32),
		edgeCount:          0,
	}

	if result == nil || len(result.Instructions) == 0 {
		return graph
	}

	// Build the dependency graph
	for _, instr := range result.Instructions {
		graph.analyzeInstruction(instr, result.Instructions)
	}

	return graph
}

// analyzeInstruction analyzes an instruction's dependencies
func (g *InstructionDependencyGraph) analyzeInstruction(instr *debugger.Instruction, allInstrs []*debugger.Instruction) {
	deps := make([]uint32, 0)

	// Analyze operands to find register dependencies
	for _, op := range instr.Operands {
		if op.Kind == debugger.OperandKindRegister && op.Register != nil {
			// Check if this register was written before
			if lastWriter, ok := g.registerLastWriter[op.Register.Name]; ok {
				deps = append(deps, lastWriter)
			}
		}
	}

	// Update register tracking for operands that are written
	// (This is a simplified analysis - a real implementation would track read vs write)
	for _, op := range instr.Operands {
		if op.Kind == debugger.OperandKindRegister && op.Register != nil {
			g.registerLastWriter[op.Register.Name] = instr.Address
		}
	}

	// Remove duplicates
	depMap := make(map[uint32]bool)
	for _, dep := range deps {
		depMap[dep] = true
	}

	uniqueDeps := make([]uint32, 0, len(depMap))
	for dep := range depMap {
		uniqueDeps = append(uniqueDeps, dep)
		// Add reverse mapping for dependents
		g.dependents[dep] = append(g.dependents[dep], instr.Address)
		g.edgeCount++
	}

	g.dependencies[instr.Address] = uniqueDeps
}

// GetDependencies returns the addresses of instructions that this instruction depends on
func (g *InstructionDependencyGraph) GetDependencies(addr uint32) []uint32 {
	if deps, ok := g.dependencies[addr]; ok {
		return deps
	}
	return []uint32{}
}

// GetDependents returns the addresses of instructions that depend on this instruction
func (g *InstructionDependencyGraph) GetDependents(addr uint32) []uint32 {
	if dependents, ok := g.dependents[addr]; ok {
		return dependents
	}
	return []uint32{}
}

// EdgeCount returns the total number of dependency edges
func (g *InstructionDependencyGraph) EdgeCount() int {
	return g.edgeCount
}

// GetCriticalPath returns the critical dependency path (longest chain of dependencies)
func (g *InstructionDependencyGraph) GetCriticalPath(startAddr uint32) []uint32 {
	path := make([]uint32, 0)
	visited := make(map[uint32]bool)

	g.buildCriticalPath(startAddr, &path, visited)
	return path
}

// buildCriticalPath recursively builds the critical path
func (g *InstructionDependencyGraph) buildCriticalPath(addr uint32, path *[]uint32, visited map[uint32]bool) {
	if visited[addr] {
		return
	}

	visited[addr] = true
	*path = append(*path, addr)

	deps := g.GetDependencies(addr)
	for _, dep := range deps {
		g.buildCriticalPath(dep, path, visited)
	}
}

// JumpGraph represents branch and jump relationships
type JumpGraph struct {
	// Map from source address to target address
	Edges map[uint32]uint32
	// Reverse map: from target to sources
	ReverseEdges map[uint32][]uint32
	// Branch types: conditional vs unconditional
	BranchTypes map[uint32]string
	// Call graph edges (for function calls)
	CallEdges map[uint32][]uint32
}

// NewJumpGraph builds a jump graph from disassembled instructions
func NewJumpGraph(result *debugger.DisasmResult) *JumpGraph {
	graph := &JumpGraph{
		Edges:        make(map[uint32]uint32),
		ReverseEdges: make(map[uint32][]uint32),
		BranchTypes:  make(map[uint32]string),
		CallEdges:    make(map[uint32][]uint32),
	}

	if result == nil || len(result.Instructions) == 0 {
		return graph
	}

	// Build from control flow graph if available
	if result.ControlFlowGraph != nil {
		for source, target := range result.ControlFlowGraph.Edges {
			graph.addEdge(source, target)
		}
	}

	// Also extract from branch instructions
	for _, instr := range result.Instructions {
		if instr.BranchTarget != nil {
			graph.addEdge(instr.Address, *instr.BranchTarget)

			// Classify branch type based on mnemonic
			branchType := classifyBranch(instr.Mnemonic)
			graph.BranchTypes[instr.Address] = branchType

			// Track function calls
			if branchType == "call" {
				graph.CallEdges[instr.Address] = append(graph.CallEdges[instr.Address], *instr.BranchTarget)
			}
		}
	}

	return graph
}

// addEdge adds an edge to the jump graph
func (g *JumpGraph) addEdge(source, target uint32) {
	g.Edges[source] = target
	g.ReverseEdges[target] = append(g.ReverseEdges[target], source)
}

// GetTargets returns the jump targets for an instruction
func (g *JumpGraph) GetTargets(addr uint32) []uint32 {
	if target, ok := g.Edges[addr]; ok {
		return []uint32{target}
	}
	return []uint32{}
}

// GetSources returns the instructions that jump to an address
func (g *JumpGraph) GetSources(addr uint32) []uint32 {
	return g.ReverseEdges[addr]
}

// GetBranchType returns the type of branch for an instruction
func (g *JumpGraph) GetBranchType(addr uint32) string {
	if bt, ok := g.BranchTypes[addr]; ok {
		return bt
	}
	return "unknown"
}

// IsLoopTarget checks if an address is a loop target (backward jump)
func (g *JumpGraph) IsLoopTarget(addr uint32) bool {
	for source, target := range g.Edges {
		if target == addr && source > addr {
			return true
		}
	}
	return false
}

// GetLoopBack returns the loop-back edges (backward jumps)
func (g *JumpGraph) GetLoopBack() map[uint32]uint32 {
	loopbacks := make(map[uint32]uint32)
	for source, target := range g.Edges {
		if source > target {
			loopbacks[source] = target
		}
	}
	return loopbacks
}

// classifyBranch classifies a branch instruction type
func classifyBranch(mnemonic string) string {
	mnemonic = strings.ToLower(mnemonic)

	switch {
	case strings.Contains(mnemonic, "call"):
		return "call"
	case strings.Contains(mnemonic, "ret"):
		return "return"
	case strings.HasPrefix(mnemonic, "b") || strings.Contains(mnemonic, "jmp"):
		if strings.Contains(mnemonic, "l") && !strings.Contains(mnemonic, "eq") {
			return "branch_link"
		}
		return "branch"
	case strings.HasPrefix(mnemonic, "j"):
		return "jump"
	default:
		return "unknown"
	}
}
