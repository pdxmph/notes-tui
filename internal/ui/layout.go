package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Layout manages the application's layout
type Layout struct {
	Width         int
	Height        int
	MarginPercent int // Percentage of width for margins
	Theme         Theme
}

// NewLayout creates a new layout manager
func NewLayout(width, height int, theme Theme) *Layout {
	return &Layout{
		Width:         width,
		Height:        height,
		MarginPercent: 15,
		Theme:         theme,
	}
}

// ContentArea returns the available content dimensions
func (l *Layout) ContentArea() (width, height int) {
	marginSize := l.Width * l.MarginPercent / 100
	width = l.Width - (marginSize * 2)
	height = l.Height
	return
}

// ApplyMargins wraps content with appropriate margins
func (l *Layout) ApplyMargins(content string) string {
	marginSize := l.Width * l.MarginPercent / 100
	contentWidth, _ := l.ContentArea()
	
	style := lipgloss.NewStyle().
		Width(contentWidth).
		MarginLeft(marginSize).
		MarginRight(marginSize)
		
	return style.Render(content)
}

// CenterPopover centers a popover on screen
func (l *Layout) CenterPopover(content string, widthPercent, heightPercent int) string {
	// The content should already be sized appropriately
	// Just center it on the screen
	centerStyle := lipgloss.NewStyle().
		Width(l.Width).
		Height(l.Height).
		Align(lipgloss.Center, lipgloss.Center)
		
	return centerStyle.Render(content)
}

// RenderScreen composes the full screen layout
func (l *Layout) RenderScreen(header, content, footer string) string {
	contentWidth, contentHeight := l.ContentArea()
	
	// Calculate heights
	headerLines := strings.Count(header, "\n") + 1
	footerLines := strings.Count(footer, "\n") + 1
	mainHeight := contentHeight - headerLines - footerLines - 2 // margins
	
	// Apply width constraints
	if header != "" {
		header = lipgloss.NewStyle().Width(contentWidth).Render(header)
	}
	if footer != "" {
		footer = lipgloss.NewStyle().Width(contentWidth).Render(footer)
	}
	
	// Build the screen
	var parts []string
	if header != "" {
		parts = append(parts, header)
	}
	if content != "" {
		// Apply height constraint to content
		contentStyle := lipgloss.NewStyle().
			Width(contentWidth).
			Height(mainHeight).
			MaxHeight(mainHeight)
		parts = append(parts, contentStyle.Render(content))
	}
	if footer != "" {
		parts = append(parts, footer)
	}
	
	// Join and apply margins
	screen := strings.Join(parts, "\n")
	return l.ApplyMargins(screen)
}

// ViewportBounds calculates visible item range for scrolling lists
func ViewportBounds(totalItems, viewHeight, cursorPos int) (start, end int) {
	if totalItems <= viewHeight {
		return 0, totalItems
	}
	
	// Keep cursor in view
	start = 0
	if cursorPos >= viewHeight {
		start = cursorPos - viewHeight + 1
	}
	
	end = start + viewHeight
	if end > totalItems {
		end = totalItems
	}
	
	return start, end
}

// TruncateText safely truncates text with ellipsis
func TruncateText(text string, maxLen int) string {
	if len(text) <= maxLen || maxLen < 4 {
		return text
	}
	return text[:maxLen-3] + "..."
}

// WrapText wraps text to fit within width
func WrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}
	
	var lines []string
	var current []string
	currentLen := 0
	
	for _, word := range words {
		wordLen := len(word)
		if currentLen > 0 && currentLen+1+wordLen > width {
			// Start new line
			lines = append(lines, strings.Join(current, " "))
			current = []string{word}
			currentLen = wordLen
		} else {
			// Add to current line
			current = append(current, word)
			if currentLen > 0 {
				currentLen++ // space
			}
			currentLen += wordLen
		}
	}
	
	if len(current) > 0 {
		lines = append(lines, strings.Join(current, " "))
	}
	
	return lines
}

// SplitIntoColumns splits items into columns
func SplitIntoColumns(items []string, columns int, columnWidth int) string {
	if columns <= 0 || len(items) == 0 {
		return ""
	}
	
	// Calculate items per column
	itemsPerColumn := (len(items) + columns - 1) / columns
	
	// Build columns
	var cols []string
	for i := 0; i < columns; i++ {
		start := i * itemsPerColumn
		end := start + itemsPerColumn
		if end > len(items) {
			end = len(items)
		}
		
		if start >= len(items) {
			break
		}
		
		// Build column
		var colItems []string
		for j := start; j < end; j++ {
			item := TruncateText(items[j], columnWidth)
			colItems = append(colItems, item)
		}
		
		col := strings.Join(colItems, "\n")
		cols = append(cols, col)
	}
	
	// Join columns horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}