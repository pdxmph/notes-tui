package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// ViewState represents the current view configuration
type ViewState struct {
	Mode            ViewMode
	Files           []string
	Filtered        []string
	Cursor          int
	Width           int
	Height          int
	Theme           Theme
	Layout          *Layout
	
	// Mode-specific data
	SearchQuery     string
	SelectedFile    string
	PreviewContent  string
	PreviewScroll   int
	DeleteTarget    string
	StatusMessage   StatusMessage
	
	// Filter states
	TaskFilter      bool
	TagFilter       bool
	TextFilter      bool
	DailyFilter     bool
	OldFilter       bool
	OldDays         int
	
	// Sort state
	CurrentSort     string
	ReversedSort    bool
	
	// Configuration
	TaskwarriorSupport bool
}

// ViewMode represents the current interaction mode
type ViewMode int

const (
	ModeNormal ViewMode = iota
	ModeSearch
	ModeCreate
	ModeTagSearch
	ModeTagCreate
	ModeTaskCreate
	ModeSort
	ModeOldFilter
	ModeDelete
	ModePreview
	ModeLoading
)

// ViewComposer handles view composition
type ViewComposer struct {
	state  ViewState
	inputs map[string]textinput.Model
}

// NewViewComposer creates a new view composer
func NewViewComposer(state ViewState) *ViewComposer {
	return &ViewComposer{
		state:  state,
		inputs: make(map[string]textinput.Model),
	}
}

// SetInput stores an input model
func (v *ViewComposer) SetInput(name string, input textinput.Model) {
	v.inputs[name] = input
}

// Render composes the appropriate view based on state
func (v *ViewComposer) Render() string {
	// Handle special modes first
	switch v.state.Mode {
	case ModePreview:
		return v.renderPreview()
	case ModeLoading:
		return v.renderLoading()
	}
	
	// Compose standard layout
	header := v.renderHeader()
	content := v.renderContent()
	footer := v.renderFooter()
	
	return v.state.Layout.RenderScreen(header, content, footer)
}

// renderHeader creates the header based on current state
func (v *ViewComposer) renderHeader() string {
	filters := v.getActiveFilters()
	sortInfo := v.getSortInfo()
	
	header := Header{
		Title:     "Notes",
		FileCount: len(v.state.Filtered),
		Filters:   filters,
		SortInfo:  sortInfo,
		Width:     v.state.Width,
		Style:     v.state.Theme.Header,
	}
	
	headerView := header.View()
	
	// Add status message if present
	if v.state.StatusMessage.Duration > 0 {
		statusView := v.state.StatusMessage.View()
		if statusView != "" {
			headerView += "\n" + statusView
		}
	}
	
	return headerView
}

// renderContent creates the main content area
func (v *ViewComposer) renderContent() string {
	switch v.state.Mode {
	case ModeSearch:
		return v.renderSearchMode()
	case ModeCreate:
		return v.renderCreateMode()
	case ModeTagSearch:
		return v.renderTagSearchMode()
	case ModeTagCreate:
		return v.renderTagCreateMode()
	case ModeTaskCreate:
		return v.renderTaskCreateMode()
	case ModeSort:
		return v.renderSortMode()
	case ModeOldFilter:
		return v.renderOldFilterMode()
	case ModeDelete:
		return v.renderDeleteMode()
	default:
		return v.renderFileList()
	}
}

// renderFooter creates the footer with help text
func (v *ViewComposer) renderFooter() string {
	if v.state.Mode == ModeNormal {
		return v.renderHelpBar()
	}
	
	// Mode-specific footers are included in content
	return ""
}

// renderFileList creates the main file list view
func (v *ViewComposer) renderFileList() string {
	contentWidth, contentHeight := v.state.Layout.ContentArea()
	
	list := ListView{
		Items:        v.state.Filtered,
		Cursor:       v.state.Cursor,
		Width:        contentWidth,
		Height:       contentHeight - 6, // Reserve space for header/footer
		ShowCursor:   true,
		EmptyMessage: "No files found.",
		Style:        v.state.Theme.List,
	}
	
	if len(v.state.Files) == 0 {
		list.EmptyMessage = "No markdown files found."
	} else if v.state.Mode == ModeSearch && v.state.SearchQuery != "" {
		list.EmptyMessage = "No files match your search."
	}
	
	return list.View()
}

