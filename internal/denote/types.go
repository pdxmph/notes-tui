package denote

import (
	"time"
)

// Note represents the basic Denote file structure
type Note struct {
	ID    string   // Denote timestamp ID (e.g., "20250704T151739")
	Title string   // Human-readable title
	Tags  []string // Tags from filename
}

// TaskMetadata represents task-specific frontmatter
type TaskMetadata struct {
	Title     string `yaml:"title"`
	TaskID    int    `yaml:"task_id"`
	Status    string `yaml:"status"`
	Priority  string `yaml:"priority"`
	DueDate   string `yaml:"due_date"`
	StartDate string `yaml:"start_date"`
	Estimate  int    `yaml:"estimate"`
	Project   string `yaml:"project"`
	Area      string `yaml:"area"`
	Assignee  string `yaml:"assignee"`
}

// ProjectMetadata represents project-specific frontmatter
type ProjectMetadata struct {
	Title      string `yaml:"title"`
	ProjectID  int    `yaml:"project_id"`
	Identifier string `yaml:"identifier"` // The actual project key used in tasks
	Status     string `yaml:"status"`
	Priority   string `yaml:"priority"`
	DueDate    string `yaml:"due_date"`
	StartDate  string `yaml:"start_date"`
	Area       string `yaml:"area"`
}

// Task combines Note info with TaskMetadata
type Task struct {
	Note
	TaskMetadata
	Path     string
	ModTime  time.Time
	Content  string // Full file content
}

// Project combines Note info with ProjectMetadata
type Project struct {
	Note
	ProjectMetadata
	Path     string
	ModTime  time.Time
	Content  string
}

// Common status values
const (
	// Task statuses
	TaskStatusOpen      = "open"
	TaskStatusDone      = "done"
	TaskStatusPaused    = "paused"
	TaskStatusDelegated = "delegated"
	TaskStatusDropped   = "dropped"

	// Project statuses
	ProjectStatusActive    = "active"
	ProjectStatusCompleted = "completed"
	ProjectStatusPaused    = "paused"
	ProjectStatusCancelled = "cancelled"

	// Priority levels
	PriorityP1 = "p1"
	PriorityP2 = "p2"
	PriorityP3 = "p3"
)

// IsValidTaskStatus checks if a status is valid for tasks
func IsValidTaskStatus(status string) bool {
	switch status {
	case TaskStatusOpen, TaskStatusDone, TaskStatusPaused, TaskStatusDelegated, TaskStatusDropped:
		return true
	}
	return false
}

// IsValidProjectStatus checks if a status is valid for projects
func IsValidProjectStatus(status string) bool {
	switch status {
	case ProjectStatusActive, ProjectStatusCompleted, ProjectStatusPaused, ProjectStatusCancelled:
		return true
	}
	return false
}

// IsValidPriority checks if a priority is valid
func IsValidPriority(priority string) bool {
	switch priority {
	case PriorityP1, PriorityP2, PriorityP3:
		return true
	}
	return false
}

// IsOverdue checks if a task/project is overdue
func IsOverdue(dueDateStr string) bool {
	if dueDateStr == "" {
		return false
	}
	dueDate, err := time.Parse("2006-01-02", dueDateStr)
	if err != nil {
		return false
	}
	now := time.Now().Truncate(24 * time.Hour)
	return dueDate.Before(now)
}

// DaysUntilDue returns the number of days until the due date
func DaysUntilDue(dueDateStr string) int {
	if dueDateStr == "" {
		return 0
	}
	dueDate, err := time.Parse("2006-01-02", dueDateStr)
	if err != nil {
		return 0
	}
	now := time.Now().Truncate(24 * time.Hour)
	return int(dueDate.Sub(now).Hours() / 24)
}

// IsDueThisWeek checks if a task is due within the next 7 days
func IsDueThisWeek(dueDateStr string) bool {
	days := DaysUntilDue(dueDateStr)
	return days >= 0 && days <= 7
}

// formatDate formats a time as YYYY-MM-DD
func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// getCurrentTime returns the current time (useful for testing)
func getCurrentTime() time.Time {
	return time.Now()
}