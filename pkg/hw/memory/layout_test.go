package memory

import (
	"testing"

	"github.com/Manu343726/cucaracha/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomLayout(t *testing.T) {
	options := LayoutOptions{
		TotalSize:            32768,
		SystemDescriptorSize: 512,
		VectorTableSize:      64,
		CodeSize:             8192,
		DataSize:             4096,
		HeapSize:             2048,
		StackSize:            8192,
		PeripheralSize:       2048,
	}

	layout := CustomLayout(options)
	require.NoError(t, layout.Validate())

	// Check calculated bases and sizes
	expectedVectorTableBase := options.SystemDescriptorSize
	expectedVectorTableSize := options.VectorTableSize
	expectedCodeBase := expectedVectorTableBase + expectedVectorTableSize
	expectedDataBase := expectedCodeBase + options.CodeSize
	expectedHeapBase := expectedDataBase + options.DataSize
	expectedStackBase := utils.NextAligned(expectedHeapBase+options.HeapSize, 4)
	expectedStackBottom := expectedStackBase + options.StackSize - 4
	expectedStackTop := expectedStackBase
	expectedPeripheralBase := expectedStackBase + options.StackSize

	assert.Equal(t, options.TotalSize, layout.TotalSize)
	assert.Equal(t, expectedCodeBase, layout.CodeBase)
	assert.Equal(t, options.DataSize, layout.DataSize)
	assert.Equal(t, expectedStackBase, layout.StackBase)
	assert.Equal(t, options.StackSize, layout.StackSize)
	assert.Equal(t, expectedStackBottom, layout.StackBottom())
	assert.Equal(t, expectedStackTop, layout.StackTop())
	assert.Equal(t, expectedPeripheralBase, layout.PeripheralBase)
	assert.Equal(t, options.PeripheralSize, layout.PeripheralSize)
}

func TestDefaultLayout(t *testing.T) {
	layout := DefaultLayout(65536) // 64KB

	require.NoError(t, layout.Validate())

	// Check peripheral space is reserved
	if layout.PeripheralSize == 0 {
		t.Error("Should have peripheral space")
	}

	// Stack should be below peripheral space
	if layout.StackBase > layout.PeripheralBase {
		t.Error("Stack should be below peripheral space")
	}

	// StackTop should be word-aligned
	if layout.StackTop()%4 != 0 {
		t.Error("Stack top should be word-aligned")
	}
}

