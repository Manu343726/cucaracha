package themes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	tvlib "github.com/rivo/tview"
	"gopkg.in/yaml.v2"
)

// CSyntaxTheme defines colors for C syntax highlighting
type CSyntaxTheme struct {
	Keyword      tcell.Color
	Type         tcell.Color
	Function     tcell.Color
	String       tcell.Color
	Comment      tcell.Color
	Preprocessor tcell.Color
	Number       tcell.Color
	Operator     tcell.Color
}

// SourceSnippetTheme defines colors for source code snippets
type SourceSnippetTheme struct {
	LineNumber       tcell.Color
	CurrentLine      tcell.Color
	BreakpointMarker tcell.Color
	C                *CSyntaxTheme
}

// MemoryDumpTheme defines colors for memory dump view
type MemoryDumpTheme struct {
	Address          tcell.Color
	HexDump          tcell.Color
	ASCII            tcell.Color
	HighlightBG      tcell.Color
	WatchpointMarker tcell.Color
}

// DisassemblyTheme defines colors for disassembly view
type DisassemblyTheme struct {
	Address                     tcell.Color
	RegisterOperand             tcell.Color
	ImmediateOperand            tcell.Color
	Mnemonic                    tcell.Color
	HighlightBG                 tcell.Color
	BreakpointMarker            tcell.Color
	CallGraphColors             []tcell.Color
	DataDependenciesGraphColors []tcell.Color
}

// RegistersTheme defines colors for register values
type RegistersTheme struct {
	RegisterName              tcell.Color
	RegisterValue_Decimal     tcell.Color
	RegisterValue_Hexadecimal tcell.Color
	RegisterValue_Binary      tcell.Color
}

// UserIOTheme defines colors for user I/O elements
type UserIOTheme struct {
	CommandPrompt tcell.Color
	Info          tcell.Color
	Error         tcell.Color
	Success       tcell.Color
	Warning       tcell.Color
}

// EventsTheme defines colors for debugger events display
type EventsTheme struct {
	Timestamp             tcell.Color
	ProgramLoaded         tcell.Color
	Stepped               tcell.Color
	BreakpointHit         tcell.Color
	WatchpointHit         tcell.Color
	ProgramTerminated     tcell.Color
	ProgramHalted         tcell.Color
	Error                 tcell.Color
	SourceLocationChanged tcell.Color
	Interrupted           tcell.Color
	Lagging               tcell.Color
	EventDetail           tcell.Color
}

// Theme defines colors for different UI elements
type Theme struct {
	Name          string
	Description   string
	UserIO        *UserIOTheme
	Events        *EventsTheme
	SourceSnippet *SourceSnippetTheme
	MemoryDump    *MemoryDumpTheme
	Disassembly   *DisassemblyTheme
	Registers     *RegistersTheme
	*tvlib.Theme
}

// Formatting functions for theme-driven output using a single sprintf call
func (t *Theme) FormatError(format string, args ...any) string {
	color := fmt.Sprintf("%06x", t.UserIO.Error.Hex())
	allArgs := append([]any{color}, args...)
	return fmt.Sprintf("[#%s]"+format+"[-]", allArgs...)
}

func (t *Theme) FormatSuccess(format string, args ...any) string {
	color := fmt.Sprintf("%06x", t.UserIO.Success.Hex())
	allArgs := append([]any{color}, args...)
	return fmt.Sprintf("[#%s]"+format+"[-]", allArgs...)
}

func (t *Theme) FormatWarning(format string, args ...any) string {
	color := fmt.Sprintf("%06x", t.UserIO.Warning.Hex())
	allArgs := append([]any{color}, args...)
	return fmt.Sprintf("[#%s]"+format+"[-]", allArgs...)
}

func (t *Theme) FormatInfo(format string, args ...any) string {
	color := fmt.Sprintf("%06x", t.UserIO.Info.Hex())
	allArgs := append([]any{color}, args...)
	return fmt.Sprintf("[#%s]"+format+"[-]", allArgs...)
}

func (t *Theme) FormatPrompt(format string, args ...any) string {
	color := fmt.Sprintf("%06x", t.UserIO.CommandPrompt.Hex())
	allArgs := append([]any{color}, args...)
	return fmt.Sprintf("[#%s]"+format+"[-]", allArgs...)
}

