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
	
	// Format title (truncate if needed)
	// Calculate available space more accurately
	// cursor(2) + status(10) + priority(7) + dates(24) + tasks(20) + spacing(6) = ~69
	maxTitleLen := v.Width - 70
	if maxTitleLen < 20 {
		maxTitleLen = 20 // Minimum title length
	}
	title := item.Title
	if len(title) > maxTitleLen {
		title = title[:maxTitleLen-3] + "..."
	}
	
	// Format status
	statusStyle := v.Theme.Status[""]
	if style, ok := v.Theme.Status[item.Status]; ok {
		statusStyle = style
	}
	status := statusStyle.Render(fmt.Sprintf("[%s]", formatStatus(item.Status)))
	
	// Format priority
	priorityStyle := v.Theme.Priority[item.Priority]
	priority := priorityStyle.Render(formatPriority(item.Priority))
	
	// Format dates
	dates := ""
	if item.StartDate != nil || item.DueDate != nil {
		var parts []string
		if item.StartDate != nil {
			parts = append(parts, item.StartDate.Format("2006-01-02"))
		} else {
			parts = append(parts, "          ")
		}
		parts = append(parts, "→")
		if item.DueDate != nil {
			parts = append(parts, item.DueDate.Format("2006-01-02"))
		} else {
			parts = append(parts, "          ")
		}
		dates = v.Theme.Dates.Render(strings.Join(parts, " "))
	}
	
	// Format task count
	taskCount := v.Theme.TaskCount.Render(
		fmt.Sprintf("(%d open, %d done)", item.OpenTasks, item.DoneTasks),
	)
	
	// Build the line without excessive padding
	line := fmt.Sprintf("%s%s  %s %s %s %s",
		cursor,
		title,
		status,
		priority,
		dates,
		taskCount,
	)
	
	// Apply selection styling
	if selected && v.ShowCursor {
		return v.Theme.Cursor.Render(line)
	}
	
	return line
}

// formatStatus formats the status for display
func formatStatus(status string) string {
	if status == "" {
		return "Unknown"
	}
	// Ensure consistent width
	return fmt.Sprintf("%-8s", strings.Title(status))
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