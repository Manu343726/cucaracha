package disassembly

import (
	"fmt"
	"strings"

	"github.com/Manu343726/cucaracha/pkg/ui/debugger"
)

// SearchType defines the type of search to perform
type SearchType int

const (
	SearchByMnemonic SearchType = iota
	SearchByOperand
	SearchByAddress
	SearchBySymbol
	SearchBySourceLine
)

// SearchResult represents a search result
type SearchResult struct {
	Address     uint32
	Instruction *debugger.Instruction
	Context     string
}

// SearchEngine handles all search operations
type SearchEngine struct {
	instructions map[uint32]*debugger.Instruction
	indexed      []*debugger.Instruction
}

// NewSearchEngine creates a new search engine
func NewSearchEngine() *SearchEngine {
	return &SearchEngine{
		instructions: make(map[uint32]*debugger.Instruction),
		indexed:      make([]*debugger.Instruction, 0),
	}
}

// Index indexes the instructions for searching
func (s *SearchEngine) Index(instructions []*debugger.Instruction) {
	s.instructions = make(map[uint32]*debugger.Instruction)
	s.indexed = make([]*debugger.Instruction, 0, len(instructions))

	for _, instr := range instructions {
		s.instructions[instr.Address] = instr
		s.indexed = append(s.indexed, instr)
	}
}

// Search performs a search and returns matching instruction addresses
func (s *SearchEngine) Search(pattern string, searchType SearchType) []uint32 {
	var results []uint32

	pattern = strings.ToLower(strings.TrimSpace(pattern))
	if pattern == "" {
		return results
	}

	for _, instr := range s.indexed {
		if s.matches(instr, pattern, searchType) {
			results = append(results, instr.Address)
		}
	}

	return results
}

// matches checks if an instruction matches the search pattern
func (s *SearchEngine) matches(instr *debugger.Instruction, pattern string, searchType SearchType) bool {
	switch searchType {
	case SearchByMnemonic:
		return strings.Contains(strings.ToLower(instr.Mnemonic), pattern)

	case SearchByOperand:
		for _, op := range instr.Operands {
			if s.operandMatches(op, pattern) {
				return true
			}
		}
		return false

	case SearchByAddress:
		return strings.Contains(fmt.Sprintf("0x%x", instr.Address), pattern)

	case SearchBySourceLine:
		if instr.SourceLine != nil {
			return strings.Contains(strings.ToLower(instr.SourceLine.Text), pattern)
		}
		return false

	case SearchBySymbol:
		if instr.BranchTargetSym != nil {
			return strings.Contains(strings.ToLower(*instr.BranchTargetSym), pattern)
		}
		return false

	default:
		return false
	}
}

// operandMatches checks if an operand matches a pattern
func (s *SearchEngine) operandMatches(op *debugger.InstructionOperand, pattern string) bool {
	switch op.Kind {
	case debugger.OperandKindRegister:
		if op.Register != nil {
			return strings.Contains(strings.ToLower(op.Register.Name), pattern)
		}

	case debugger.OperandKindImmediate:
		if op.Immediate != nil {
			return strings.Contains(fmt.Sprintf("0x%x", *op.Immediate), pattern) ||
				strings.Contains(fmt.Sprintf("%d", *op.Immediate), pattern)
		}
	}

	return false
}

// SearchMultiple performs multiple searches and returns the intersection
func (s *SearchEngine) SearchMultiple(patterns []string, searchTypes []SearchType) []uint32 {
	if len(patterns) == 0 {
		return []uint32{}
	}

	// First search
	results := s.Search(patterns[0], searchTypes[0])
	if len(results) == 0 {
		return results
	}

	// Intersect with subsequent searches
	resultMap := make(map[uint32]bool)
	for _, addr := range results {
		resultMap[addr] = true
	}

	for i := 1; i < len(patterns); i++ {
		searchResults := s.Search(patterns[i], searchTypes[i])
		newResultMap := make(map[uint32]bool)

		for _, addr := range searchResults {
			if resultMap[addr] {
				newResultMap[addr] = true
			}
		}

		resultMap = newResultMap
		if len(resultMap) == 0 {
			return []uint32{}
		}
	}

	// Convert back to slice
	results = make([]uint32, 0, len(resultMap))
	for addr := range resultMap {
		results = append(results, addr)
	}

	return results
}

// GetInstructionsByMnemonic returns all instructions with a specific mnemonic
func (s *SearchEngine) GetInstructionsByMnemonic(mnemonic string) []*debugger.Instruction {
	var results []*debugger.Instruction

	mnemonic = strings.ToLower(mnemonic)
	for _, instr := range s.indexed {
		if strings.ToLower(instr.Mnemonic) == mnemonic {
			results = append(results, instr)
		}
	}

	return results
}

// GetInstructionCount returns the total number of indexed instructions
func (s *SearchEngine) GetInstructionCount() int {
	return len(s.indexed)
}
