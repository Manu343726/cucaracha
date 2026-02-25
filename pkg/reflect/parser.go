package reflect

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"path/filepath"
	"strings"
)

// ParseFile parses a single Go file and extracts all types, functions, and constants
// It uses default parsing options.
func ParseFile(filePath string) (*File, error) {
	return ParseFileWithOptions(filePath, DefaultParsingOptions())
}

// ParseFileWithOptions parses a single Go file with custom parsing options
func ParseFileWithOptions(filePath string, opts ParsingOptions) (*File, error) {
	log().Debug("Parsing file", slog.String("path", filePath))
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, log().Errorf("failed to parse file %s: %w", filePath, err)
	}

	file := &File{
		Name:      filepath.Base(filePath),
		Path:      filePath,
		Package:   astFile.Name.Name,
		Types:     make(map[string]*Type),
		Functions: []*Function{},
		Constants: []*Constant{},
	}

	log().Debug("File package identified", slog.String("file", file.Name), slog.String("package", file.Package))

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

	log().Debug("Imports parsed", slog.String("file", file.Name), slog.Int("count", len(file.Imports)))

	// Parse top-level declarations
	for _, decl := range astFile.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			parseGenDecl(file, decl)
		case *ast.FuncDecl:
			if fn := parseFuncDecl(decl, file.Package); fn != nil {
				log().Debug("Function found", slog.String("file", file.Name), slog.String("name", fn.Name))
				file.Functions = append(file.Functions, fn)
			}
		}
	}

	log().Debug("Declarations parsed", slog.String("file", file.Name), slog.Int("types", len(file.Types)), slog.Int("functions", len(file.Functions)), slog.Int("constants", len(file.Constants)))

	// Second pass: resolve type references within the file
	resolveFileTypeReferences(file)

	log().Info("File parse complete", slog.String("path", filepath.Base(filePath)))
	return file, nil
}

// parseGenDecl handles const, type, and import declarations
func parseGenDecl(file *File, decl *ast.GenDecl) {
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
			}
		}

	case token.VAR:
		for _, spec := range decl.Specs {
			valueSpec := spec.(*ast.ValueSpec)
			for _, name := range valueSpec.Names {
				const_ := &Constant{
					Name:  name.Name,
					Doc:   decl.Doc.Text(),
					Value: &Value{},
				}

				// Try to extract value
				if len(valueSpec.Values) > 0 {
					if lit, ok := valueSpec.Values[0].(*ast.BasicLit); ok {
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
			}
		}
	}
}

// parseTypeSpec extracts information about a type declaration
func parseTypeSpec(typeSpec *ast.TypeSpec, decl *ast.GenDecl) *Type {
	typ := &Type{
		Name:       typeSpec.Name.Name,
		Doc:        decl.Doc.Text(),
		Underlying: typeSpec.Type,
		SourcePos:  typeSpec.Pos(),
	}

	// Determine type kind
	switch typeSpec.Type.(type) {
	case *ast.StructType:
		typ.Kind = TypeKindStruct
		if structType, ok := typeSpec.Type.(*ast.StructType); ok {
			typ.Fields = parseStructFields(structType)
		}

	case *ast.InterfaceType:
		typ.Kind = TypeKindInterface
		if iface, ok := typeSpec.Type.(*ast.InterfaceType); ok {
			typ.Methods = parseInterfaceMethods(iface)
		}

	default:
		// Distinguish between alias (with =) and typedef (without =)
		if typeSpec.Assign != 0 { // Assign != 0 means there's an = sign
			typ.Kind = TypeKindAlias
			log().Debug("Type is an alias", slog.String("name", typ.Name))
		} else {
			typ.Kind = TypeKindTypedef
			log().Debug("Type is a typedef", slog.String("name", typ.Name))
		}
	}

	return typ
}