// renderSearchMode creates the search interface
func (v *ViewComposer) renderSearchMode() string {
	input, ok := v.inputs["search"]
	if !ok {
		return "Search input not initialized"
	}
	
	modal := InputModal{
		Title:    "",
		Prompt:   "Search:",
		Input:    input,
		HelpText: "[Enter] apply filter [Esc] cancel",
		Width:    v.state.Width * 70 / 100,
		Style:    v.state.Theme.Modal,
	}
	
	// Show filtered results below
	listView := v.renderFileList()
	
	return modal.View() + "\n\n" + listView
}

// renderCreateMode creates the note creation interface
func (v *ViewComposer) renderCreateMode() string {
	input, ok := v.inputs["create"]
	if !ok {
		return "Create input not initialized"
	}
	
	modal := InputModal{
		Title:    "Create New Note",
		Prompt:   "Title:",
		Input:    input,
		HelpText: "[Enter] create [Esc] cancel",
		Width:    v.state.Width * 70 / 100,
		Style:    v.state.Theme.Modal,
	}
	
	return modal.View()
}

// renderTagSearchMode creates the tag search interface
func (v *ViewComposer) renderTagSearchMode() string {
	input, ok := v.inputs["tag"]
	if !ok {
		return "Tag input not initialized"
	}
	
	modal := InputModal{
		Title:    "Search by Tag",
		Prompt:   "Tag:",
		Input:    input,
		HelpText: "[Enter] search [Esc] cancel",
		Width:    v.state.Width * 70 / 100,
		Style:    v.state.Theme.Modal,
	}
	
	return modal.View()
}

// renderTagCreateMode creates the tag input interface
func (v *ViewComposer) renderTagCreateMode() string {
	input, ok := v.inputs["tagcreate"]
	if !ok {
		return "Tag create input not initialized"
	}
	
	modal := InputModal{
		Title:    "Add Tags to New Note", 
		Prompt:   "Tags:",
		Input:    input,
		HelpText: "[Enter] create note [Esc] create without tags",
		Width:    v.state.Width * 70 / 100,
		Style:    v.state.Theme.Modal,
	}
	
	return modal.View()
}

// renderTaskCreateMode creates the task creation interface
func (v *ViewComposer) renderTaskCreateMode() string {
	input, ok := v.inputs["task"]
	if !ok {
		return "Task input not initialized"
	}
	
	modal := InputModal{
		Title:    "Create TaskWarrior Task",
		Prompt:   "Task:",
		Input:    input,
		HelpText: "[Enter] create task [Esc] cancel",
		Width:    v.state.Width * 70 / 100,
		Style:    v.state.Theme.Modal,
	}
	
	return modal.View()
}

// renderSortMode creates the sort selection interface
func (v *ViewComposer) renderSortMode() string {
	contentWidth, _ := v.state.Layout.ContentArea()
	
	var content string
	content += v.state.Theme.Modal.Title.Render("Sort Files") + "\n\n"
	content += "Choose sort method:\n"
	content += "[d] Date  [m] Modified  [t] Title  [i] Denote  [r] Reverse\n\n"
	content += v.state.Theme.Modal.Help.Render("[Esc] cancel")
	
	return lipgloss.NewStyle().Width(contentWidth).Render(content)
}

// renderOldFilterMode creates the days filter interface
func (v *ViewComposer) renderOldFilterMode() string {
	input, ok := v.inputs["old"]
	if !ok {
		return "Days input not initialized"
	}
	
	modal := InputModal{
		Title:    "Filter by Days Old",
		Prompt:   "Days back:",
		Input:    input,
		HelpText: "[Enter] apply filter [Esc] cancel",
		Width:    v.state.Width * 70 / 100,
		Style:    v.state.Theme.Modal,
	}
	
	return modal.View()
}