func TestRanges(t *testing.T) {
	layout := CustomLayout(LayoutOptions{
		TotalSize:            65536,
		SystemDescriptorSize: 256,
		VectorTableSize:      32 * 4,
		CodeSize:             16384,
		DataSize:             8192,
		HeapSize:             4096,
		StackSize:            8192,
		PeripheralSize:       4096,
	})

	systemDescriptorRange := layout.SystemDescriptor()
	assert.Equal(t, uint32(0), systemDescriptorRange.Start)
	assert.Equal(t, layout.SystemDescriptorSize, systemDescriptorRange.Size)

	vectorTableRange := layout.VectorTable()
	assert.Equal(t, layout.VectorTableBase, vectorTableRange.Start)
	assert.Equal(t, layout.VectorTableSize, vectorTableRange.Size)

	dataRange := layout.Data()
	assert.Equal(t, layout.DataBase, dataRange.Start)
	assert.Equal(t, layout.DataSize, dataRange.Size)

	codeRange := layout.Code()
	assert.Equal(t, layout.CodeBase, codeRange.Start)
	assert.Equal(t, layout.CodeSize, codeRange.Size)

	heapRange := layout.Heap()
	assert.Equal(t, layout.HeapBase, heapRange.Start)
	assert.Equal(t, layout.HeapSize, heapRange.Size)

	stackRange := layout.Stack()
	assert.Equal(t, layout.StackBase, stackRange.Start)
	assert.Equal(t, layout.StackSize, stackRange.Size)

	peripheralsRange := layout.Peripherals()
	assert.Equal(t, layout.PeripheralBase, peripheralsRange.Start)
	assert.Equal(t, layout.PeripheralSize, peripheralsRange.Size)

	ranges := layout.Ranges()
	require.Len(t, ranges, 7)
	assert.Equal(t, layout.SystemDescriptor(), ranges[0])
	assert.Equal(t, layout.VectorTable(), ranges[1])
	assert.Equal(t, layout.Code(), ranges[2])
	assert.Equal(t, layout.Data(), ranges[3])
	assert.Equal(t, layout.Heap(), ranges[4])
	assert.Equal(t, layout.Stack(), ranges[5])
	assert.Equal(t, layout.Peripherals(), ranges[6])
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name          string
		layout        MemoryLayout
		expectedError string
	}{
		{
			name: "System descriptor is after vector table",
			layout: MemoryLayout{
				TotalSize:            1024,
				SystemDescriptorBase: 900,
				SystemDescriptorSize: 256,
				VectorTableBase:      800,
				VectorTableSize:      64,
				CodeBase:             1200,
				CodeSize:             512,
				DataBase:             1700,
				DataSize:             256,
				HeapBase:             2000,
				HeapSize:             128,
				StackBase:            2200,
				StackSize:            256,
				PeripheralBase:       2500,
				PeripheralSize:       512,
			},
			expectedError: "system descriptor is after vector table",
		},
		{
			name: "System descriptor overlaps with vector table",
			layout: MemoryLayout{
				TotalSize:            1024,
				SystemDescriptorBase: 800,
				SystemDescriptorSize: 256,
				VectorTableBase:      900,
				VectorTableSize:      64,
				CodeBase:             1200,
				CodeSize:             512,
				DataBase:             1700,
				DataSize:             256,
				HeapBase:             2000,
				HeapSize:             128,
				StackBase:            2200,
				StackSize:            256,
				PeripheralBase:       2500,
				PeripheralSize:       512,
			},
			expectedError: "system descriptor overlaps with vector table",
		},
		{
			name: "System descriptor too large",
			layout: CustomLayout(LayoutOptions{
				TotalSize:            1024,
				SystemDescriptorSize: 2048, // Too large
				VectorTableSize:      64,
				CodeSize:             512,
			}),
			expectedError: "system descriptor exceeds total size",
		},
		{
			name: "Vector table is after code",
			layout: MemoryLayout{
				TotalSize:            2048,
				SystemDescriptorBase: 0,
				SystemDescriptorSize: 256,
				VectorTableBase:      2000,
				VectorTableSize:      64,
				CodeBase:             1800,
				CodeSize:             512,
				DataBase:             2500,
				DataSize:             256,
				HeapBase:             3000,
				HeapSize:             128,
				StackBase:            3200,
				StackSize:            256,
				PeripheralBase:       3500,
				PeripheralSize:       512,
			},
			expectedError: "vector table is after code",
		},
		{
			name: "Vector table overlaps with code",
			layout: MemoryLayout{
				TotalSize:            20048,
				SystemDescriptorBase: 0,
				SystemDescriptorSize: 256,
				VectorTableBase:      1500,
				VectorTableSize:      64,
				CodeBase:             1532,
				CodeSize:             512,
				DataBase:             2500,
				DataSize:             256,
				HeapBase:             3000,
				HeapSize:             128,
				StackBase:            3200,
				StackSize:            256,
				PeripheralBase:       3500,
				PeripheralSize:       512,
			},
			expectedError: "vector table overlaps with code",
		},
		{
			name: "Vector table too large",
			layout: CustomLayout(LayoutOptions{
				TotalSize:            2048,
				SystemDescriptorSize: 256,
				VectorTableSize:      2048, // Too large
				CodeSize:             512,
			}),
			expectedError: "vector table exceeds total size",
		},
		{
			name: "Code region is after data",
			layout: MemoryLayout{
				TotalSize:            4096,
				SystemDescriptorBase: 0,
				SystemDescriptorSize: 256,
				VectorTableBase:      256,
				VectorTableSize:      64,
				CodeBase:             3500,
				CodeSize:             512,
				DataBase:             3000,
				DataSize:             256,
				HeapBase:             4000,
				HeapSize:             128,
				StackBase:            4200,
				StackSize:            256,
				PeripheralBase:       4500,
				PeripheralSize:       512,
			},
			expectedError: "code region is after data",
		},
		{
			name: "Code region overlaps with data",
			layout: MemoryLayout{
				TotalSize:            4096,
				SystemDescriptorBase: 0,
				SystemDescriptorSize: 256,
				VectorTableBase:      256,
				VectorTableSize:      64,
				CodeBase:             2800,
				CodeSize:             512,
				DataBase:             3000,
				DataSize:             512,
				HeapBase:             4000,
				HeapSize:             128,
				StackBase:            4200,
				StackSize:            256,
				PeripheralBase:       4500,
				PeripheralSize:       512,
			},
			expectedError: "code region overlaps with data",
		},
		{
			name: "Code region too large",
			layout: CustomLayout(LayoutOptions{
				TotalSize:            4096,
				SystemDescriptorSize: 256,
				VectorTableSize:      64,
				CodeSize:             8192, // Too large
			}),
			expectedError: "code region exceeds total size",
		},
		{
			name: "Data region is after heap",
			layout: MemoryLayout{
				TotalSize:            8192,
				SystemDescriptorBase: 0,
				SystemDescriptorSize: 256,
				VectorTableBase:      256,
				VectorTableSize:      64,
				CodeBase:             1024,
				CodeSize:             2048,
				DataBase:             7000,
				DataSize:             512,
				HeapBase:             6000,
				HeapSize:             512,
				StackBase:            7500,
				StackSize:            512,
				PeripheralBase:       8000,
				PeripheralSize:       512,
			},
			expectedError: "data region is after heap",
		},
		{
			name: "Data region overlaps with heap",
			layout: MemoryLayout{
				TotalSize:            8192,
				SystemDescriptorBase: 0,
				SystemDescriptorSize: 256,
				VectorTableBase:      256,
				VectorTableSize:      64,
				CodeBase:             1024,
				CodeSize:             2048,
				DataBase:             5500,
				DataSize:             1024,
				HeapBase:             6000,
				HeapSize:             512,
				StackBase:            7500,
				StackSize:            512,
				PeripheralBase:       8000,
				PeripheralSize:       512,
			},
			expectedError: "data region overlaps with heap",
		},
		{
			name: "Data region too large",
			layout: CustomLayout(LayoutOptions{
				TotalSize:            8192,
				SystemDescriptorSize: 256,
				VectorTableSize:      64,
				CodeSize:             2048,
				DataSize:             8192, // Too large
			}),
			expectedError: "data region exceeds total size",
		},
		{
			name: "Heap is after stack",
			layout: MemoryLayout{
				TotalSize:            16384,
				SystemDescriptorBase: 0,
				SystemDescriptorSize: 256,
				VectorTableBase:      256,
				VectorTableSize:      64,
				CodeBase:             1024,
				CodeSize:             4096,
				DataBase:             5120,
				DataSize:             2048,
				HeapBase:             14000,
				HeapSize:             512,
				StackBase:            12000,
				StackSize:            128,
				PeripheralBase:       15000,
				PeripheralSize:       512,
			},
			expectedError: "heap is after stack",
		},
		{
			name: "Heap overlaps with stack",
			layout: MemoryLayout{
				TotalSize:            16384,
				SystemDescriptorBase: 0,
				SystemDescriptorSize: 256,
				VectorTableBase:      256,
				VectorTableSize:      64,
				CodeBase:             1024,
				CodeSize:             4096,
				DataBase:             5120,
				DataSize:             2048,
				HeapBase:             11500,
				HeapSize:             2000,
				StackBase:            12000,
				StackSize:            128,
				PeripheralBase:       15000,
				PeripheralSize:       512,
			},
			expectedError: "heap overlaps with stack",
		},
		{
			name: "Heap is too large",
			layout: CustomLayout(LayoutOptions{
				TotalSize:            16384,
				SystemDescriptorSize: 256,
				VectorTableSize:      64,
				CodeSize:             4096,
				DataSize:             2048,
				HeapSize:             20000, // Too large
				StackSize:            2048,
			}),
			expectedError: "heap exceeds total size",
		},
		{
			name: "Stack is after peripheral region",
			layout: MemoryLayout{
				TotalSize:            32768,
				SystemDescriptorBase: 0,
				SystemDescriptorSize: 256,
				VectorTableBase:      256,
				VectorTableSize:      64,
				CodeBase:             1024,
				CodeSize:             8192,
				DataBase:             9216,
				DataSize:             4096,
				HeapBase:             13312,
				HeapSize:             2048,
				StackBase:            30000,
				StackSize:            4096,
				PeripheralBase:       28000,
				PeripheralSize:       4096,
			},
			expectedError: "stack is after peripherals",
		},
		{
			name: "Stack overlaps with peripheral region",
			layout: MemoryLayout{
				TotalSize:            32768,
				SystemDescriptorBase: 0,
				SystemDescriptorSize: 256,
				VectorTableBase:      256,
				VectorTableSize:      64,
				CodeBase:             1024,
				CodeSize:             8192,
				DataBase:             9216,
				DataSize:             4096,
				HeapBase:             13312,
				HeapSize:             2048,
				StackBase:            27000,
				StackSize:            4096,
				PeripheralBase:       28000,
				PeripheralSize:       4096,
			},
			expectedError: "stack overlaps with peripherals",
		},
		{
			name: "Stack too large",
			layout: CustomLayout(LayoutOptions{
				TotalSize:            32768,
				SystemDescriptorSize: 256,
				VectorTableSize:      64,
				CodeSize:             8192,
				DataSize:             4096,
				HeapSize:             2048,
				StackSize:            40000, // Too large
				PeripheralSize:       2048,
			}),
			expectedError: "stack exceeds total size",
		},
		{
			name: "Peripheral region too large",
			layout: CustomLayout(LayoutOptions{
				TotalSize:            32768,
				SystemDescriptorSize: 256,
				VectorTableSize:      64,
				CodeSize:             8192,
				DataSize:             4096,
				HeapSize:             2048,
				StackSize:            4096,
				PeripheralSize:       40000, // Too large
			}),
			expectedError: "peripheral region exceeds total size",
		},
		{
			name: "Valid layout",
			layout: CustomLayout(LayoutOptions{
				TotalSize:            32768,
				SystemDescriptorSize: 256,
				VectorTableSize:      64,
				CodeSize:             8192,
				DataSize:             4096,
				HeapSize:             2048,
				StackSize:            8192,
				PeripheralSize:       2048,
			}),
			expectedError: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.layout.Validate()
			if test.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedError)
			}
		})
	}
}
