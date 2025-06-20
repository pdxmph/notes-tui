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
	"github.com/charmbracelet/glamour"
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
	previewContent string       // content of the currently selected file
	width       int             // terminal width
	height      int             // terminal height
	lastPreviewedFile string   // track which file is currently previewed
}

// Message for preview content
type previewLoadedMsg struct {
	content string
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
	return tea.Batch(
		tea.WindowSize(),
		m.loadPreview(), // Load initial preview
	)
}

// Load the content of the selected file
func (m *model) loadPreview() tea.Cmd {
	// Don't reload if we're already showing this file
	if m.cursor >= len(m.filtered) || m.cursor < 0 {
		return nil
	}
	
	filepath := m.filtered[m.cursor]
	if filepath == m.lastPreviewedFile {
		return nil // Already showing this file
	}
	
	return func() tea.Msg {
		content, err := os.ReadFile(filepath)
		if err != nil {
			return previewLoadedMsg{
				content: fmt.Sprintf("Error reading file: %v", err),
				filepath: filepath,
			}
		}
		
		// Render markdown with glamour
		renderer, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(m.width * 60 / 100), // 60% of terminal width
		)
		
		rendered, err := renderer.Render(string(content))
		if err != nil {
			return previewLoadedMsg{
				content: string(content), // Fall back to raw content
				filepath: filepath,
			}
		}
		
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
		// Reload preview with new width
		m.lastPreviewedFile = "" // Force reload with new dimensions
		return m, m.loadPreview()

	case previewLoadedMsg:
		m.previewContent = msg.content
		m.lastPreviewedFile = msg.filepath
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.searchMode {
				// Exit search mode on q
				m.searchMode = false
				m.search.SetValue("")
				m.filtered = m.files
				m.cursor = 0
				m.lastPreviewedFile = "" // Clear preview cache
				return m, m.loadPreview()
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
				m.lastPreviewedFile = "" // Clear preview cache
				return m, m.loadPreview()
			}
			return m, tea.Quit

		case "e", "ctrl+e":
			// Open in external editor
			if m.cursor < len(m.filtered) {
				m.selected = m.filtered[m.cursor]
				// We'll handle the actual editor opening after we return
				return m, tea.ExecProcess(m.openInEditor(), func(err error) tea.Msg {
					return m.loadPreview()
				})
			}

		case "esc":
			if m.searchMode {
				// Exit search mode
				m.searchMode = false
				m.search.SetValue("")
				m.filtered = m.files
				m.cursor = 0
				cmds = append(cmds, m.loadPreview())
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
				cmds = append(cmds, m.loadPreview())
			}
			if m.taskFilter {
				// Clear task filter
				m.taskFilter = false
				m.filtered = m.files
				m.cursor = 0
				cmds = append(cmds, m.loadPreview())
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
						return m, m.loadPreview()
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
					return m, m.loadPreview()
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
					cmds = append(cmds, m.loadPreview())
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
						cmds = append(cmds, m.loadPreview())
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
						return m, m.loadPreview()
					}
				}
				// Exit create mode
				m.createMode = false
				m.createInput.SetValue("")
			} else if m.searchMode {
				// Exit search mode on enter
				m.searchMode = false
			} else if m.cursor < len(m.filtered) {
				m.selected = m.filtered[m.cursor]
				return m, m.loadPreview()
			}
		}
	}

	// Handle search input
	if m.searchMode {
		m.search, cmd = m.search.Update(msg)
		query := m.search.Value()
		m.filtered = filterFiles(m.files, query)
		oldCursor := m.cursor
		m.cursor = 0 // Reset cursor when filtering
		if oldCursor != m.cursor && len(m.filtered) > 0 {
			cmds = append(cmds, cmd, m.loadPreview())
		} else {
			cmds = append(cmds, cmd)
		}
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

	// Calculate pane widths
	leftWidth := m.width * 40 / 100
	rightWidth := m.width - leftWidth - 1 // -1 for border

	// Create styles
	leftPaneStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Height(m.height - 1) // -1 for status line

	rightPaneStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Height(m.height - 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		Padding(1)

	// Build left pane (file list)
	var leftPane strings.Builder
	
	// Header
	header := fmt.Sprintf("Notes (%d files)", len(m.filtered))
	if m.taskFilter {
		header += " - [Tasks]"
	}
	leftPane.WriteString(header + "\n\n")

	if m.tagMode {
		// Tag search mode
		leftPane.WriteString("Search by Tag\n\n")
		leftPane.WriteString(fmt.Sprintf("Tag: %s\n\n", m.tagInput.View()))
		leftPane.WriteString("[Enter] search [Esc] cancel")
	} else if m.createMode {
		// Create mode
		leftPane.WriteString("Create New Note\n\n")
		leftPane.WriteString(fmt.Sprintf("Title: %s\n\n", m.createInput.View()))
		leftPane.WriteString("[Enter] create [Esc] cancel")
	} else if len(m.filtered) == 0 && m.searchMode && m.search.Value() != "" {
		leftPane.WriteString("No files match your search.\n\n")
	} else if len(m.filtered) == 0 && !m.searchMode {
		leftPane.WriteString("No files found.\n\n")
	} else if len(m.files) == 0 {
		leftPane.WriteString("No markdown files found.\n\n")
	} else {
		// Show file list
		maxVisible := m.height - 7 // Leave room for header and bottom help
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
			// Truncate if too long for left pane
			maxLen := leftWidth - 3
			if len(displayName) > maxLen {
				displayName = displayName[:maxLen-3] + "..."
			}
			leftPane.WriteString(fmt.Sprintf("%s%s\n", cursor, displayName))
		}
		
		if len(m.filtered) > maxVisible {
			remaining := len(m.filtered) - startIdx - maxVisible
			if remaining > 0 {
				leftPane.WriteString(fmt.Sprintf("\n... %d more\n", remaining))
			}
		}
	}

	// Add search field at bottom of left pane
	leftPane.WriteString("\n")
	if m.searchMode {
		leftPane.WriteString(fmt.Sprintf("Search: %s", m.search.View()))
	}

	// Build right pane (preview)
	rightPane := m.previewContent
	if rightPane == "" {
		rightPane = "Welcome to notes-tui!\n\n• Navigate with ↑↓ or j/k\n• Press Enter to preview a file\n• Press e to edit in external editor\n• Press / to search\n• Press q to quit"
	}

	// Combine panes
	combined := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPaneStyle.Render(leftPane.String()),
		rightPaneStyle.Render(rightPane),
	)

	// Add status line at bottom
	statusLine := "\n"
	if !m.createMode && !m.tagMode {
		if m.searchMode {
			statusLine = "[Esc] cancel | [Enter] preview"
		} else {
			statusLine = "[↑↓/jk] navigate | [Enter] preview | [/] search | [#] tags | [Ctrl+T] tasks | [Ctrl+N] new | [e] edit | [q] quit"
		}
	}

	return combined + "\n" + statusLine
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