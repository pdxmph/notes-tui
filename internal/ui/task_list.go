package ui

import (
	"fmt"
	"strconv"
	"strings"
	
	"github.com/charmbracelet/lipgloss"
)

// TaskListView is a specialized list view for tasks with proper color support
type TaskListView struct {
	Items         []TaskItem
	Cursor        int
	Width         int
	Height        int
	ShowCursor    bool
	EmptyMessage  string
	Theme         TaskTheme
}

// TaskItem represents a parsed task for display
type TaskItem struct {
	Status    string
	Priority  string
	Title     string
	Project   string
	Area      string
	Estimate  int
	DueDate   string
	IsOverdue bool
	DueDays   int
	Raw       string // Original item string
}

// TaskTheme defines colors for task display
type TaskTheme struct {
	// Status colors
	StatusOpen      lipgloss.Style
	StatusDone      lipgloss.Style
	StatusPaused    lipgloss.Style
	StatusDelegated lipgloss.Style
	StatusDropped   lipgloss.Style
	
	// Priority colors
	PriorityP1 lipgloss.Style
	PriorityP2 lipgloss.Style
	PriorityP3 lipgloss.Style
	
	// Metadata colors
	Project   lipgloss.Style
	Area      lipgloss.Style
	Estimate  lipgloss.Style
	DueNormal lipgloss.Style
	DueOverdue lipgloss.Style
	
	// List colors
	Cursor    lipgloss.Style
	Normal    lipgloss.Style
	EmptyMsg  lipgloss.Style
}

// DefaultTaskTheme returns the default color theme for tasks
func DefaultTaskTheme() TaskTheme {
	return TaskTheme{
		// Status colors
		StatusOpen:      lipgloss.NewStyle().Foreground(lipgloss.Color("6")),    // Cyan
		StatusDone:      lipgloss.NewStyle().Foreground(lipgloss.Color("2")),    // Green
		StatusPaused:    lipgloss.NewStyle().Foreground(lipgloss.Color("3")),    // Yellow
		StatusDelegated: lipgloss.NewStyle().Foreground(lipgloss.Color("4")),    // Blue
		StatusDropped:   lipgloss.NewStyle().Foreground(lipgloss.Color("240")),  // Gray
		
		// Priority colors
		PriorityP1: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9")), // Bright Red
		PriorityP2: lipgloss.NewStyle().Foreground(lipgloss.Color("3")),            // Yellow
		PriorityP3: lipgloss.NewStyle().Foreground(lipgloss.Color("4")),            // Blue
		
		// Metadata colors
		Project:    lipgloss.NewStyle().Foreground(lipgloss.Color("4")),   // Blue
		Area:       lipgloss.NewStyle().Foreground(lipgloss.Color("5")),   // Magenta
		Estimate:   lipgloss.NewStyle().Foreground(lipgloss.Color("240")), // Gray
		DueNormal:  lipgloss.NewStyle().Foreground(lipgloss.Color("3")),   // Yellow
		DueOverdue: lipgloss.NewStyle().Foreground(lipgloss.Color("9")),   // Bright Red
		
		// List colors
		Cursor:   lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true),
		Normal:   lipgloss.NewStyle(),
		EmptyMsg: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
	}
}

// ParseTaskItem parses a formatted task string into components
func ParseTaskItem(formatted string) TaskItem {
	item := TaskItem{Raw: formatted}
	
	// Parse the formatted string to extract components
	// Expected format: "status priority title @project #area ~estimate (due info)"
	
	parts := strings.Fields(formatted)
	if len(parts) == 0 {
		return item
	}
	
	idx := 0
	
	// Parse status indicator
	if idx < len(parts) {
		switch parts[idx] {
		case "○", "✓", "⏸", "→", "✗":
			item.Status = parts[idx]
			idx++
		}
	}
	
	// Parse priority
	if idx < len(parts) && strings.HasPrefix(parts[idx], "[P") && strings.HasSuffix(parts[idx], "]") {
		item.Priority = parts[idx]
		idx++
	}
	
	// Skip task ID if present (e.g., "#5")
	if idx < len(parts) && strings.HasPrefix(parts[idx], "#") && len(parts[idx]) > 1 {
		// Check if it's a task ID (all digits after #)
		idStr := strings.TrimPrefix(parts[idx], "#")
		if _, err := strconv.Atoi(idStr); err == nil {
			// It's a task ID, skip it
			idx++
		}
	}
	
	// Collect title until we hit metadata
	titleParts := []string{}
	for idx < len(parts) {
		part := parts[idx]
		if strings.HasPrefix(part, "@") || strings.HasPrefix(part, "#") || 
		   strings.HasPrefix(part, "~") || strings.HasPrefix(part, "(") {
			break
		}
		titleParts = append(titleParts, part)
		idx++
	}
	item.Title = strings.Join(titleParts, " ")
	
	// Parse remaining metadata
	for idx < len(parts) {
		part := parts[idx]
		
		if strings.HasPrefix(part, "@") {
			item.Project = strings.TrimPrefix(part, "@")
		} else if strings.HasPrefix(part, "#") {
			item.Area = strings.TrimPrefix(part, "#")
		} else if strings.HasPrefix(part, "~") {
			if est, err := fmt.Sscanf(part, "~%d", &item.Estimate); err == nil {
				item.Estimate = est
			}
		} else if strings.HasPrefix(part, "(") {
			// Parse due date info
			dueText := part
			// Collect all parts of due date (might be multiple words)
			for idx+1 < len(parts) {
				idx++
				dueText += " " + parts[idx]
				if strings.HasSuffix(parts[idx], ")") {
					break
				}
			}
			item.DueDate = dueText
			
			// Check if overdue
			if strings.Contains(dueText, "OVERDUE") || strings.Contains(dueText, "overdue") {
				item.IsOverdue = true
			}
		}
		idx++
	}
	
	return item
}

