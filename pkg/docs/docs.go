// Package docs provides structured documentation generation and indexing.
//
// It uses the reflect package to parse types, functions, packages, and other Go
// constructs, then builds a comprehensive, indexed documentation system.
//
// The documentation index is structured into searchable and cross-linked entries,
// where each entry includes:
// - A title and summary
// - Detailed documentation
// - Code examples
// - Links to related entries (showing dependencies, implementations, etc.)
//
// Example usage:
//
//	pkgs := []*reflect.Package{...}
//	builder := docs.NewBuilder()
//	index, err := builder.Build(pkgs)
//	if err != nil {
//		log.Fatal(err)
//	}
//	// index now contains all documented items across packages
package docs
