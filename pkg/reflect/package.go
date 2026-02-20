package reflect

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ParsePackage parses all Go files in a package directory and extracts metadata
func ParsePackage(packagePath string) (*Package, error) {
	fset := token.NewFileSet()

	// Read all Go files in the directory
	entries, err := os.ReadDir(packagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package directory %s: %w", packagePath, err)
	}

	var goFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") &&
			!strings.HasSuffix(entry.Name(), "_test.go") {
			goFiles = append(goFiles, filepath.Join(packagePath, entry.Name()))
		}
	}

	if len(goFiles) == 0 {
		return nil, fmt.Errorf("no Go files found in package directory %s", packagePath)
	}

	// Parse all files
	pkgs, err := parser.ParseDir(fset, packagePath, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse package directory %s: %w", packagePath, err)
	}

	// Should have exactly one package
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found in %s", packagePath)
	}

	var pkgName string
	var astPkg *ast.Package
	for name, pkg := range pkgs {
		// Skip test packages
		if strings.HasSuffix(name, "_test") {
			continue
		}
		pkgName = name
		astPkg = pkg
		break
	}

	if astPkg == nil {
		return nil, fmt.Errorf("failed to find package in %s", packagePath)
	}

	// Create Package structure
	pkg := &Package{
		Name:      pkgName,
		Path:      packagePath,
		Files:     []*File{},
		Types:     make(map[string]*Type),
		Functions: []*Function{},
		Constants: []*Constant{},
		Enums:     make(map[string]*Enum),
		FileSet:   fset,
	}

	// Parse each file
	for filename, astFile := range astPkg.Files {
		file := &File{
			Name:      filepath.Base(filename),
			Path:      filename,
			Package:   pkgName,
			Types:     make(map[string]*Type),
			Functions: []*Function{},
			Constants: []*Constant{},
		}

		// Parse imports
		for _, importSpec := range astFile.Imports {
			importPath := strings.Trim(importSpec.Path.Value, "\"")
			importName := filepath.Base(importPath)
			if importSpec.Name != nil {
				importName = importSpec.Name.Name
			}
			alias := ""
			if importSpec.Name != nil && importSpec.Name.Name != filepath.Base(importPath) {
				alias = importSpec.Name.Name
			}

			file.Imports = append(file.Imports, Import{
				Name:  importName,
				Path:  importPath,
				Alias: alias,
			})
		}

		// Parse top-level declarations
		for _, decl := range astFile.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				parseGenDeclWithFileContext(file, pkg, decl)
			case *ast.FuncDecl:
				if fn := parseFuncDecl(decl, pkgName); fn != nil {
					file.Functions = append(file.Functions, fn)
					pkg.Functions = append(pkg.Functions, fn)
				}
			}
		}

		pkg.Files = append(pkg.Files, file)
	}

	// Merge types from all files
	for _, file := range pkg.Files {
		for name, typ := range file.Types {
			pkg.Types[name] = typ
		}
	}

	// Post-process to identify enums
	identifyEnums(pkg)

	return pkg, nil
}

// parseGenDeclWithFileContext handles const, type, and import declarations with package context
func parseGenDeclWithFileContext(file *File, pkg *Package, decl *ast.GenDecl) {
	switch decl.Tok {
	case token.TYPE:
		for _, spec := range decl.Specs {
			typeSpec := spec.(*ast.TypeSpec)
			typ := parseTypeSpec(typeSpec, decl)
			file.Types[typeSpec.Name.Name] = typ
		}

	case token.CONST:
		for _, spec := range decl.Specs {
			valueSpec := spec.(*ast.ValueSpec)
			for i, name := range valueSpec.Names {
				const_ := &Constant{
					Name: name.Name,
					Doc:  decl.Doc.Text(),
				}

				// Try to extract value
				if len(valueSpec.Values) > i {
					if lit, ok := valueSpec.Values[i].(*ast.BasicLit); ok {
						const_.Value = lit.Value
					}
				}

				// Extract type
				if valueSpec.Type != nil {
					const_.Type = typeToString(valueSpec.Type)
				}

				file.Constants = append(file.Constants, const_)
				pkg.Constants = append(pkg.Constants, const_)
			}
		}

	case token.VAR:
		for _, spec := range decl.Specs {
			valueSpec := spec.(*ast.ValueSpec)
			for _, name := range valueSpec.Names {
				const_ := &Constant{
					Name: name.Name,
					Doc:  decl.Doc.Text(),
				}

				// Extract type
				if valueSpec.Type != nil {
					const_.Type = typeToString(valueSpec.Type)
				}

				file.Constants = append(file.Constants, const_)
				pkg.Constants = append(pkg.Constants, const_)
			}
		}
	}
}

