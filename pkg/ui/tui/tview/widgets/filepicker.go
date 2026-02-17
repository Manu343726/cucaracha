package widgets

import (
	"os"
	"path/filepath"

	"github.com/Manu343726/cucaracha/pkg/ui/tui/tview/themes"
	tvlib "github.com/rivo/tview"
)

// FilePicker is a widget for selecting files
type FilePicker struct {
	*tvlib.List
	currentDir  string
	files       []string
	selectedIdx int
}

// NewFilePicker creates a new FilePicker widget
func NewFilePicker(startDir string) *FilePicker {
	list := tvlib.NewList()
	list.SetBorder(true)
	list.SetTitle("Select File")
	fp := &FilePicker{
		List:        list,
		currentDir:  startDir,
		files:       []string{},
		selectedIdx: 0,
	}
	fp.loadDirectory()
	return fp
}

// loadDirectory loads the files from the current directory
func (fp *FilePicker) loadDirectory() {
	fp.files = []string{}
	fp.List.Clear()

	entries, err := os.ReadDir(fp.currentDir)
	if err != nil {
		fp.List.AddItem("..", "Parent directory", 0, nil)
		return
	}

	// Add parent directory option
	fp.List.AddItem("..", "Go to parent directory", 0, nil)

	// Add directories and files
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		fp.files = append(fp.files, name)
		fp.List.AddItem(name, "", 0, nil)
	}
}

// SelectedFile returns the currently selected file path
func (fp *FilePicker) SelectedFile() string {
	idx := fp.List.GetCurrentItem()
	if idx > 0 && idx-1 < len(fp.files) {
		file := fp.files[idx-1]
		fullPath := filepath.Join(fp.currentDir, file)
		return fullPath
	}
	return ""
}

// SetTheme applies the theme to the FilePicker widget
func (fp *FilePicker) SetTheme(theme *themes.Theme) *FilePicker {
	fp.List.SetBackgroundColor(theme.PrimitiveBackgroundColor)
	fp.List.SetMainTextColor(theme.PrimaryTextColor)
	fp.List.SetBorderColor(theme.BorderColor)
	return fp
}
