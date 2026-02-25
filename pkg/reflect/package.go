package reflect

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// ParsePackage parses all Go files in a package directory and extracts metadata
// It uses default parsing options.
func ParsePackage(packagePath string) (*Package, error) {
	return ParsePackageWithOptions(packagePath, DefaultParsingOptions())
}

// ParsePackageWithOptions parses all Go files in a package directory with custom parsing options
func ParsePackageWithOptions(packagePath string, opts ParsingOptions) (*Package, error) {
	log().Info("Starting package parse", slog.String("path", packagePath))
	fset := token.NewFileSet()

	// Read all Go files in the directory
	entries, err := os.ReadDir(packagePath)
	if err != nil {
		return nil, log().Errorf("failed to read package directory %s: %w", packagePath, err)
	}

	var goFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") &&
			!strings.HasSuffix(entry.Name(), "_test.go") {
			goFiles = append(goFiles, filepath.Join(packagePath, entry.Name()))
		}
	}

	if len(goFiles) == 0 {
		return nil, log().Errorf("no Go files found in package directory %s", packagePath)
	}

	log().Debug("Found Go files", slog.Int("count", len(goFiles)), slog.String("path", packagePath))

	// Parse all files
	pkgs, err := parser.ParseDir(fset, packagePath, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, log().Errorf("failed to parse package directory %s: %w", packagePath, err)
	}

	// Should have exactly one package
	if len(pkgs) == 0 {
		return nil, log().Errorf("no packages found in %s", packagePath)
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
		return nil, log().Errorf("failed to find package in %s", packagePath)
	}

	log().Debug("Found package", slog.String("name", pkgName), slog.String("path", packagePath))

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
		log().Debug("Parsing file", slog.String("file", filepath.Base(filename)))
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
					log().Debug("Found function", slog.String("name", fn.Name), slog.String("file", file.Name))
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
			log().Debug("Found type", slog.String("name", name), slog.String("kind", string(typ.Kind)))
			pkg.Types[name] = typ
		}
	}

	// Post-process to identify enums
	log().Debug("Identifying enums in package")
	identifyEnums(pkg)
	log().Debug("Enum identification complete", slog.Int("count", len(pkg.Enums)))

	// Second pass: resolve all type references
	log().Debug("Resolving type references")
	resolveTypeReferencesWithOptions(pkg, opts, 0)

	log().Info("Package parse complete", slog.String("name", pkgName), slog.Int("types", len(pkg.Types)), slog.Int("functions", len(pkg.Functions)), slog.Int("enums", len(pkg.Enums)))
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
					Name:  name.Name,
					Doc:   decl.Doc.Text(),
					Value: &Value{},
				}

				// Try to extract value
				if len(valueSpec.Values) > i {
					if lit, ok := valueSpec.Values[i].(*ast.BasicLit); ok {
						const_.Value.Value = lit.Value
					}
				}

				// Extract type
				if valueSpec.Type != nil {
					typeStr := typeToString(valueSpec.Type)
					const_.Value.Type = &TypeReference{
						Name: typeStr,
						Type: nil,
					}
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
					typeStr := typeToString(valueSpec.Type)
					const_.Value = &Value{
						Type: &TypeReference{
							Name: typeStr,
							Type: nil,
						},
					}
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
		if const_.Value.Type != nil {
			typeToConstants[const_.Value.Type.Name] = append(typeToConstants[const_.Value.Type.Name], const_)
		}
	}

	// Check which types in the package correspond to enum types
	for _, typ := range pkg.Types {
		log().Debug("Checking if type is enum", slog.String("typeName", typ.Name), slog.String("typeKind", string(typ.Kind)))
		// Look for a constant group with this type
		if constants, ok := typeToConstants[typ.Name]; ok && len(constants) > 1 {
			log().Info("Identified enum", slog.String("name", typ.Name), slog.String("kind", string(typ.Kind)), slog.Int("values", len(constants)))
			for i := range constants {
				constants[i].Value.Type = MakeTypeReference(typ)
			}

			// This looks like an enum
			enum := &Enum{
				Type:      MakeTypeReference(typ),
				Values:    constants,
				SourcePos: typ.SourcePos,
			}

			// Check if there's a String() method on this type
			for _, method := range typ.Methods {
				if method.Name == "String" && len(method.Results) > 0 && method.Results[0].Type.Name == "string" {
					enum.StringMethod = true
					break
				}
			}

			pkg.Enums[typ.Name] = enum
		}
	}
}

