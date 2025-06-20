package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	files       []string        // all markdown files
	filtered    []string        // filtered results
	cursor      int             // which file is selected
	selected    string          // selected file
	searchMode  bool            // are we in search mode?
	search      textinput.Model // search input
	createMode  bool            // are we in create mode?
	createInput textinput.Model // create note input
	tagMode     bool            // are we in tag search mode?
	tagInput    textinput.Model // tag search input
	taskFilter  bool            // are we showing only files with tasks?
	cwd         string          // current working directory
	width       int             // terminal width
	height      int             // terminal height
	// Preview popover state
	previewMode    bool            // are we showing preview popover?
	previewContent string          // content for preview popover
	previewFile    string          // file being previewed
	previewScroll  int             // scroll position in preview
}

// Message for preview content
type previewLoadedMsg struct {
	content  string
	filepath string
}

func initialModel() model {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	
	files, err := findMarkdownFiles(cwd)
	if err != nil {
		log.Fatal(err)
	}

	// Create search input
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 100
	ti.Width = 50

	// Create note input
	ci := textinput.New()
	ci.Placeholder = "Note title..."
	ci.CharLimit = 100
	ci.Width = 50

	// Create tag input
	tagi := textinput.New()
	tagi.Placeholder = "Enter tag..."
	tagi.CharLimit = 50
	tagi.Width = 30

	return model{
		files:       files,
		filtered:    files, // Initially show all files
		search:      ti,
		createInput: ci,
		tagInput:    tagi,
		cwd:         cwd,
	}
}

func findMarkdownFiles(dir string) ([]string, error) {
	var files []string
	
	// Debug: log what directory we're searching
	// fmt.Printf("Searching for markdown files in: %s\n", dir)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories (but not the root if it's hidden)
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != dir {
			return filepath.SkipDir
		}

		// Check if it's a markdown file
		if !info.IsDir() && (strings.HasSuffix(strings.ToLower(info.Name()), ".md") || strings.HasSuffix(strings.ToLower(info.Name()), ".markdown")) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// Helper to get display name for a file
func getDisplayName(fullPath, cwd string) string {
	rel, err := filepath.Rel(cwd, fullPath)
	if err != nil {
		return filepath.Base(fullPath)
	}
	return rel
}

// Filter files based on search query
func filterFiles(files []string, query string) []string {
	if query == "" {
		return files
	}

	query = strings.ToLower(query)
	var filtered []string
	
	for _, file := range files {
		if strings.Contains(strings.ToLower(file), query) {
			filtered = append(filtered, file)
		}
	}
	
	return filtered
}

// Convert title to filename
func titleToFilename(title string) string {
	// Convert to lowercase
	filename := strings.ToLower(title)
	
	// Replace spaces with hyphens
	filename = strings.ReplaceAll(filename, " ", "-")
	
	// Remove any characters that aren't alphanumeric or hyphens
	var result strings.Builder
	for _, r := range filename {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	
	// Clean up multiple hyphens
	cleaned := result.String()
	for strings.Contains(cleaned, "--") {
		cleaned = strings.ReplaceAll(cleaned, "--", "-")
	}
	
	// Trim hyphens from start and end
	cleaned = strings.Trim(cleaned, "-")
	
	// Add .md extension
	if cleaned == "" {
		cleaned = "untitled"
	}
	
	return cleaned + ".md"
}

// Get today's daily note filename
func getDailyNoteFilename() string {
	today := time.Now().Format("2006-01-02")
	return today + "-daily.md"
}

// Search for files containing a specific tag using ripgrep
func searchTag(dir, tag string) ([]string, error) {
	// Remove # if present at the start
	tag = strings.TrimPrefix(tag, "#")
	
	// We'll search for the tag in different contexts:
	// 1. Inline hashtag: #tag
	// 2. In YAML frontmatter tags field (various formats)
	//    - tags: [tag1, tag2]
	//    - tags: ["tag1", "tag2"]  
	//    - tags:
	//      - tag1
	//      - tag2
	
	// Run ripgrep with multiple patterns
	cmd := exec.Command("rg", "-l", "--type", "md", 
		"-e", fmt.Sprintf("#%s\\b", tag),                    // #tag with word boundary
		"-e", fmt.Sprintf(`tags:.*\b%s\b`, tag),            // tags: line containing the tag
		"-e", fmt.Sprintf(`^\s*-\s+%s\b`, tag),             // - tag in YAML lists
		dir)
	
	output, err := cmd.Output()
	if err != nil {
		// Check if it's just "no matches found" (exit code 1)
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []string{}, nil
		}
		return nil, err
	}
	
	// Parse output - each line is a file path
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var files []string
	for _, line := range lines {
		if line != "" {
			files = append(files, line)
		}
	}
	
	return files, nil
}

// Search for files containing open tasks using ripgrep
func searchTasks(dir string) ([]string, error) {
	// Search for unchecked checkbox pattern: "- [ ]"
	// Use -F for fixed strings and -- to stop flag parsing
	cmd := exec.Command("rg", "-l", "--type", "md", "-F", "--", "- [ ]", dir)
	output, err := cmd.Output()
	if err != nil {
		// Check if it's just "no matches found" (exit code 1)
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []string{}, nil
		}
		return nil, err
	}
	
	// Parse output - each line is a file path
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var files []string
	for _, line := range lines {
		if line != "" {
			files = append(files, line)
		}
	}
	
	return files, nil
}
func (m model) Init() tea.Cmd {
	return tea.WindowSize()
}