// View renders the task list with proper color handling
func (v TaskListView) View() string {
	if len(v.Items) == 0 {
		return v.Theme.EmptyMsg.Render(v.EmptyMessage)
	}
	
	var content strings.Builder
	maxVisible := v.Height
	if maxVisible <= 0 {
		maxVisible = 1
	}
	
	// Simple approach for small heights
	if maxVisible < 3 && len(v.Items) > maxVisible {
		startIdx := v.Cursor
		if startIdx > len(v.Items) - maxVisible {
			startIdx = len(v.Items) - maxVisible
		}
		if startIdx < 0 {
			startIdx = 0
		}
		
		for i := startIdx; i < len(v.Items) && i < startIdx+maxVisible; i++ {
			item := v.Items[i]
			line := v.renderTaskLine(item, i == v.Cursor && v.ShowCursor)
			content.WriteString(line)
			if i < startIdx+maxVisible-1 {
				content.WriteString("\n")
			}
		}
		return content.String()
	}
	
	// Normal case with indicators
	startIdx := 0
	endIdx := len(v.Items)
	
	// If list is scrollable, calculate viewport
	if len(v.Items) > maxVisible {
		// Keep cursor in the middle third of the view when possible
		preferredStart := v.Cursor - maxVisible/3
		
		// Adjust bounds
		if preferredStart < 0 {
			startIdx = 0
		} else if preferredStart + maxVisible > len(v.Items) {
			startIdx = len(v.Items) - maxVisible
		} else {
			startIdx = preferredStart
		}
		
		endIdx = startIdx + maxVisible
	}
	
	// Show top indicator if needed
	if startIdx > 0 {
		indicator := fmt.Sprintf("... %d items above", startIdx)
		content.WriteString(v.Theme.EmptyMsg.Render(indicator))
		content.WriteString("\n")
		startIdx++ // Skip one item to make room for indicator
	}
	
	// Show bottom indicator if needed
	showBottomIndicator := endIdx < len(v.Items)
	if showBottomIndicator {
		endIdx-- // Reserve space for bottom indicator
	}
	
	for i := startIdx; i < endIdx && i < len(v.Items); i++ {
		item := v.Items[i]
		
		// Build the line with proper styling
		line := v.renderTaskLine(item, i == v.Cursor && v.ShowCursor)
		
		content.WriteString(line)
		if i < endIdx-1 {
			content.WriteString("\n")
		}
	}
	
	// Add bottom indicator if needed
	if showBottomIndicator {
		remaining := len(v.Items) - endIdx
		indicator := fmt.Sprintf("\n... %d more items", remaining)
		content.WriteString(v.Theme.EmptyMsg.Render(indicator))
	}
	
	return content.String()
}

// renderTaskLine renders a single task line with colors
func (v TaskListView) renderTaskLine(item TaskItem, isSelected bool) string {
	var parts []string
	
	// Add cursor
	cursor := "  "
	if v.ShowCursor && isSelected {
		cursor = "> "
	}
	
	// Status with color
	if item.Status != "" {
		var statusStyle lipgloss.Style
		switch item.Status {
		case "✓":
			statusStyle = v.Theme.StatusDone
		case "⏸":
			statusStyle = v.Theme.StatusPaused
		case "→":
			statusStyle = v.Theme.StatusDelegated
		case "✗":
			statusStyle = v.Theme.StatusDropped
		default:
			statusStyle = v.Theme.StatusOpen
		}
		parts = append(parts, statusStyle.Render(item.Status))
	}
	
	// Priority with color
	if item.Priority != "" {
		var priorityStyle lipgloss.Style
		switch item.Priority {
		case "[P1]":
			priorityStyle = v.Theme.PriorityP1
		case "[P2]":
			priorityStyle = v.Theme.PriorityP2
		case "[P3]":
			priorityStyle = v.Theme.PriorityP3
		default:
			priorityStyle = v.Theme.Normal
		}
		parts = append(parts, priorityStyle.Render(item.Priority))
	}
	
	// Title (no special color)
	if item.Title != "" {
		parts = append(parts, item.Title)
	}
	
	// Project with color
	if item.Project != "" {
		parts = append(parts, v.Theme.Project.Render("@"+item.Project))
	}
	
	// Area with color
	if item.Area != "" {
		parts = append(parts, v.Theme.Area.Render("#"+item.Area))
	}
	
	// Estimate with color
	if item.Estimate > 0 {
		parts = append(parts, v.Theme.Estimate.Render(fmt.Sprintf("~%d", item.Estimate)))
	}
	
	// Due date with color
	if item.DueDate != "" {
		if item.IsOverdue {
			parts = append(parts, v.Theme.DueOverdue.Render(item.DueDate))
		} else {
			parts = append(parts, v.Theme.DueNormal.Render(item.DueDate))
		}
	}
	
	// Build the complete line
	line := cursor + strings.Join(parts, " ")
	
	// Apply selection styling if needed
	if isSelected && v.ShowCursor {
		// For selected lines, we apply a subtle background or bold
		// without overriding the individual component colors
		return lipgloss.NewStyle().Bold(true).Render(line)
	}
	
	return line
}