package reflect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStructFieldsParsed verifies that struct fields are properly parsed and attached to struct types
func TestStructFieldsParsed(t *testing.T) {
	// Create a temporary test package
	tmpDir := t.TempDir()

	// Write a test file with a struct
	testCode := `package testpkg

type Person struct {
	// Name is the person's full name
	Name string
	
	// Age is the person's age in years
	Age int
	
	// Email is optional
	Email string
}
`

	testFile := filepath.Join(tmpDir, "person.go")
	err := os.WriteFile(testFile, []byte(testCode), 0644)
	require.NoError(t, err)

	// Parse the package
	pkg, _, err := ParsePackageWithOptions(tmpDir, DefaultParsingOptions())
	require.NoError(t, err)

	// Get the Person type
	personType, exists := pkg.Types["Person"]
	require.True(t, exists, "Person type should be in package")
	require.Equal(t, TypeKindStruct, personType.Kind, "Person should be a struct")

	// Check fields
	require.NotNil(t, personType.Fields, "Struct should have fields slice")
	require.Equal(t, 3, len(personType.Fields), "Person should have 3 fields")

	// Check first field
	assert.Equal(t, "Name", personType.Fields[0].Name)
	assert.Equal(t, "string", personType.Fields[0].Type.Name)
	assert.NotEmpty(t, personType.Fields[0].Doc, "Name field should have documentation")

	// Check second field
	assert.Equal(t, "Age", personType.Fields[1].Name)
	assert.Equal(t, "int", personType.Fields[1].Type.Name)
	assert.NotEmpty(t, personType.Fields[1].Doc, "Age field should have documentation")

	// Check third field
	assert.Equal(t, "Email", personType.Fields[2].Name)
	assert.Equal(t, "string", personType.Fields[2].Type.Name)
	assert.NotEmpty(t, personType.Fields[2].Doc, "Email field should have documentation")
}

// TestStructMethodsAttached verifies that struct methods are parsed and attached to their types
func TestStructMethodsAttached(t *testing.T) {
	// Create a temporary test package
	tmpDir := t.TempDir()

	// Write a test file with a struct and methods
	testCode := `package testpkg

type Calculator struct {
	value int
}

// Add adds a number to the calculator's value
func (c *Calculator) Add(x int) int {
	c.value += x
	return c.value
}

// Subtract subtracts a number from the calculator's value
func (c *Calculator) Subtract(x int) int {
	c.value -= x
	return c.value
}

// GetValue returns the current value
func (c Calculator) GetValue() int {
	return c.value
}
`

	testFile := filepath.Join(tmpDir, "calculator.go")
	err := os.WriteFile(testFile, []byte(testCode), 0644)
	require.NoError(t, err)

	// Parse the package
	pkg, _, err := ParsePackageWithOptions(tmpDir, DefaultParsingOptions())
	require.NoError(t, err)

	// Get the Calculator type
	calcType, exists := pkg.Types["Calculator"]
	require.True(t, exists, "Calculator type should be in package")
	require.Equal(t, TypeKindStruct, calcType.Kind, "Calculator should be a struct")

	// Check that methods are attached
	require.NotNil(t, calcType.Methods, "Struct should have methods")
	require.Equal(t, 3, len(calcType.Methods), "Calculator should have 3 methods")

	// Check method names and documentation
	methodNames := make(map[string]*Method)
	for _, method := range calcType.Methods {
		methodNames[method.Name] = method
	}

	assert.Contains(t, methodNames, "Add", "Calculator should have Add method")
	assert.Contains(t, methodNames, "Subtract", "Calculator should have Subtract method")
	assert.Contains(t, methodNames, "GetValue", "Calculator should have GetValue method")

	// Verify documentation
	assert.NotEmpty(t, methodNames["Add"].Doc, "Add method should have documentation")
	assert.NotEmpty(t, methodNames["Subtract"].Doc, "Subtract method should have documentation")
	assert.NotEmpty(t, methodNames["GetValue"].Doc, "GetValue method should have documentation")

	// Verify method signatures
	addMethod := methodNames["Add"]
	require.Greater(t, len(addMethod.Args), 0, "Add method should have parameters")
	require.Greater(t, len(addMethod.Results), 0, "Add method should have results")
}

