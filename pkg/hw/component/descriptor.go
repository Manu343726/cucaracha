package component

import (
	"fmt"
	"sync"
)

// =============================================================================
// Port Descriptor
// =============================================================================

// PortDescriptor describes a port's properties
type PortDescriptor struct {
	// Name is the port identifier
	Name string

	// Description explains the port's purpose
	Description string

	// Width is the number of bits
	Width int

	// Direction indicates input/output/bidirectional
	Direction Direction

	// Optional indicates if the port can be left unconnected
	Optional bool
}

// =============================================================================
// Component Descriptor
// =============================================================================

// ComponentDescriptor provides metadata about a component type
type ComponentDescriptor struct {
	// Name is the unique identifier for this component type
	Name string

	// DisplayName is a human-readable name
	DisplayName string

	// Description explains what the component does
	Description string

	// Category groups related components (e.g., "logic", "arithmetic", "memory")
	Category string

	// Version is the component version
	Version string

	// Inputs describes the input ports
	Inputs []PortDescriptor

	// Outputs describes the output ports
	Outputs []PortDescriptor

	// Parameters describes configurable parameters
	Parameters []ParameterDescriptor

	// Factory creates an instance of this component
	Factory ComponentFactory
}

// ParameterDescriptor describes a configurable parameter
type ParameterDescriptor struct {
	// Name is the parameter identifier
	Name string

	// Description explains the parameter
	Description string

	// Type is the parameter type ("int", "bool", "string", "float")
	Type string

	// Default is the default value
	Default interface{}

	// Required indicates if the parameter must be provided
	Required bool

	// Constraints holds validation constraints
	Constraints ParameterConstraints
}

// ParameterConstraints defines validation rules for a parameter
type ParameterConstraints struct {
	// Min is the minimum value (for numeric types)
	Min *float64

	// Max is the maximum value (for numeric types)
	Max *float64

	// Enum lists allowed values
	Enum []interface{}
}

// ComponentFactory is a function that creates a component instance
type ComponentFactory func(name string, params map[string]interface{}) (Component, error)

// =============================================================================
// Component Registry
// =============================================================================

// Registry stores component descriptors and allows querying/instantiation
type Registry struct {
	descriptors map[string]*ComponentDescriptor
	categories  map[string][]string // category -> component names
	mu          sync.RWMutex
}

// NewRegistry creates a new component registry
func NewRegistry() *Registry {
	return &Registry{
		descriptors: make(map[string]*ComponentDescriptor),
		categories:  make(map[string][]string),
	}
}

// Register adds a component descriptor to the registry
func (r *Registry) Register(desc *ComponentDescriptor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if desc.Name == "" {
		return fmt.Errorf("component descriptor must have a name")
	}

	if _, exists := r.descriptors[desc.Name]; exists {
		return fmt.Errorf("component %q already registered", desc.Name)
	}

	r.descriptors[desc.Name] = desc

	// Index by category
	if desc.Category != "" {
		r.categories[desc.Category] = append(r.categories[desc.Category], desc.Name)
	}

	return nil
}

// Unregister removes a component descriptor from the registry
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	desc, exists := r.descriptors[name]
	if !exists {
		return fmt.Errorf("component %q not found", name)
	}

	delete(r.descriptors, name)

	// Remove from category index
	if desc.Category != "" {
		names := r.categories[desc.Category]
		for i, n := range names {
			if n == name {
				r.categories[desc.Category] = append(names[:i], names[i+1:]...)
				break
			}
		}
	}

	return nil
}

// Get returns a component descriptor by name
func (r *Registry) Get(name string) (*ComponentDescriptor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	desc, exists := r.descriptors[name]
	if !exists {
		return nil, fmt.Errorf("component %q not found", name)
	}

	return desc, nil
}

// List returns all registered component names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.descriptors))
	for name := range r.descriptors {
		names = append(names, name)
	}
	return names
}

// ListByCategory returns component names in a category
func (r *Registry) ListByCategory(category string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := r.categories[category]
	result := make([]string, len(names))
	copy(result, names)
	return result
}

// Categories returns all registered categories
func (r *Registry) Categories() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cats := make([]string, 0, len(r.categories))
	for cat := range r.categories {
		cats = append(cats, cat)
	}
	return cats
}

// Create instantiates a component by name
func (r *Registry) Create(componentName, instanceName string, params map[string]interface{}) (Component, error) {
	r.mu.RLock()
	desc, exists := r.descriptors[componentName]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("component %q not found", componentName)
	}

	if desc.Factory == nil {
		return nil, fmt.Errorf("component %q has no factory", componentName)
	}

	// Validate and apply defaults
	resolvedParams, err := r.resolveParams(desc, params)
	if err != nil {
		return nil, fmt.Errorf("parameter error for %q: %w", componentName, err)
	}

	return desc.Factory(instanceName, resolvedParams)
}

// resolveParams validates parameters and applies defaults
func (r *Registry) resolveParams(desc *ComponentDescriptor, params map[string]interface{}) (map[string]interface{}, error) {
	resolved := make(map[string]interface{})

	// Copy provided params
	for k, v := range params {
		resolved[k] = v
	}

	// Check required params and apply defaults
	for _, p := range desc.Parameters {
		if _, provided := resolved[p.Name]; !provided {
			if p.Required {
				return nil, fmt.Errorf("required parameter %q not provided", p.Name)
			}
			if p.Default != nil {
				resolved[p.Name] = p.Default
			}
		}
	}

	return resolved, nil
}

