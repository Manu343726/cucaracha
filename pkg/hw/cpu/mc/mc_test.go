package mc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocumentation(t *testing.T) {
	assert.Equal(t, ``, Descriptor.Documentation(0))
}
