package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderSimpleMarkdown renders markdown content for preview
func RenderSimpleMarkdown(content string, width int) string {
	lines := strings.Split(content, "\n")
	var result []string
	
	// Skip YAML frontmatter
	startIdx := 0
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		// Find the closing ---
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				startIdx = i + 1
				break
			}
		}
	}
	
	// Define styles
	h1Style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	h2Style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	h3Style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	codeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	bulletStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	
	inCodeBlock := false
	
	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		// Handle code blocks
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			if inCodeBlock {
				result = append(result, codeStyle.Render("─────────────────────"))
			} else {
				result = append(result, codeStyle.Render("─────────────────────"))
			}
			continue
		}
		
		if inCodeBlock {
			result = append(result, codeStyle.Render(line))
			continue
		}
		
		// Handle headers
		if strings.HasPrefix(line, "# ") {
			result = append(result, h1Style.Render(strings.TrimPrefix(line, "# ")))
		} else if strings.HasPrefix(line, "## ") {
			result = append(result, h2Style.Render(strings.TrimPrefix(line, "## ")))
		} else if strings.HasPrefix(line, "### ") {
			result = append(result, h3Style.Render(strings.TrimPrefix(line, "### ")))
		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			// Handle bullet points
			bullet := bulletStyle.Render("•")
			content := strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* ")
			result = append(result, fmt.Sprintf("%s %s", bullet, content))
		} else if strings.HasPrefix(line, "> ") {
			// Handle blockquotes
			quoteLine := strings.TrimPrefix(line, "> ")
			result = append(result, lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("8")).Render("│ " + quoteLine))
		} else if strings.TrimSpace(line) == "---" || strings.TrimSpace(line) == "***" {
			// Handle horizontal rules
			result = append(result, strings.Repeat("─", width-6))
		} else {
			// Handle inline formatting
			formatted := line
			
			// Bold
			for strings.Contains(formatted, "**") {
				start := strings.Index(formatted, "**")
				if start == -1 {
					break
				}
				end := strings.Index(formatted[start+2:], "**")
				if end == -1 {
					break
				}
				end += start + 2
				
				before := formatted[:start]
				bold := lipgloss.NewStyle().Bold(true).Render(formatted[start+2:end])
				after := formatted[end+2:]
				formatted = before + bold + after
			}
			
			// Italic
			for strings.Contains(formatted, "*") && !strings.Contains(formatted, "**") {
				start := strings.Index(formatted, "*")
				if start == -1 {
					break
				}
				end := strings.Index(formatted[start+1:], "*")
				if end == -1 {
					break
				}
				end += start + 1
				
				before := formatted[:start]
				italic := lipgloss.NewStyle().Italic(true).Render(formatted[start+1:end])
				after := formatted[end+1:]
				formatted = before + italic + after
			}
			
			// Inline code
			for strings.Contains(formatted, "`") {
				start := strings.Index(formatted, "`")
				if start == -1 {
					break
				}
				end := strings.Index(formatted[start+1:], "`")
				if end == -1 {
					break
				}
				end += start + 1
				
				before := formatted[:start]
				code := codeStyle.Render(formatted[start+1:end])
				after := formatted[end+1:]
				formatted = before + code + after
			}
			
			result = append(result, formatted)
		}
	}
	
	return strings.Join(result, "\n")
}