// Search finds components matching a query
func (r *Registry) Search(query SearchQuery) []*ComponentDescriptor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*ComponentDescriptor

	for _, desc := range r.descriptors {
		if query.Matches(desc) {
			results = append(results, desc)
		}
	}

	return results
}

// SearchQuery defines search criteria for components
type SearchQuery struct {
	// NameContains matches components whose name contains this string
	NameContains string

	// Category filters by category
	Category string

	// MinInputs filters by minimum number of inputs
	MinInputs int

	// MaxInputs filters by maximum number of inputs (0 = no limit)
	MaxInputs int

	// MinOutputs filters by minimum number of outputs
	MinOutputs int

	// MaxOutputs filters by maximum number of outputs (0 = no limit)
	MaxOutputs int
}

// Matches checks if a descriptor matches the query
func (q SearchQuery) Matches(desc *ComponentDescriptor) bool {
	if q.NameContains != "" {
		if !containsIgnoreCase(desc.Name, q.NameContains) &&
			!containsIgnoreCase(desc.DisplayName, q.NameContains) &&
			!containsIgnoreCase(desc.Description, q.NameContains) {
			return false
		}
	}

	if q.Category != "" && desc.Category != q.Category {
		return false
	}

	if q.MinInputs > 0 && len(desc.Inputs) < q.MinInputs {
		return false
	}

	if q.MaxInputs > 0 && len(desc.Inputs) > q.MaxInputs {
		return false
	}

	if q.MinOutputs > 0 && len(desc.Outputs) < q.MinOutputs {
		return false
	}

	if q.MaxOutputs > 0 && len(desc.Outputs) > q.MaxOutputs {
		return false
	}

	return true
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (substr == "" ||
		findIgnoreCase(s, substr) >= 0)
}

func findIgnoreCase(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			sc := s[i+j]
			pc := substr[j]
			// Convert to lowercase for comparison
			if sc >= 'A' && sc <= 'Z' {
				sc += 'a' - 'A'
			}
			if pc >= 'A' && pc <= 'Z' {
				pc += 'a' - 'A'
			}
			if sc != pc {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// =============================================================================
// Descriptor Builders
// =============================================================================

// DescriptorBuilder provides a fluent API for building component descriptors
type DescriptorBuilder struct {
	desc *ComponentDescriptor
}

// NewDescriptor creates a new descriptor builder
func NewDescriptor(name string) *DescriptorBuilder {
	return &DescriptorBuilder{
		desc: &ComponentDescriptor{
			Name:       name,
			Inputs:     make([]PortDescriptor, 0),
			Outputs:    make([]PortDescriptor, 0),
			Parameters: make([]ParameterDescriptor, 0),
		},
	}
}

// DisplayName sets the display name
func (b *DescriptorBuilder) DisplayName(name string) *DescriptorBuilder {
	b.desc.DisplayName = name
	return b
}

// Description sets the description
func (b *DescriptorBuilder) Description(desc string) *DescriptorBuilder {
	b.desc.Description = desc
	return b
}

// Category sets the category
func (b *DescriptorBuilder) Category(cat string) *DescriptorBuilder {
	b.desc.Category = cat
	return b
}

// Version sets the version
func (b *DescriptorBuilder) Version(ver string) *DescriptorBuilder {
	b.desc.Version = ver
	return b
}

// Input adds an input port descriptor
func (b *DescriptorBuilder) Input(name string, width int, description string) *DescriptorBuilder {
	b.desc.Inputs = append(b.desc.Inputs, PortDescriptor{
		Name:        name,
		Width:       width,
		Direction:   Input,
		Description: description,
	})
	return b
}

// OptionalInput adds an optional input port descriptor
func (b *DescriptorBuilder) OptionalInput(name string, width int, description string) *DescriptorBuilder {
	b.desc.Inputs = append(b.desc.Inputs, PortDescriptor{
		Name:        name,
		Width:       width,
		Direction:   Input,
		Description: description,
		Optional:    true,
	})
	return b
}

// Output adds an output port descriptor
func (b *DescriptorBuilder) Output(name string, width int, description string) *DescriptorBuilder {
	b.desc.Outputs = append(b.desc.Outputs, PortDescriptor{
		Name:        name,
		Width:       width,
		Direction:   Output,
		Description: description,
	})
	return b
}

// Param adds a parameter descriptor
func (b *DescriptorBuilder) Param(name, typ string, defaultVal interface{}, description string) *DescriptorBuilder {
	b.desc.Parameters = append(b.desc.Parameters, ParameterDescriptor{
		Name:        name,
		Type:        typ,
		Default:     defaultVal,
		Description: description,
	})
	return b
}

// RequiredParam adds a required parameter descriptor
func (b *DescriptorBuilder) RequiredParam(name, typ, description string) *DescriptorBuilder {
	b.desc.Parameters = append(b.desc.Parameters, ParameterDescriptor{
		Name:        name,
		Type:        typ,
		Required:    true,
		Description: description,
	})
	return b
}

// Factory sets the component factory
func (b *DescriptorBuilder) Factory(f ComponentFactory) *DescriptorBuilder {
	b.desc.Factory = f
	return b
}

// Build returns the completed descriptor
func (b *DescriptorBuilder) Build() *ComponentDescriptor {
	return b.desc
}
