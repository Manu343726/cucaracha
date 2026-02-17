package themes

// Themes is the global map of all available themes
var Themes = map[string]*Theme{
	"monokai-dark":   MonokaiDark,
	"dracula":        Dracula,
	"nord":           Nord,
	"solarized-dark": SolarizedDark,
	"gruvbox-dark":   GruvboxDark,
	"one-dark":       OneDark,
}
