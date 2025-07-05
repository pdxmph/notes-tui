package ui

import (
	"bufio"
	"os"
	"strings"
)

// ExtractNoteTitle reads a markdown file and extracts its title
// It looks for:
// 1. title: field in YAML frontmatter
// 2. First level 1 heading (# Title)
func ExtractNoteTitle(filepath string) string {
	// Read the first 20 lines of the file to find title
	file, err := os.Open(filepath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	inFrontmatter := false
	
	for scanner.Scan() && lineCount < 20 {
		line := scanner.Text()
		lineCount++
		
		// Check for YAML frontmatter
		if lineCount == 1 && line == "---" {
			inFrontmatter = true
			continue
		}
		
		// Look for title in frontmatter
		if inFrontmatter {
			if line == "---" {
				inFrontmatter = false
				continue
			}
			if strings.HasPrefix(line, "title:") {
				title := strings.TrimSpace(strings.TrimPrefix(line, "title:"))
				// Remove quotes if present
				title = strings.Trim(title, "\"'")
				if title != "" {
					return title
				}
			}
		}
		
		// Look for first level 1 heading
		if !inFrontmatter && strings.HasPrefix(line, "# ") {
			title := strings.TrimSpace(strings.TrimPrefix(line, "# "))
			if title != "" {
				return title
			}
		}
	}
	
	return ""
}