// Simple markdown renderer for fast preview
func renderSimpleMarkdown(content string, width int) string {
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

// Load preview content for popover (with simple markdown rendering)
func (m *model) loadPreviewForPopover() tea.Cmd {
	if m.cursor >= len(m.filtered) || m.cursor < 0 {
		return nil
	}
	
	filepath := m.filtered[m.cursor]
	width := m.width // Capture width for the closure
	
	return func() tea.Msg {
		content, err := os.ReadFile(filepath)
		if err != nil {
			return previewLoadedMsg{
				content: fmt.Sprintf("Error reading file: %v", err),
				filepath: filepath,
			}
		}
		
		// Use simple markdown renderer
		popoverContentWidth := (width * 80 / 100) - 6
		rendered := renderSimpleMarkdown(string(content), popoverContentWidth)
		
		return previewLoadedMsg{
			content: rendered,
			filepath: filepath,
		}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// If in preview mode, reload with new dimensions
		if m.previewMode {
			return m, m.loadPreviewForPopover()
		}
		return m, nil

	case previewLoadedMsg:
		m.previewContent = msg.content
		return m, nil

	case tea.KeyMsg:
		// Handle preview mode separately
		if m.previewMode {
			switch msg.String() {
			case "esc", "q":
				m.previewMode = false
				m.previewContent = ""
				m.previewScroll = 0
				return m, nil
			
			case "e", "ctrl+e":
				// Open in editor from preview
				m.previewMode = false
				m.previewContent = ""
				m.previewScroll = 0
				m.selected = m.previewFile
				return m, tea.ExecProcess(m.openInEditor(), func(err error) tea.Msg {
					return nil
				})
			
			case "up", "k":
				if m.previewScroll > 0 {
					m.previewScroll--
				}
				return m, nil
			
			case "down", "j":
				// Check if we can scroll down
				lines := strings.Split(m.previewContent, "\n")
				contentHeight := (m.height * 80 / 100) - 8
				if m.previewScroll < len(lines)-contentHeight {
					m.previewScroll++
				}
				return m, nil
			
			case "pgup":
				contentHeight := (m.height * 80 / 100) - 8
				m.previewScroll -= contentHeight
				if m.previewScroll < 0 {
					m.previewScroll = 0
				}
				return m, nil
				
			case "pgdown", " ": // Space bar also pages down
				lines := strings.Split(m.previewContent, "\n")
				contentHeight := (m.height * 80 / 100) - 8
				m.previewScroll += contentHeight
				maxScroll := len(lines) - contentHeight
				if m.previewScroll > maxScroll {
					m.previewScroll = maxScroll
				}
				if m.previewScroll < 0 {
					m.previewScroll = 0
				}
				return m, nil
			}
			return m, nil
		}

		// Normal mode key handling
		switch msg.String() {
		case "ctrl+c", "q":
			if m.searchMode {
				// Exit search mode on q
				m.searchMode = false
				m.search.SetValue("")
				m.filtered = m.files
				m.cursor = 0
				return m, nil
			}
			if m.createMode {
				// Exit create mode on q
				m.createMode = false
				m.createInput.SetValue("")
				return m, nil
			}
			if m.tagMode {
				// Exit tag mode on q
				m.tagMode = false
				m.tagInput.SetValue("")
				m.filtered = m.files
				m.cursor = 0
				return m, nil
			}
			return m, tea.Quit

		case "e", "ctrl+e":
			// Open in external editor
			if !m.searchMode && !m.createMode && !m.tagMode && m.cursor < len(m.filtered) {
				m.selected = m.filtered[m.cursor]
				// We'll handle the actual editor opening after we return
				return m, tea.ExecProcess(m.openInEditor(), func(err error) tea.Msg {
					return nil
				})
			}

		case "esc":
			if m.searchMode {
				// Exit search mode
				m.searchMode = false
				m.search.SetValue("")
				m.filtered = m.files
				m.cursor = 0
			}
			if m.createMode {
				// Exit create mode
				m.createMode = false
				m.createInput.SetValue("")
			}
			if m.tagMode {
				// Exit tag mode
				m.tagMode = false
				m.tagInput.SetValue("")
				m.filtered = m.files
				m.cursor = 0
			}
			if m.taskFilter {
				// Clear task filter
				m.taskFilter = false
				m.filtered = m.files
				m.cursor = 0
			}

		case "ctrl+n":
			if !m.searchMode && !m.createMode {
				// Enter create mode
				m.createMode = true
				m.createInput.Focus()
				return m, nil
			}

		case "ctrl+d":
			if !m.searchMode && !m.createMode {
				// Create or open daily note
				filename := getDailyNoteFilename()
				fullPath := filepath.Join(m.cwd, filename)
				
				// Check if file exists
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					// Create the daily note with a header
					today := time.Now().Format("Monday, January 2, 2006")
					content := fmt.Sprintf("# Daily Note - %s\n\n## Tasks\n\n## Notes\n\n", today)
					if err := os.WriteFile(fullPath, []byte(content), 0644); err == nil {
						m.selected = fullPath
						// Refresh file list to include new file
						files, _ := findMarkdownFiles(m.cwd)
						m.files = files
						m.filtered = files
						// Find and select the new file
						for i, f := range m.filtered {
							if f == fullPath {
								m.cursor = i
								break
							}
						}
						return m, nil
					}
				} else {
					// File exists, just open it
					m.selected = fullPath
					// Find and select the file
					for i, f := range m.filtered {
						if f == fullPath {
							m.cursor = i
							break
						}
					}
					return m, nil
				}
			}

		case "/":
			if !m.createMode && !m.tagMode {
				// Enter search mode
				m.searchMode = true
				m.search.Focus()
				return m, nil
			}

		case "#":
			if !m.searchMode && !m.createMode && !m.tagMode {
				// Enter tag search mode
				m.tagMode = true
				m.tagInput.Focus()
				return m, nil
			}

		case "ctrl+t":
			if !m.searchMode && !m.createMode && !m.tagMode {
				// Search for tasks
				if files, err := searchTasks(m.cwd); err == nil {
					m.filtered = files
					m.cursor = 0
					m.taskFilter = true
				}
			}

		case "up", "k":
			if !m.searchMode && !m.createMode && !m.tagMode && m.cursor > 0 {
				m.cursor--
				// Don't auto-load preview on cursor movement
			}

		case "down", "j":
			if !m.searchMode && !m.createMode && !m.tagMode && m.cursor < len(m.filtered)-1 {
				m.cursor++
				// Don't auto-load preview on cursor movement
			}

		case "enter":
			if m.tagMode {
				// Search for the tag
				tag := m.tagInput.Value()
				if tag != "" {
					if files, err := searchTag(m.cwd, tag); err == nil {
						m.filtered = files
						m.cursor = 0
					}
				}
				// Exit tag mode
				m.tagMode = false
				m.tagInput.SetValue("")
			} else if m.createMode {
				// Create the new note
				title := m.createInput.Value()
				if title != "" {
					filename := titleToFilename(title)
					fullPath := filepath.Join(m.cwd, filename)
					
					// Create the file with the title as the first line
					content := fmt.Sprintf("# %s\n\n", title)
					if err := os.WriteFile(fullPath, []byte(content), 0644); err == nil {
						m.selected = fullPath
						// Refresh file list to include new file
						files, _ := findMarkdownFiles(m.cwd)
						m.files = files
						m.filtered = files
						// Find and select the new file
						for i, f := range m.filtered {
							if f == fullPath {
								m.cursor = i
								break
							}
						}
						// Exit create mode
						m.createMode = false
						m.createInput.SetValue("")
						return m, nil
					}
				}
				// Exit create mode
				m.createMode = false
				m.createInput.SetValue("")
			} else if m.searchMode {
				// Exit search mode on enter
				m.searchMode = false
			} else if m.cursor < len(m.filtered) {
				// Show preview popover
				m.selected = m.filtered[m.cursor]
				m.previewFile = m.selected
				m.previewMode = true
				m.previewScroll = 0
				return m, m.loadPreviewForPopover()
			}
		}
	}

	// Handle search input
	if m.searchMode {
		m.search, cmd = m.search.Update(msg)
		query := m.search.Value()
		m.filtered = filterFiles(m.files, query)
		m.cursor = 0 // Reset cursor when filtering
		cmds = append(cmds, cmd)
	}

	// Handle create input
	if m.createMode {
		m.createInput, cmd = m.createInput.Update(msg)
	}

	// Handle tag input
	if m.tagMode {
		m.tagInput, cmd = m.tagInput.Update(msg)
	}

	return m, tea.Batch(cmds...)
}