// renderDeleteMode creates the delete confirmation dialog
func (v *ViewComposer) renderDeleteMode() string {
	dialog := ConfirmDialog{
		Title:   "Delete Note",
		Message: fmt.Sprintf("Delete '%s'?", v.state.DeleteTarget),
		Options: []DialogOption{
			{Key: "y", Label: "yes"},
			{Key: "n", Label: "no"},
			{Key: "Esc", Label: "cancel"},
		},
		Width: v.state.Width * 60 / 100,
		Style: v.state.Theme.Dialog,
	}
	
	return dialog.View()
}

// renderPreview creates the preview popover
func (v *ViewComposer) renderPreview() string {
	popover := PreviewPopover{
		Title:     v.state.SelectedFile,
		Content:   v.state.PreviewContent,
		ScrollPos: v.state.PreviewScroll,
		Width:     v.state.Width * 80 / 100,
		Height:    v.state.Height * 80 / 100,
		Style:     v.state.Theme.Popover,
	}
	
	return v.state.Layout.CenterPopover(popover.View(), 80, 80)
}

// renderLoading creates a loading screen
func (v *ViewComposer) renderLoading() string {
	indicator := LoadingIndicator{
		Message: "Loading",
		Style:   v.state.Theme.Modal.Title,
	}
	
	contentWidth, contentHeight := v.state.Layout.ContentArea()
	centerStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Height(contentHeight).
		Align(lipgloss.Center, lipgloss.Center)
		
	return v.state.Layout.ApplyMargins(centerStyle.Render(indicator.View()))
}

// renderHelpBar creates the context-sensitive help bar
func (v *ViewComposer) renderHelpBar() string {
	contentWidth, _ := v.state.Layout.ContentArea()
	
	// Split help into two lines like the original
	// Line 1: Search, preview, and filters
	line1Items := []HelpItem{
		{Key: "/", Desc: "search"},
		{Key: "Enter", Desc: "preview"},
		{Key: "D", Desc: "all [D]aily"},
		{Key: "t", Desc: "open [t]asks"},
		{Key: "#", Desc: "tags"},
		{Key: "o", Desc: "s[o]rt"},
		{Key: "O", Desc: "days [O]ld"},
	}
	
	// Line 2: File operations
	line2Items := []HelpItem{
		{Key: "e", Desc: "[e]dit"},
		{Key: "n", Desc: "[n]ew note"},
		{Key: "d", Desc: "[d]aily note"},
	}
	
	// Add TaskWarrior if enabled
	if v.state.TaskwarriorSupport {
		line2Items = append(line2Items, HelpItem{Key: "Ctrl+K", Desc: "task"})
	}
	
	// Add remaining operations
	line2Items = append(line2Items,
		HelpItem{Key: "R", Desc: "Denote [R]ename"},
		HelpItem{Key: "X", Desc: "delete"},
		HelpItem{Key: "q", Desc: "[q]uit"},
	)
	
	// Build the two help bars
	help1 := HelpBar{
		Items: line1Items,
		Width: contentWidth,
		Style: v.state.Theme.Help,
	}
	
	help2 := HelpBar{
		Items: line2Items,
		Width: contentWidth,
		Style: v.state.Theme.Help,
	}
	
	// Join the two lines
	return help1.View() + "\n" + help2.View()
}

// Helper methods

func (v *ViewComposer) getActiveFilters() []string {
	var filters []string
	if v.state.TaskFilter {
		filters = append(filters, "Tasks")
	}
	if v.state.TagFilter {
		filters = append(filters, "Tag")
	}
	if v.state.TextFilter {
		filters = append(filters, "Search")
	}
	if v.state.DailyFilter {
		filters = append(filters, "Daily")
	}
	if v.state.OldFilter {
		filters = append(filters, fmt.Sprintf("Last %d days", v.state.OldDays))
	}
	return filters
}

func (v *ViewComposer) getSortInfo() string {
	if v.state.CurrentSort == "" {
		return ""
	}
	
	sortLabel := ""
	switch v.state.CurrentSort {
	case "date":
		sortLabel = "Date"
	case "modified":
		sortLabel = "Modified"
	case "title":
		sortLabel = "Title"
	case "denote":
		sortLabel = "Denote"
	}
	
	if v.state.ReversedSort {
		sortLabel += " (reversed)"
	}
	
	return sortLabel
}