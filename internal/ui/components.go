package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// Component interface for all UI components
type Component interface {
	View() string
	Width() int
	Height() int
}

// ListView component for file browsing
type ListView struct {
	Items        []string
	Cursor       int
	Width        int
	Height       int
	ShowCursor   bool
	EmptyMessage string
	Style        ListStyle
	ItemFormatter func(string) string // Optional custom formatter
}

type ListStyle struct {
	Cursor      lipgloss.Style
	Item        lipgloss.Style
	EmptyMsg    lipgloss.Style
}

func (l ListView) View() string {
	if len(l.Items) == 0 {
		return l.Style.EmptyMsg.Render(l.EmptyMessage)
	}

	var content strings.Builder
	maxVisible := l.Height
	if maxVisible <= 0 {
		maxVisible = 1
	}
	
	// Simple approach: just show items without indicators if height is too small
	if maxVisible < 3 && len(l.Items) > maxVisible {
		// Not enough room for indicators, just show items around cursor
		startIdx := l.Cursor
		if startIdx > len(l.Items) - maxVisible {
			startIdx = len(l.Items) - maxVisible
		}
		if startIdx < 0 {
			startIdx = 0
		}
		
		for i := startIdx; i < len(l.Items) && i < startIdx+maxVisible; i++ {
			cursor := "  "
			if l.ShowCursor && l.Cursor == i {
				cursor = "> "
			}

			item := l.Items[i]
			if l.ItemFormatter != nil {
				item = l.ItemFormatter(item)
			}
			
			maxLen := l.Width - 3
			if len(item) > maxLen && maxLen > 3 {
				item = item[:maxLen-3] + "..."
			}

			line := fmt.Sprintf("%s%s", cursor, item)
			if l.ShowCursor && l.Cursor == i {
				content.WriteString(l.Style.Cursor.Render(line))
			} else {
				content.WriteString(l.Style.Item.Render(line))
			}
			
			if i < startIdx+maxVisible-1 {
				content.WriteString("\n")
			}
		}
		return content.String()
	}
	
	// Normal case with indicators
	startIdx := 0
	endIdx := len(l.Items)
	
	// If list is scrollable, calculate viewport
	if len(l.Items) > maxVisible {
		// Keep cursor in the middle third of the view when possible
		preferredStart := l.Cursor - maxVisible/3
		
		// Adjust bounds
		if preferredStart < 0 {
			startIdx = 0
		} else if preferredStart + maxVisible > len(l.Items) {
			startIdx = len(l.Items) - maxVisible
		} else {
			startIdx = preferredStart
		}
		
		endIdx = startIdx + maxVisible
	}
	
	// Show top indicator if needed
	if startIdx > 0 {
		indicator := fmt.Sprintf("... %d items above", startIdx)
		content.WriteString(l.Style.EmptyMsg.Render(indicator))
		content.WriteString("\n")
		startIdx++ // Skip one item to make room for indicator
	}
	
	// Show bottom indicator if needed
	showBottomIndicator := endIdx < len(l.Items)
	if showBottomIndicator {
		endIdx-- // Reserve space for bottom indicator
	}

	// Render visible items
	for i := startIdx; i < endIdx && i < len(l.Items); i++ {
		cursor := "  "
		if l.ShowCursor && l.Cursor == i {
			cursor = "> "
		}

		item := l.Items[i]
		
		// Apply custom formatter if provided
		if l.ItemFormatter != nil {
			item = l.ItemFormatter(item)
		}
		
		// Truncate if too long
		maxLen := l.Width - 3
		if len(item) > maxLen && maxLen > 3 {
			item = item[:maxLen-3] + "..."
		}

		line := fmt.Sprintf("%s%s", cursor, item)
		if l.ShowCursor && l.Cursor == i {
			content.WriteString(l.Style.Cursor.Render(line))
		} else {
			content.WriteString(l.Style.Item.Render(line))
		}
		
		if i < endIdx-1 {
			content.WriteString("\n")
		}
	}

	// Add bottom indicator if needed
	if showBottomIndicator {
		remaining := len(l.Items) - endIdx
		indicator := fmt.Sprintf("\n... %d more items", remaining)
		content.WriteString(l.Style.EmptyMsg.Render(indicator))
	}

	return content.String()
}

// InputModal component for various input modes
type InputModal struct {
	Title       string
	Prompt      string
	Input       textinput.Model
	HelpText    string
	Width       int
	Style       ModalStyle
}

type ModalStyle struct {
	Title    lipgloss.Style
	Prompt   lipgloss.Style
	Help     lipgloss.Style
	Border   lipgloss.Style
}

func (m InputModal) View() string {
	var content strings.Builder
	
	if m.Title != "" {
		content.WriteString(m.Style.Title.Render(m.Title))
		content.WriteString("\n\n")
	}
	
	if m.Prompt != "" {
		content.WriteString(m.Style.Prompt.Render(m.Prompt))
		content.WriteString(" ")
	}
	
	content.WriteString(m.Input.View())
	
	if m.HelpText != "" {
		content.WriteString("\n\n")
		content.WriteString(m.Style.Help.Render(m.HelpText))
	}
	
	return m.Style.Border.Width(m.Width).Render(content.String())
}

// Header component
type Header struct {
	Title      string
	FileCount  int
	Filters    []string
	SortInfo   string
	Width      int
	Style      HeaderStyle
}

type HeaderStyle struct {
	Title   lipgloss.Style
	Filter  lipgloss.Style
	Sort    lipgloss.Style
}

