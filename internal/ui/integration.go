package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// ModelIntegration integrates the new UI system with the existing model
type ModelIntegration struct {
	// Core state
	Files       []string
	Filtered    []string
	Cursor      int
	Selected    string
	CWD         string
	Width       int
	Height      int

	// Mode flags
	SearchMode      bool
	CreateMode      bool
	TagMode         bool
	TagCreateMode   bool
	TaskCreateMode  bool
	PreviewMode     bool
	DeleteMode      bool
	SortMode        bool
	OldMode         bool
	RenameMode      bool
	TaskFilterMode  bool

	// Mode-specific data
	Search          textinput.Model
	CreateInput     textinput.Model
	TagInput        textinput.Model
	TagCreateInput  textinput.Model
	TaskCreateInput textinput.Model
	OldInput        textinput.Model

	// Preview state
	PreviewContent string
	PreviewFile    string
	PreviewScroll  int

	// Other state
	DeleteFile     string
	RenameFile     string
	PendingTitle   string
	CurrentSort    string
	ReversedSort   bool
	OldDays        int
	TextFilter     bool
	TagFilter      bool
	TaskFilter     bool
	DailyFilter    bool
	OldFilter      bool

	// Status message
	StatusMsg      StatusMessage
	StatusDuration int // frames remaining

	// Configuration
	ShowTitles         bool
	DenoteFilenames    bool
	TaskwarriorSupport bool
	DenoteTasksSupport bool
	ThemeName          string
	
	// Task mode
	TaskModeActive     bool
	TaskSortBy         string
	TaskFormatter      func(string) string // Function to format task lines
	TaskAreaContext    string              // Current area context (empty = all areas)
	TaskStatusFilter   string              // Current status filter within area

	// UI Components
	theme    Theme
	layout   *Layout
	composer *ViewComposer
}

// Initialize sets up the UI components
func (m *ModelIntegration) Initialize() {
	m.theme = GetTheme(m.ThemeName)
	m.layout = NewLayout(m.Width, m.Height, m.theme)
	m.createComposer()
}

// UpdateSize updates the UI dimensions
func (m *ModelIntegration) UpdateSize(width, height int) {
	m.Width = width
	m.Height = height
	if m.layout != nil {
		m.layout.Width = width
		m.layout.Height = height
	}
}

// SetStatus sets a status message
func (m *ModelIntegration) SetStatus(msg string, msgType StatusType, duration int) {
	m.StatusMsg = StatusMessage{
		Message:  msg,
		Type:     msgType,
		Duration: duration,
		Style:    m.theme.Status,
	}
	m.StatusDuration = duration
}

// UpdateStatus decrements the status duration
func (m *ModelIntegration) UpdateStatus() {
	if m.StatusDuration > 0 {
		m.StatusDuration--
		m.StatusMsg.Duration = m.StatusDuration
	}
}

// HandleStatusMsg handles status messages
func (m *ModelIntegration) HandleStatusMsg(msg StatusMsg) tea.Cmd {
	// Convert duration to frames (assuming ~60fps)
	frames := int(msg.Duration.Seconds() * 60)
	m.SetStatus(msg.Message, msg.Type, frames)
	return nil
}

// HandleClearStatusMsg clears the status
func (m *ModelIntegration) HandleClearStatusMsg() {
	m.StatusDuration = 0
	m.StatusMsg.Duration = 0
}

// Render produces the UI output
func (m *ModelIntegration) Render() string {
	if m.Width == 0 || m.Height == 0 {
		return "Loading..."
	}

	// Initialize if needed
	if m.composer == nil {
		m.Initialize()
	}

	// Update status duration
	m.UpdateStatus()

	// Update composer state
	m.updateComposerState()

	// Render
	return m.composer.Render()
}

// createComposer initializes the view composer
func (m *ModelIntegration) createComposer() {
	state := m.createViewState()
	m.composer = NewViewComposer(state)

	// Register inputs
	m.composer.SetInput("search", m.Search)
	m.composer.SetInput("create", m.CreateInput)
	m.composer.SetInput("tag", m.TagInput)
	m.composer.SetInput("tagcreate", m.TagCreateInput)
	m.composer.SetInput("task", m.TaskCreateInput)
	m.composer.SetInput("old", m.OldInput)
}

// updateComposerState updates the composer with current state
func (m *ModelIntegration) updateComposerState() {
	m.composer.state = m.createViewState()
	
	// Update inputs
	m.composer.SetInput("search", m.Search)
	m.composer.SetInput("create", m.CreateInput)
	m.composer.SetInput("tag", m.TagInput)
	m.composer.SetInput("tagcreate", m.TagCreateInput)
	m.composer.SetInput("task", m.TaskCreateInput)
	m.composer.SetInput("old", m.OldInput)
}

