# docs

docks package implements structured documentation using reflect entries as source.
It uses standard godoc features for parsing, returning documentation entries with well defined title, summary, details, examples, etc.

# godoc syntax

See https://go.dev/doc/comment for details.

Godoc documentation entries ([`Doc`](https://go.dev/doc/comment)) are composed of a list of documentation blocks and a list of links to other documentation entries. Blocks can be paragraphs, bullet lists, code blocks, etc.
In cucaracha, we tried to simplify the documentation structure a bit, by following these rules:
 - First paragraph is the summary of the documentation. Cucaracha's `DocumentationEntry` exposes this as a single string in its `Summary` field.
 - All following paragraphs are the details of the documentation. Cucaracha's `DocumentationEntry` exposes this as a single string in its `Details` field.
 - Code blocks are treated as examples of usage, collected in a single `Examples` array.

godoc API is probably better suited, for example cucaracha `docs` model is not well suited for displaying in display-width aware situations, like terminals, while godoc `Printer` API does (godoc's `Doc` stores paragraphs as slices of text lines).