// TestInterfaceMethodsAttached verifies that interface methods are properly attached
func TestInterfaceMethodsAttached(t *testing.T) {
	// Create a temporary test package
	tmpDir := t.TempDir()

	// Write a test file with an interface
	testCode := `package testpkg

type Reader interface {
	// Read reads data into the buffer
	Read(p []byte) (n int, err error)
	
	// Close closes the reader
	Close() error
}

type Writer interface {
	// Write writes data from the buffer
	Write(p []byte) (n int, err error)
}
`

	testFile := filepath.Join(tmpDir, "interfaces.go")
	err := os.WriteFile(testFile, []byte(testCode), 0644)
	require.NoError(t, err)

	// Parse the package
	pkg, _, err := ParsePackageWithOptions(tmpDir, DefaultParsingOptions())
	require.NoError(t, err)

	// Get the Reader interface
	readerType, exists := pkg.Types["Reader"]
	require.True(t, exists, "Reader type should be in package")
	require.Equal(t, TypeKindInterface, readerType.Kind, "Reader should be an interface")

	// Check methods
	require.Equal(t, 2, len(readerType.Methods), "Reader should have 2 methods")

	methodNames := make(map[string]*Method)
	for _, method := range readerType.Methods {
		methodNames[method.Name] = method
	}

	assert.Contains(t, methodNames, "Read", "Reader should have Read method")
	assert.Contains(t, methodNames, "Close", "Reader should have Close method")

	// Verify documentation
	assert.NotEmpty(t, methodNames["Read"].Doc, "Read method should have documentation")
	assert.NotEmpty(t, methodNames["Close"].Doc, "Close method should have documentation")

	// Get the Writer interface
	writerType, exists := pkg.Types["Writer"]
	require.True(t, exists, "Writer type should be in package")
	require.Equal(t, TypeKindInterface, writerType.Kind, "Writer should be an interface")
	require.Equal(t, 1, len(writerType.Methods), "Writer should have 1 method")

	assert.Equal(t, "Write", writerType.Methods[0].Name)
	assert.NotEmpty(t, writerType.Methods[0].Doc, "Write method should have documentation")
}

// TestPrivateFieldsAndMethods verifies that both public and private fields and methods are parsed
func TestPrivateFieldsAndMethods(t *testing.T) {
	// Create a temporary test package
	tmpDir := t.TempDir()

	// Write a test file with public and private fields/methods
	testCode := `package testpkg

type Container struct {
	// Public field
	Value string
	
	// private field - lowercase
	secret int
}

// Public method
func (c *Container) Public() string {
	return c.Value
}

// private method - lowercase
func (c *Container) private() {
	// private implementation
}
`

	testFile := filepath.Join(tmpDir, "container.go")
	err := os.WriteFile(testFile, []byte(testCode), 0644)
	require.NoError(t, err)

	// Parse the package - reflect parses all items, filtering is done elsewhere
	pkg, _, err := ParsePackageWithOptions(tmpDir, DefaultParsingOptions())
	require.NoError(t, err)

	containerType := pkg.Types["Container"]
	require.NotNil(t, containerType)

	// Reflect package parses all fields
	require.Equal(t, 2, len(containerType.Fields), "Should have both public and private fields")

	fieldsByName := make(map[string]*Field)
	for _, field := range containerType.Fields {
		fieldsByName[field.Name] = field
	}

	assert.Contains(t, fieldsByName, "Value", "Should have public field Value")
	assert.Contains(t, fieldsByName, "secret", "Should have private field secret")

	// Reflect package parses all methods
	require.Equal(t, 2, len(containerType.Methods), "Should have both public and private methods")

	methodsByName := make(map[string]*Method)
	for _, method := range containerType.Methods {
		methodsByName[method.Name] = method
	}

	assert.Contains(t, methodsByName, "Public", "Should have public method Public")
	assert.Contains(t, methodsByName, "private", "Should have private method private")
}