// Open file in external editor
func (m model) openInEditor() *exec.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// Try common editors
		editors := []string{"vim", "nvim", "emacs", "nano"}
		for _, e := range editors {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}
	
	if editor == "" {
		editor = "vi" // fallback
	}
	
	cmd := exec.Command(editor, m.selected)
	return cmd
}
func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// If in preview mode, show the popover
	if m.previewMode {
		return m.renderPreviewPopover()
	}

	// Calculate margins (15% each side)
	marginSize := m.width * 15 / 100
	contentWidth := m.width - (marginSize * 2)
	
	// Create main content style with margins
	contentStyle := lipgloss.NewStyle().
		Width(contentWidth).
		MarginLeft(marginSize).
		MarginRight(marginSize)

	// Build content
	var content strings.Builder
	
	// Header
	header := fmt.Sprintf("Notes (%d files)", len(m.filtered))
	if m.taskFilter {
		header += " - [Tasks]"
	}
	content.WriteString(lipgloss.NewStyle().Bold(true).Render(header) + "\n\n")

	if m.tagMode {
		// Tag search mode
		content.WriteString("Search by Tag\n\n")
		content.WriteString(fmt.Sprintf("Tag: %s\n\n", m.tagInput.View()))
		content.WriteString("[Enter] search [Esc] cancel")
	} else if m.createMode {
		// Create mode
		content.WriteString("Create New Note\n\n")
		content.WriteString(fmt.Sprintf("Title: %s\n\n", m.createInput.View()))
		content.WriteString("[Enter] create [Esc] cancel")
	} else if len(m.filtered) == 0 && m.searchMode && m.search.Value() != "" {
		content.WriteString("No files match your search.\n\n")
	} else if len(m.filtered) == 0 && !m.searchMode {
		content.WriteString("No files found.\n\n")
	} else if len(m.files) == 0 {
		content.WriteString("No markdown files found.\n\n")
	} else {
		// Show file list
		maxVisible := m.height - 8 // Leave room for header, search, and help
		startIdx := 0
		
		// Adjust view window if cursor is outside
		if m.cursor >= maxVisible {
			startIdx = m.cursor - maxVisible + 1
		}
		
		for i := startIdx; i < len(m.filtered) && i < startIdx+maxVisible; i++ {
			cursor := "  "
			if m.cursor == i {
				cursor = "> "
			}
			
			displayName := getDisplayName(m.filtered[i], m.cwd)
			// Truncate if too long
			maxLen := contentWidth - 3
			if len(displayName) > maxLen {
				displayName = displayName[:maxLen-3] + "..."
			}
			content.WriteString(fmt.Sprintf("%s%s\n", cursor, displayName))
		}
		
		if len(m.filtered) > maxVisible {
			remaining := len(m.filtered) - startIdx - maxVisible
			if remaining > 0 {
				content.WriteString(fmt.Sprintf("\n... %d more files\n", remaining))
			}
		}
	}

	// Add search field at bottom
	if !m.tagMode && !m.createMode {
		content.WriteString("\n")
		if m.searchMode {
			content.WriteString(fmt.Sprintf("Search: %s", m.search.View()))
		} else {
			// Help text
			helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			help := "[/] search  [Enter] preview  [e] edit  [Ctrl+N] new  [Ctrl+D] daily  [q] quit"
			content.WriteString("\n" + helpStyle.Render(help))
		}
	}

	return contentStyle.Render(content.String())
}