// LoadTheme loads a theme from a JSON or YAML file
// The file extension determines the format (.json or .yaml/.yml)
func LoadTheme(filePath string) (*Theme, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read theme file: %w", err)
	}

	// Determine file format from extension
	ext := strings.ToLower(filepath.Ext(filePath))

	theme := &Theme{
		Theme: &tvlib.Theme{},
	}

	switch ext {
	case ".json":
		err = json.Unmarshal(data, theme)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, theme)
	default:
		return nil, fmt.Errorf("unsupported file format: %s (use .json, .yaml, or .yml)", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse theme file: %w", err)
	}

	return theme, nil
}

// vscodeTheme represents a VSCode theme from the marketplace
type vscodeTheme struct {
	Name        string             `json:"name"`
	Colors      map[string]string  `json:"colors"`
	TokenColors []vscodeTokenColor `json:"tokenColors"`
}

// vscodeTokenColor represents a VSCode token color rule
type vscodeTokenColor struct {
	Scope    interface{}            `json:"scope"`
	Settings map[string]interface{} `json:"settings"`
}

// LoadVSCodeThemeFromURL loads a VSCode theme from a URL
// The URL should point to a VSCode theme JSON file
// Example: https://raw.githubusercontent.com/user/theme-repo/main/theme.json
func LoadVSCodeThemeFromURL(url string) (*Theme, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Fetch the theme file
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch theme from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch theme: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return parseVSCodeTheme(data)
}

// LoadVSCodeThemeFromFile loads a VSCode theme from a local JSON file
func LoadVSCodeThemeFromFile(filePath string) (*Theme, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read VSCode theme file: %w", err)
	}

	return parseVSCodeTheme(data)
}

// parseVSCodeTheme parses a VSCode theme JSON and converts it to our Theme format
func parseVSCodeTheme(data []byte) (*Theme, error) {
	var vscodeTheme vscodeTheme
	if err := json.Unmarshal(data, &vscodeTheme); err != nil {
		return nil, fmt.Errorf("failed to parse VSCode theme: %w", err)
	}

	// Helper function to parse color string
	parseColor := func(colorStr string) tcell.Color {
		if colorStr == "" {
			return tcell.ColorDefault
		}
		// Remove # if present
		if strings.HasPrefix(colorStr, "#") {
			colorStr = colorStr[1:]
		}
		// Pad with zeros if needed
		if len(colorStr) == 6 {
			colorStr = "#" + colorStr
		}
		return tcell.GetColor(colorStr)
	}

	theme := &Theme{
		Name:        vscodeTheme.Name,
		Description: fmt.Sprintf("Imported from VSCode: %s", vscodeTheme.Name),
		UserIO:      &UserIOTheme{},
		SourceSnippet: &SourceSnippetTheme{
			C: &CSyntaxTheme{},
		},
		MemoryDump:  &MemoryDumpTheme{},
		Disassembly: &DisassemblyTheme{},
		Registers:   &RegistersTheme{},
		Theme:       &tvlib.Theme{},
	}

	// Map VSCode colors to our theme
	// Editor colors
	if bg, ok := vscodeTheme.Colors["editor.background"]; ok {
		theme.Theme.PrimitiveBackgroundColor = parseColor(bg)
	}
	if fg, ok := vscodeTheme.Colors["editor.foreground"]; ok {
		theme.Theme.PrimaryTextColor = parseColor(fg)
	}
	if lineNum, ok := vscodeTheme.Colors["editorLineNumber.foreground"]; ok {
		theme.SourceSnippet.LineNumber = parseColor(lineNum)
	}
	if currentLine, ok := vscodeTheme.Colors["editor.lineHighlightBackground"]; ok {
		theme.SourceSnippet.CurrentLine = parseColor(currentLine)
	}

	// Border and UI colors
	if border, ok := vscodeTheme.Colors["editorBracketMatch.border"]; ok {
		theme.Theme.BorderColor = parseColor(border)
	} else if editorGroup, ok := vscodeTheme.Colors["editorGroup.border"]; ok {
		theme.Theme.BorderColor = parseColor(editorGroup)
	}

	// Extract token colors for syntax highlighting
	for _, tokenColor := range vscodeTheme.TokenColors {
		scope := ""
		if scopeStr, ok := tokenColor.Scope.(string); ok {
			scope = scopeStr
		}

		if settings, ok := tokenColor.Settings["foreground"]; ok {
			if colorStr, ok := settings.(string); ok {
				color := parseColor(colorStr)

				// Map scopes to syntax theme colors
				switch {
				case strings.Contains(scope, "keyword"):
					theme.SourceSnippet.C.Keyword = color
				case strings.Contains(scope, "type"):
					theme.SourceSnippet.C.Type = color
				case strings.Contains(scope, "function"):
					theme.SourceSnippet.C.Function = color
				case strings.Contains(scope, "string"):
					theme.SourceSnippet.C.String = color
				case strings.Contains(scope, "comment"):
					theme.SourceSnippet.C.Comment = color
				case strings.Contains(scope, "preprocessor"):
					theme.SourceSnippet.C.Preprocessor = color
				case strings.Contains(scope, "constant") || strings.Contains(scope, "number"):
					theme.SourceSnippet.C.Number = color
				case strings.Contains(scope, "operator"):
					theme.SourceSnippet.C.Operator = color
				}
			}
		}
	}

	// Set default colors if not found
	if theme.SourceSnippet.LineNumber == tcell.ColorDefault {
		theme.SourceSnippet.LineNumber = parseColor("#858585")
	}
	if theme.SourceSnippet.C.Keyword == tcell.ColorDefault {
		theme.SourceSnippet.C.Keyword = parseColor("#569cd6")
	}
	if theme.Theme.BorderColor == tcell.ColorDefault {
		theme.Theme.BorderColor = parseColor("#cccccc")
	}

	// Set fallback values for required fields
	if theme.Theme.PrimitiveBackgroundColor == tcell.ColorDefault {
		theme.Theme.PrimitiveBackgroundColor = parseColor("#ffffff")
	}
	if theme.Theme.PrimaryTextColor == tcell.ColorDefault {
		theme.Theme.PrimaryTextColor = parseColor("#000000")
	}

	return theme, nil
}

