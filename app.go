package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"supercalc/internal/calc"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const maxRecentFiles = 10

// App struct
type App struct {
	ctx         context.Context
	recentFiles []string
}

// NewApp creates a new App application struct
func NewApp() *App {
	app := &App{}
	app.loadRecentFiles()
	return app
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// getConfigPath returns the path to the config directory
func getConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}
	return filepath.Join(configDir, "supercalc")
}

// loadRecentFiles loads recent files from config
func (a *App) loadRecentFiles() {
	configPath := filepath.Join(getConfigPath(), "recent.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		a.recentFiles = []string{}
		return
	}
	json.Unmarshal(data, &a.recentFiles)
}

// saveRecentFiles saves recent files to config
func (a *App) saveRecentFiles() {
	configDir := getConfigPath()
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "recent.json")
	data, _ := json.Marshal(a.recentFiles)
	os.WriteFile(configPath, data, 0644)
}

// GetRecentFiles returns the list of recent files
func (a *App) GetRecentFiles() []string {
	return a.recentFiles
}

// AddRecentFile adds a file to the recent files list
func (a *App) AddRecentFile(path string) {
	// Remove if already exists
	filtered := []string{}
	for _, f := range a.recentFiles {
		if f != path {
			filtered = append(filtered, f)
		}
	}
	// Add to front
	a.recentFiles = append([]string{path}, filtered...)
	// Limit size
	if len(a.recentFiles) > maxRecentFiles {
		a.recentFiles = a.recentFiles[:maxRecentFiles]
	}
	a.saveRecentFiles()
}

// EvalResult represents a single line evaluation result
type EvalResult struct {
	LineNum int    `json:"lineNum"`
	Input   string `json:"input"`
	Output  string `json:"output"`
}

// Evaluate evaluates all lines and returns results
func (a *App) Evaluate(text string) []EvalResult {
	lines := strings.Split(text, "\n")
	results := calc.EvalLines(lines)

	evalResults := make([]EvalResult, len(results))
	for i, r := range results {
		evalResults[i] = EvalResult{
			LineNum: i + 1,
			Input:   lines[i],
			Output:  r.Output,
		}
	}
	return evalResults
}

// GetVersion returns the app version
func (a *App) GetVersion() string {
	return version
}

// OpenFileDialog opens a file dialog and returns the selected path
func (a *App) OpenFileDialog() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open File",
		Filters: []runtime.FileFilter{
			{DisplayName: "SuperCalc Files", Pattern: "*.txt;*.sc"},
			{DisplayName: "All Files", Pattern: "*"},
		},
	})
}

// SaveFileDialog opens a save dialog and returns the selected path
func (a *App) SaveFileDialog() (string, error) {
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save File",
		DefaultFilename: "untitled.txt",
		Filters: []runtime.FileFilter{
			{DisplayName: "Text Files", Pattern: "*.txt"},
			{DisplayName: "SuperCalc Files", Pattern: "*.sc"},
			{DisplayName: "All Files", Pattern: "*"},
		},
	})
}

// ReadFile reads a file and returns its contents
func (a *App) ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteFile writes content to a file
func (a *App) WriteFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// CopyWithResolvedRefs copies text with references replaced by values
func (a *App) CopyWithResolvedRefs(text string) string {
	return calc.ReplaceRefsWithValues(text)
}