// parseStructFields extracts fields from a struct type
func parseStructFields(structType *ast.StructType) []*Field {
	var fields []*Field

	if structType.Fields == nil {
		return fields
	}

	for _, field := range structType.Fields.List {
		typeStr := typeToString(field.Type)
		fieldType := astExprToType(field.Type)
		tag := ""
		if field.Tag != nil {
			tag = strings.Trim(field.Tag.Value, "`")
		}

		// Handle embedded fields (no field name)
		if len(field.Names) == 0 {
			fields = append(fields, &Field{
				Name: typeStr,
				Type: &TypeReference{
					Name: typeStr,
					Type: fieldType,
				},
				Tag:        tag,
				Doc:        field.Doc.Text(),
				IsEmbedded: true,
			})
		} else {
			for _, name := range field.Names {
				fields = append(fields, &Field{
					Name: name.Name,
					Type: &TypeReference{
						Name: typeStr,
						Type: fieldType,
					},
					Tag: tag,
					Doc: field.Doc.Text(),
				})
			}
		}
	}

	return fields
}

// parseInterfaceMethods extracts methods from an interface type
func parseInterfaceMethods(iface *ast.InterfaceType) []*Method {
	var methods []*Method

	if iface.Methods == nil {
		return methods
	}

	for _, methodField := range iface.Methods.List {
		if len(methodField.Names) == 0 {
			// Embedded interface - skip for now
			continue
		}

		if funcType, ok := methodField.Type.(*ast.FuncType); ok {
			method := &Method{
				Name:      methodField.Names[0].Name,
				Doc:       methodField.Doc.Text(),
				Signature: typeToString(methodField.Type),
				SourcePos: methodField.Pos(),
			}

			// Parse parameters and results
			method.Args = parseParameters(funcType.Params)
			method.Results = parseParameters(funcType.Results)

			methods = append(methods, method)
		}
	}

	return methods
}

// parseFuncDecl extracts information from a function declaration
func parseFuncDecl(decl *ast.FuncDecl, pkgName string) *Function {
	fn := &Function{
		Name:      decl.Name.Name,
		Package:   pkgName,
		Doc:       decl.Doc.Text(),
		Signature: typeToString(decl.Type),
		SourcePos: decl.Pos(),
	}

	// Check if it's a method (has receiver)
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		// Skip methods when parsing functions - they should be handled separately
		return nil
	}

	// Parse parameters and results
	if decl.Type.Params != nil {
		fn.Args = parseParameters(decl.Type.Params)
	}
	if decl.Type.Results != nil {
		fn.Results = parseParameters(decl.Type.Results)
	}

	return fn
}

// parseParameters extracts parameters from a parameter list, creating proper Type structures for composite types
func parseParameters(params *ast.FieldList) []*Parameter {
	var result []*Parameter

	if params == nil || params.List == nil {
		return result
	}

	for _, field := range params.List {
		typeStr := typeToString(field.Type)
		// Create a proper Type structure for composite types
		paramType := astExprToType(field.Type)

		if len(field.Names) == 0 {
			// Unnamed parameter (common in results)
			result = append(result, &Parameter{
				Name: "",
				Type: &TypeReference{
					Name: typeStr,
					Type: paramType,
				},
			})
		} else {
			for _, name := range field.Names {
				result = append(result, &Parameter{
					Name: name.Name,
					Type: &TypeReference{
						Name: typeStr,
						Type: paramType,
					},
				})
			}
		}
	}

	return result
}

