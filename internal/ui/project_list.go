package ui

import (
	"fmt"
	"strings"
	"time"
	
	"github.com/charmbracelet/lipgloss"
	"github.com/pdxmph/notes-tui/internal/denote"
)

// ProjectItem represents a project with display metadata
type ProjectItem struct {
	Project      *denote.Project
	Title        string
	Status       string
	Priority     int
	StartDate    *time.Time
	DueDate      *time.Time
	OpenTasks    int
	DoneTasks    int
}

// ProjectListView renders a list of projects with metadata
type ProjectListView struct {
	Items        []ProjectItem
	Cursor       int
	Width        int
	Height       int
	ShowCursor   bool
	EmptyMessage string
	Theme        ProjectTheme
}

// ProjectTheme defines colors for project list
type ProjectTheme struct {
	Cursor       lipgloss.Style
	Title        lipgloss.Style
	Status       map[string]lipgloss.Style
	Priority     map[int]lipgloss.Style
	Dates        lipgloss.Style
	TaskCount    lipgloss.Style
	EmptyMessage lipgloss.Style
}

// DefaultProjectTheme returns the default project theme
func DefaultProjectTheme() ProjectTheme {
	return ProjectTheme{
		Cursor: lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("240")),
		Title: lipgloss.NewStyle().
			Bold(true),
		Status: map[string]lipgloss.Style{
			"active":  lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
			"on-hold": lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
			"done":    lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
			"":        lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
		},
		Priority: map[int]lipgloss.Style{
			1: lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true), // High - red
			2: lipgloss.NewStyle().Foreground(lipgloss.Color("3")),            // Medium - yellow
			3: lipgloss.NewStyle().Foreground(lipgloss.Color("2")),            // Low - green
			0: lipgloss.NewStyle().Foreground(lipgloss.Color("8")),            // None - gray
		},
		Dates: lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")),
		TaskCount: lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")),
		EmptyMessage: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true),
	}
}

// View renders the project list
func (v ProjectListView) View() string {
	if len(v.Items) == 0 {
		return v.Theme.EmptyMessage.
			Width(v.Width).
			Height(v.Height).
			AlignHorizontal(lipgloss.Center).
			AlignVertical(lipgloss.Center).
			Render(v.EmptyMessage)
	}
	
	var lines []string
	
	// Calculate visible items
	start := 0
	if v.Cursor >= v.Height {
		start = v.Cursor - v.Height + 1
	}
	end := start + v.Height
	if end > len(v.Items) {
		end = len(v.Items)
	}
	
	// Render each visible project
	for i := start; i < end; i++ {
		line := v.renderProjectLine(v.Items[i], i == v.Cursor)
		lines = append(lines, line)
	}
	
	// Fill remaining space
	for len(lines) < v.Height {
		lines = append(lines, "")
	}
	
	return strings.Join(lines, "\n")
}

// renderProjectLine renders a single project line
func (v ProjectListView) renderProjectLine(item ProjectItem, selected bool) string {
	// Start with cursor or padding
	cursor := "  "
	if v.ShowCursor && selected {
		cursor = "▸ "
	}
	
	// Fixed column widths for better alignment
	const (
		titleWidth    = 25  // Project name column
		statusWidth   = 12  // Status column  
		priorityWidth = 7   // Priority column
	)
	
	// Format title with fixed width
	title := item.Title
	if len(title) > titleWidth {
		title = title[:titleWidth-3] + "..."
	}
	// Pad to fixed width
	title = fmt.Sprintf("%-*s", titleWidth, title)
	
	// Format status with fixed width
	statusStyle := v.Theme.Status[""]
	if style, ok := v.Theme.Status[item.Status]; ok {
		statusStyle = style
	}
	statusText := formatStatus(item.Status)
	// Pad the text before styling to ensure consistent width
	paddedStatus := fmt.Sprintf("%-*s", statusWidth, fmt.Sprintf("[%s]", statusText))
	status := statusStyle.Render(paddedStatus)
	
	// Format priority with fixed width
	priorityStyle := v.Theme.Priority[item.Priority]
	priorityText := formatPriority(item.Priority)
	// Pad before styling
	paddedPriority := fmt.Sprintf("%-*s", priorityWidth, priorityText)
	priority := priorityStyle.Render(paddedPriority)
	
	// Format dates more compactly
	var dateStr string
	if item.DueDate != nil {
		// Show just due date when present
		dateStr = fmt.Sprintf("→ %s", item.DueDate.Format("2006-01-02"))
	} else if item.StartDate != nil {
		// Show start date with arrow if no due date
		dateStr = fmt.Sprintf("%s →", item.StartDate.Format("2006-01-02"))
	} else {
		// No dates
		dateStr = ""
	}
	dates := v.Theme.Dates.Render(dateStr)
	
	// Format task count with fixed width
	const taskCountWidth = 20 // "(99 open, 99 done)" should be enough
	taskCountText := fmt.Sprintf("(%d open, %d done)", item.OpenTasks, item.DoneTasks)
	taskCountPadded := fmt.Sprintf("%-*s", taskCountWidth, taskCountText)
	taskCount := v.Theme.TaskCount.Render(taskCountPadded)
	
	// Build the line with tighter spacing
	line := fmt.Sprintf("%s%s %s %s %s  %s",
		cursor,
		title,
		status,
		priority,
		taskCount,
		dates,
	)
	
	// Apply selection styling
	if selected && v.ShowCursor {
		return v.Theme.Cursor.Render(line)
	}
	
	return line
}

// formatStatus formats the status for display
func formatStatus(status string) string {
	switch status {
	case "active":
		return "Active"
	case "on-hold":
		return "On Hold"
	case "done", "completed":
		return "Done"
	case "paused":
		return "Paused"
	case "cancelled":
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// formatPriority formats the priority for display
func formatPriority(priority int) string {
	switch priority {
	case 1:
		return "P:High"
	case 2:
		return "P:Med "
	case 3:
		return "P:Low "
	default:
		return "P:--- "
	}
}