// TestFieldTypes verifies that field types are correctly resolved
func TestFieldTypes(t *testing.T) {
	// Create a temporary test package
	tmpDir := t.TempDir()

	// Write a test file with various field types
	testCode := `package testpkg

type Point struct {
	X int
	Y int
}

type Shape struct {
	// Primitive types
	Name string
	
	// Pointer to custom type
	Location *Point
	
	// Slice type
	Vertices []Point
	
	// Map type
	Properties map[string]interface{}
}
`

	testFile := filepath.Join(tmpDir, "shape.go")
	err := os.WriteFile(testFile, []byte(testCode), 0644)
	require.NoError(t, err)

	// Parse the package
	pkg, _, err := ParsePackageWithOptions(tmpDir, DefaultParsingOptions())
	require.NoError(t, err)

	shapeType, exists := pkg.Types["Shape"]
	require.True(t, exists, "Shape type should be in package")

	fieldsByName := make(map[string]*Field)
	for _, field := range shapeType.Fields {
		fieldsByName[field.Name] = field
	}

	// Check Name field (string)
	assert.Equal(t, "string", fieldsByName["Name"].Type.Name)

	// Check Location field (pointer)
	assert.Equal(t, "*Point", fieldsByName["Location"].Type.Name)

	// Check Vertices field (slice)
	assert.Equal(t, "[]Point", fieldsByName["Vertices"].Type.Name)

	// Check Properties field (map)
	assert.Equal(t, "map[string]interface{}", fieldsByName["Properties"].Type.Name)
}

// TestEmbeddedFields verifies that embedded (anonymous) fields are handled
func TestEmbeddedFields(t *testing.T) {
	// Create a temporary test package
	tmpDir := t.TempDir()

	// Write a test file with embedded fields
	testCode := `package testpkg

type Base struct {
	Id int
}

type Extended struct {
	Base
	Name string
}
`

	testFile := filepath.Join(tmpDir, "embedded.go")
	err := os.WriteFile(testFile, []byte(testCode), 0644)
	require.NoError(t, err)

	// Parse the package
	pkg, _, err := ParsePackageWithOptions(tmpDir, DefaultParsingOptions())
	require.NoError(t, err)

	extType, exists := pkg.Types["Extended"]
	require.True(t, exists, "Extended type should be in package")
	require.Equal(t, 2, len(extType.Fields), "Extended should have 2 fields (embedded Base and Name)")

	// First field should be embedded Base
	assert.Equal(t, "Base", extType.Fields[0].Name)
	assert.True(t, extType.Fields[0].IsEmbedded, "Base should be marked as embedded")

	// Second field should be Name
	assert.Equal(t, "Name", extType.Fields[1].Name)
	assert.False(t, extType.Fields[1].IsEmbedded, "Name should not be embedded")
}

// TestMethodReceiver verifies that method receivers are correctly identified
func TestMethodReceiver(t *testing.T) {
	// Create a temporary test package
	tmpDir := t.TempDir()

	// Write a test file with value and pointer receivers
	testCode := `package testpkg

type Vector struct {
	x, y float64
}

// Value receiver
func (v Vector) Magnitude() float64 {
	return 0
}

// Pointer receiver
func (v *Vector) Translate(dx, dy float64) {
	v.x += dx
	v.y += dy
}
`

	testFile := filepath.Join(tmpDir, "vector.go")
	err := os.WriteFile(testFile, []byte(testCode), 0644)
	require.NoError(t, err)

	// Parse the package
	pkg, _, err := ParsePackageWithOptions(tmpDir, DefaultParsingOptions())
	require.NoError(t, err)

	vectorType, exists := pkg.Types["Vector"]
	require.True(t, exists, "Vector type should be in package")
	require.Equal(t, 2, len(vectorType.Methods), "Vector should have 2 methods")

	methodsByName := make(map[string]*Method)
	for _, method := range vectorType.Methods {
		methodsByName[method.Name] = method
	}

	// Both methods should have Vector as receiver (pointer indicator removed)
	assert.NotNil(t, methodsByName["Magnitude"].Receiver)
	assert.Equal(t, "Vector", methodsByName["Magnitude"].Receiver.Name)

	assert.NotNil(t, methodsByName["Translate"].Receiver)
	assert.Equal(t, "Vector", methodsByName["Translate"].Receiver.Name)
}

// isExported is a helper function to check if a name is exported
func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}
