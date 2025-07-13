package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
)

// ModelAdapter adapts the existing model to use the new view system
type ModelAdapter struct {
	// Reference to original model fields
	Files        []string
	Filtered     []string
	Cursor       int
	Selected     string
	SearchMode   bool
	CreateMode   bool
	TagMode      bool
	PreviewMode  bool
	DeleteMode   bool
	SortMode     bool
	OldMode      bool
	TagCreateMode bool
	TaskCreateMode bool
	
	// Inputs
	SearchInput     textinput.Model
	CreateInput     textinput.Model
	TagInput        textinput.Model
	TagCreateInput  textinput.Model
	TaskCreateInput textinput.Model
	OldInput        textinput.Model
	
	// State
	PreviewContent string
	PreviewFile    string
	PreviewScroll  int
	DeleteFile     string
	PendingTitle   string
	CurrentSort    string
	ReversedSort   bool
	OldDays        int
	
	// Filters
	TaskFilter   bool
	TagFilter    bool
	TextFilter   bool
	DailyFilter  bool
	OldFilter    bool
	
	// Display
	Width  int
	Height int
	CWD    string
	
	// Configuration
	ShowTitles bool
	
	// UI Components
	Theme    Theme
	Layout   *Layout
	Composer *ViewComposer
}

// InitializeUI sets up the UI components
func (m *ModelAdapter) InitializeUI() {
	m.Theme = DefaultTheme()
	m.Layout = NewLayout(m.Width, m.Height, m.Theme)
	
	// Create view state
	state := m.createViewState()
	m.Composer = NewViewComposer(state)
	
	// Register inputs
	m.Composer.SetInput("search", m.SearchInput)
	m.Composer.SetInput("create", m.CreateInput)
	m.Composer.SetInput("tag", m.TagInput)
	m.Composer.SetInput("tagcreate", m.TagCreateInput)
	m.Composer.SetInput("task", m.TaskCreateInput)
	m.Composer.SetInput("old", m.OldInput)
}

// UpdateDimensions updates layout dimensions
func (m *ModelAdapter) UpdateDimensions(width, height int) {
	m.Width = width
	m.Height = height
	if m.Layout != nil {
		m.Layout.Width = width
		m.Layout.Height = height
	}
}

// View renders the current view
func (m *ModelAdapter) View() string {
	if m.Width == 0 || m.Height == 0 {
		return "Loading..."
	}
	
	// Initialize UI if needed
	if m.Composer == nil {
		m.InitializeUI()
	}
	
	// Update view state
	m.Composer.state = m.createViewState()
	
	// Render
	return m.Composer.Render()
}

// createViewState converts model state to view state
func (m *ModelAdapter) createViewState() ViewState {
	mode := m.getCurrentMode()
	
	return ViewState{
		Mode:           mode,
		Files:          m.Files,
		Filtered:       m.Filtered,
		Cursor:         m.Cursor,
		Width:          m.Width,
		Height:         m.Height,
		Theme:          m.Theme,
		Layout:         m.Layout,
		
		SearchQuery:    m.SearchInput.Value(),
		SelectedFile:   m.getDisplayName(m.PreviewFile),
		PreviewContent: m.PreviewContent,
		PreviewScroll:  m.PreviewScroll,
		DeleteTarget:   m.getDisplayName(m.DeleteFile),
		
		TaskFilter:     m.TaskFilter,
		TagFilter:      m.TagFilter,
		TextFilter:     m.TextFilter,
		DailyFilter:    m.DailyFilter,
		OldFilter:      m.OldFilter,
		OldDays:        m.OldDays,
		
		CurrentSort:    m.CurrentSort,
		ReversedSort:   m.ReversedSort,
	}
}

// getCurrentMode determines the current view mode
func (m *ModelAdapter) getCurrentMode() ViewMode {
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
	if m.OldMode {
		return ModeOldFilter
	}
	return ModeNormal
}

// getDisplayName returns the display name for a file
func (m *ModelAdapter) getDisplayName(filepath string) string {
	// This would call the actual getEnhancedDisplayName function
	// For now, return the path
	return filepath
}

// Example of how to integrate into existing model:
//
// In the main model struct, add:
//   uiAdapter *ModelAdapter
//
// In Init:
//   m.uiAdapter = &ModelAdapter{
//       Files: m.files,
//       Filtered: m.filtered,
//       // ... map all fields
//   }
//   m.uiAdapter.InitializeUI()
//
// In Update for WindowSizeMsg:
//   m.uiAdapter.UpdateDimensions(msg.Width, msg.Height)
//
// In View:
//   return m.uiAdapter.View()