// identifyEnums finds groups of constants that form logical enums based on naming patterns
// An enum is identified when:
// 1. Multiple constants have the same underlying type (typically int or string)
// 2. There's a corresponding type definition in the package
// 3. Optional: a String() method exists on the type
func identifyEnums(pkg *Package) {
	// Group constants by type
	typeToConstants := make(map[string][]*Constant)
	for _, const_ := range pkg.Constants {
		if const_.Type != "" {
			typeToConstants[const_.Type] = append(typeToConstants[const_.Type], const_)
		}
	}

	// Check which types in the package correspond to enum types
	for _, typ := range pkg.Types {
		// Look for a constant group with this type
		if constants, ok := typeToConstants[typ.Name]; ok && len(constants) > 1 {
			// This looks like an enum
			enum := &Enum{
				Name:      typ.Name,
				Type:      typ.Name,
				Doc:       typ.Doc,
				Values:    constants,
				SourcePos: typ.SourcePos,
			}

			// Check if there's a String() method on this type
			for _, method := range typ.Methods {
				if method.Name == "String" && len(method.Results) > 0 && method.Results[0].Type == "string" {
					enum.StringMethod = true
					break
				}
			}

			pkg.Enums[typ.Name] = enum
		}
	}
}

// FindType searches for a type by name across the package
func (p *Package) FindType(name string) *Type {
	if typ, ok := p.Types[name]; ok {
		return typ
	}
	return nil
}

// FindFunction searches for a function by name across the package
func (p *Package) FindFunction(name string) *Function {
	for _, fn := range p.Functions {
		if fn.Name == name {
			return fn
		}
	}
	return nil
}

// FindEnum searches for a enum by name across the package
func (p *Package) FindEnum(name string) *Enum {
	if enum, ok := p.Enums[name]; ok {
		return enum
	}
	return nil
}

// FindTypesByKind returns all types of a specific kind (struct, interface, alias, etc.)
func (p *Package) FindTypesByKind(kind TypeKind) []*Type {
	var result []*Type
	for _, typ := range p.Types {
		if typ.Kind == kind {
			result = append(result, typ)
		}
	}
	return result
}

// FindFunctionsByPrefix returns all functions whose name starts with the given prefix
func (p *Package) FindFunctionsByPrefix(prefix string) []*Function {
	var result []*Function
	for _, fn := range p.Functions {
		if strings.HasPrefix(fn.Name, prefix) {
			result = append(result, fn)
		}
	}
	return result
}

// GetMethodsForType returns all methods defined on a specific type
func (p *Package) GetMethodsForType(typeName string) []*Method {
	typ := p.FindType(typeName)
	if typ == nil {
		return nil
	}
	return typ.Methods
}

// GetStructFields returns the fields of a struct type by name
func (p *Package) GetStructFields(structName string) []*Field {
	typ := p.FindType(structName)
	if typ == nil || typ.Kind != TypeKindStruct {
		return nil
	}
	return typ.Fields
}

// FindStructsWithField finds all struct types that have a field of the given name
func (p *Package) FindStructsWithField(fieldName string) []*Type {
	var result []*Type
	for _, typ := range p.Types {
		if typ.Kind != TypeKindStruct {
			continue
		}
		for _, field := range typ.Fields {
			if field.Name == fieldName {
				result = append(result, typ)
				break
			}
		}
	}
	return result
}

// FindStructsWithFieldType finds all struct types that have a field of the given type
func (p *Package) FindStructsWithFieldType(fieldType string) []*Type {
	var result []*Type
	for _, typ := range p.Types {
		if typ.Kind != TypeKindStruct {
			continue
		}
		for _, field := range typ.Fields {
			if field.Type == fieldType {
				result = append(result, typ)
				break
			}
		}
	}
	return result
}

// GetMethodsByName finds all methods with a specific name across all types
func (p *Package) GetMethodsByName(methodName string) []*Method {
	var result []*Method
	for _, typ := range p.Types {
		for _, method := range typ.Methods {
			if method.Name == methodName {
				result = append(result, method)
			}
		}
	}
	return result
}

// GetInterfaceMethods returns the methods defined in an interface type
func (p *Package) GetInterfaceMethods(interfaceName string) []*Method {
	typ := p.FindType(interfaceName)
	if typ == nil || typ.Kind != TypeKindInterface {
		return nil
	}
	return typ.Methods
}

// FindConstantsByType returns all constants of a specific type
func (p *Package) FindConstantsByType(typeName string) []*Constant {
	var result []*Constant
	for _, const_ := range p.Constants {
		if const_.Type == typeName {
			result = append(result, const_)
		}
	}
	return result
}

// FindConstantByName returns a constant by name
func (p *Package) FindConstantByName(name string) *Constant {
	for _, const_ := range p.Constants {
		if const_.Name == name {
			return const_
		}
	}
	return nil
}

// GetPublicTypes returns all public (exported) types in the package
func (p *Package) GetPublicTypes() []*Type {
	var result []*Type
	for name, typ := range p.Types {
		// In Go, exported names start with an uppercase letter
		if len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z' {
			result = append(result, typ)
		}
	}
	return result
}

// GetPublicFunctions returns all public (exported) functions in the package
func (p *Package) GetPublicFunctions() []*Function {
	var result []*Function
	for _, fn := range p.Functions {
		// In Go, exported names start with an uppercase letter
		if len(fn.Name) > 0 && fn.Name[0] >= 'A' && fn.Name[0] <= 'Z' {
			result = append(result, fn)
		}
	}
	return result
}

// GetFileByName returns a file by its name
func (p *Package) GetFileByName(fileName string) *File {
	for _, file := range p.Files {
		if file.Name == fileName {
			return file
		}
	}
	return nil
}
