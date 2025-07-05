package denote

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// UpdateTaskStatus updates the status field in a task file's frontmatter
func UpdateTaskStatus(filepath string, newStatus string) error {
	// Validate status
	if !IsValidTaskStatus(newStatus) {
		return fmt.Errorf("invalid status: %s", newStatus)
	}
	
	// Read file
	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Update status in frontmatter
	updated := updateFrontmatterField(string(content), "status", newStatus)
	
	// Write back
	if err := os.WriteFile(filepath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

// UpdateTaskPriority updates the priority field in a task file's frontmatter
func UpdateTaskPriority(filepath string, newPriority string) error {
	// Validate priority
	if newPriority != "" && !IsValidPriority(newPriority) {
		return fmt.Errorf("invalid priority: %s", newPriority)
	}
	
	// Read file
	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Update or add priority in frontmatter
	updated := updateFrontmatterField(string(content), "priority", newPriority)
	
	// Write back
	if err := os.WriteFile(filepath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

// updateFrontmatterField updates or adds a field in YAML frontmatter
func updateFrontmatterField(content, field, value string) string {
	// Check if file has frontmatter
	if !strings.HasPrefix(content, "---\n") {
		// Add frontmatter
		return fmt.Sprintf("---\n%s: %s\n---\n\n%s", field, value, content)
	}
	
	// Find the end of YAML frontmatter more robustly
	// Look for a line that contains only "---" and is preceded by valid YAML
	lines := strings.Split(content, "\n")
	frontmatterEndLine := -1
	inFrontmatter := false
	
	for i, line := range lines {
		if i == 0 && line == "---" {
			inFrontmatter = true
			continue
		}
		
		if inFrontmatter && line == "---" {
			// Check if this looks like the end of frontmatter by verifying
			// the content between start and this line looks like YAML
			possibleYAML := strings.Join(lines[1:i], "\n")
			if looksLikeYAML(possibleYAML) {
				frontmatterEndLine = i
				break
			}
		}
	}
	
	if frontmatterEndLine == -1 {
		// Malformed frontmatter
		return content
	}
	
	// Extract frontmatter and rest
	frontmatter := strings.Join(lines[1:frontmatterEndLine], "\n")
	rest := strings.Join(lines[frontmatterEndLine+1:], "\n")
	
	// Format the value with quotes if needed
	formattedValue := value
	if value != "" && needsQuotes(field, value) {
		formattedValue = fmt.Sprintf(`"%s"`, value)
	}
	
	// Try to update existing field
	fieldPattern := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(field) + `:\s*.*$`)
	if fieldPattern.MatchString(frontmatter) {
		// Update existing field
		newValue := field + ": " + formattedValue
		if value == "" {
			newValue = field + ":"
		}
		frontmatter = fieldPattern.ReplaceAllString(frontmatter, newValue)
	} else {
		// Add new field
		if !strings.HasSuffix(frontmatter, "\n") {
			frontmatter += "\n"
		}
		frontmatter += field + ": " + formattedValue + "\n"
	}
	
	// Reconstruct the document
	result := "---\n" + frontmatter + "\n---\n"
	if rest != "" && !strings.HasPrefix(rest, "\n") {
		result += "\n"
	}
	result += rest
	
	return result
}

// looksLikeYAML checks if content appears to be valid YAML frontmatter
func looksLikeYAML(content string) bool {
	if content == "" {
		return true // Empty frontmatter is valid
	}
	
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		
		// Check for basic YAML patterns
		// Must be either:
		// 1. key: value
		// 2. - list item  
		// 3. Indented continuation
		if strings.HasPrefix(trimmed, "-") && len(trimmed) > 1 && trimmed[1] == ' ' {
			// List item
			continue
		}
		
		if strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t") {
			// Indented line (continuation or nested)
			continue
		}
		
		// Check for key: value pattern
		colonIndex := strings.Index(line, ":")
		if colonIndex > 0 && colonIndex < len(line) {
			key := strings.TrimSpace(line[:colonIndex])
			// Key should be alphanumeric with underscores
			if isValidYAMLKey(key) {
				continue
			}
		}
		
		// This line doesn't match any YAML pattern
		return false
	}
	
	return true
}

// isValidYAMLKey checks if a string is a valid YAML key
func isValidYAMLKey(key string) bool {
	if key == "" {
		return false
	}
	
	// Check if key contains only valid characters
	for _, ch := range key {
		if !((ch >= 'a' && ch <= 'z') || 
			 (ch >= 'A' && ch <= 'Z') || 
			 (ch >= '0' && ch <= '9') || 
			 ch == '_' || ch == '-') {
			return false
		}
	}
	
	return true
}

// UpdateTaskDueDate updates the due_date field in a task file's frontmatter
func UpdateTaskDueDate(filepath string, dueDate string) error {
	// Read file
	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Update or add due_date in frontmatter
	updated := updateFrontmatterField(string(content), "due_date", dueDate)
	
	// Write back
	if err := os.WriteFile(filepath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

// UpdateTaskStartDate updates the start_date field in a task file's frontmatter
func UpdateTaskStartDate(filepath string, startDate string) error {
	// Read file
	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Update or add start_date in frontmatter
	updated := updateFrontmatterField(string(content), "start_date", startDate)
	
	// Write back
	if err := os.WriteFile(filepath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

// UpdateTaskEstimate updates the estimate field in a task file's frontmatter
func UpdateTaskEstimate(filepath string, estimate int) error {
	// Validate estimate
	if estimate != 0 && !IsValidEstimate(estimate) {
		return fmt.Errorf("invalid estimate: %d (must be 0, 1, 2, 3, 5, 8, or 13)", estimate)
	}
	
	// Read file
	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Update or add estimate in frontmatter
	value := ""
	if estimate > 0 {
		value = fmt.Sprintf("%d", estimate)
	}
	updated := updateFrontmatterField(string(content), "estimate", value)
	
	// Write back
	if err := os.WriteFile(filepath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

// UpdateTaskProject updates the project field in a task file's frontmatter
func UpdateTaskProject(filepath string, project string) error {
	// Read file
	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Update or add project in frontmatter
	updated := updateFrontmatterField(string(content), "project", project)
	
	// Write back
	if err := os.WriteFile(filepath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

// UpdateTaskArea updates the area field in a task file's frontmatter
func UpdateTaskArea(filepath string, area string) error {
	// Read file
	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Update or add area in frontmatter
	updated := updateFrontmatterField(string(content), "area", area)
	
	// Write back
	if err := os.WriteFile(filepath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

// UpdateTaskTags updates the tags field in a task file's frontmatter
func UpdateTaskTags(filepath string, tags []string) error {
	// Read file
	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Parse frontmatter to update tags as YAML array
	contentStr := string(content)
	if !strings.HasPrefix(contentStr, "---\n") {
		return fmt.Errorf("no frontmatter found")
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
			possibleYAML := strings.Join(lines[1:i], "\n")
			if looksLikeYAML(possibleYAML) {
				frontmatterEndLine = i
				break
			}
		}
	}
	
	if frontmatterEndLine == -1 {
		return fmt.Errorf("frontmatter not properly closed")
	}
	
	frontmatter := strings.Join(lines[1:frontmatterEndLine], "\n")
	rest := strings.Join(lines[frontmatterEndLine+1:], "\n")
	
	// Remove existing tags field
	tagPattern := regexp.MustCompile(`(?m)^tags:.*\n(?:  - .*\n)*`)
	frontmatter = tagPattern.ReplaceAllString(frontmatter, "")
	
	// Add new tags
	if len(tags) > 0 {
		tagsYAML := "tags:\n"
		for _, tag := range tags {
			tagsYAML += fmt.Sprintf("  - %s\n", tag)
		}
		frontmatter = strings.TrimRight(frontmatter, "\n") + "\n" + tagsYAML
	}
	
	// Write back
	updated := "---\n" + frontmatter + "---\n" + rest
	if err := os.WriteFile(filepath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

// IsValidEstimate checks if an estimate value is valid (Fibonacci)
func IsValidEstimate(estimate int) bool {
	validEstimates := []int{1, 2, 3, 5, 8, 13}
	for _, v := range validEstimates {
		if estimate == v {
			return true
		}
	}
	return false
}

// needsQuotes determines if a YAML value needs quotes
func needsQuotes(field, value string) bool {
	// String fields that should be quoted
	stringFields := map[string]bool{
		"project": true,
		"area":    true,
		"title":   true,
	}
	
	// If it's a known string field, quote it
	if stringFields[field] {
		return true
	}
	
	// Don't quote dates, numbers, or priority values
	if field == "due_date" || field == "start_date" {
		return false
	}
	if field == "priority" && (value == "p1" || value == "p2" || value == "p3") {
		return false
	}
	if field == "estimate" {
		return false
	}
	
	// Quote if contains special characters
	if strings.ContainsAny(value, ": {}[]|>\"'") {
		return true
	}
	
	return false
}

// BulkUpdateTaskStatus updates status for multiple tasks
func BulkUpdateTaskStatus(filepaths []string, newStatus string) error {
	for _, filepath := range filepaths {
		if err := UpdateTaskStatus(filepath, newStatus); err != nil {
			return fmt.Errorf("failed to update %s: %w", filepath, err)
		}
	}
	return nil
}