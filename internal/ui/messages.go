package ui

import (
	"time"
	tea "github.com/charmbracelet/bubbletea"
)

// StatusMsg is sent to display temporary status messages
type StatusMsg struct {
	Message  string
	Type     StatusType
	Duration time.Duration
}

// ClearStatusMsg clears the current status message
type ClearStatusMsg struct{}

// ShowStatus creates a command to show a status message
func ShowStatus(message string, msgType StatusType) tea.Cmd {
	return func() tea.Msg {
		return StatusMsg{
			Message:  message,
			Type:     msgType,
			Duration: 3 * time.Second,
		}
	}
}

// ShowError shows an error status message
func ShowError(message string) tea.Cmd {
	return ShowStatus(message, StatusError)
}

// ShowSuccess shows a success status message
func ShowSuccess(message string) tea.Cmd {
	return ShowStatus(message, StatusSuccess)
}

// ShowInfo shows an info status message
func ShowInfo(message string) tea.Cmd {
	return ShowStatus(message, StatusInfo)
}

// ShowWarning shows a warning status message
func ShowWarning(message string) tea.Cmd {
	return ShowStatus(message, StatusWarning)
}

// ClearStatusAfter clears status after duration
func ClearStatusAfter(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(time.Time) tea.Msg {
		return ClearStatusMsg{}
	})
}

// Common status messages
var (
	MsgFileCreated   = "Note created successfully"
	MsgFileDeleted   = "Note deleted"
	MsgFileRenamed   = "File renamed to Denote format"
	MsgTaskCreated   = "Task created successfully"
	MsgNoEditor      = "No editor configured"
	MsgNoFiles       = "No files found"
	MsgSearchApplied = "Search filter applied"
	MsgFilterCleared = "Filter cleared"
	MsgSortApplied   = "Sort applied"
)

// Error messages
var (
	ErrFileCreate   = "Failed to create note"
	ErrFileDelete   = "Failed to delete file"
	ErrFileRename   = "Failed to rename file"
	ErrFileRead     = "Failed to read file"
	ErrTaskCreate   = "Failed to create task"
	ErrInvalidInput = "Invalid input"
	ErrNoRipgrep    = "ripgrep not found"
)