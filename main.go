package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/BurntSushi/toml"
)

// Config holds application configuration
type Config struct {
	NotesDirectory string `toml:"notes_directory"`
	Editor         string `toml:"editor"`
	PreviewCommand string `toml:"preview_command"`
	AddFrontmatter bool   `toml:"add_frontmatter"`
	InitialSort    string `toml:"initial_sort"`
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() Config {
	homeDir, _ := os.UserHomeDir()
	notesDir := filepath.Join(homeDir, "notes")
	
	// If ~/notes doesn't exist, use current directory
	if _, err := os.Stat(notesDir); os.IsNotExist(err) {
		cwd, _ := os.Getwd()
		notesDir = cwd
	}
	
	return Config{
		NotesDirectory: notesDir,
		Editor:         "", // Will fall back to $EDITOR
		PreviewCommand: "", // Will use internal preview
		AddFrontmatter: false, // Default to simple markdown headers
	}
}

// LoadConfig loads configuration from file with fallbacks
func LoadConfig() Config {
	config := DefaultConfig()
	
	// Try to find config file
	configPath := getConfigPath()
	if configPath == "" {
		return config
	}
	
	// Try to load and parse config file
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		// If config file has errors, fall back to defaults
		// Could log this in the future
		return DefaultConfig()
	}
	
	return config
}

