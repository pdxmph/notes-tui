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
	// Parse date in local timezone to avoid timezone issues
	loc := time.Now().Location()
	dueDate, err := time.ParseInLocation("2006-01-02", dueDateStr, loc)
	if err != nil {
		return false
	}
	// Get current time at start of day in local timezone
	now := time.Now().In(loc)
	nowStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	dueStart := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, loc)
	
	return dueStart.Before(nowStart)
}

// DaysUntilDue returns the number of days until the due date
func DaysUntilDue(dueDateStr string) int {
	if dueDateStr == "" {
		return 0
	}
	// Parse date in local timezone to avoid timezone issues
	loc := time.Now().Location()
	dueDate, err := time.ParseInLocation("2006-01-02", dueDateStr, loc)
	if err != nil {
		return 0
	}
	// Get current time at start of day in local timezone
	now := time.Now().In(loc)
	nowStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	dueStart := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, loc)
	
	return int(dueStart.Sub(nowStart).Hours() / 24)
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

// GetParsedStartDate returns the parsed start date for a project
func (p *Project) GetParsedStartDate() *time.Time {
	if p.StartDate == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", p.StartDate)
	if err != nil {
		return nil
	}
	return &t
}

// GetParsedDueDate returns the parsed due date for a project
func (p *Project) GetParsedDueDate() *time.Time {
	if p.DueDate == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", p.DueDate)
	if err != nil {
		return nil
	}
	return &t
}

// GetPriorityInt returns the priority as an integer (1=high, 2=med, 3=low)
func (p *Project) GetPriorityInt() int {
	switch p.Priority {
	case "p1", "1":
		return 1
	case "p2", "2":
		return 2
	case "p3", "3":
		return 3
	default:
		return 0
	}
}