// typeToString converts an AST expression to a string representation
func typeToString(expr ast.Expr) string {
	if expr == nil {
		return ""
	}

	switch v := expr.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.StarExpr:
		return "*" + typeToString(v.X)
	case *ast.SelectorExpr:
		return typeToString(v.X) + "." + v.Sel.Name
	case *ast.ArrayType:
		if v.Len == nil {
			return "[]" + typeToString(v.Elt)
		}
		// For fixed-size arrays, try to get the size
		if lit, ok := v.Len.(*ast.BasicLit); ok {
			return "[" + lit.Value + "]" + typeToString(v.Elt)
		}
		return "[]" + typeToString(v.Elt)
	case *ast.MapType:
		return "map[" + typeToString(v.Key) + "]" + typeToString(v.Value)
	case *ast.FuncType:
		params := ""
		if v.Params != nil {
			paramStrs := []string{}
			for _, field := range v.Params.List {
				paramStrs = append(paramStrs, typeToString(field.Type))
			}
			params = "(" + strings.Join(paramStrs, ", ") + ")"
		}

		results := ""
		if v.Results != nil {
			resultStrs := []string{}
			for _, field := range v.Results.List {
				resultStrs = append(resultStrs, typeToString(field.Type))
			}
			results = "(" + strings.Join(resultStrs, ", ") + ")"
		}

		return "func" + params + " " + results
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	case *ast.ChanType:
		arrow := "<-"
		if v.Dir == ast.SEND {
			arrow = ""
		}
		return arrow + "chan " + typeToString(v.Value)
	default:
		return ""
	}
}

// astExprToType converts an AST expression to a Type structure, properly handling composite types
func astExprToType(expr ast.Expr) *Type {
	if expr == nil {
		return nil
	}

	switch v := expr.(type) {
	case *ast.StarExpr:
		// Pointer type: *T
		innerType := astExprToType(v.X)
		if innerType == nil {
			// If inner couldn't be parsed as a type, create a basic pointer type
			innerType = &Type{
				Name: typeToString(v.X),
				Kind: TypeKindBasic,
			}
		}
		return &Type{
			Name: typeToString(expr),
			Kind: TypeKindPointer,
			Elem: &TypeReference{
				Name: innerType.Name,
				Type: innerType,
			},
		}

	case *ast.ArrayType:
		// Array or slice type: [N]T or []T
		if v.Len == nil {
			// Slice type: []T
			elemType := astExprToType(v.Elt)
			if elemType == nil {
				elemType = &Type{
					Name: typeToString(v.Elt),
					Kind: TypeKindBasic,
				}
			}
			return &Type{
				Name: typeToString(expr),
				Kind: TypeKindSlice,
				Elem: &TypeReference{
					Name: elemType.Name,
					Type: elemType,
				},
			}
		}

		// Array type: [N]T
		size := 0
		if lit, ok := v.Len.(*ast.BasicLit); ok {
			fmt.Sscanf(lit.Value, "%d", &size)
		}
		elemType := astExprToType(v.Elt)
		if elemType == nil {
			elemType = &Type{
				Name: typeToString(v.Elt),
				Kind: TypeKindBasic,
			}
		}
		return &Type{
			Name: typeToString(expr),
			Kind: TypeKindArray,
			Size: size,
			Elem: &TypeReference{
				Name: elemType.Name,
				Type: elemType,
			},
		}

	case *ast.MapType:
		// Map type: map[K]V
		keyType := astExprToType(v.Key)
		if keyType == nil {
			keyType = &Type{
				Name: typeToString(v.Key),
				Kind: TypeKindBasic,
			}
		}
		valueType := astExprToType(v.Value)
		if valueType == nil {
			valueType = &Type{
				Name: typeToString(v.Value),
				Kind: TypeKindBasic,
			}
		}
		return &Type{
			Name: typeToString(expr),
			Kind: TypeKindMap,
			Key: &TypeReference{
				Name: keyType.Name,
				Type: keyType,
			},
			Value: &TypeReference{
				Name: valueType.Name,
				Type: valueType,
			},
		}

	case *ast.ChanType:
		// Channel type: chan T, <-chan T, chan<- T
		elemType := astExprToType(v.Value)
		if elemType == nil {
			elemType = &Type{
				Name: typeToString(v.Value),
				Kind: TypeKindBasic,
			}
		}

		chanDir := ChanBidirectional
		if v.Dir == ast.RECV {
			chanDir = ChanRecv
		} else if v.Dir == ast.SEND {
			chanDir = ChanSend
		}

		return &Type{
			Name:    typeToString(expr),
			Kind:    TypeKindChannel,
			ChanDir: chanDir,
			Elem: &TypeReference{
				Name: elemType.Name,
				Type: elemType,
			},
		}

	case *ast.FuncType:
		// Function type: func(args...) (results...)
		args := []*TypeReference{}
		if v.Params != nil && v.Params.List != nil {
			for _, field := range v.Params.List {
				fieldType := astExprToType(field.Type)
				if fieldType == nil {
					fieldType = &Type{
						Name: typeToString(field.Type),
						Kind: TypeKindBasic,
					}
				}
				args = append(args, &TypeReference{
					Name: fieldType.Name,
					Type: fieldType,
				})
			}
		}

		results := []*TypeReference{}
		if v.Results != nil && v.Results.List != nil {
			for _, field := range v.Results.List {
				fieldType := astExprToType(field.Type)
				if fieldType == nil {
					fieldType = &Type{
						Name: typeToString(field.Type),
						Kind: TypeKindBasic,
					}
				}
				results = append(results, &TypeReference{
					Name: fieldType.Name,
					Type: fieldType,
				})
			}
		}

		return &Type{
			Name:    typeToString(expr),
			Kind:    TypeKindFunction,
			Args:    args,
			Results: results,
		}

	case *ast.Ident:
		// Basic or named type
		return &Type{
			Name: v.Name,
			Kind: TypeKindBasic, // Could be basic or named, but we don't know yet
		}

	case *ast.SelectorExpr:
		// Package-qualified type: pkg.Type
		return &Type{
			Name: typeToString(expr),
			Kind: TypeKindBasic,
		}

	case *ast.InterfaceType:
		return &Type{
			Name: "interface{}",
			Kind: TypeKindBasic,
		}

	case *ast.StructType:
		return &Type{
			Name: "struct{}",
			Kind: TypeKindBasic,
		}

	default:
		return nil
	}
}