func (m model) renderPreviewPopover() string {
	// Create a popover style
	popoverStyle := lipgloss.NewStyle().
		Width(m.width * 80 / 100).          // 80% of terminal width
		Height(m.height * 80 / 100).        // 80% of terminal height
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)
	
	// Position in center
	centerStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)
	
	// Header with filename
	header := lipgloss.NewStyle().
		Bold(true).
		MarginBottom(1).
		Render(getDisplayName(m.previewFile, m.cwd))
	
	// Footer with controls
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(1).
		Render("[Esc] close  [↑↓/jk] scroll  [e] edit")
	
	// Calculate available height for content
	contentHeight := (m.height * 80 / 100) - 8 // Account for borders, padding, header, footer
	
	// Split content into lines and handle scrolling
	lines := strings.Split(m.previewContent, "\n")
	visibleLines := lines
	
	if len(lines) > contentHeight {
		// Apply scrolling
		end := m.previewScroll + contentHeight
		if end > len(lines) {
			end = len(lines)
		}
		visibleLines = lines[m.previewScroll:end]
		
		// Add scroll indicator
		scrollInfo := fmt.Sprintf(" (line %d/%d)", m.previewScroll+1, len(lines))
		header += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(scrollInfo)
	}
	
	// Join the visible content
	content := strings.Join(visibleLines, "\n")
	
	// Combine everything
	popoverContent := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)
	
	popover := popoverStyle.Render(popoverContent)
	return centerStyle.Render(popover)
}

func main() {
	// If a directory is specified as an argument, change to it
	if len(os.Args) > 1 {
		dir := os.Args[1]
		if err := os.Chdir(dir); err != nil {
			log.Fatal(err)
		}
	}

	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	// If a file was selected, open it in $EDITOR
	if m, ok := m.(model); ok && m.selected != "" {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			// Try common editors
			editors := []string{"vim", "nvim", "nano", "emacs", "code"}
			for _, e := range editors {
				if _, err := exec.LookPath(e); err == nil {
					editor = e
					break
				}
			}
		}
		
		if editor == "" {
			fmt.Println("No editor found. Please set $EDITOR environment variable.")
			fmt.Printf("Selected file: %s\n", m.selected)
			return
		}

		// Open the file in the editor
		cmd := exec.Command(editor, m.selected)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error opening editor: %v\n", err)
			fmt.Printf("Selected file: %s\n", m.selected)
		}
	}
}