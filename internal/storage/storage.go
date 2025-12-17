package storage

import (
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

const (
	PrefLastFile    = "lastFile"
	PrefRecentFiles = "recentFiles"
	MaxRecentFiles  = 10
)

// Manager handles file operations and preferences.
type Manager struct {
	app         fyne.App
	window      fyne.Window
	currentFile string
	onLoad      func(content string)
	getContent  func() string
	onNew       func()
}

// NewManager creates a new storage manager.
func NewManager(app fyne.App, window fyne.Window, onLoad func(string), getContent func() string, onNew func()) *Manager {
	return &Manager{
		app:        app,
		window:     window,
		onLoad:     onLoad,
		getContent: getContent,
		onNew:      onNew,
	}
}

// CurrentFile returns the current file path.
func (m *Manager) CurrentFile() string {
	return m.currentFile
}

// New clears the current document.
func (m *Manager) New() {
	m.currentFile = ""
	m.onNew()
	m.window.SetTitle("SuperCalc - Untitled")
}

// Open shows a file dialog and loads the selected file.
func (m *Manager) Open() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()

		data, err := os.ReadFile(reader.URI().Path())
		if err != nil {
			dialog.ShowError(err, m.window)
			return
		}

		m.currentFile = reader.URI().Path()
		m.onLoad(string(data))
		m.updateTitle()
		m.saveLastFile()
		m.addToRecent(m.currentFile)
	}, m.window)
}

// OpenFile loads a specific file by path.
func (m *Manager) OpenFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	m.currentFile = path
	m.onLoad(string(data))
	m.updateTitle()
	m.saveLastFile()
	m.addToRecent(path)
	return nil
}

// Save saves to the current file, or shows SaveAs if no file is set.
func (m *Manager) Save() {
	if m.currentFile == "" {
		m.SaveAs()
		return
	}
	content := m.getContent()
	if err := os.WriteFile(m.currentFile, []byte(content), 0644); err != nil {
		dialog.ShowError(err, m.window)
		return
	}
	m.saveLastFile()
}

// SaveAs shows a file dialog to save to a new file.
func (m *Manager) SaveAs() {
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil || writer == nil {
			return
		}
		defer writer.Close()

		content := m.getContent()
		if _, err := writer.Write([]byte(content)); err != nil {
			dialog.ShowError(err, m.window)
			return
		}

		m.currentFile = writer.URI().Path()
		m.updateTitle()
		m.saveLastFile()
		m.addToRecent(m.currentFile)
	}, m.window)
}

// LoadLastFile loads the last opened file if it exists.
func (m *Manager) LoadLastFile() bool {
	path := m.app.Preferences().String(PrefLastFile)
	if path == "" {
		return false
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return m.OpenFile(path) == nil
}

// GetRecentFiles returns the list of recent files.
func (m *Manager) GetRecentFiles() []string {
	raw := m.app.Preferences().String(PrefRecentFiles)
	if raw == "" {
		return nil
	}
	return strings.Split(raw, "\n")
}

func (m *Manager) updateTitle() {
	if m.currentFile == "" {
		m.window.SetTitle("SuperCalc - Untitled")
	} else {
		m.window.SetTitle("SuperCalc - " + filepath.Base(m.currentFile))
	}
}

func (m *Manager) saveLastFile() {
	m.app.Preferences().SetString(PrefLastFile, m.currentFile)
}

func (m *Manager) addToRecent(path string) {
	recent := m.GetRecentFiles()

	// Remove if already exists
	filtered := make([]string, 0, len(recent))
	for _, r := range recent {
		if r != path && r != "" {
			filtered = append(filtered, r)
		}
	}

	// Add to front
	filtered = append([]string{path}, filtered...)

	// Limit size
	if len(filtered) > MaxRecentFiles {
		filtered = filtered[:MaxRecentFiles]
	}

	m.app.Preferences().SetString(PrefRecentFiles, strings.Join(filtered, "\n"))
}

// CreateOpenURI creates a URI for the file dialog starting location.
func CreateOpenURI(path string) fyne.ListableURI {
	if path == "" {
		home, _ := os.UserHomeDir()
		path = home
	} else {
		path = filepath.Dir(path)
	}
	uri := storage.NewFileURI(path)
	listable, _ := storage.ListerForURI(uri)
	return listable
}