// getConfigPath returns the path to the config file, or empty string if not found
func getConfigPath() string {
	// Try XDG config dir first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		path := filepath.Join(xdgConfig, "notes-tui", "config.toml")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	
	// Try ~/.config/notes-tui/config.toml
	if homeDir, err := os.UserHomeDir(); err == nil {
		path := filepath.Join(homeDir, ".config", "notes-tui", "config.toml")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	
	return ""
}

// parseCommand splits a command string into command and args, handling quoted arguments
func parseCommand(cmdStr string) (string, []string) {
	if cmdStr == "" {
		return "", nil
	}
	
	// Simple parsing that handles quoted arguments
	var parts []string
	var current strings.Builder
	inQuotes := false
	
	for i, r := range cmdStr {
		switch {
		case r == '"' && (i == 0 || cmdStr[i-1] != '\\'):
			inQuotes = !inQuotes
		case r == ' ' && !inQuotes:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	
	if len(parts) == 0 {
		return "", nil
	}
	
	return parts[0], parts[1:]
}

type model struct {
	files       []string        // all markdown files
	filtered    []string        // filtered results
	cursor      int             // which file is selected
	selected    string          // selected file
	searchMode  bool            // are we in search mode?
	search      textinput.Model // search input
	textFilter  bool            // are we showing only files matching text search?
	createMode  bool            // are we in create mode?
	createInput textinput.Model // create note input
	tagMode     bool            // are we in tag search mode?
	tagInput    textinput.Model // tag search input
	tagFilter   bool            // are we showing only files with a specific tag?
	taskFilter  bool            // are we showing only files with tasks?
	dailyFilter bool            // are we showing only daily note files?
	deleteMode  bool            // are we in delete confirmation mode?
	deleteFile  string          // file to be deleted
	// Sorting state
	sortMode     bool            // are we in sort selection mode?
	currentSort  string          // current sort method: "date", "modified", "title", or ""
	// Days old filter
	oldMode      bool            // are we in days old mode?
	oldInput     textinput.Model // days old input
	oldFilter    bool            // are we showing only files from last N days?
	oldDays      int             // number of days for old filter
	cwd         string          // current working directory
	width       int             // terminal width
	height      int             // terminal height
	config      Config          // application configuration
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

// Message to clear selected file state
type clearSelectedMsg struct{}


func initialModel(startupTag string) model {
	// Load configuration
	config := LoadConfig()
	
	// Use configured notes directory
	cwd := config.NotesDirectory
	if cwd == "" {
		// Fallback to current working directory
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
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

	// Create days old input
	oldi := textinput.New()
	oldi.Placeholder = "Days back..."
	oldi.CharLimit = 3
	oldi.Width = 15

	m := model{
		files:       files,
		filtered:    files, // Initially show all files
		search:      ti,
		createInput: ci,
		tagInput:    tagi,
		oldInput:    oldi,
		cwd:         cwd,
		config:      config,
	}

	// Apply initial sort if configured
	if config.InitialSort != "" {
		switch config.InitialSort {
		case "date", "modified", "title":
			m.currentSort = config.InitialSort
			m.files = m.applySorting(files)
			m.filtered = m.files
		}
	}

	// If a startup tag was provided, apply tag filter
	if startupTag != "" {
		if tagFiles, err := searchTag(cwd, startupTag); err == nil {
			m.filtered = tagFiles
			m.tagFilter = true
			m.cursor = 0
		}
	}

	return m
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

// Generate note content based on configuration
func generateNoteContent(title string, config Config) string {
	if config.AddFrontmatter {
		// YAML frontmatter format
		today := time.Now().Format("2006-01-02")
		return fmt.Sprintf("---\ntitle: %s\ndate: %s\n---\n\n", title, today)
	} else {
		// Simple markdown header format
		return fmt.Sprintf("# %s\n\n", title)
	}
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

// Search for daily note files (matching *-daily.md pattern)
func searchDailyNotes(dir string) ([]string, error) {
	// Get all markdown files first
	allFiles, err := findMarkdownFiles(dir)
	if err != nil {
		return nil, err
	}
	
	// Filter for files matching the daily note pattern
	var dailyFiles []string
	for _, file := range allFiles {
		filename := filepath.Base(file)
		if strings.HasSuffix(filename, "-daily.md") {
			dailyFiles = append(dailyFiles, file)
		}
	}
	
	return dailyFiles, nil
}

// Sort files by different criteria
func sortFilesByDate(files []string) []string {
	sorted := make([]string, len(files))
	copy(sorted, files)
	
	sort.Slice(sorted, func(i, j int) bool {
		// Try to extract date from filename first (YYYY-MM-DD format)
		nameI := filepath.Base(sorted[i])
		nameJ := filepath.Base(sorted[j])
		
		// Check for YYYY-MM-DD pattern at start of filename
		datePattern := `^(\d{4}-\d{2}-\d{2})`
		if matched, _ := regexp.MatchString(datePattern, nameI); matched {
			if matched2, _ := regexp.MatchString(datePattern, nameJ); matched2 {
				return nameI > nameJ // Newer dates first
			}
		}
		
		// Fall back to file modification time
		statI, errI := os.Stat(sorted[i])
		statJ, errJ := os.Stat(sorted[j])
		if errI != nil || errJ != nil {
			return sorted[i] < sorted[j] // Fallback to name sort
		}
		return statI.ModTime().After(statJ.ModTime()) // Newer files first
	})
	
	return sorted
}

func sortFilesByModified(files []string) []string {
	sorted := make([]string, len(files))
	copy(sorted, files)
	
	sort.Slice(sorted, func(i, j int) bool {
		statI, errI := os.Stat(sorted[i])
		statJ, errJ := os.Stat(sorted[j])
		if errI != nil || errJ != nil {
			return sorted[i] < sorted[j] // Fallback to name sort
		}
		return statI.ModTime().After(statJ.ModTime()) // Newer files first
	})
	
	return sorted
}

func sortFilesByTitle(files []string) []string {
	sorted := make([]string, len(files))
	copy(sorted, files)
	
	sort.Slice(sorted, func(i, j int) bool {
		nameI := filepath.Base(sorted[i])
		nameJ := filepath.Base(sorted[j])
		return strings.ToLower(nameI) < strings.ToLower(nameJ)
	})
	
	return sorted
}

// Apply current sort to file list
func (m *model) applySorting(files []string) []string {
	switch m.currentSort {
	case "date":
		return sortFilesByDate(files)
	case "modified":
		return sortFilesByModified(files)
	case "title":
		return sortFilesByTitle(files)
	default:
		return files // No sorting
	}
}

// Filter files by days old (files modified within last N days)
func filterFilesByDaysOld(files []string, days int) []string {
	if days <= 0 {
		return files
	}
	
	cutoff := time.Now().AddDate(0, 0, -days)
	var filtered []string
	
	for _, file := range files {
		if stat, err := os.Stat(file); err == nil {
			if stat.ModTime().After(cutoff) {
				filtered = append(filtered, file)
			}
		}
	}
	
	return filtered
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

	case clearSelectedMsg:
		m.selected = ""
		return m, nil


	case tea.KeyMsg:
		// Handle preview mode separately
		if m.previewMode {
			switch msg.String() {
			case "esc", "q":
				m.previewMode = false
				m.previewContent = ""
				m.previewScroll = 0
				m.selected = "" // Clear selected file to prevent editor opening on quit
				return m, nil
			
			case "e", "ctrl+e":
				// Open in editor from preview
				m.previewMode = false
				m.previewContent = ""
				m.previewScroll = 0
				m.selected = m.previewFile
				return m, tea.ExecProcess(m.openInEditor(), func(err error) tea.Msg {
					return clearSelectedMsg{}
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
			if m.deleteMode {
				// Exit delete mode on q
				m.deleteMode = false
				m.deleteFile = ""
				return m, nil
			}
			return m, tea.Quit

		case "e", "ctrl+e":
			// Open in external editor
			if !m.searchMode && !m.createMode && !m.tagMode && !m.deleteMode && m.cursor < len(m.filtered) {
				m.selected = m.filtered[m.cursor]
				// We'll handle the actual editor opening after we return
				return m, tea.ExecProcess(m.openInEditor(), func(err error) tea.Msg {
					return clearSelectedMsg{}
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
			if m.deleteMode {
				// Exit delete mode
				m.deleteMode = false
				m.deleteFile = ""
			}
			if m.sortMode {
				// Exit sort mode
				m.sortMode = false
			}
			if m.oldMode {
				// Exit old mode
				m.oldMode = false
				m.oldInput.SetValue("")
			}
			if m.taskFilter {
				// Clear task filter
				m.taskFilter = false
				m.filtered = m.files
				m.cursor = 0
			}
			if m.tagFilter {
				// Clear tag filter
				m.tagFilter = false
				m.filtered = m.files
				m.cursor = 0
			}
			if m.textFilter {
				// Clear text filter
				m.textFilter = false
				m.filtered = m.files
				m.cursor = 0
			}
			if m.dailyFilter {
				// Clear daily filter
				m.dailyFilter = false
				m.filtered = m.files
				m.cursor = 0
			}
			if m.oldFilter {
				// Clear old filter
				m.oldFilter = false
				m.filtered = m.files
				m.cursor = 0
			}

		case "n":
			if m.deleteMode {
				// Cancel deletion
				m.deleteMode = false
				m.deleteFile = ""
			} else if !m.searchMode && !m.createMode && !m.tagMode && !m.deleteMode {
				// Enter create mode
				m.createMode = true
				m.createInput.Focus()
				return m, nil
			}

		case "/":
			if !m.createMode && !m.tagMode && !m.deleteMode {
				// Enter search mode
				m.searchMode = true
				m.search.Focus()
				return m, nil
			}

		case "#":
			if !m.searchMode && !m.createMode && !m.tagMode && !m.deleteMode {
				// Enter tag search mode
				m.tagMode = true
				m.tagInput.Focus()
				return m, nil
			}

		case "D":
			if !m.searchMode && !m.createMode && !m.tagMode && !m.deleteMode && !m.sortMode && !m.oldMode {
				// Search for daily notes
				if files, err := searchDailyNotes(m.cwd); err == nil {
					m.filtered = files
					m.cursor = 0
					m.dailyFilter = true
					m.taskFilter = false // Clear task filter when switching to daily filter
					m.tagFilter = false // Clear tag filter when switching to daily filter
					m.textFilter = false // Clear text filter when switching to daily filter
					m.oldFilter = false // Clear old filter when switching to daily filter
				}
			}

		case "X":
			if !m.searchMode && !m.createMode && !m.tagMode && !m.deleteMode && !m.sortMode && !m.oldMode && m.cursor < len(m.filtered) {
				// Enter delete confirmation mode
				m.deleteMode = true
				m.deleteFile = m.filtered[m.cursor]
			}

		case "o":
			if !m.searchMode && !m.createMode && !m.tagMode && !m.deleteMode && !m.oldMode {
				// Enter sort mode
				m.sortMode = true
			}

		case "O":
			if !m.searchMode && !m.createMode && !m.tagMode && !m.deleteMode && !m.sortMode {
				// Enter old (days back) mode
				m.oldMode = true
				m.oldInput.Focus()
				return m, nil
			}

		case "d":
			if m.sortMode {
				// Sort by date
				m.currentSort = "date"
				m.files = m.applySorting(m.files)
				m.filtered = m.applySorting(m.filtered)
				m.sortMode = false
				m.cursor = 0
			} else if !m.searchMode && !m.createMode && !m.tagMode && !m.deleteMode && !m.sortMode && !m.oldMode {
				// Create or open daily note (existing functionality)
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
						return m, tea.ExecProcess(m.openInEditor(), func(err error) tea.Msg {
							return clearSelectedMsg{}
						})
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
					return m, tea.ExecProcess(m.openInEditor(), func(err error) tea.Msg {
						return clearSelectedMsg{}
					})
				}
			}

		case "m":
			if m.sortMode {
				// Sort by modified
				m.currentSort = "modified"
				m.files = m.applySorting(m.files)
				m.filtered = m.applySorting(m.filtered)
				m.sortMode = false
				m.cursor = 0
			}

		case "t":
			if m.sortMode {
				// Sort by title
				m.currentSort = "title"
				m.files = m.applySorting(m.files)
				m.filtered = m.applySorting(m.filtered)
				m.sortMode = false
				m.cursor = 0
			} else if !m.searchMode && !m.createMode && !m.tagMode && !m.deleteMode && !m.sortMode && !m.oldMode {
				// Search for tasks (existing functionality)
				if files, err := searchTasks(m.cwd); err == nil {
					m.filtered = files
					m.cursor = 0
					m.taskFilter = true
					m.tagFilter = false // Clear tag filter when switching to task filter
					m.textFilter = false // Clear text filter when switching to task filter
					m.dailyFilter = false // Clear daily filter when switching to task filter
					m.oldFilter = false // Clear old filter when switching to task filter
				}
			}

		case "y":
			if m.deleteMode {
				// Confirm deletion
				if err := os.Remove(m.deleteFile); err == nil {
					// Successfully deleted, refresh file list
					files, _ := findMarkdownFiles(m.cwd)
					m.files = files
					
					// If we had filters applied, reapply them
					if m.taskFilter {
						if taskFiles, err := searchTasks(m.cwd); err == nil {
							m.filtered = taskFiles
						} else {
							m.filtered = files
						}
					} else if m.dailyFilter {
						if dailyFiles, err := searchDailyNotes(m.cwd); err == nil {
							m.filtered = dailyFiles
						} else {
							m.filtered = files
						}
					} else if m.search.Value() != "" {
						m.filtered = filterFiles(files, m.search.Value())
					} else {
						m.filtered = files
					}
					
					// Adjust cursor position
					if m.cursor >= len(m.filtered) {
						if len(m.filtered) > 0 {
							m.cursor = len(m.filtered) - 1
						} else {
							m.cursor = 0
						}
					}
				}
				// Exit delete mode regardless of success/failure
				m.deleteMode = false
				m.deleteFile = ""
			}

		case "up", "k":
			if !m.searchMode && !m.createMode && !m.tagMode && !m.deleteMode && m.cursor > 0 {
				m.cursor--
				// Don't auto-load preview on cursor movement
			}

		case "down", "j":
			if !m.searchMode && !m.createMode && !m.tagMode && !m.deleteMode && m.cursor < len(m.filtered)-1 {
				m.cursor++
				// Don't auto-load preview on cursor movement
			}

		case "enter":
			if m.deleteMode {
				// Don't delete on enter - require explicit 'y' confirmation
				return m, nil
			} else if m.tagMode {
				// Search for the tag
				tag := m.tagInput.Value()
				if tag != "" {
					if files, err := searchTag(m.cwd, tag); err == nil {
						m.filtered = files
						m.cursor = 0
						m.tagFilter = true // Set tag filter active
						m.taskFilter = false // Clear task filter when switching to tag filter
						m.textFilter = false // Clear text filter when switching to tag filter
						m.dailyFilter = false // Clear daily filter when switching to tag filter
						m.oldFilter = false // Clear old filter when switching to tag filter
					}
				}
				// Exit tag mode
				m.tagMode = false
				m.tagInput.SetValue("")
			} else if m.oldMode {
				// Apply days old filter
				daysStr := m.oldInput.Value()
				if daysStr != "" {
					if days, err := strconv.Atoi(daysStr); err == nil && days > 0 {
						filteredFiles := filterFilesByDaysOld(m.files, days)
						m.filtered = filteredFiles
						m.cursor = 0
						m.oldFilter = true
						m.oldDays = days
						// Clear other filters when switching to old filter
						m.taskFilter = false
						m.tagFilter = false
						m.textFilter = false
						m.dailyFilter = false
					}
				}
				// Exit old mode
				m.oldMode = false
				m.oldInput.SetValue("")
			} else if m.createMode {
				// Create the new note
				title := m.createInput.Value()
				if title != "" {
					filename := titleToFilename(title)
					fullPath := filepath.Join(m.cwd, filename)
					
					// Create the file with templated content
					content := generateNoteContent(title, m.config)
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
						// Exit create mode and open editor
						m.createMode = false
						m.createInput.SetValue("")
						return m, tea.ExecProcess(m.openInEditor(), func(err error) tea.Msg {
							return clearSelectedMsg{}
						})
					}
				}
				// Exit create mode
				m.createMode = false
				m.createInput.SetValue("")
			} else if m.searchMode {
				// Exit search mode on enter, but keep the filter active
				m.searchMode = false
				// If there's a search query, mark text filter as active
				if m.search.Value() != "" {
					m.textFilter = true
					m.taskFilter = false // Clear other filters
					m.tagFilter = false
					m.dailyFilter = false
				}
			} else if !m.deleteMode && m.cursor < len(m.filtered) {
				// Preview: use external if configured, otherwise internal
				m.selected = m.filtered[m.cursor]
				if m.config.PreviewCommand != "" {
					// Use external preview
					return m, tea.ExecProcess(m.openInPreview(), func(err error) tea.Msg {
						return clearSelectedMsg{}
					})
				} else {
					// Use internal preview popover
					m.previewFile = m.selected
					m.previewMode = true
					m.previewScroll = 0
					return m, m.loadPreviewForPopover()
				}
			}
		}
	}

	// Handle search input
	if m.searchMode {
		m.search, cmd = m.search.Update(msg)
		query := m.search.Value()
		m.filtered = filterFiles(m.files, query)
		m.cursor = 0 // Reset cursor when filtering
		// Clear other filters when doing a text search
		m.taskFilter = false
		m.tagFilter = false
		m.textFilter = false // Also clear text filter since we're in live search mode
		m.dailyFilter = false
		m.oldFilter = false
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

	// Handle old input
	if m.oldMode {
		m.oldInput, cmd = m.oldInput.Update(msg)
	}

	return m, tea.Batch(cmds...)
}

// Open file in external editor
func (m model) openInEditor() *exec.Cmd {
	var editor string
	var args []string
	
	// Use configured editor first
	if m.config.Editor != "" {
		editor, args = parseCommand(m.config.Editor)
	} else {
		// Fall back to $EDITOR environment variable
		editor = os.Getenv("EDITOR")
		if editor == "" {
			// Try common editors as last resort
			editors := []string{"vim", "nvim", "emacs", "nano"}
			for _, e := range editors {
				if _, err := exec.LookPath(e); err == nil {
					editor = e
					break
				}
			}
		}
	}
	
	if editor == "" {
		editor = "vi" // ultimate fallback
	}
	
	// Add the filename to the arguments
	args = append(args, m.selected)
	
	cmd := exec.Command(editor, args...)
	return cmd
}

// Open file in external preview command
func (m model) openInPreview() *exec.Cmd {
	if m.config.PreviewCommand == "" {
		return nil
	}
	
	command, args := parseCommand(m.config.PreviewCommand)
	if command == "" {
		return nil
	}
	
	// Add the filename to the arguments
	args = append(args, m.selected)
	
	cmd := exec.Command(command, args...)
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
	if m.tagFilter {
		header += " - [Tag]"
	}
	if m.textFilter {
		header += " - [Search]"
	}
	if m.dailyFilter {
		header += " - [Daily]"
	}
	if m.oldFilter {
		header += fmt.Sprintf(" - [Last %d days]", m.oldDays)
	}
	if m.currentSort != "" {
		switch m.currentSort {
		case "date":
			header += " - [Sort: Date]"
		case "modified":
			header += " - [Sort: Modified]"
		case "title":
			header += " - [Sort: Title]"
		}
	}
	content.WriteString(lipgloss.NewStyle().Bold(true).Render(header) + "\n\n")

	if m.tagMode {
		// Tag search mode
		content.WriteString("Search by Tag\n\n")
		content.WriteString(fmt.Sprintf("Tag: %s\n\n", m.tagInput.View()))
		content.WriteString("[Enter] search [Esc] cancel")
	} else if m.sortMode {
		// Sort selection mode
		content.WriteString("Sort Files\n\n")
		content.WriteString("Choose sort method:\n")
		content.WriteString("[d] Date  [m] Modified  [t] Title\n\n")
		content.WriteString("[Esc] cancel")
	} else if m.oldMode {
		// Days old mode
		content.WriteString("Filter by Days Old\n\n")
		content.WriteString(fmt.Sprintf("Days back: %s\n\n", m.oldInput.View()))
		content.WriteString("[Enter] apply filter [Esc] cancel")
	} else if m.createMode {
		// Create mode
		content.WriteString("Create New Note\n\n")
		content.WriteString(fmt.Sprintf("Title: %s\n\n", m.createInput.View()))
		content.WriteString("[Enter] create [Esc] cancel")
	} else if m.deleteMode {
		// Delete confirmation mode
		filename := getDisplayName(m.deleteFile, m.cwd)
		content.WriteString("Delete Note\n\n")
		content.WriteString(fmt.Sprintf("Delete '%s'?\n\n", filename))
		content.WriteString("[y] yes [n] no [Esc] cancel")
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
	if !m.tagMode && !m.createMode && !m.deleteMode {
		content.WriteString("\n")
		if m.searchMode {
			content.WriteString(fmt.Sprintf("Search: %s", m.search.View()))
		} else {
			// Help text
			helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			help := "[/] search  [Enter] preview  [e] edit  [X] delete  [n] new  [d] daily  [D] all daily  [t] tasks  [#] tags  [o] sort  [O] days old  [q] quit"
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
	// Parse command line flags
	var tag = flag.String("tag", "", "Filter notes by tag (e.g., --tag=@mikeh)")
	flag.Parse()

	// Load config first
	config := LoadConfig()
	
	// Handle directory argument (remaining args after flags)
	args := flag.Args()
	if len(args) > 0 {
		dir := args[0]
		if err := os.Chdir(dir); err != nil {
			log.Fatal(err)
		}
		// Update config to use the specified directory
		if cwd, err := os.Getwd(); err == nil {
			config.NotesDirectory = cwd
		}
	} else if config.NotesDirectory != "" {
		// Use configured directory
		if err := os.Chdir(config.NotesDirectory); err != nil {
			log.Printf("Warning: Could not change to configured directory %s: %v", config.NotesDirectory, err)
			// Continue with current directory
		}
	}

	p := tea.NewProgram(initialModel(*tag), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	// If a file was selected, open it in editor
	if m, ok := m.(model); ok && m.selected != "" {
		var editor string
		var args []string
		
		// Use configured editor first
		if m.config.Editor != "" {
			editor, args = parseCommand(m.config.Editor)
		} else {
			// Fall back to $EDITOR environment variable
			editor = os.Getenv("EDITOR")
			if editor == "" {
				// Try common editors as last resort
				editors := []string{"vim", "nvim", "nano", "emacs", "code"}
				for _, e := range editors {
					if _, err := exec.LookPath(e); err == nil {
						editor = e
						break
					}
				}
			}
		}
		
		if editor == "" {
			fmt.Println("No editor found. Please set $EDITOR environment variable or configure editor in config file.")
			fmt.Printf("Selected file: %s\n", m.selected)
			return
		}

		// Add the filename to the arguments
		args = append(args, m.selected)

		// Open the file in the editor
		cmd := exec.Command(editor, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error opening editor: %v\n", err)
			fmt.Printf("Selected file: %s\n", m.selected)
		}
	}
}