// VSCodeThemeInfo represents a theme from the VSCode marketplace
type VSCodeThemeInfo struct {
	Name        string
	Publisher   string
	ID          string
	Version     string
	Description string
	DownloadURL string
}

// vscodeMarketplaceResponse represents the marketplace API response
type vscodeMarketplaceResponse struct {
	Results []struct {
		Publisher struct {
			DisplayName string `json:"displayName"`
		} `json:"publisher"`
		ExtensionID      string `json:"extensionId"`
		DisplayName      string `json:"displayName"`
		Version          string `json:"version"`
		ShortDescription string `json:"shortDescription"`
		Files            []struct {
			AssetType string `json:"assetType"`
			Source    string `json:"source"`
		} `json:"files"`
	} `json:"results"`
}

// ListVSCodeThemes lists available color themes from the VSCode marketplace
func ListVSCodeThemes() ([]VSCodeThemeInfo, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// VSCode marketplace API endpoint for theme search
	url := "https://marketplace.visualstudio.com/api/gallery/queryallpublished"

	reqBody := strings.NewReader(`{
		"filters": [
			{
				"criteria": [
					{
						"filterType": 8,
						"value": "Microsoft.VisualStudio.Code.Themes"
					}
				]
			}
		],
		"pageNumber": 1,
		"pageSize": 50,
		"sortBy": 4,
		"sortOrder": 0
	}`)

	req, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create marketplace request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json;api-version=7.2-preview.1")
	req.Header.Set("User-Agent", "cucaracha-debugger")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch themes from marketplace: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("marketplace returned HTTP %d", resp.StatusCode)
	}

	var marketplaceResp vscodeMarketplaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&marketplaceResp); err != nil {
		return nil, fmt.Errorf("failed to parse marketplace response: %w", err)
	}

	var themes []VSCodeThemeInfo
	for _, result := range marketplaceResp.Results {
		// Find the theme file URL
		var themeURL string
		for _, file := range result.Files {
			if strings.Contains(file.AssetType, "Microsoft.VisualStudio.Code.Themes") {
				themeURL = file.Source
				break
			}
		}

		if themeURL != "" {
			themes = append(themes, VSCodeThemeInfo{
				Name:        result.DisplayName,
				Publisher:   result.Publisher.DisplayName,
				ID:          result.ExtensionID,
				Version:     result.Version,
				Description: result.ShortDescription,
				DownloadURL: themeURL,
			})
		}
	}

	return themes, nil
}

// LoadVSCodeThemeByName loads a VSCode theme by name from the marketplace
// If multiple themes have the same name, the first match is returned
func LoadVSCodeThemeByName(themeName string) (*Theme, error) {
	themes, err := ListVSCodeThemes()
	if err != nil {
		return nil, fmt.Errorf("failed to list marketplace themes: %w", err)
	}

	// Search for the theme by name (case-insensitive)
	lowerName := strings.ToLower(themeName)
	for _, themeInfo := range themes {
		if strings.ToLower(themeInfo.Name) == lowerName {
			return LoadVSCodeThemeFromURL(themeInfo.DownloadURL)
		}
	}

	return nil, fmt.Errorf("theme %q not found in marketplace", themeName)
}