// resolveTypeReferences performs a second pass to resolve all TypeReference pointers in the package
// This links each TypeReference to its corresponding Type definition if it exists in the package
func resolveTypeReferences(pkg *Package) {
	// Resolve types in all function signatures
	for _, fn := range pkg.Functions {
		for _, param := range fn.Args {
			if param.Type != nil {
				param.Type.Type = resolveType(param.Type.Name, pkg)
			}
		}
		for _, result := range fn.Results {
			if result.Type != nil {
				result.Type.Type = resolveType(result.Type.Name, pkg)
			}
		}
	}

	// Resolve types in all type definitions
	for _, typ := range pkg.Types {
		// Resolve field types
		if typ.Kind == TypeKindStruct {
			for _, field := range typ.Fields {
				if field.Type != nil {
					field.Type.Type = resolveType(field.Type.Name, pkg)
				}
			}
		}

		// Resolve method parameter and result types
		for _, method := range typ.Methods {
			for _, param := range method.Args {
				if param.Type != nil {
					param.Type.Type = resolveType(param.Type.Name, pkg)
				}
			}
			for _, result := range method.Results {
				if result.Type != nil {
					result.Type.Type = resolveType(result.Type.Name, pkg)
				}
			}
		}
	}

	// Resolve types in all constants
	for _, const_ := range pkg.Constants {
		if const_.Value.Type != nil {
			const_.Value.Type.Type = resolveType(const_.Value.Type.Name, pkg)
		}
	}
}

// resolveType extracts the base type name from a potentially complex type string
// (e.g., "*MyType" -> "MyType", "[]MyType" -> "MyType") and looks it up in the package
// or in the global basic types
func resolveType(typeName string, pkg *Package) *Type {
	if typeName == "" {
		return nil
	}

	// Extract base type name by removing pointers, slices, maps, and channels
	baseName := extractBaseTypeName(typeName)

	// First check if it's a basic type
	if basicType := GetBasicType(baseName); basicType != nil {
		return basicType
	}

	// Look up the type in the package
	if typ, ok := pkg.Types[baseName]; ok {
		return typ
	}

	// Type not found in package or basic types (could be external)
	return nil
}

