package denote

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	// Denote filename pattern: YYYYMMDDTHHMMSS-title__tags.md or YYYYMMDDTHHMMSS--title__tags.md
	denotePattern = regexp.MustCompile(`^(\d{8}T\d{6})-{1,2}([^_]+)(?:__(.+))?\.md$`)
)

// ParseFilename extracts Denote components from a filename
func ParseFilename(filename string) (*Note, error) {
	base := filepath.Base(filename)
	matches := denotePattern.FindStringSubmatch(base)
	if len(matches) < 3 {
		return nil, fmt.Errorf("not a valid denote filename: %s", base)
	}

	note := &Note{
		ID:    matches[1],
		Title: titleFromSlug(matches[2]), // Convert slug to readable title as fallback
		Tags:  []string{},
	}

	// Parse tags if present
	if len(matches) > 3 && matches[3] != "" {
		note.Tags = strings.Split(matches[3], "_")
	}

	return note, nil
}

// ParseTaskFile reads and parses a task file
func ParseTaskFile(path string) (*Task, error) {
	// Parse filename first
	note, err := ParseFilename(path)
	if err != nil {
		return nil, err
	}

	// Check if it's a task file
	if !contains(note.Tags, "task") {
		return nil, fmt.Errorf("not a task file: %s", path)
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	task := &Task{
		Note:    *note,
		Path:    path,
		ModTime: info.ModTime(),
		Content: string(content),
	}

	// Parse frontmatter
	if meta, _, err := parseFrontmatter(content); err == nil {
		if taskMeta, ok := meta.(TaskMetadata); ok {
			task.TaskMetadata = taskMeta
		} else if m, ok := meta.(map[string]interface{}); ok {
			// Try to extract task metadata from generic map
			task.TaskMetadata = extractTaskMetadata(m)
		}
	}

	// Set default status if not specified
	if task.Status == "" {
		task.Status = TaskStatusOpen
	}
	
	// Use metadata title if available, otherwise fall back to filename title
	if task.TaskMetadata.Title != "" {
		task.Note.Title = task.TaskMetadata.Title
	}

	return task, nil
}

// ParseProjectFile reads and parses a project file
func ParseProjectFile(path string) (*Project, error) {
	// Parse filename first
	note, err := ParseFilename(path)
	if err != nil {
		return nil, err
	}

	// Check if it's a project file
	if !contains(note.Tags, "project") {
		return nil, fmt.Errorf("not a project file: %s", path)
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	project := &Project{
		Note:    *note,
		Path:    path,
		ModTime: info.ModTime(),
		Content: string(content),
	}

	// Parse frontmatter
	if meta, _, err := parseFrontmatter(content); err == nil {
		if projMeta, ok := meta.(ProjectMetadata); ok {
			project.ProjectMetadata = projMeta
		} else if m, ok := meta.(map[string]interface{}); ok {
			// Try to extract project metadata from generic map
			project.ProjectMetadata = extractProjectMetadata(m)
		}
	}

	// Set default status if not specified
	if project.Status == "" {
		project.Status = ProjectStatusActive
	}
	
	// Use metadata title if available, otherwise fall back to filename title
	if project.ProjectMetadata.Title != "" {
		project.Note.Title = project.ProjectMetadata.Title
	}

	return project, nil
}

// parseFrontmatter extracts YAML frontmatter from file content
func parseFrontmatter(content []byte) (interface{}, string, error) {
	contentStr := string(content)
	if !strings.HasPrefix(contentStr, "---\n") {
		return nil, contentStr, fmt.Errorf("no frontmatter found")
	}

	// Use robust frontmatter parsing
	lines := strings.Split(contentStr, "\n")
	frontmatterEndLine := -1
	inFrontmatter := false
	
	for i, line := range lines {
		if i == 0 && line == "---" {
			inFrontmatter = true
			continue
		}
		
		if inFrontmatter && line == "---" {
			// Check if this looks like the end of frontmatter
			possibleYAML := strings.Join(lines[1:i], "\n")
			if looksLikeValidFrontmatter(possibleYAML) {
				frontmatterEndLine = i
				break
			}
		}
	}
	
	if frontmatterEndLine == -1 {
		return nil, contentStr, fmt.Errorf("frontmatter not properly closed")
	}

	// Extract frontmatter YAML and remaining content
	frontmatterStr := strings.Join(lines[1:frontmatterEndLine], "\n")
	remaining := strings.Join(lines[frontmatterEndLine+1:], "\n")

	// Try to unmarshal as specific types first
	var taskMeta TaskMetadata
	if err := yaml.Unmarshal([]byte(frontmatterStr), &taskMeta); err == nil && taskMeta.TaskID > 0 {
		return taskMeta, remaining, nil
	}

	var projMeta ProjectMetadata
	if err := yaml.Unmarshal([]byte(frontmatterStr), &projMeta); err == nil && projMeta.ProjectID > 0 {
		return projMeta, remaining, nil
	}

	// Fall back to generic map
	var meta map[string]interface{}
	if err := yaml.Unmarshal([]byte(frontmatterStr), &meta); err != nil {
		return nil, contentStr, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return meta, remaining, nil
}

// extractTaskMetadata converts a generic map to TaskMetadata
func extractTaskMetadata(m map[string]interface{}) TaskMetadata {
	meta := TaskMetadata{}
	
	if v, ok := m["title"].(string); ok {
		meta.Title = v
	}
	if v, ok := m["task_id"].(int); ok {
		meta.TaskID = v
	}
	if v, ok := m["status"].(string); ok {
		meta.Status = v
	}
	if v, ok := m["priority"].(string); ok {
		meta.Priority = v
	}
	if v, ok := m["due_date"].(string); ok {
		meta.DueDate = v
	}
	if v, ok := m["start_date"].(string); ok {
		meta.StartDate = v
	}
	if v, ok := m["estimate"].(int); ok {
		meta.Estimate = v
	}
	if v, ok := m["project"].(string); ok {
		meta.Project = v
	}
	if v, ok := m["area"].(string); ok {
		meta.Area = v
	}
	if v, ok := m["assignee"].(string); ok {
		meta.Assignee = v
	}
	
	return meta
}

// extractProjectMetadata converts a generic map to ProjectMetadata
func extractProjectMetadata(m map[string]interface{}) ProjectMetadata {
	meta := ProjectMetadata{}
	
	if v, ok := m["title"].(string); ok {
		meta.Title = v
	}
	if v, ok := m["project_id"].(int); ok {
		meta.ProjectID = v
	}
	if v, ok := m["identifier"].(string); ok {
		meta.Identifier = v
	}
	if v, ok := m["status"].(string); ok {
		meta.Status = v
	}
	if v, ok := m["priority"].(string); ok {
		meta.Priority = v
	}
	if v, ok := m["due_date"].(string); ok {
		meta.DueDate = v
	}
	if v, ok := m["start_date"].(string); ok {
		meta.StartDate = v
	}
	if v, ok := m["area"].(string); ok {
		meta.Area = v
	}
	
	return meta
}

// titleFromSlug converts a kebab-case slug to a title
func titleFromSlug(slug string) string {
	// Simply replace hyphens with spaces, no capitalization
	return strings.ReplaceAll(slug, "-", " ")
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// looksLikeValidFrontmatter checks if content appears to be valid YAML frontmatter
// This is a simpler version than in update.go since we're just reading
func looksLikeValidFrontmatter(content string) bool {
	if content == "" {
		return true // Empty frontmatter is technically valid
	}
	
	// At least check that it has some YAML-like structure
	// Must contain at least one "key: value" pattern
	lines := strings.Split(content, "\n")
	hasValidLine := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		
		// Check for key: value pattern
		if colonIndex := strings.Index(trimmed, ":"); colonIndex > 0 {
			key := strings.TrimSpace(trimmed[:colonIndex])
			// Basic check that key looks valid
			if key != "" && !strings.ContainsAny(key, "{}[]|><") {
				hasValidLine = true
				break
			}
		}
		
		// Also accept list items
		if strings.HasPrefix(trimmed, "- ") {
			hasValidLine = true
			break
		}
	}
	
	return hasValidLine
}