// resolveFileTypeReferences resolves type references within a single file
// It links TypeReference pointers to their corresponding Type definitions within the file
func resolveFileTypeReferences(file *File) {
	// Resolve types in all function signatures
	for _, fn := range file.Functions {
		for _, param := range fn.Args {
			if param.Type != nil {
				param.Type.Type = resolveFileType(param.Type.Name, file)
			}
		}
		for _, result := range fn.Results {
			if result.Type != nil {
				result.Type.Type = resolveFileType(result.Type.Name, file)
			}
		}
	}

	// Resolve types in all type definitions
	for _, typ := range file.Types {
		// Resolve field types
		if typ.Kind == TypeKindStruct {
			for _, field := range typ.Fields {
				if field.Type != nil {
					field.Type.Type = resolveFileType(field.Type.Name, file)
				}
			}
		}

		// Resolve method parameter and result types
		for _, method := range typ.Methods {
			for _, param := range method.Args {
				if param.Type != nil {
					param.Type.Type = resolveFileType(param.Type.Name, file)
				}
			}
			for _, result := range method.Results {
				if result.Type != nil {
					result.Type.Type = resolveFileType(result.Type.Name, file)
				}
			}
		}
	}

	// Resolve types in all constants
	for _, const_ := range file.Constants {
		if const_.Value.Type != nil {
			const_.Value.Type.Type = resolveFileType(const_.Value.Type.Name, file)
		}
	}
}

// resolveFileType looks up a type name within a single file's type definitions
// or in the global basic types
func resolveFileType(typeName string, file *File) *Type {
	if typeName == "" {
		return nil
	}

	// Extract base type name by removing pointers, slices, maps, and channels
	baseName := extractBaseTypeName(typeName)

	// First check if it's a basic type
	if basicType := GetBasicType(baseName); basicType != nil {
		return basicType
	}

	// Look up the type in the file
	if typ, ok := file.Types[baseName]; ok {
		return typ
	}

	// Type not found in file (could be external or from another file in the package)
	return nil
}