// extractBaseTypeName removes type modifiers (pointers, slices, maps, channels) to get the base type name
func extractBaseTypeName(typeName string) string {
	// Remove leading channel notation first
	if strings.HasPrefix(typeName, "<-chan ") {
		typeName = strings.TrimPrefix(typeName, "<-chan ")
	}
	if strings.HasPrefix(typeName, "chan ") {
		typeName = strings.TrimPrefix(typeName, "chan ")
	}

	// Remove leading pointer symbols
	for strings.HasPrefix(typeName, "*") {
		typeName = strings.TrimPrefix(typeName, "*")
	}

	// Remove leading slice notation
	for strings.HasPrefix(typeName, "[]") {
		typeName = strings.TrimPrefix(typeName, "[]")
	}

	// Remove map notation (map[keyType]valType -> extract valType)
	if strings.HasPrefix(typeName, "map[") {
		// Find the closing bracket of the key type
		bracketCount := 0
		for i, ch := range typeName {
			if ch == '[' {
				bracketCount++
			} else if ch == ']' {
				bracketCount--
				if bracketCount == 0 {
					typeName = typeName[i+1:]
					break
				}
			}
		}
	}

	// Remove any remaining leading pointer symbols (after processing collections)
	for strings.HasPrefix(typeName, "*") {
		typeName = strings.TrimPrefix(typeName, "*")
	}

	// Handle qualified names (e.g., "package.Type" -> "Type")
	// Only extract the type name after the last dot if it's a selector
	if idx := strings.LastIndex(typeName, "."); idx >= 0 {
		typeName = typeName[idx+1:]
	}

	return strings.TrimSpace(typeName)
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
			if field.Type != nil && field.Type.Name == fieldType {
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
		if const_.Value.Type != nil && const_.Value.Type.Name == typeName {
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

// resolveTypeReferencesWithOptions resolves type references with support for external type resolution
// It respects the MaxResolutionDepth limit to prevent infinite recursion
func resolveTypeReferencesWithOptions(pkg *Package, opts ParsingOptions, depth int) {
	// Resolve types in all function signatures
	for _, fn := range pkg.Functions {
		for _, param := range fn.Args {
			if param.Type != nil {
				param.Type.Type = resolveTypeWithOptions(param.Type.Name, pkg, opts, depth)
			}
		}
		for _, result := range fn.Results {
			if result.Type != nil {
				result.Type.Type = resolveTypeWithOptions(result.Type.Name, pkg, opts, depth)
			}
		}
	}

	// Resolve types in all type definitions
	for _, typ := range pkg.Types {
		// Resolve field types
		if typ.Kind == TypeKindStruct {
			for _, field := range typ.Fields {
				if field.Type != nil {
					field.Type.Type = resolveTypeWithOptions(field.Type.Name, pkg, opts, depth)
				}
			}
		}

		// Resolve method parameter and result types
		for _, method := range typ.Methods {
			for _, param := range method.Args {
				if param.Type != nil {
					param.Type.Type = resolveTypeWithOptions(param.Type.Name, pkg, opts, depth)
				}
			}
			for _, result := range method.Results {
				if result.Type != nil {
					result.Type.Type = resolveTypeWithOptions(result.Type.Name, pkg, opts, depth)
				}
			}
		}

		// Resolve types in function/method return types
		if typ.Kind == TypeKindFunction {
			for _, arg := range typ.Args {
				if arg.Type != nil {
					arg.Type = resolveTypeWithOptions(arg.Name, pkg, opts, depth)
				}
			}
			for _, result := range typ.Results {
				if result.Type != nil {
					result.Type = resolveTypeWithOptions(result.Name, pkg, opts, depth)
				}
			}
		}
	}

	// Resolve types in all constants
	for _, const_ := range pkg.Constants {
		if const_.Value.Type != nil {
			const_.Value.Type.Type = resolveTypeWithOptions(const_.Value.Type.Name, pkg, opts, depth)
		}
	}
}

// resolveTypeWithOptions resolves a type name, optionally including external packages
// based on the provided options and current recursion depth
func resolveTypeWithOptions(typeName string, pkg *Package, opts ParsingOptions, depth int) *Type {
	if typeName == "" {
		return nil
	}

	// Extract base type name by removing pointers, slices, maps, and channels
	baseName := extractBaseTypeName(typeName)

	// First check if it's a basic type
	if basicType := GetBasicType(baseName); basicType != nil {
		return basicType
	}

	// Look up the type in the package
	if typ, ok := pkg.Types[baseName]; ok {
		return typ
	}

	// If external resolution is not enabled or we've hit the recursion limit, return nil
	if !opts.ResolveExternalTypes {
		return nil
	}

	// Check if we've exceeded the maximum recursion depth
	if opts.MaxResolutionDepth >= 0 && depth >= opts.MaxResolutionDepth {
		return nil
	}

	// Try to resolve from external packages
	return resolveExternalType(baseName, pkg, opts, depth+1)
}

// resolveExternalType attempts to resolve a type from external packages
func resolveExternalType(typeName string, pkg *Package, opts ParsingOptions, depth int) *Type {
	// If it contains a dot, it's a qualified name (e.g., "io.Reader")
	if idx := strings.Index(typeName, "."); idx >= 0 {
		importPath := typeName[:idx]
		unqualifiedName := typeName[idx+1:]

		// Try to resolve the import path
		extPkg, err := resolveImportToPackage(importPath, pkg)
		if err != nil || extPkg == nil {
			return nil
		}

		// Recursively resolve the unqualified type in the external package
		resolveTypeReferencesWithOptions(extPkg, opts, depth)

		if typ, ok := extPkg.Types[unqualifiedName]; ok {
			return typ
		}
	}

	return nil
}

// resolveImportToPackage attempts to resolve an import alias/name to the actual package
func resolveImportToPackage(importName string, pkg *Package) (*Package, error) {
	// Find the import in the current package
	var importPath string
	for _, file := range pkg.Files {
		for _, imp := range file.Imports {
			if imp.Name == importName || imp.Alias == importName {
				importPath = imp.Path
				break
			}
		}
		if importPath != "" {
			break
		}
	}

	if importPath == "" {
		// Try to use the name directly as a path (for standard library)
		importPath = importName
	}

	// Try to parse the external package from the import path
	return ParsePackageFromImportInDir(importPath, pkg.Path)
}
