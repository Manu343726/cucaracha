package instructions

import (
	"testing"

	"github.com/Manu343726/cucaracha/pkg/hw/cpu/mc"
	"github.com/stretchr/testify/assert"
)

func TestAddFactory(t *testing.T) {
	instr := Add(1, 2, 3)

	assert.Equal(t, instr.OperandValues, []uint64{1, 2, 3})

	t.Logf("%v\n\n%v", instr, instr.PrettyPrint(0))

	binaryRepresentation := instr.Encode()
	decodedInstr, err := mc.DecodeInstruction(binaryRepresentation)

	assert.Nil(t, err)
	assert.NotNil(t, decodedInstr)
	assert.Equal(t, instr, decodedInstr)
}