// createViewState converts model state to view state
func (m *ModelIntegration) createViewState() ViewState {
	// Process filenames for display
	displayFiles := make([]string, len(m.Filtered))
	
	// In task mode, pass paths directly - the formatter will handle display
	if m.TaskModeActive && m.TaskFormatter != nil {
		// Create a proper copy of the filtered paths
		for i, file := range m.Filtered {
			displayFiles[i] = file
		}
	} else {
		for i, file := range m.Filtered {
			displayFiles[i] = m.getEnhancedDisplayName(file)
		}
	}

	return ViewState{
		Mode:           m.getCurrentMode(),
		Files:          displayFiles,
		Filtered:       displayFiles,
		Cursor:         m.Cursor,
		Width:          m.Width,
		Height:         m.Height,
		Theme:          m.theme,
		Layout:         m.layout,
		
		SearchQuery:    m.Search.Value(),
		SelectedFile:   m.getEnhancedDisplayName(m.PreviewFile),
		PreviewContent: m.PreviewContent,
		PreviewScroll:  m.PreviewScroll,
		DeleteTarget:   m.getEnhancedDisplayName(m.DeleteFile),
		StatusMessage:  m.StatusMsg,
		
		TaskFilter:     m.TaskFilter,
		TagFilter:      m.TagFilter,
		TextFilter:     m.TextFilter,
		DailyFilter:    m.DailyFilter,
		OldFilter:      m.OldFilter,
		OldDays:        m.OldDays,
		
		CurrentSort:    m.CurrentSort,
		ReversedSort:   m.ReversedSort,
		
		TaskwarriorSupport: m.TaskwarriorSupport,
		DenoteTasksSupport: m.DenoteTasksSupport,
		TaskModeActive:     m.TaskModeActive,
		TaskSortBy:         m.TaskSortBy,
		TaskFormatter:      m.TaskFormatter,
		TaskAreaContext:    m.TaskAreaContext,
		TaskStatusFilter:   m.TaskStatusFilter,
	}
}

// getCurrentMode determines the current view mode
func (m *ModelIntegration) getCurrentMode() ViewMode {
	if m.PreviewMode {
		return ModePreview
	}
	if m.DeleteMode {
		return ModeDelete
	}
	if m.SortMode {
		return ModeSort
	}
	if m.SearchMode {
		return ModeSearch
	}
	if m.CreateMode {
		return ModeCreate
	}
	if m.TagMode {
		return ModeTagSearch
	}
	if m.TagCreateMode {
		return ModeTagCreate
	}
	if m.TaskCreateMode {
		return ModeTaskCreate
	}
	if m.OldMode {
		return ModeOldFilter
	}
	if m.TaskFilterMode {
		return ModeTaskFilter
	}
	if m.SortMode {
		return ModeSort
	}
	return ModeNormal
}

// getEnhancedDisplayName returns display name for a file
func (m *ModelIntegration) getEnhancedDisplayName(fullPath string) string {
	if fullPath == "" {
		return ""
	}

	// Get relative path
	rel, err := filepath.Rel(m.CWD, fullPath)
	if err != nil {
		rel = filepath.Base(fullPath)
	}

	// If not showing titles, return relative path
	if !m.ShowTitles {
		return rel
	}

	// Try to extract title from file content
	title := ExtractNoteTitle(fullPath)
	
	// If we got a title from the file content, use it
	if title != "" {
		// Check if it's a Denote file to add date
		filename := filepath.Base(fullPath)
		if m.DenoteFilenames && len(filename) > 16 && filename[8] == 'T' {
			// Extract date from Denote filename
			if filename[15] == '-' || filename[15] == '_' {
				date := filename[:8]
				year := date[:4]
				month := date[4:6]
				day := date[6:8]
				return fmt.Sprintf("%s (%s-%s-%s)", title, year, month, day)
			}
		}
		return title
	}

	// If no title in content, try parsing Denote filename
	filename := filepath.Base(fullPath)
	if m.DenoteFilenames && len(filename) > 16 && filename[8] == 'T' {
		// Try to parse Denote format
		if filename[15] == '-' || filename[15] == '_' {
			// Extract title part
			titleStart := 16
			titleEnd := strings.IndexAny(filename[titleStart:], "_.")
			if titleEnd > 0 {
				title := filename[titleStart:titleStart+titleEnd]
				// Convert hyphens to spaces (no capitalization)
				title = strings.ReplaceAll(title, "-", " ")
				
				// Add date
				date := filename[:8]
				year := date[:4]
				month := date[4:6]
				day := date[6:8]
				
				return fmt.Sprintf("%s (%s-%s-%s)", title, year, month, day)
			}
		}
	}

	// Fall back to relative path
	return rel
}