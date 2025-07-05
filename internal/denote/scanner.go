package denote

import (
	"fmt"
	"path/filepath"
	"sort"
)

// Scanner finds and loads Denote files
type Scanner struct {
	BaseDir string
}

// NewScanner creates a new scanner for the given directory
func NewScanner(dir string) *Scanner {
	return &Scanner{BaseDir: dir}
}

// FindTasks finds all task files in the directory
func (s *Scanner) FindTasks() ([]*Task, error) {
	pattern := filepath.Join(s.BaseDir, "*__task*.md")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob task files: %w", err)
	}

	var tasks []*Task
	for _, file := range files {
		task, err := ParseTaskFile(file)
		if err != nil {
			// Skip files that fail to parse
			continue
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// FindProjects finds all project files in the directory
func (s *Scanner) FindProjects() ([]*Project, error) {
	pattern := filepath.Join(s.BaseDir, "*__project*.md")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob project files: %w", err)
	}

	var projects []*Project
	for _, file := range files {
		project, err := ParseProjectFile(file)
		if err != nil {
			// Skip files that fail to parse
			continue
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// SortTasks sorts tasks by various criteria
func SortTasks(tasks []*Task, sortBy string, reverse bool) {
	switch sortBy {
	case "priority":
		sort.Slice(tasks, func(i, j int) bool {
			// P1 < P2 < P3 < no priority
			pi := priorityValue(tasks[i].Priority)
			pj := priorityValue(tasks[j].Priority)
			if pi != pj {
				return pi < pj
			}
			// Secondary sort by due date
			return tasks[i].DueDate < tasks[j].DueDate
		})
	
	case "due":
		sort.Slice(tasks, func(i, j int) bool {
			// Tasks with due dates come before those without
			if tasks[i].DueDate == "" && tasks[j].DueDate != "" {
				return false
			}
			if tasks[i].DueDate != "" && tasks[j].DueDate == "" {
				return true
			}
			return tasks[i].DueDate < tasks[j].DueDate
		})
	
	case "status":
		sort.Slice(tasks, func(i, j int) bool {
			// Open < Paused < Delegated < Done < Dropped
			si := statusValue(tasks[i].Status)
			sj := statusValue(tasks[j].Status)
			if si != sj {
				return si < sj
			}
			// Secondary sort by priority
			return priorityValue(tasks[i].Priority) < priorityValue(tasks[j].Priority)
		})
	
	case "id":
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].TaskID < tasks[j].TaskID
		})
	
	case "created":
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].ID < tasks[j].ID
		})
	
	case "modified":
		fallthrough
	default:
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].ModTime.After(tasks[j].ModTime)
		})
	}

	if reverse {
		reverseTaskSlice(tasks)
	}
}

// Helper functions for sorting

func priorityValue(p string) int {
	switch p {
	case PriorityP1:
		return 1
	case PriorityP2:
		return 2
	case PriorityP3:
		return 3
	default:
		return 4
	}
}

func statusValue(s string) int {
	switch s {
	case TaskStatusOpen:
		return 1
	case TaskStatusPaused:
		return 2
	case TaskStatusDelegated:
		return 3
	case TaskStatusDone:
		return 4
	case TaskStatusDropped:
		return 5
	default:
		return 6
	}
}

func reverseTaskSlice(tasks []*Task) {
	for i, j := 0, len(tasks)-1; i < j; i, j = i+1, j-1 {
		tasks[i], tasks[j] = tasks[j], tasks[i]
	}
}

// FilterTasks filters tasks based on various criteria
func FilterTasks(tasks []*Task, filterType string, filterValue string) []*Task {
	var filtered []*Task
	
	switch filterType {
	case "all":
		return tasks
		
	case "open":
		for _, task := range tasks {
			if task.Status == TaskStatusOpen {
				filtered = append(filtered, task)
			}
		}
		
	case "done":
		for _, task := range tasks {
			if task.Status == TaskStatusDone {
				filtered = append(filtered, task)
			}
		}
		
	case "active":
		// Open, paused, or delegated tasks
		for _, task := range tasks {
			if task.Status == TaskStatusOpen || 
			   task.Status == TaskStatusPaused || 
			   task.Status == TaskStatusDelegated {
				filtered = append(filtered, task)
			}
		}
		
	case "area":
		// Filter by specific area
		for _, task := range tasks {
			if task.Area == filterValue {
				filtered = append(filtered, task)
			}
		}
		
	case "project":
		// Filter by specific project
		for _, task := range tasks {
			if task.Project == filterValue {
				filtered = append(filtered, task)
			}
		}
		
	case "overdue":
		// Tasks with due dates in the past
		for _, task := range tasks {
			if task.DueDate != "" && IsOverdue(task.DueDate) && task.Status != TaskStatusDone {
				filtered = append(filtered, task)
			}
		}
		
	case "today":
		// Tasks due today
		today := formatDate(getCurrentTime())
		for _, task := range tasks {
			if task.DueDate == today && task.Status != TaskStatusDone {
				filtered = append(filtered, task)
			}
		}
		
	case "week":
		// Tasks due this week
		for _, task := range tasks {
			if task.DueDate != "" && IsDueThisWeek(task.DueDate) && task.Status != TaskStatusDone {
				filtered = append(filtered, task)
			}
		}
		
	case "priority":
		// Filter by specific priority
		for _, task := range tasks {
			if task.Priority == filterValue {
				filtered = append(filtered, task)
			}
		}
	}
	
	return filtered
}

// GetUniqueAreas returns all unique areas from tasks
func GetUniqueAreas(tasks []*Task) []string {
	areaMap := make(map[string]bool)
	for _, task := range tasks {
		if task.Area != "" {
			areaMap[task.Area] = true
		}
	}
	
	var areas []string
	for area := range areaMap {
		areas = append(areas, area)
	}
	sort.Strings(areas)
	return areas
}

// GetUniqueProjects returns all unique project names from tasks
func GetUniqueProjects(tasks []*Task) []string {
	projectMap := make(map[string]bool)
	for _, task := range tasks {
		if task.Project != "" {
			projectMap[task.Project] = true
		}
	}
	
	var projects []string
	for project := range projectMap {
		projects = append(projects, project)
	}
	sort.Strings(projects)
	return projects
}