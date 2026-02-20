package reflect

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

// ParseFile parses a single Go file and extracts all types, functions, and constants
func ParseFile(filePath string) (*File, error) {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	file := &File{
		Name:      filepath.Base(filePath),
		Path:      filePath,
		Package:   astFile.Name.Name,
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
			parseGenDecl(file, decl)
		case *ast.FuncDecl:
			if fn := parseFuncDecl(decl, file.Package); fn != nil {
				file.Functions = append(file.Functions, fn)
			}
		}
	}

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
		typ.Kind = TypeKindAlias
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
		tag := ""
		if field.Tag != nil {
			tag = strings.Trim(field.Tag.Value, "`")
		}

		// Handle embedded fields (no field name)
		if len(field.Names) == 0 {
			fields = append(fields, &Field{
				Name:       typeStr,
				Type:       typeStr,
				Tag:        tag,
				Doc:        field.Doc.Text(),
				IsEmbedded: true,
			})
		} else {
			for _, name := range field.Names {
				fields = append(fields, &Field{
					Name: name.Name,
					Type: typeStr,
					Tag:  tag,
					Doc:  field.Doc.Text(),
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
		fn.IsMethod = true
		fn.Receiver = typeToString(decl.Recv.List[0].Type)
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

// parseParameters extracts parameters from a parameter list
func parseParameters(params *ast.FieldList) []*Parameter {
	var result []*Parameter

	if params == nil || params.List == nil {
		return result
	}

	for _, field := range params.List {
		typeStr := typeToString(field.Type)

		if len(field.Names) == 0 {
			// Unnamed parameter (common in results)
			result = append(result, &Parameter{
				Name: "",
				Type: typeStr,
			})
		} else {
			for _, name := range field.Names {
				result = append(result, &Parameter{
					Name: name.Name,
					Type: typeStr,
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