func (h Header) View() string {
	title := fmt.Sprintf("%s (%d files)", h.Title, h.FileCount)
	
	// Add active filters
	for _, filter := range h.Filters {
		title += fmt.Sprintf(" - %s", h.Style.Filter.Render("["+filter+"]"))
	}
	
	// Add sort info
	if h.SortInfo != "" {
		title += fmt.Sprintf(" - %s", h.Style.Sort.Render("[Sort: "+h.SortInfo+"]"))
	}
	
	return h.Style.Title.Width(h.Width).Render(title)
}

// HelpBar component
type HelpBar struct {
	Items   []HelpItem
	Width   int
	Style   HelpStyle
}

type HelpItem struct {
	Key   string
	Desc  string
}

type HelpStyle struct {
	Key       lipgloss.Style
	Desc      lipgloss.Style
	Separator lipgloss.Style
}

func (h HelpBar) View() string {
	if len(h.Items) == 0 {
		return ""
	}

	sep := h.Style.Separator.Render(" • ")
	
	items := make([]string, len(h.Items))
	for i, item := range h.Items {
		// Special handling for keys that should appear mid-word
		var formatted string
		if strings.Contains(item.Desc, "[") && strings.Contains(item.Desc, "]") {
			// Description contains the key position marker
			formatted = h.Style.Desc.Render(item.Desc)
			// Replace [X] with styled key
			formatted = strings.Replace(formatted, "["+item.Key+"]", h.Style.Key.Render("["+item.Key+"]"), 1)
		} else {
			// Standard format: [key] description
			key := h.Style.Key.Render("[" + item.Key + "]")
			desc := h.Style.Desc.Render(item.Desc)
			formatted = fmt.Sprintf("%s %s", key, desc)
		}
		items[i] = formatted
	}
	
	return strings.Join(items, sep)
}

// PreviewPopover component
type PreviewPopover struct {
	Title       string
	Content     string
	ScrollPos   int
	Width       int
	Height      int
	Style       PopoverStyle
}

type PopoverStyle struct {
	Border    lipgloss.Style
	Title     lipgloss.Style
	ScrollBar lipgloss.Style
	Help      lipgloss.Style
}

func (p PreviewPopover) View() string {
	// Calculate content area dimensions
	contentHeight := p.Height - 4 // Account for borders and padding
	
	// Header with scroll info
	header := p.Style.Title.Render(p.Title)
	
	// Split content into lines for scrolling
	lines := strings.Split(p.Content, "\n")
	visibleLines := lines
	
	if len(lines) > contentHeight {
		// Apply scrolling
		end := p.ScrollPos + contentHeight
		if end > len(lines) {
			end = len(lines)
		}
		visibleLines = lines[p.ScrollPos:end]
		
		// Add scroll indicator
		scrollInfo := fmt.Sprintf(" (line %d/%d)", p.ScrollPos+1, len(lines))
		header += p.Style.Title.Foreground(lipgloss.Color("240")).Render(scrollInfo)
	}
	
	// Join visible content
	content := strings.Join(visibleLines, "\n")
	
	// Footer
	footer := p.Style.Help.Render("[Esc] close  [↑↓/jk] scroll  [e] edit")
	
	// Combine all parts
	fullContent := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)
	
	return p.Style.Border.
		Width(p.Width).
		Height(p.Height).
		Render(fullContent)
}

// ConfirmDialog component
type ConfirmDialog struct {
	Title    string
	Message  string
	Options  []DialogOption
	Width    int
	Style    DialogStyle
}

type DialogOption struct {
	Key   string
	Label string
}

type DialogStyle struct {
	Title   lipgloss.Style
	Message lipgloss.Style
	Option  lipgloss.Style
	Border  lipgloss.Style
}

func (d ConfirmDialog) View() string {
	var content strings.Builder
	
	content.WriteString(d.Style.Title.Render(d.Title))
	content.WriteString("\n\n")
	content.WriteString(d.Style.Message.Render(d.Message))
	content.WriteString("\n\n")
	
	options := make([]string, len(d.Options))
	for i, opt := range d.Options {
		options[i] = d.Style.Option.Render(fmt.Sprintf("[%s] %s", opt.Key, opt.Label))
	}
	content.WriteString(strings.Join(options, "  "))
	
	return d.Style.Border.Width(d.Width).Render(content.String())
}

// LoadingIndicator component
type LoadingIndicator struct {
	Message string
	Style   lipgloss.Style
}

func (l LoadingIndicator) View() string {
	return l.Style.Render("⣾ " + l.Message + "...")
}

// StatusMessage component for temporary notifications
type StatusMessage struct {
	Message  string
	Type     StatusType
	Duration int // frames to display
	Style    StatusStyle
}

type StatusType int

const (
	StatusInfo StatusType = iota
	StatusSuccess
	StatusWarning
	StatusError
)

type StatusStyle struct {
	Info    lipgloss.Style
	Success lipgloss.Style
	Warning lipgloss.Style
	Error   lipgloss.Style
}

func (s StatusMessage) View() string {
	if s.Message == "" || s.Duration <= 0 {
		return ""
	}
	
	var style lipgloss.Style
	switch s.Type {
	case StatusSuccess:
		style = s.Style.Success
	case StatusWarning:
		style = s.Style.Warning
	case StatusError:
		style = s.Style.Error
	default:
		style = s.Style.Info
	}
	
	return style.Render(s.Message)
}