package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsciiFrame_NoFields(t *testing.T) {
	fields := []AsciiFrameField{}

	actual, err := AsciiFrame(fields, 16, "bits", AsciiFrameUnitLayout_RightToLeft, 0)
	assert.NoError(t, err)

	assert.Equal(t, ""+
		`15            0
+-------------+
|  (unused)   |
+-------------+
 <- 16 bits -> 
`,
		actual)
}

func TestAsciiFrame_SingleField(t *testing.T) {
	fields := []AsciiFrameField{
		{
			Name:  "first field",
			Begin: 0,
			Width: 16,
		},
	}

	actual, err := AsciiFrame(fields, 16, "bits", AsciiFrameUnitLayout_RightToLeft, 0)
	assert.NoError(t, err)

	assert.Equal(t, ""+
		`15            0
+-------------+
| first field |
+-------------+
 <- 16 bits -> 
`,
		actual)
}

func TestAsciiFrame_SingleField_NotFittingFullFrame(t *testing.T) {
	fields := []AsciiFrameField{
		{
			Name:  "first field",
			Begin: 0,
			Width: 16,
		},
	}

	actual, err := AsciiFrame(fields, 32, "bits", AsciiFrameUnitLayout_RightToLeft, 0)
	assert.NoError(t, err)

	assert.Equal(t, ""+
		`31            15            0
+-------------+-------------+
|  (unused)   | first field |
+-------------+-------------+
 <- 16 bits -> <- 16 bits -> 
`,
		actual)
}

func TestAsciiFrame_SingleField_WithTextPadding(t *testing.T) {
	fields := []AsciiFrameField{
		{
			Name:  "first field",
			Begin: 0,
			Width: 16,
		},
	}

	actual, err := AsciiFrame(fields, 16, "bits", AsciiFrameUnitLayout_RightToLeft, 4)
	assert.NoError(t, err)

	assert.Equal(t, ""+
		`    15            0
    +-------------+
    | first field |
    +-------------+
     <- 16 bits -> 
`,
		actual)
}

func TestAsciiFrame_AVeryLooooooongField(t *testing.T) {
	fields := []AsciiFrameField{
		{
			Name:  "a very loooooooooong field",
			Begin: 0,
			Width: 16,
		},
	}

	actual, err := AsciiFrame(fields, 16, "bits", AsciiFrameUnitLayout_RightToLeft, 0)

	assert.NoError(t, err)

	assert.Equal(t, ""+
		`15                           0
+----------------------------+
| a very loooooooooong field |
+----------------------------+
 <-------- 16 bits ---------> 
`,
		actual)
}

func TestAsciiFrame_TwoConsecutiveFields(t *testing.T) {
	fields := []AsciiFrameField{
		{
			Name:  "first field",
			Begin: 0,
			Width: 16,
		},
		{
			Name:  "second field",
			Begin: 16,
			Width: 16,
		},
	}

	actual, err := AsciiFrame(fields, 32, "bits", AsciiFrameUnitLayout_RightToLeft, 0)
	assert.NoError(t, err)

	assert.Equal(t, ""+
		`31             15            0
+--------------+-------------+
| second field | first field |
+--------------+-------------+
 <- 16 bits --> <- 16 bits -> 
`,
		actual)
}

func TestAsciiFrame_TwoConsecutiveFields_LeftToRight(t *testing.T) {
	fields := []AsciiFrameField{
		{
			Name:  "first field",
			Begin: 0,
			Width: 16,
		},
		{
			Name:  "second field",
			Begin: 16,
			Width: 16,
		},
	}

	actual, err := AsciiFrame(fields, 32, "bits", AsciiFrameUnitLayout_LeftToRight, 0)
	assert.NoError(t, err)

	assert.Equal(t, ""+
		`0             16             31
+-------------+--------------+
| first field | second field |
+-------------+--------------+
 <- 16 bits -> <- 16 bits --> 
`,
		actual)
}

func TestAsciiFrame_TwoFieldsWithAGap(t *testing.T) {
	fields := []AsciiFrameField{
		{
			Name:  "first field",
			Begin: 0,
			Width: 16,
		},
		{
			Name:  "second field",
			Begin: 20,
			Width: 16,
		},
	}

	actual, err := AsciiFrame(fields, 36, "bits", AsciiFrameUnitLayout_LeftToRight, 0)
	assert.NoError(t, err)

	assert.Equal(t, ""+
		`0             16           20             35
+-------------+------------+--------------+
| first field |  (unused)  | second field |
+-------------+------------+--------------+
 <- 16 bits -> <- 4 bits -> <- 16 bits --> 
`,
		actual)
}
