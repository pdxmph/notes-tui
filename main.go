package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
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
	"github.com/pdxmph/notes-tui/internal/ui"
	"github.com/pdxmph/notes-tui/internal/denote"
)

// Config holds application configuration
type Config struct {
	NotesDirectory     string `toml:"notes_directory"`
	TasksDirectory     string `toml:"tasks_directory"`
	Editor             string `toml:"editor"`
	PreviewCommand     string `toml:"preview_command"`
	AddFrontmatter     bool   `toml:"add_frontmatter"`
	InitialSort        string `toml:"initial_sort"`
	InitialReverseSort bool   `toml:"initial_reverse_sort"`
	DenoteFilenames    bool   `toml:"denote_filenames"`
	ShowTitles         bool   `toml:"show_titles"`
	PromptForTags      bool   `toml:"prompt_for_tags"`
	TaskwarriorSupport bool   `toml:"taskwarrior_support"`
	DenoteTasksSupport bool   `toml:"denote_tasks_support"`
	Theme              string `toml:"theme"`
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
		NotesDirectory:     notesDir,
		TasksDirectory:     "", // Will default to NotesDirectory if empty
		Editor:             "", // Will fall back to $EDITOR
		PreviewCommand:     "", // Will use internal preview
		AddFrontmatter:     false, // Default to simple markdown headers
		InitialReverseSort: false, // Default to normal sort order
		TaskwarriorSupport: false, // Default to disabled
		DenoteTasksSupport: false, // Default to disabled
		Theme:              "default", // Default theme
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
	// Tag creation state
	tagCreateMode  bool            // are we prompting for tags during note creation?
	tagCreateInput textinput.Model // tag input during note creation
	pendingTitle   string          // title waiting for tags before creating note
	// Task creation state
	taskCreateMode  bool            // are we creating a TaskWarrior task?
	taskCreateInput textinput.Model // task description input
	// Sorting state
	sortMode     bool            // are we in sort selection mode?
	currentSort  string          // current sort method: "date", "modified", "title", "denote", or ""
	reversedSort bool            // is the current sort reversed?
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
	// Rename state
	renameMode     bool            // are we renaming a file to Denote format?
	renameFile     string          // file being renamed
	// Navigation state
	waitingForSecondG bool          // waiting for second 'g' in 'gg' sequence
	// UI integration
	ui              *ui.ModelIntegration
	// Task mode state
	taskModeActive  bool            // are we in task management mode?
	tasks           []*denote.Task  // loaded tasks when in task mode
	projects        []*denote.Project // loaded projects when showing projects
	taskSortBy      string          // task-specific sort: "priority", "due", "status", "id"
	taskFilterMode  bool            // are we in task filter selection mode?
	taskFilterType  string          // current task filter: "all", "open", "projects", etc.
	taskFilterProject string        // specific project to filter by
	projectsMode    bool            // are we in projects list mode?
	// Area context - persists across other filters
	taskAreaContext string          // current area context (empty means all areas)
	taskStatusFilter string         // status filter applied within area context
	// Area selection mode
	areaSelectMode  bool            // are we in area selection mode?
	availableAreas  []string        // list of available areas
	areaSelectCursor int            // cursor position in area list
	// Task edit mode
	taskEditMode    bool            // are we editing task metadata?
	taskEditField   string          // which field is being edited: "due", "start", "estimate", "priority", "project", "area"
	taskEditInput   textinput.Model // input for current field
	taskBeingEdited *denote.Task    // pointer to task being edited
}

// Message for preview content
type previewLoadedMsg struct {
	content  string
	filepath string
}

// Message to clear selected file state
type clearSelectedMsg struct{}

// getTasksDirectory returns the effective tasks directory
func (m model) getTasksDirectory() string {
	if m.config.TasksDirectory != "" {
		return m.config.TasksDirectory
	}
	return m.cwd // Fall back to notes directory
}

func initialModel(config Config, startupTag string, startInTaskMode bool, startupArea string) model {
	
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

	// Create tag creation input
	tagci := textinput.New()
	tagci.Placeholder = "Tags (comma-separated, optional)..."
	tagci.CharLimit = 200
	tagci.Width = 50

	// Create days old input
	oldi := textinput.New()
	oldi.Placeholder = "Days back..."
	oldi.CharLimit = 3
	oldi.Width = 15
	
	// Create task input
	taski := textinput.New()
	taski.Placeholder = "Task description..."
	taski.CharLimit = 200
	taski.Width = 50
	
	// Create task edit input
	taskEditi := textinput.New()
	taskEditi.CharLimit = 100
	taskEditi.Width = 40

	m := model{
		files:          files,
		filtered:       files, // Initially show all files
		search:         ti,
		createInput:    ci,
		tagInput:       tagi,
		tagCreateInput: tagci,
		oldInput:       oldi,
		taskCreateInput: taski,
		taskEditInput:  taskEditi,
		cwd:            cwd,
		config:         config,
		reversedSort:   config.InitialReverseSort,
	}

	// Apply initial sort if configured
	if config.InitialSort != "" {
		switch config.InitialSort {
		case "date", "modified", "title", "denote":
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

	// Initialize UI integration
	m.ui = &ui.ModelIntegration{
		Files:              m.files,
		Filtered:           m.filtered,
		Cursor:             m.cursor,
		CWD:                m.cwd,
		ShowTitles:         config.ShowTitles,
		DenoteFilenames:    config.DenoteFilenames,
		TaskwarriorSupport: config.TaskwarriorSupport,
		DenoteTasksSupport: config.DenoteTasksSupport,
		ThemeName:          config.Theme,
		Search:             m.search,
		CreateInput:        m.createInput,
		TagInput:           m.tagInput,
		TagCreateInput:     m.tagCreateInput,
		TaskCreateInput:    m.taskCreateInput,
		OldInput:           m.oldInput,
	}

	// Initialize UI
	m.ui.Initialize()

	// Start in task mode if requested AND if task support is enabled
	if startInTaskMode && config.DenoteTasksSupport {
		// Load tasks directly
		scanner := denote.NewScanner(m.getTasksDirectory())
		tasks, err := scanner.FindTasks()
		
		if err == nil && len(tasks) > 0 {
			m.taskModeActive = true
			m.tasks = tasks
			m.taskSortBy = "priority"
			denote.SortTasks(m.tasks, m.taskSortBy, m.reversedSort)
			
			// Update file lists
			m.files = make([]string, len(m.tasks))
			m.filtered = make([]string, len(m.tasks))
			for i, task := range m.tasks {
				m.files[i] = task.Path
				m.filtered[i] = task.Path
			}
			m.cursor = 0
			
			// Set task formatter
			m.ui.TaskFormatter = func(path string) string {
				task := m.findTaskByPath(path)
				if task != nil {
					return m.formatTaskLine(task)
				}
				return filepath.Base(path)
			}
			m.ui.TaskModeActive = true
			
			// Apply area filter if specified
			if startupArea != "" {
				// Check if the area exists in tasks
				areaExists := false
				for _, task := range m.tasks {
					if task.Area == startupArea {
						areaExists = true
						break
					}
				}
				
				if areaExists {
					m.taskAreaContext = startupArea
					m.refreshTaskView()
					// Update UI components with area context
					m.ui.TaskAreaContext = startupArea
				}
			}
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

// Extract title from a note file
func extractNoteTitle(filepath string) string {
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

// Generate Denote-style filename from title
func generateDenoteName(title string, tags []string, timestamp time.Time) (filename string, identifier string) {
	// Use provided timestamp
	identifier = timestamp.Format("20060102T150405")
	
	// Sanitize title for filename
	// Convert to lowercase
	sanitized := strings.ToLower(title)
	
	// Replace spaces with hyphens
	sanitized = strings.ReplaceAll(sanitized, " ", "-")
	
	// Remove special characters, keep only alphanumeric and hyphens
	var result strings.Builder
	for _, ch := range sanitized {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
			result.WriteRune(ch)
		}
	}
	
	// Clean up multiple hyphens
	cleaned := regexp.MustCompile(`-+`).ReplaceAllString(result.String(), "-")
	cleaned = strings.Trim(cleaned, "-")
	
	// If empty after sanitization, use "untitled"
	if cleaned == "" {
		cleaned = "untitled"
	}
	
	// Process tags if provided
	if len(tags) > 0 {
		var sanitizedTags []string
		for _, tag := range tags {
			// Sanitize each tag similar to title
			tagLower := strings.ToLower(tag)
			tagLower = strings.ReplaceAll(tagLower, " ", "-")
			
			var tagResult strings.Builder
			for _, ch := range tagLower {
				if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
					tagResult.WriteRune(ch)
				}
			}
			
			tagCleaned := regexp.MustCompile(`-+`).ReplaceAllString(tagResult.String(), "-")
			tagCleaned = strings.Trim(tagCleaned, "-")
			
			if tagCleaned != "" {
				sanitizedTags = append(sanitizedTags, tagCleaned)
			}
		}
		
		// Join tags with single underscore and append with double underscore
		if len(sanitizedTags) > 0 {
			tagString := strings.Join(sanitizedTags, "_")
			filename = fmt.Sprintf("%s-%s__%s.md", identifier, cleaned, tagString)
			return filename, identifier
		}
	}
	
	filename = fmt.Sprintf("%s-%s.md", identifier, cleaned)
	return filename, identifier
}

// Parse Denote filename to extract title
func parseDenoteFilename(filename string) (title string, timestamp time.Time) {
	base := strings.TrimSuffix(filename, ".md")
	
	// Expected format: YYYYMMDDTHHMMSS-title
	if len(base) < 16 { // Minimum length for timestamp + hyphen
		return filename, time.Time{}
	}
	
	// Check if it matches Denote pattern
	if base[8] == 'T' && base[15] == '-' {
		timestampStr := base[:15]
		titlePart := base[16:]
		
		// Try to parse timestamp
		t, err := time.Parse("20060102T150405", timestampStr)
		if err == nil {
			// Convert hyphens back to spaces (no capitalization)
			title := strings.ReplaceAll(titlePart, "-", " ")
			return title, t
		}
	}
	
	return filename, time.Time{}
}

// Rename a file to Denote format
func renameToDenoteName(filepath string, config Config) (string, error) {
	// Check if file already has Denote format
	filename := path.Base(filepath)
	denotePattern := regexp.MustCompile(`^\d{8}T\d{6}-`)
	var existingTimestamp string
	
	if denotePattern.MatchString(filename) {
		// Extract existing timestamp
		existingTimestamp = filename[:15]
	}
	
	// Read file to extract title and tags
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	
	// Extract title - use the extractNoteTitle function
	title := extractNoteTitle(filepath)
	
	// Extract tags from frontmatter
	var tags []string
	var yamlDate *time.Time
	lines := strings.Split(string(content), "\n")
	inFrontmatter := false
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Check for frontmatter boundaries
		if i == 0 && trimmed == "---" {
			inFrontmatter = true
			continue
		}
		if inFrontmatter && trimmed == "---" {
			break
		}
		
		// Look for date in frontmatter
		if inFrontmatter && strings.HasPrefix(trimmed, "date:") {
			datePart := strings.TrimPrefix(trimmed, "date:")
			datePart = strings.TrimSpace(datePart)
			
			// Try to parse the date - support multiple formats
			dateFormats := []string{
				"2006-01-02T15:04:05Z07:00", // ISO 8601 with timezone
				"2006-01-02T15:04:05",       // ISO 8601 without timezone
				"2006-01-02 15:04:05",       // Space separated datetime
				"2006-01-02",                // Date only
				"01/02/2006",                // US format
				"02/01/2006",                // European format
			}
			
			for _, format := range dateFormats {
				if parsedDate, err := time.Parse(format, datePart); err == nil {
					yamlDate = &parsedDate
					break
				}
			}
		}
		
		// Look for tags in frontmatter
		if inFrontmatter && strings.HasPrefix(trimmed, "tags:") {
			// Handle array format: tags: [tag1, tag2, tag3]
			if strings.Contains(trimmed, "[") {
				// Extract tags from array format
				tagsPart := strings.TrimPrefix(trimmed, "tags:")
				tagsPart = strings.TrimSpace(tagsPart)
				tagsPart = strings.Trim(tagsPart, "[]")
				
				if tagsPart != "" {
					// Split by comma and clean each tag
					tagList := strings.Split(tagsPart, ",")
					for _, tag := range tagList {
						cleaned := strings.TrimSpace(tag)
						if cleaned != "" {
							tags = append(tags, cleaned)
						}
					}
				}
			}
		}
	}
	
	// Generate new filename
	var newFilename string
	var identifier string
	
	if existingTimestamp != "" {
		// Preserve existing timestamp
		identifier = existingTimestamp
		
		// Sanitize title for filename
		titleLower := strings.ToLower(title)
		titleLower = strings.ReplaceAll(titleLower, " ", "-")
		
		var result strings.Builder
		for _, ch := range titleLower {
			if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
				result.WriteRune(ch)
			}
		}
		
		cleaned := regexp.MustCompile(`-+`).ReplaceAllString(result.String(), "-")
		cleaned = strings.Trim(cleaned, "-")
		
		if cleaned == "" {
			cleaned = "untitled"
		}
		
		// Build filename with tags if present
		if len(tags) > 0 {
			var sanitizedTags []string
			for _, tag := range tags {
				tagLower := strings.ToLower(tag)
				tagLower = strings.ReplaceAll(tagLower, " ", "-")
				
				var tagResult strings.Builder
				for _, ch := range tagLower {
					if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
						tagResult.WriteRune(ch)
					}
				}
				
				tagCleaned := regexp.MustCompile(`-+`).ReplaceAllString(tagResult.String(), "-")
				tagCleaned = strings.Trim(tagCleaned, "-")
				
				if tagCleaned != "" {
					sanitizedTags = append(sanitizedTags, tagCleaned)
				}
			}
			
			if len(sanitizedTags) > 0 {
				tagString := strings.Join(sanitizedTags, "_")
				newFilename = fmt.Sprintf("%s-%s__%s.md", identifier, cleaned, tagString)
			} else {
				newFilename = fmt.Sprintf("%s-%s.md", identifier, cleaned)
			}
		} else {
			newFilename = fmt.Sprintf("%s-%s.md", identifier, cleaned)
		}
	} else {
		// Get file modification time to preserve original timestamp
		fileInfo, err := os.Stat(filepath)
		if err != nil {
			return "", fmt.Errorf("failed to get file info: %w", err)
		}
		
		// Use modification time as the default timestamp
		timestamp := fileInfo.ModTime()
		
		// If we have a YAML date and it's earlier than the mod time, use it
		if yamlDate != nil && yamlDate.Before(timestamp) {
			timestamp = *yamlDate
		}
		
		// Use the earlier timestamp for the identifier
		newFilename, identifier = generateDenoteName(title, tags, timestamp)
	}
	
	// Create new path
	dir := path.Dir(filepath)
	newPath := path.Join(dir, newFilename)
	
	// Check if target already exists
	if filepath != newPath {
		if _, err := os.Stat(newPath); err == nil {
			return "", fmt.Errorf("file already exists: %s", newFilename)
		}
		
		// Rename the file
		if err := os.Rename(filepath, newPath); err != nil {
			return "", fmt.Errorf("failed to rename file: %w", err)
		}
	}
	
	return newPath, nil
}

// Get enhanced display name for a file (with title extraction if enabled)
func getEnhancedDisplayName(fullPath, cwd string, showTitles bool) string {
	// Get the basic display name first
	basicName := getDisplayName(fullPath, cwd)
	
	if !showTitles {
		return basicName
	}
	
	// Try to extract title
	title := extractNoteTitle(fullPath)
	
	// If we got a title from the file content, use it
	if title != "" {
		// Check if it's a Denote file to add date
		filename := filepath.Base(fullPath)
		if _, timestamp := parseDenoteFilename(filename); !timestamp.IsZero() {
			return fmt.Sprintf("%s (%s)", title, timestamp.Format("2006-01-02"))
		}
		return title
	}
	
	// If no title in content, try parsing Denote filename
	filename := filepath.Base(fullPath)
	if denoteTitle, timestamp := parseDenoteFilename(filename); denoteTitle != filename && !timestamp.IsZero() {
		return fmt.Sprintf("%s (%s)", denoteTitle, timestamp.Format("2006-01-02"))
	}
	
	// Fall back to basic display name
	return basicName
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
func generateNoteContent(title string, config Config, identifier string, tags []string) string {
	if config.AddFrontmatter {
		// YAML frontmatter format
		today := time.Now().Format("2006-01-02")
		
		// Start building frontmatter
		frontmatter := "---\n"
		frontmatter += fmt.Sprintf("title: %s\n", title)
		frontmatter += fmt.Sprintf("date: %s\n", today)
		
		// Add identifier if using Denote style
		if config.DenoteFilenames && identifier != "" {
			frontmatter += fmt.Sprintf("identifier: %s\n", identifier)
		}
		
		// Add tags if provided
		if len(tags) > 0 {
			// Format as YAML array
			tagList := strings.Join(tags, ", ")
			frontmatter += fmt.Sprintf("tags: [%s]\n", tagList)
		}
		
		frontmatter += "---\n\n"
		return frontmatter
	} else {
		// Simple markdown header format
		return fmt.Sprintf("# %s\n\n", title)
	}
}

// generateTaskContent generates content for a new task file
func generateTaskContent(title string, config Config, identifier string, taskID int) string {
	today := time.Now().Format("2006-01-02")
	
	// Always use frontmatter for tasks
	frontmatter := "---\n"
	frontmatter += fmt.Sprintf("title: %s\n", title)
	frontmatter += fmt.Sprintf("date: %s\n", today)
	
	// Add identifier if using Denote style
	if config.DenoteFilenames && identifier != "" {
		frontmatter += fmt.Sprintf("identifier: %s\n", identifier)
	}
	
	// Task-specific fields
	frontmatter += "tags: [task]\n"
	frontmatter += "status: open\n"
	frontmatter += "priority: p2\n"
	
	// Use the provided sequential task ID
	frontmatter += fmt.Sprintf("task_id: %d\n", taskID)
	
	frontmatter += "---\n\n"
	frontmatter += fmt.Sprintf("# %s\n\n", title)
	frontmatter += "## Description\n\n"
	frontmatter += "## Tasks\n\n"
	frontmatter += "- [ ] \n\n"
	frontmatter += "## Notes\n\n"
	
	return frontmatter
}

// Get today's daily note filename
// Generate daily note filename based on config
func getDailyNoteFilename(config Config) (string, string) {
	if config.DenoteFilenames {
		// Use Denote format for daily notes
		// Tags could be "daily" by default
		filename, identifier := generateDenoteName("daily", []string{"daily"}, time.Now())
		return filename, identifier
	}
	// Traditional format
	today := time.Now().Format("2006-01-02")
	return today + "-daily.md", ""
}

// Find existing daily note for today regardless of format
func findTodaysDailyNote(dir string) (string, error) {
	today := time.Now()
	
	// Pattern 1: Traditional format (YYYY-MM-DD-daily.md)
	traditionalName := today.Format("2006-01-02") + "-daily.md"
	traditionalPath := filepath.Join(dir, traditionalName)
	if _, err := os.Stat(traditionalPath); err == nil {
		return traditionalPath, nil
	}
	
	// Pattern 2: Denote format (YYYYMMDDTHHMMSS-daily*.md)
	// Need to search for files starting with today's date in Denote format
	todayDenote := today.Format("20060102")
	
	// Get all markdown files
	files, err := findMarkdownFiles(dir)
	if err != nil {
		return "", err
	}
	
	// Check each file for Denote daily note pattern
	denotePattern := regexp.MustCompile(fmt.Sprintf(`^%sT\d{6}-daily.*\.md$`, todayDenote))
	
	for _, file := range files {
		filename := filepath.Base(file)
		if denotePattern.MatchString(filename) {
			return file, nil
		}
	}
	
	// No daily note found for today
	return "", os.ErrNotExist
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

// Search for daily note files (matching *-daily.md pattern or Denote files with __daily tag)
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
		// Check for traditional daily note format (*-daily.md)
		if strings.HasSuffix(filename, "-daily.md") {
			dailyFiles = append(dailyFiles, file)
		} else if strings.Contains(filename, "__") && strings.Contains(filename, "_daily") {
			// Check for Denote format with daily tag (e.g., 20250623T094530-daily__daily.md)
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

func sortFilesByDenoteIdentifier(files []string) []string {
	sorted := make([]string, len(files))
	copy(sorted, files)
	
	sort.Slice(sorted, func(i, j int) bool {
		nameI := filepath.Base(sorted[i])
		nameJ := filepath.Base(sorted[j])
		
		// Extract Denote identifier from filename (YYYYMMDDTHHMMSS)
		denotePattern := regexp.MustCompile(`^(\d{8}T\d{6})-`)
		
		matchI := denotePattern.FindStringSubmatch(nameI)
		matchJ := denotePattern.FindStringSubmatch(nameJ)
		
		// If both have Denote identifiers, compare them
		if len(matchI) > 1 && len(matchJ) > 1 {
			return matchI[1] > matchJ[1] // Newer identifiers first
		}
		
		// If only one has identifier, it comes first
		if len(matchI) > 1 && len(matchJ) <= 1 {
			return true
		}
		if len(matchI) <= 1 && len(matchJ) > 1 {
			return false
		}
		
		// If neither has identifier, fall back to filename sort
		return strings.ToLower(nameI) < strings.ToLower(nameJ)
	})
	
	return sorted
}

// Apply current sort to file list
func (m *model) applySorting(files []string) []string {
	var sorted []string
	
	switch m.currentSort {
	case "date":
		sorted = sortFilesByDate(files)
	case "modified":
		sorted = sortFilesByModified(files)
	case "title":
		sorted = sortFilesByTitle(files)
	case "denote":
		sorted = sortFilesByDenoteIdentifier(files)
	default:
		sorted = files // No sorting
	}
	
	// Apply reverse if enabled
	if m.reversedSort && len(sorted) > 0 {
		reversed := make([]string, len(sorted))
		for i, file := range sorted {
			reversed[len(sorted)-1-i] = file
		}
		return reversed
	}
	
	return sorted
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
		rendered := ui.RenderSimpleMarkdown(string(content), popoverContentWidth)
		
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
		// Update UI dimensions
		if m.ui != nil {
			m.ui.UpdateSize(msg.Width, msg.Height)
		}
		// If in preview mode, reload with new dimensions
		if m.previewMode {
			return m, m.loadPreviewForPopover()
		}
		return m, nil

	case previewLoadedMsg:
		m.previewContent = msg.content
		return m, nil

	case ui.StatusMsg:
		if m.ui != nil {
			m.ui.HandleStatusMsg(msg)
		}
		return m, nil

	case ui.ClearStatusMsg:
		if m.ui != nil {
			m.ui.HandleClearStatusMsg()
		}
		return m, nil

	case clearSelectedMsg:
		// Remember the file that was just edited (if any)
		previousFile := m.selected
		m.selected = ""
		
		// Refresh based on current mode
		if m.taskModeActive {
			// Reload tasks
			scanner := denote.NewScanner(m.getTasksDirectory())
			tasks, err := scanner.FindTasks()
			if err != nil {
				return m, nil
			}
			
			m.tasks = tasks
			
			// Apply current task sorting
			if m.taskSortBy != "" {
				denote.SortTasks(m.tasks, m.taskSortBy, m.reversedSort)
			}
			
			// Refresh the task view to apply any filters
			m.refreshTaskView()
			
			// Ensure task formatter is set
			m.ui.TaskFormatter = func(path string) string {
				task := m.findTaskByPath(path)
				if task != nil {
					return m.formatTaskLine(task)
				}
				return filepath.Base(path)
			}
		} else {
			// Normal mode - refresh markdown files
			files, err := findMarkdownFiles(m.cwd)
			if err != nil {
				return m, nil
			}
			
			// Apply current sort
			m.files = m.applySorting(files)
			
			// Reapply any active filters
			if m.taskFilter {
				if taskFiles, err := searchTasks(m.cwd); err == nil {
					m.filtered = taskFiles
				}
			} else if m.tagFilter && m.tagInput.Value() != "" {
				if tagFiles, err := searchTag(m.cwd, m.tagInput.Value()); err == nil {
					m.filtered = tagFiles
				}
			} else if m.textFilter && m.search.Value() != "" {
				m.filtered = filterFiles(m.files, m.search.Value())
			} else if m.dailyFilter {
				if dailyFiles, err := searchDailyNotes(m.cwd); err == nil {
					m.filtered = dailyFiles
				}
			} else if m.oldFilter {
				m.filtered = filterFilesByDaysOld(m.files, m.oldDays)
			} else {
				// No filter active, use all files
				m.filtered = m.files
			}
			
			// Apply sorting to filtered list
			m.filtered = m.applySorting(m.filtered)
		}
		
		// Try to maintain cursor position on the edited file
		if previousFile != "" {
			for i, f := range m.filtered {
				if f == previousFile {
					m.cursor = i
					break
				}
			}
		}
		
		// Ensure cursor is within bounds
		if m.cursor >= len(m.filtered) && len(m.filtered) > 0 {
			m.cursor = len(m.filtered) - 1
		}
		
		// Sync UI state
		m.syncUIState()
		
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
		
		// Task edit mode handling
		if m.taskEditMode {
			switch msg.String() {
			case "esc":
				// Cancel task editing
				m.taskEditMode = false
				m.taskEditField = ""
				m.taskEditInput.SetValue("")
				m.taskEditInput.Blur()
				m.taskBeingEdited = nil
				return m, nil
				
			case "enter":
				if m.taskEditField == "" {
					// Field selection is handled below in individual key cases
					return m, nil
				} else {
					// Process the field update
					cmd := m.processTaskFieldUpdate()
					return m, cmd
				}
				
			case "d":
				if m.taskEditField == "" {
					m.taskEditField = "due"
					m.taskEditInput.Placeholder = "YYYY-MM-DD or relative (today, tomorrow, 3d, 1w)"
					if m.taskBeingEdited.DueDate != "" {
						m.taskEditInput.SetValue(m.taskBeingEdited.DueDate)
					}
					return m, nil
				}
				
			case "s":
				if m.taskEditField == "" {
					m.taskEditField = "start"
					m.taskEditInput.Placeholder = "YYYY-MM-DD or relative (today, tomorrow, 3d, 1w)"
					if m.taskBeingEdited.StartDate != "" {
						m.taskEditInput.SetValue(m.taskBeingEdited.StartDate)
					}
					return m, nil
				}
				
			case "e":
				if m.taskEditField == "" {
					m.taskEditField = "estimate"
					m.taskEditInput.Placeholder = "1, 2, 3, 5, 8, or 13 (Fibonacci)"
					if m.taskBeingEdited.Estimate > 0 {
						m.taskEditInput.SetValue(fmt.Sprintf("%d", m.taskBeingEdited.Estimate))
					}
					return m, nil
				}
				
			case "p":
				if m.taskEditField == "" {
					m.taskEditField = "priority"
					m.taskEditInput.Placeholder = "1, 2, 3 or p1, p2, p3"
					if m.taskBeingEdited.Priority != "" {
						m.taskEditInput.SetValue(m.taskBeingEdited.Priority)
					}
					return m, nil
				}
				
			case "P":
				if m.taskEditField == "" {
					m.taskEditField = "project"
					m.taskEditInput.Placeholder = "Project name"
					if m.taskBeingEdited.Project != "" {
						m.taskEditInput.SetValue(m.taskBeingEdited.Project)
					}
					return m, nil
				}
				
			case "a":
				if m.taskEditField == "" {
					m.taskEditField = "area"
					m.taskEditInput.Placeholder = "Area (work, personal, etc.)"
					if m.taskBeingEdited.Area != "" {
						m.taskEditInput.SetValue(m.taskBeingEdited.Area)
					}
					return m, nil
				}
				
			case "t":
				if m.taskEditField == "" {
					m.taskEditField = "tags"
					m.taskEditInput.Placeholder = "Comma-separated tags (e.g., bug, urgent, frontend)"
					// Convert tags array to comma-delimited string
					if len(m.taskBeingEdited.Tags) > 0 {
						m.taskEditInput.SetValue(strings.Join(m.taskBeingEdited.Tags, ", "))
					}
					return m, nil
				}
			}
			
			// If we have a field selected, update the input
			if m.taskEditField != "" {
				var cmd tea.Cmd
				m.taskEditInput, cmd = m.taskEditInput.Update(msg)
				return m, cmd
			}
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
			if m.tagCreateMode {
				// Exit tag create mode on q
				m.tagCreateMode = false
				m.tagCreateInput.SetValue("")
				m.pendingTitle = ""
				return m, nil
			}
			if m.taskCreateMode {
				// Exit task create mode on q
				m.taskCreateMode = false
				m.taskCreateInput.SetValue("")
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
			if !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.taskCreateMode && m.cursor < len(m.filtered) {
				m.selected = m.filtered[m.cursor]
				// We'll handle the actual editor opening after we return
				return m, tea.ExecProcess(m.openInEditor(), func(err error) tea.Msg {
					return clearSelectedMsg{}
				})
			}

		case "ctrl+k":
			// Create TaskWarrior task from current note (if enabled)
			if m.config.TaskwarriorSupport && !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.taskCreateMode && m.cursor < len(m.filtered) {
				// Extract Denote identifier from current file
				currentFile := m.filtered[m.cursor]
				filename := filepath.Base(currentFile)
				
				// Check if it's a Denote-formatted file
				if len(filename) > 16 && filename[8] == 'T' && (filename[15] == '-' || filename[15] == '_') {
					// Enter task creation mode
					m.taskCreateMode = true
					m.taskCreateInput.Focus()
				} else {
					// File doesn't have a Denote identifier
					// Could show an error or handle differently
					// For now, we'll just do nothing
				}
			}

		case "esc":
			// Reset waiting for second g state
			m.waitingForSecondG = false
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
			if m.tagCreateMode {
				// Exit tag create mode and create note without tags
				title := m.pendingTitle
				var filename string
				var identifier string
				if m.config.DenoteFilenames {
					filename, identifier = generateDenoteName(title, []string{}, time.Now())
				} else {
					filename = titleToFilename(title)
					identifier = ""
				}
				fullPath := filepath.Join(m.cwd, filename)
				
				// Create the file without tags
				content := generateNoteContent(title, m.config, identifier, nil)
				if err := os.WriteFile(fullPath, []byte(content), 0644); err == nil {
					m.selected = fullPath
					// Refresh file list to include new file
					files, _ := findMarkdownFiles(m.cwd)
					m.files = m.applySorting(files)
					m.filtered = m.files
					// Find and select the new file
					for i, f := range m.filtered {
						if f == fullPath {
							m.cursor = i
							break
						}
					}
					// Exit tag create mode and open editor
					m.tagCreateMode = false
					m.tagCreateInput.SetValue("")
					m.pendingTitle = ""
					return m, tea.ExecProcess(m.openInEditor(), func(err error) tea.Msg {
						return clearSelectedMsg{}
					})
				}
				// If file creation failed, still exit the mode
				m.tagCreateMode = false
				m.tagCreateInput.SetValue("")
				m.pendingTitle = ""
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
			if m.taskCreateMode {
				// Exit task create mode
				m.taskCreateMode = false
				m.taskCreateInput.SetValue("")
			}
			if m.taskFilterMode {
				// Exit task filter mode
				m.taskFilterMode = false
			}
			if m.areaSelectMode {
				// Exit area select mode
				m.areaSelectMode = false
				m.availableAreas = nil
			}
			if m.projectsMode {
				// Exit projects mode back to task list
				m.projectsMode = false
				m.ui.ProjectsMode = false
				// Restore task view
				m.refreshTaskView()
				m.areaSelectCursor = 0
			}
			var filterCleared bool
			if m.taskFilter {
				// Clear task filter
				m.taskFilter = false
				m.filtered = m.files
				m.cursor = 0
				filterCleared = true
			}
			if m.tagFilter {
				// Clear tag filter
				m.tagFilter = false
				m.filtered = m.files
				m.cursor = 0
				filterCleared = true
			}
			if m.textFilter {
				// Clear text filter
				m.textFilter = false
				m.filtered = m.files
				m.cursor = 0
				filterCleared = true
			}
			if m.dailyFilter {
				// Clear daily filter
				m.dailyFilter = false
				m.filtered = m.files
				m.cursor = 0
				filterCleared = true
			}
			if m.oldFilter {
				// Clear old filter
				m.oldFilter = false
				m.filtered = m.files
				m.cursor = 0
				filterCleared = true
			}
			// Show filter cleared message if any filter was active
			if filterCleared {
				cmds = append(cmds, ui.ShowInfo(ui.MsgFilterCleared))
			}
			if m.renameMode {
				// Exit rename mode
				m.renameMode = false
				m.renameFile = ""
			}

		case "n":
			if m.deleteMode {
				// Cancel deletion
				m.deleteMode = false
				m.deleteFile = ""
			} else if !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.taskCreateMode {
				// Enter create mode
				m.createMode = true
				m.createInput.Focus()
				// Set the prompt based on mode
				if m.taskModeActive {
					m.createInput.Placeholder = "Task title..."
				} else {
					m.createInput.Placeholder = "Note title..."
				}
				return m, nil
			}

		case "/":
			if !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.taskCreateMode {
				// Enter search mode
				m.searchMode = true
				m.search.Focus()
				return m, nil
			}

		case "#":
			if !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.taskCreateMode {
				// Enter tag search mode
				m.tagMode = true
				m.tagInput.Focus()
				return m, nil
			}

		case "D":
			if !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.sortMode && !m.oldMode && !m.taskCreateMode {
				// Search for daily notes
				if files, err := searchDailyNotes(m.cwd); err == nil {
					m.filtered = files
					m.cursor = 0
					m.dailyFilter = true
					m.taskFilter = false // Clear task filter when switching to daily filter
					m.tagFilter = false // Clear tag filter when switching to daily filter
					m.textFilter = false // Clear text filter when switching to daily filter
					m.oldFilter = false // Clear old filter when switching to daily filter
					
					// Show status message
					var cmds []tea.Cmd
					if len(files) == 0 {
						cmds = append(cmds, ui.ShowInfo(ui.MsgNoMatches))
					} else {
						cmds = append(cmds, ui.ShowSuccess(ui.MsgDailyFilter))
					}
					return m, tea.Batch(cmds...)
				}
			}

		case "X":
			if !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.sortMode && !m.oldMode && !m.taskCreateMode && m.cursor < len(m.filtered) {
				// Enter delete confirmation mode
				m.deleteMode = true
				m.deleteFile = m.filtered[m.cursor]
			}


		case "O":
			if !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.sortMode && !m.taskCreateMode {
				// Enter old (days back) mode
				m.oldMode = true
				m.oldInput.Focus()
				return m, nil
			}

		case "R":
			if !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.sortMode && !m.oldMode && !m.taskCreateMode && m.cursor < len(m.filtered) {
				// Enter rename mode - rename file to Denote format
				m.renameMode = true
				m.renameFile = m.filtered[m.cursor]
				
				// Perform the rename immediately
				if newPath, err := renameToDenoteName(m.renameFile, m.config); err == nil {
					// Refresh file list after successful rename
					files, _ := findMarkdownFiles(m.cwd)
					m.files = m.applySorting(files)
					m.filtered = m.files
					
					// Try to maintain cursor position on the renamed file
					for i, f := range m.filtered {
						if f == newPath {
							m.cursor = i
							break
						}
					}
				}
				
				// Always exit rename mode immediately
				m.renameMode = false
				m.renameFile = ""
			}

		case "T":
			if !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.sortMode && !m.oldMode && !m.taskCreateMode {
				// Only allow task mode if enabled in config
				if m.config.DenoteTasksSupport {
					// Toggle task mode
					cmd := m.toggleTaskMode()
					return m, cmd
				}
			}

		case "P":
			if m.taskModeActive && !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.sortMode && !m.oldMode && !m.taskCreateMode && !m.taskFilterMode && !m.projectsMode {
				// Enter projects mode
				m.projectsMode = true
				m.ui.ProjectsMode = true
				
				// Load projects if not already loaded
				if m.projects == nil {
					scanner := denote.NewScanner(m.getTasksDirectory())
					projects, err := scanner.FindProjects()
					if err != nil || len(projects) == 0 {
						m.projectsMode = false
						m.ui.ProjectsMode = false
						return m, ui.ShowError("No projects found")
					}
					m.projects = projects
				}
				
				// Convert projects to display items with task counts
				projectItems := m.prepareProjectItems()
				
				// Update UI state
				m.ui.Projects = make([]interface{}, len(projectItems))
				for i, item := range projectItems {
					itemCopy := item // Important: create a copy
					m.ui.Projects[i] = &itemCopy
				}
				m.ui.ProjectsCursor = 0
				
				// Update file lists for navigation - only include filtered projects
				filteredProjects := make([]*denote.Project, 0)
				for i := range projectItems {
					if projectItem, ok := m.ui.Projects[i].(*ui.ProjectItem); ok && projectItem.Project != nil {
						filteredProjects = append(filteredProjects, projectItem.Project)
					}
				}
				
				m.files = make([]string, len(filteredProjects))
				m.filtered = make([]string, len(filteredProjects))
				for i, project := range filteredProjects {
					m.files[i] = project.Path
					m.filtered[i] = project.Path
				}
				m.cursor = 0
				
				// Show message if area filter is active
				if m.taskAreaContext != "" {
					return m, ui.ShowInfo(fmt.Sprintf("Showing projects in area: %s", m.taskAreaContext))
				}
				
				return m, nil
			}

		case "d":
			if m.sortMode {
				if m.taskModeActive {
					// In task mode, sort by due date
					m.taskSortBy = "due"
					m.applyTaskSorting()
				} else {
					// Normal mode, sort by date
					m.currentSort = "date"
					m.files = m.applySorting(m.files)
					m.filtered = m.applySorting(m.filtered)
				}
				m.sortMode = false
				m.cursor = 0
				return m, nil
			} else if m.taskModeActive && !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.sortMode && !m.oldMode && !m.taskCreateMode && m.cursor < len(m.filtered) {
				// In task mode, 'd' marks task as done
				if task := m.findTaskByPath(m.filtered[m.cursor]); task != nil {
					if err := denote.UpdateTaskStatus(task.Path, denote.TaskStatusDone); err == nil {
						// Update task in memory
						task.Status = denote.TaskStatusDone
						// Show success message
						return m, ui.ShowSuccess("Task marked as done")
					} else {
						return m, ui.ShowError("Failed to update task status")
					}
				}
			} else if !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.sortMode && !m.oldMode && !m.taskCreateMode {
				// First, check if a daily note already exists for today
				if existingDaily, err := findTodaysDailyNote(m.cwd); err == nil {
					// Daily note exists, open it
					m.selected = existingDaily
					// Find and select the file in the list
					for i, f := range m.filtered {
						if f == existingDaily {
							m.cursor = i
							break
						}
					}
					return m, tea.ExecProcess(m.openInEditor(), func(err error) tea.Msg {
						return clearSelectedMsg{}
					})
				} else {
					// No daily note exists, create one in the current format
					filename, identifier := getDailyNoteFilename(m.config)
					fullPath := filepath.Join(m.cwd, filename)
					
					// Create the daily note with proper formatting
					today := time.Now().Format("Monday, January 2, 2006")
					title := fmt.Sprintf("Daily Note - %s", today)
					
					// Generate content with proper format
					// For daily notes, we pass the "daily" tag automatically
					var tags []string
					if m.config.DenoteFilenames {
						tags = []string{"daily"}
					}
					content := generateNoteContent(title, m.config, identifier, tags)
					// Add daily note sections after frontmatter/title
					content += "## Tasks\n\n## Notes\n\n"
					if err := os.WriteFile(fullPath, []byte(content), 0644); err == nil {
						m.selected = fullPath
						// Refresh file list to include new file
						files, _ := findMarkdownFiles(m.cwd)
						m.files = m.applySorting(files)
						m.filtered = m.files
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
				}
			}

		case "m":
			if m.sortMode {
				if m.taskModeActive {
					// In task mode, 'm' sorts by modified
					m.taskSortBy = "modified"
					m.applyTaskSorting()
				} else {
					// Sort by modified
					m.currentSort = "modified"
					m.files = m.applySorting(m.files)
					m.filtered = m.applySorting(m.filtered)
				}
				m.sortMode = false
				m.cursor = 0
				return m, nil
			}

		case "p":
			if m.sortMode && m.taskModeActive {
				// In task mode sort, 'p' sorts by priority
				m.taskSortBy = "priority"
				m.applyTaskSorting()
				m.sortMode = false
				m.cursor = 0
				return m, nil
			} else if m.taskModeActive && !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.sortMode && !m.oldMode && !m.taskCreateMode && m.cursor < len(m.filtered) {
				// In task mode, 'p' pauses/unpauses task
				if task := m.findTaskByPath(m.filtered[m.cursor]); task != nil {
					newStatus := denote.TaskStatusPaused
					if task.Status == denote.TaskStatusPaused {
						newStatus = denote.TaskStatusOpen
					}
					if err := denote.UpdateTaskStatus(task.Path, newStatus); err == nil {
						task.Status = newStatus
						return m, ui.ShowSuccess(fmt.Sprintf("Task %s", newStatus))
					}
				}
			}

		case "1", "2", "3":
			if m.taskModeActive && !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.sortMode && !m.oldMode && !m.taskCreateMode && m.cursor < len(m.filtered) {
				// In task mode, number keys set priority
				if task := m.findTaskByPath(m.filtered[m.cursor]); task != nil {
					priority := "p" + msg.String()
					if err := denote.UpdateTaskPriority(task.Path, priority); err == nil {
						task.Priority = priority
						return m, ui.ShowSuccess(fmt.Sprintf("Priority set to %s", priority))
					}
				}
			}

		case "u":
			if m.taskModeActive && !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.sortMode && !m.oldMode && !m.taskCreateMode && !m.taskEditMode && m.cursor < len(m.filtered) {
				// In task mode, 'u' starts task metadata editing
				if task := m.findTaskByPath(m.filtered[m.cursor]); task != nil {
					m.taskEditMode = true
					m.taskBeingEdited = task
					m.taskEditField = "" // No field selected yet
					m.taskEditInput.SetValue("")
					m.taskEditInput.Focus()
				}
			}

		case "t":
			if m.sortMode {
				// Sort by title
				m.currentSort = "title"
				m.files = m.applySorting(m.files)
				m.filtered = m.applySorting(m.filtered)
				m.sortMode = false
				m.cursor = 0
			} else if !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.sortMode && !m.oldMode && !m.taskCreateMode {
				// Search for tasks (existing functionality)
				if files, err := searchTasks(m.cwd); err == nil {
					m.filtered = files
					m.cursor = 0
					m.taskFilter = true
					m.tagFilter = false // Clear tag filter when switching to task filter
					m.textFilter = false // Clear text filter when switching to task filter
					m.dailyFilter = false // Clear daily filter when switching to task filter
					m.oldFilter = false // Clear old filter when switching to task filter
					
					// Show status message
					var cmds []tea.Cmd
					if len(files) == 0 {
						cmds = append(cmds, ui.ShowInfo(ui.MsgNoMatches))
					} else {
						cmds = append(cmds, ui.ShowSuccess(ui.MsgTaskFilter))
					}
					return m, tea.Batch(cmds...)
				}
			}

		case "i":
			if m.sortMode {
				// Sort by Denote identifier
				m.currentSort = "denote"
				m.files = m.applySorting(m.files)
				m.filtered = m.applySorting(m.filtered)
				m.sortMode = false
				m.cursor = 0
			}

		case "r":
			if m.sortMode {
				// Reverse current sort
				m.reversedSort = !m.reversedSort
				m.files = m.applySorting(m.files)
				m.filtered = m.applySorting(m.filtered)
				m.sortMode = false
				m.cursor = 0
			}

		case "s":
			if m.sortMode && m.taskModeActive {
				// In task mode sort, 's' sorts by status
				m.taskSortBy = "status"
				m.applyTaskSorting()
				m.sortMode = false
				m.cursor = 0
				return m, nil
			}

		case "y":
			if m.deleteMode {
				// Confirm deletion
				deletedFile := filepath.Base(m.deleteFile)
				if err := os.Remove(m.deleteFile); err == nil {
					// Successfully deleted, refresh appropriate list
					if m.taskModeActive {
						// Reload tasks through denote scanner
						scanner := denote.NewScanner(m.getTasksDirectory())
						if tasks, err := scanner.FindTasks(); err == nil {
							m.tasks = tasks
							// Apply current sorting
							if m.taskSortBy != "" {
								denote.SortTasks(m.tasks, m.taskSortBy, m.reversedSort)
							}
							// Update file lists
							m.files = make([]string, len(m.tasks))
							m.filtered = make([]string, len(m.tasks))
							for i, task := range m.tasks {
								m.files[i] = task.Path
								m.filtered[i] = task.Path
							}
							// Update task formatter
							m.ui.TaskFormatter = func(path string) string {
								task := m.findTaskByPath(path)
								if task != nil {
									return m.formatTaskLine(task)
								}
								return filepath.Base(path)
							}
							// Sync UI state
							m.syncUIState()
						}
					} else {
						// Normal mode - refresh markdown files
						files, _ := findMarkdownFiles(m.cwd)
						m.files = m.applySorting(files)
						
						// If we had filters applied, reapply them
						if m.taskFilter {
							if taskFiles, err := searchTasks(m.cwd); err == nil {
								m.filtered = m.applySorting(taskFiles)
							} else {
								m.filtered = m.files
							}
						} else if m.dailyFilter {
						if dailyFiles, err := searchDailyNotes(m.cwd); err == nil {
							m.filtered = m.applySorting(dailyFiles)
						} else {
							m.filtered = m.files
						}
						} else if m.search.Value() != "" {
							m.filtered = m.applySorting(filterFiles(m.files, m.search.Value()))
						} else {
							m.filtered = m.files
						}
					}
					
					// Adjust cursor position
					if m.cursor >= len(m.filtered) {
						if len(m.filtered) > 0 {
							m.cursor = len(m.filtered) - 1
						} else {
							m.cursor = 0
						}
					}
					// Show success message
					cmds = append(cmds, ui.ShowSuccess(fmt.Sprintf("Deleted %s", deletedFile)))
				} else {
					// Show error message
					cmds = append(cmds, ui.ShowError(fmt.Sprintf("Failed to delete %s", deletedFile)))
				}
				// Exit delete mode regardless of success/failure
				m.deleteMode = false
				m.deleteFile = ""
			}

		case "f":
			if m.taskModeActive && !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.sortMode && !m.oldMode && !m.taskCreateMode {
				// Enter task filter mode
				m.taskFilterMode = true
			}

		case "a":
			if m.taskFilterMode {
				// Show all tasks
				m.applyTaskFilter("all", "")
				m.taskFilterMode = false
			}

		case "o":
			if m.taskFilterMode {
				// Show only open tasks
				m.applyTaskFilter("open", "")
				m.taskFilterMode = false
			} else if !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.oldMode && !m.taskCreateMode {
				// Enter sort mode
				m.sortMode = true
			}

		case "c":
			if m.taskFilterMode {
				// Show active tasks (open, paused, delegated)
				m.applyTaskFilter("active", "")
				m.taskFilterMode = false
			}

		case "v":
			if m.taskFilterMode {
				// Show overdue tasks
				m.applyTaskFilter("overdue", "")
				m.taskFilterMode = false
			}

		case "w":
			if m.taskFilterMode {
				// Show tasks due this week
				m.applyTaskFilter("week", "")
				m.taskFilterMode = false
			}
			
		case "x":
			if m.taskFilterMode && m.taskAreaContext != "" {
				// Clear area context
				m.taskAreaContext = ""
				m.refreshTaskView()
				m.taskFilterMode = false
				return m, ui.ShowSuccess("Area filter cleared")
			}

		case "A":
			if m.taskFilterMode {
				// Show area selection submenu
				areas := denote.GetUniqueAreas(m.tasks)
				if len(areas) > 0 {
					m.taskFilterMode = false
					m.areaSelectMode = true
					m.availableAreas = areas
					m.areaSelectCursor = 0
				} else {
					m.taskFilterMode = false
					return m, ui.ShowInfo("No areas found")
				}
			}



		case "backspace":
			if m.taskModeActive {
				// Handle different contexts - check project filter first
				if m.taskFilterType == "project" && m.taskFilterProject != "" {
					// Was viewing project tasks, go back to projects mode
					m.projectsMode = true
					m.ui.ProjectsMode = true
					
					// Restore projects display
					projectItems := m.prepareProjectItems()
					m.ui.Projects = make([]interface{}, len(projectItems))
					for i, item := range projectItems {
						itemCopy := item
						m.ui.Projects[i] = &itemCopy
					}
					
					// Update file lists - only include filtered projects
					filteredProjects := make([]*denote.Project, 0)
					for i := range projectItems {
						if projectItem, ok := m.ui.Projects[i].(*ui.ProjectItem); ok && projectItem.Project != nil {
							filteredProjects = append(filteredProjects, projectItem.Project)
						}
					}
					
					m.files = make([]string, len(filteredProjects))
					m.filtered = make([]string, len(filteredProjects))
					for i, project := range filteredProjects {
						m.files[i] = project.Path
						m.filtered[i] = project.Path
					}
					
					// Clear project filter
					m.taskFilterType = ""
					m.taskFilterProject = ""
					
					return m, ui.ShowInfo("Back to projects")
				} else if m.taskAreaContext != "" && m.taskStatusFilter != "" {
					// Clear status filter but keep area context
					m.taskStatusFilter = ""
					m.taskFilterType = "all"
					m.refreshTaskView()
					return m, ui.ShowInfo(fmt.Sprintf("Showing all tasks in area: %s", m.taskAreaContext))
				} else if m.taskAreaContext != "" {
					// Clear area context
					m.taskAreaContext = ""
					m.refreshTaskView()
					return m, ui.ShowInfo("Area filter cleared")
				} else if m.taskFilterType == "projects" {
					// Was viewing projects, go back to all tasks
					// Need to reload tasks since they were set to nil
					scanner := denote.NewScanner(m.getTasksDirectory())
					tasks, err := scanner.FindTasks()
					if err == nil && len(tasks) > 0 {
						m.tasks = tasks
						m.applyTaskFilter("all", "")
						return m, ui.ShowInfo("Showing all tasks")
					} else {
						return m, ui.ShowError("No tasks found")
					}
				} else if m.taskFilterType != "" && m.taskFilterType != "all" {
					// Clear filter and show all tasks
					m.applyTaskFilter("all", "")
					return m, ui.ShowInfo("Showing all tasks")
				}
			}

		case "up", "k":
			if m.areaSelectMode {
				if m.areaSelectCursor > 0 {
					m.areaSelectCursor--
				}
			} else if !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.taskCreateMode && m.cursor > 0 {
				m.cursor--
				// Don't auto-load preview on cursor movement
			}
			// Reset waiting for second g state on other navigation
			m.waitingForSecondG = false

		case "down", "j":
			if m.areaSelectMode {
				if m.areaSelectCursor < len(m.availableAreas)-1 {
					m.areaSelectCursor++
				}
			} else if !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.taskCreateMode && m.cursor < len(m.filtered)-1 {
				m.cursor++
				// Don't auto-load preview on cursor movement
			}
			// Reset waiting for second g state on other navigation
			m.waitingForSecondG = false

		case "G":
			// Jump to bottom of list
			if !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.taskCreateMode && len(m.filtered) > 0 {
				m.cursor = len(m.filtered) - 1
			}
			// Reset waiting for second g state
			m.waitingForSecondG = false

		case "g":
			// Handle gg sequence for jump to top
			if !m.searchMode && !m.createMode && !m.tagMode && !m.tagCreateMode && !m.deleteMode && !m.taskCreateMode {
				if m.waitingForSecondG {
					// Second g - jump to top
					m.cursor = 0
					m.waitingForSecondG = false
				} else {
					// First g - start waiting for second g
					m.waitingForSecondG = true
				}
			}

		case "enter":
			// Reset waiting for second g state
			m.waitingForSecondG = false
			if m.deleteMode {
				// Don't delete on enter - require explicit 'y' confirmation
				return m, nil
			} else if m.areaSelectMode {
				// Select the area under cursor
				if m.areaSelectCursor < len(m.availableAreas) {
					selectedArea := m.availableAreas[m.areaSelectCursor]
					m.areaSelectMode = false
					m.availableAreas = nil
					m.areaSelectCursor = 0
					m.applyTaskFilter("area", selectedArea)
					return m, ui.ShowSuccess(fmt.Sprintf("Filtering by area: %s", selectedArea))
				}
			} else if m.projectsMode && m.cursor < len(m.projects) {
				// View tasks for selected project
				project := m.projects[m.cursor]
				projectKey := project.ProjectMetadata.Identifier
				if projectKey == "" {
					projectKey = project.ProjectMetadata.Title
				}
				
				// Exit projects mode and show project tasks
				m.projectsMode = false
				m.ui.ProjectsMode = false
				m.applyTaskFilter("project", projectKey)
				return m, ui.ShowSuccess(fmt.Sprintf("Showing tasks for project: %s", project.ProjectMetadata.Title))
			} else if m.tagMode {
				// Search for the tag
				tag := m.tagInput.Value()
				var cmds []tea.Cmd
				if tag != "" {
					if files, err := searchTag(m.cwd, tag); err == nil {
						m.filtered = files
						m.cursor = 0
						m.tagFilter = true // Set tag filter active
						m.taskFilter = false // Clear task filter when switching to tag filter
						m.textFilter = false // Clear text filter when switching to tag filter
						m.dailyFilter = false // Clear daily filter when switching to tag filter
						m.oldFilter = false // Clear old filter when switching to tag filter
						
						// Show status message
						if len(files) == 0 {
							cmds = append(cmds, ui.ShowInfo(ui.MsgNoMatches))
						} else {
							cmds = append(cmds, ui.ShowSuccess(fmt.Sprintf(ui.MsgTagFilter, tag)))
						}
					}
				}
				// Exit tag mode
				m.tagMode = false
				m.tagInput.SetValue("")
				if len(cmds) > 0 {
					return m, tea.Batch(cmds...)
				}
			} else if m.oldMode {
				// Apply days old filter
				daysStr := m.oldInput.Value()
				var cmds []tea.Cmd
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
						
						// Show status message
						if len(filteredFiles) == 0 {
							cmds = append(cmds, ui.ShowInfo(ui.MsgNoMatches))
						} else {
							cmds = append(cmds, ui.ShowSuccess(fmt.Sprintf(ui.MsgDaysFilter, days)))
						}
					}
				}
				// Exit old mode
				m.oldMode = false
				m.oldInput.SetValue("")
				if len(cmds) > 0 {
					return m, tea.Batch(cmds...)
				}
			} else if m.createMode {
				// Get the title
				title := m.createInput.Value()
				if title != "" {
					if m.taskModeActive {
						// Create a task file
						var filename string
						var identifier string
						if m.config.DenoteFilenames {
							// Tasks always have the "task" tag
							filename, identifier = generateDenoteName(title, []string{"task"}, time.Now())
						} else {
							filename = titleToFilename(title) 
							// Ensure it ends with -task.md for non-denote mode
							filename = strings.TrimSuffix(filename, ".md") + "-task.md"
							identifier = ""
						}
						fullPath := filepath.Join(m.getTasksDirectory(), filename)
						
						// Get next task ID from counter
						counter, err := denote.GetIDCounter(m.getTasksDirectory())
						if err != nil {
							return m, ui.ShowError(fmt.Sprintf("Failed to get ID counter: %v", err))
						}
						taskID, err := counter.NextID()
						if err != nil {
							return m, ui.ShowError(fmt.Sprintf("Failed to get next task ID: %v", err))
						}
						
						// Create the task file with task template
						content := generateTaskContent(title, m.config, identifier, taskID)
						if err := os.WriteFile(fullPath, []byte(content), 0644); err == nil {
							m.selected = fullPath
							// Reload tasks
							scanner := denote.NewScanner(m.getTasksDirectory())
							if tasks, err := scanner.FindTasks(); err == nil {
								m.tasks = tasks
								// Apply current sorting
								if m.taskSortBy != "" {
									denote.SortTasks(m.tasks, m.taskSortBy, m.reversedSort)
								}
								// Update file lists
								m.files = make([]string, len(m.tasks))
								m.filtered = make([]string, len(m.tasks))
								for i, task := range m.tasks {
									m.files[i] = task.Path
									m.filtered[i] = task.Path
								}
								// Find and select the new task
								for i, f := range m.filtered {
									if f == fullPath {
										m.cursor = i
										break
									}
								}
							}
							// Exit create mode and open editor
							m.createMode = false
							m.createInput.SetValue("")
							m.createInput.Placeholder = "Note title..."
							// Show success message with task ID
							cmds := []tea.Cmd{
								ui.ShowSuccess(fmt.Sprintf("Created task #%d: %s", taskID, title)),
								tea.ExecProcess(m.openInEditor(), func(err error) tea.Msg {
									return clearSelectedMsg{}
								}),
							}
							return m, tea.Batch(cmds...)
						}
					} else {
						// Check if we should prompt for tags (normal note creation)
						if m.config.AddFrontmatter && m.config.PromptForTags {
							// Transition to tag input mode
							m.pendingTitle = title
							m.createMode = false
							m.createInput.SetValue("")
							m.tagCreateMode = true
							m.tagCreateInput.Focus()
							return m, nil
						} else {
							// Create note without tags
							var filename string
							var identifier string
							if m.config.DenoteFilenames {
								filename, identifier = generateDenoteName(title, []string{}, time.Now())
							} else {
								filename = titleToFilename(title)
								identifier = ""
							}
							fullPath := filepath.Join(m.cwd, filename)
							
							// Create the file with templated content
							content := generateNoteContent(title, m.config, identifier, nil)
							if err := os.WriteFile(fullPath, []byte(content), 0644); err == nil {
								m.selected = fullPath
								// Refresh file list to include new file
								files, _ := findMarkdownFiles(m.cwd)
								m.files = m.applySorting(files)
								m.filtered = m.files
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
					}
				}
				// Exit create mode if title is empty
				m.createMode = false
				m.createInput.SetValue("")
			} else if m.tagCreateMode {
				// Process tags and create the note
				tagInput := m.tagCreateInput.Value()
				var tags []string
				if tagInput != "" {
					// Parse comma-separated tags
					parts := strings.Split(tagInput, ",")
					for _, tag := range parts {
						trimmed := strings.TrimSpace(tag)
						if trimmed != "" {
							tags = append(tags, trimmed)
						}
					}
				}
				
				// Now create the note with the pending title and tags
				title := m.pendingTitle
				var filename string
				var identifier string
				if m.config.DenoteFilenames {
					filename, identifier = generateDenoteName(title, tags, time.Now())
				} else {
					filename = titleToFilename(title)
					identifier = ""
				}
				fullPath := filepath.Join(m.cwd, filename)
				
				// Create the file with templated content including tags
				content := generateNoteContent(title, m.config, identifier, tags)
				if err := os.WriteFile(fullPath, []byte(content), 0644); err == nil {
					m.selected = fullPath
					// Refresh file list to include new file
					files, _ := findMarkdownFiles(m.cwd)
					m.files = m.applySorting(files)
					m.filtered = m.files
					// Find and select the new file
					for i, f := range m.filtered {
						if f == fullPath {
							m.cursor = i
							break
						}
					}
					// Exit tag create mode and open editor
					m.tagCreateMode = false
					m.tagCreateInput.SetValue("")
					m.pendingTitle = ""
					return m, tea.ExecProcess(m.openInEditor(), func(err error) tea.Msg {
						return clearSelectedMsg{}
					})
				}
			} else if m.searchMode {
				// Exit search mode on enter, but keep the filter active
				m.searchMode = false
				// If there's a search query, mark text filter as active
				var cmds []tea.Cmd
				if m.search.Value() != "" {
					m.textFilter = true
					m.taskFilter = false // Clear other filters
					m.tagFilter = false
					
					// Show status message
					if len(m.filtered) == 0 {
						cmds = append(cmds, ui.ShowInfo(ui.MsgNoMatches))
					} else {
						cmds = append(cmds, ui.ShowSuccess(ui.MsgSearchApplied))
					}
					m.dailyFilter = false
				}
				if len(cmds) > 0 {
					return m, tea.Batch(cmds...)
				}
			} else if m.taskCreateMode {
				// Create TaskWarrior task with current note's identifier
				taskDesc := m.taskCreateInput.Value()
				if taskDesc != "" {
					// Get current file and extract identifier
					currentFile := m.filtered[m.cursor]
					filename := filepath.Base(currentFile)
					
					// Extract the Denote identifier (first 15 characters)
					if len(filename) >= 15 {
						identifier := filename[:15]
						
						// Create the task using TaskWarrior CLI
						cmd := exec.Command("task", "add", taskDesc, "notesid:"+identifier)
						output, err := cmd.CombinedOutput()
						
						if err == nil {
							// Task created successfully
							// Could show a success message in the future
						} else {
							// Task creation failed
							// Could show an error message in the future
							// For now, just log to stderr
							fmt.Fprintf(os.Stderr, "Failed to create task: %v\n%s\n", err, output)
						}
					}
				}
				// Exit task create mode
				m.taskCreateMode = false
				m.taskCreateInput.SetValue("")
			} else if !m.deleteMode && m.cursor < len(m.filtered) {
				// Handle special case for project view
				if m.taskModeActive && m.taskFilterType == "projects" && len(m.projects) > 0 {
					// When in project view, enter shows tasks for that project
					projectPath := m.filtered[m.cursor]
					for _, project := range m.projects {
						if project.Path == projectPath {
							// Show tasks for this project
							// Use the identifier if available, otherwise use title
							var projectKey string
							
							// First priority: use identifier field if available
							if project.ProjectMetadata.Identifier != "" {
								projectKey = project.ProjectMetadata.Identifier
							} else {
								// Otherwise use the title
								projectKey = project.ProjectMetadata.Title
								if projectKey == "" {
									// Fallback to note title if project-specific title not set
									projectKey = project.Note.Title
								}
								if projectKey == "" {
									// Last resort: extract from filename
									projectSlug := strings.TrimSuffix(filepath.Base(project.Path), ".md")
									parts := strings.SplitN(projectSlug, "-", 2)
									if len(parts) >= 2 {
										titlePart := parts[1]
										if tagIdx := strings.Index(titlePart, "__"); tagIdx >= 0 {
											titlePart = titlePart[:tagIdx]
										}
										projectKey = titlePart
									}
								}
							}
							
							cmd := m.showProjectTasks(projectKey)
							return m, cmd
						}
					}
					// If we get here, something went wrong
					return m, ui.ShowError("Could not find project")
				}
				
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

	// Handle tag creation input
	if m.tagCreateMode {
		m.tagCreateInput, cmd = m.tagCreateInput.Update(msg)
	}

	// Handle old input
	if m.oldMode {
		m.oldInput, cmd = m.oldInput.Update(msg)
	}

	// Handle task creation input
	if m.taskCreateMode {
		m.taskCreateInput, cmd = m.taskCreateInput.Update(msg)
		cmds = append(cmds, cmd)
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
	// Handle area selection mode separately
	if m.areaSelectMode {
		return m.renderAreaSelectMode()
	}
	
	// Handle task edit mode
	if m.taskEditMode {
		return m.renderTaskEditMode()
	}
	
	// Sync state with UI integration
	m.syncUIState()
	
	// Use new UI system
	return m.ui.Render()
}

// renderAreaSelectMode renders the area selection interface
func (m model) renderAreaSelectMode() string {
	var content strings.Builder
	
	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Render("Select Area")
	content.WriteString(header + "\n\n")
	
	// List areas with cursor
	for i, area := range m.availableAreas {
		cursor := "  "
		if i == m.areaSelectCursor {
			cursor = "> "
		}
		
		line := fmt.Sprintf("%s%s", cursor, area)
		if i == m.areaSelectCursor {
			// Highlight selected line
			line = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Bold(true).
				Render(line)
		}
		content.WriteString(line + "\n")
	}
	
	// Footer with help
	content.WriteString("\n")
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("[Enter] select  [Esc] cancel  [↑↓/jk] navigate")
	content.WriteString(help)
	
	// Center the content
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content.String())
}

// renderTaskEditMode renders the task metadata edit interface
func (m model) renderTaskEditMode() string {
	if m.taskBeingEdited == nil {
		return "Error: No task selected"
	}
	
	var content strings.Builder
	theme := ui.GetTheme(m.config.Theme)
	
	// Title - use metadata title if available, otherwise note title
	taskTitle := m.taskBeingEdited.Note.Title
	if m.taskBeingEdited.TaskMetadata.Title != "" {
		taskTitle = m.taskBeingEdited.TaskMetadata.Title
	}
	title := fmt.Sprintf("Task #%d: %s", m.taskBeingEdited.TaskID, taskTitle)
	titleStyle := theme.Modal.Title
	content.WriteString(titleStyle.Render(title))
	content.WriteString("\n\n")
	
	if m.taskEditField == "" {
		// Show field selection menu
		content.WriteString("Select field to edit:\n\n")
		
		fields := []struct {
			key   string
			label string
			value string
		}{
			{"d", "Due date", m.taskBeingEdited.DueDate},
			{"s", "Start date", m.taskBeingEdited.StartDate},
			{"e", "Estimate", fmt.Sprintf("%d", m.taskBeingEdited.Estimate)},
			{"p", "Priority", m.taskBeingEdited.Priority},
			{"P", "Project", m.taskBeingEdited.Project},
			{"a", "Area", m.taskBeingEdited.Area},
			{"t", "Tags", strings.Join(m.taskBeingEdited.Tags, ", ")},
		}
		
		for _, field := range fields {
			value := field.value
			if value == "" || value == "0" {
				value = "(not set)"
			}
			line := fmt.Sprintf("[%s] %-12s: %s\n", 
				theme.Help.Key.Render(field.key),
				field.label,
				value)
			content.WriteString(line)
		}
		
		content.WriteString("\n")
		content.WriteString(theme.Help.Key.Render("[Esc]"))
		content.WriteString(" Done")
	} else {
		// Show input for selected field
		fieldName := ""
		switch m.taskEditField {
		case "due":
			fieldName = "Due date"
		case "start":
			fieldName = "Start date"
		case "estimate":
			fieldName = "Estimate"
		case "priority":
			fieldName = "Priority"
		case "project":
			fieldName = "Project"
		case "area":
			fieldName = "Area"
		case "tags":
			fieldName = "Tags"
		}
		
		content.WriteString(fmt.Sprintf("Editing %s:\n\n", fieldName))
		content.WriteString(m.taskEditInput.View())
		content.WriteString("\n\n")
		content.WriteString(theme.Help.Key.Render("[Enter]"))
		content.WriteString(" Save  ")
		content.WriteString(theme.Help.Key.Render("[Esc]"))
		content.WriteString(" Cancel")
	}
	
	// Create modal box
	box := theme.Modal.Border.
		Width(60).
		Padding(1, 2).
		Render(content.String())
	
	// Center on screen
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(box)
}

// syncUIState synchronizes the model state with the UI integration
func (m *model) syncUIState() {
	if m.ui == nil {
		return
	}
	
	// Always update files and filtered lists to ensure UI updates
	m.ui.Files = make([]string, len(m.files))
	copy(m.ui.Files, m.files)
	m.ui.Filtered = make([]string, len(m.filtered))
	copy(m.ui.Filtered, m.filtered)
	m.ui.Cursor = m.cursor
	m.ui.Selected = m.selected
	m.ui.Width = m.width
	m.ui.Height = m.height
	
	// Update mode flags
	m.ui.SearchMode = m.searchMode
	m.ui.CreateMode = m.createMode
	m.ui.TagMode = m.tagMode
	m.ui.TagCreateMode = m.tagCreateMode
	m.ui.TaskCreateMode = m.taskCreateMode
	m.ui.PreviewMode = m.previewMode
	m.ui.DeleteMode = m.deleteMode
	m.ui.SortMode = m.sortMode
	m.ui.OldMode = m.oldMode
	m.ui.RenameMode = m.renameMode
	m.ui.TaskFilterMode = m.taskFilterMode
	// Area selection mode is handled separately, not passed to UI
	
	// Update inputs
	m.ui.Search = m.search
	m.ui.CreateInput = m.createInput
	m.ui.TagInput = m.tagInput
	m.ui.TagCreateInput = m.tagCreateInput
	m.ui.TaskCreateInput = m.taskCreateInput
	m.ui.OldInput = m.oldInput
	
	// Update other state
	m.ui.PreviewContent = m.previewContent
	m.ui.PreviewFile = m.previewFile
	m.ui.PreviewScroll = m.previewScroll
	m.ui.DeleteFile = m.deleteFile
	m.ui.RenameFile = m.renameFile
	m.ui.PendingTitle = m.pendingTitle
	m.ui.CurrentSort = m.currentSort
	m.ui.ReversedSort = m.reversedSort
	m.ui.OldDays = m.oldDays
	m.ui.TextFilter = m.textFilter
	m.ui.TagFilter = m.tagFilter
	m.ui.TaskFilter = m.taskFilter
	m.ui.DailyFilter = m.dailyFilter
	m.ui.OldFilter = m.oldFilter
	
	// Task mode state
	m.ui.TaskModeActive = m.taskModeActive
	m.ui.TaskSortBy = m.taskSortBy
	m.ui.TaskAreaContext = m.taskAreaContext
	m.ui.TaskStatusFilter = m.taskStatusFilter
	
	// Set task formatter function based on what we're showing
	if m.taskModeActive {
		if m.taskFilterType == "projects" && m.projects != nil {
			// Create a closure that captures the projects slice
			capturedProjects := m.projects
			m.ui.TaskFormatter = func(path string) string {
				// Find the project by path
				for _, project := range capturedProjects {
					if project.Path == path {
						// Format project display
						status := ""
						switch project.Status {
						case denote.ProjectStatusActive:
							status = "● "
						case denote.ProjectStatusCompleted:
							status = "✓ "
						case denote.ProjectStatusPaused:
							status = "⏸ "
						case denote.ProjectStatusCancelled:
							status = "✗ "
						}
						
						priority := ""
						switch project.Priority {
						case denote.PriorityP1:
							priority = "[P1] "
						case denote.PriorityP2:
							priority = "[P2] "
						case denote.PriorityP3:
							priority = "[P3] "
						}
						
						due := ""
						if project.DueDate != "" {
							if denote.IsOverdue(project.DueDate) {
								due = " (overdue)"
							} else {
								days := denote.DaysUntilDue(project.DueDate)
								if days == 0 {
									due = " (today)"
								} else if days == 1 {
									due = " (tomorrow)"
								} else if days <= 7 {
									due = fmt.Sprintf(" (%dd)", days)
								}
							}
						}
						
						// Use title from metadata if available
						title := project.Note.Title
						if title == "" {
							// Fall back to just the filename without extension
							title = strings.TrimSuffix(filepath.Base(path), ".md")
						}
						
						// Add area prefix if available
						if project.Area != "" {
							title = fmt.Sprintf("%s / %s", project.Area, title)
						}
						
						return fmt.Sprintf("%s%s%s%s", status, priority, title, due)
					}
				}
				// If not found in projects, just return basename
				return strings.TrimSuffix(filepath.Base(path), ".md")
			}
		} else {
			// Regular task formatter
			m.ui.TaskFormatter = func(path string) string {
				task := m.findTaskByPath(path)
				if task != nil {
					return m.formatTaskLine(task)
				}
				return filepath.Base(path)
			}
		}
	} else {
		m.ui.TaskFormatter = nil
	}
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
		Render(getEnhancedDisplayName(m.previewFile, m.cwd, m.config.ShowTitles))
	
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

// toggleTaskMode switches between normal and task mode
func (m *model) toggleTaskMode() tea.Cmd {
	m.taskModeActive = !m.taskModeActive
	
	if m.taskModeActive {
		// Load tasks
		tasksDir := m.getTasksDirectory()
		scanner := denote.NewScanner(tasksDir)
		tasks, err := scanner.FindTasks()
		if err != nil {
			m.taskModeActive = false
			return ui.ShowError(fmt.Sprintf("Failed to load tasks from %s: %v", tasksDir, err))
		}
		
		if len(tasks) == 0 {
			m.taskModeActive = false
			return ui.ShowError(fmt.Sprintf("No tasks found in %s (looking for *__task*.md files)", tasksDir))
		}
		
		m.tasks = tasks
		
		// Apply default task sorting
		if m.taskSortBy == "" {
			m.taskSortBy = "priority"
		}
		denote.SortTasks(m.tasks, m.taskSortBy, m.reversedSort)
		
		// Update both files and filtered lists with task filenames
		m.files = make([]string, len(m.tasks))
		m.filtered = make([]string, len(m.tasks))
		for i, task := range m.tasks {
			m.files[i] = task.Path
			m.filtered[i] = task.Path
		}
		
		// Reset cursor and filters
		m.cursor = 0
		m.taskFilter = false
		m.tagFilter = false
		m.textFilter = false
		m.dailyFilter = false
		m.oldFilter = false
		
		// Set task formatter
		m.ui.TaskFormatter = func(path string) string {
			task := m.findTaskByPath(path)
			if task != nil {
				return m.formatTaskLine(task)
			}
			return filepath.Base(path)
		}
		m.ui.TaskModeActive = true
		
		// Apply area context if it exists
		if m.taskAreaContext != "" {
			m.refreshTaskView()
			return ui.ShowInfo(fmt.Sprintf("Task mode: %d tasks in area %s", len(m.filtered), m.taskAreaContext))
		}
		
		return ui.ShowInfo(fmt.Sprintf("Task mode: %d tasks", len(tasks)))
	} else {
		// Return to normal mode - reload all markdown files
		m.tasks = nil
		files, err := findMarkdownFiles(m.cwd)
		if err == nil {
			m.files = files
			m.filtered = files
		}
		m.cursor = 0
		
		// Clear task formatter
		m.ui.TaskFormatter = nil
		m.ui.TaskModeActive = false
		m.taskFilterType = ""
		m.taskAreaContext = ""
		m.taskStatusFilter = ""
		
		return ui.ShowInfo("Normal mode")
	}
}

// findTaskByPath finds a task by its file path
func (m model) findTaskByPath(path string) *denote.Task {
	for _, task := range m.tasks {
		if task.Path == path {
			return task
		}
	}
	return nil
}

// applyTaskSorting sorts tasks and updates the file lists
func (m *model) applyTaskSorting() {
	if m.taskSortBy == "" {
		m.taskSortBy = "priority" // default
	}
	
	// If we're showing projects, don't sort tasks
	if m.taskFilterType == "projects" {
		// Re-show projects with the current sort
		m.showProjectsOnly()
		return
	}
	
	// Sort the tasks
	denote.SortTasks(m.tasks, m.taskSortBy, m.reversedSort)
	
	// Refresh the view to apply filters
	m.refreshTaskView()
	
	// Update the task formatter to use the current sorted tasks
	m.ui.TaskFormatter = func(path string) string {
		task := m.findTaskByPath(path)
		if task != nil {
			return m.formatTaskLine(task)
		}
		return filepath.Base(path)
	}
	
	// Sync UI state to ensure view updates
	m.syncUIState()
}

// formatTaskLine formats a task for display in the file list
func (m model) formatTaskLine(task *denote.Task) string {
	// Status indicator
	status := ""
	switch task.Status {
	case denote.TaskStatusDone:
		status = "✓ "
	case denote.TaskStatusPaused:
		status = "⏸ "
	case denote.TaskStatusDelegated:
		status = "→ "
	case denote.TaskStatusDropped:
		status = "✗ "
	default:
		status = "○ "
	}
	
	// Priority
	priority := ""
	if task.Priority != "" {
		switch task.Priority {
		case denote.PriorityP1:
			priority = "[P1] "
		case denote.PriorityP2:
			priority = "[P2] "
		case denote.PriorityP3:
			priority = "[P3] "
		}
	}
	
	// Task ID
	taskIDStr := ""
	if task.TaskID > 0 {
		taskIDStr = fmt.Sprintf("#%d ", task.TaskID)
	}
	
	// Title (use metadata title if available)
	title := task.Note.Title
	if task.TaskMetadata.Title != "" {
		title = task.TaskMetadata.Title
	}
	
	// Project
	project := ""
	if task.Project != "" {
		project = " @" + task.Project
	}
	
	// Area
	area := ""
	if task.Area != "" {
		area = " #" + task.Area
	}
	
	// Estimate
	estimate := ""
	if task.Estimate > 0 {
		estimate = fmt.Sprintf(" ~%d", task.Estimate)
	}
	
	// Due date
	due := ""
	if task.DueDate != "" {
		if denote.IsOverdue(task.DueDate) {
			days := -denote.DaysUntilDue(task.DueDate)
			due = fmt.Sprintf(" (OVERDUE %d days)", days)
		} else {
			days := denote.DaysUntilDue(task.DueDate)
			if days == 0 {
				due = " (due today)"
			} else if days == 1 {
				due = " (due tomorrow)"
			} else if days <= 7 {
				due = fmt.Sprintf(" (due in %d days)", days)
			} else {
				due = fmt.Sprintf(" (due %s)", task.DueDate)
			}
		}
	}
	
	// Build the line
	return fmt.Sprintf("%s%s%s%s%s%s%s%s", status, priority, taskIDStr, title, project, area, estimate, due)
}

// applyTaskFilter applies a filter to the task list
func (m *model) applyTaskFilter(filterType string, filterValue string) {
	// Reload tasks if they're nil (e.g., coming back from projects view)
	if m.tasks == nil {
		scanner := denote.NewScanner(m.getTasksDirectory())
		tasks, err := scanner.FindTasks()
		if err != nil || len(tasks) == 0 {
			// No tasks found, update UI accordingly
			m.files = []string{}
			m.filtered = []string{}
			m.taskFilterType = filterType
			m.cursor = 0
			return
		}
		m.tasks = tasks
	}
	
	// Special handling for area filter - set context
	if filterType == "area" {
		m.taskAreaContext = filterValue
		m.taskStatusFilter = "" // Reset status filter when changing area
		m.refreshTaskView()
		return
	}
	
	// Special handling for project filter
	if filterType == "project" {
		m.taskFilterType = filterType
		m.taskFilterProject = filterValue
		m.refreshTaskView()
		return
	}
	
	// For other filters, set the appropriate filter
	if filterType == "all" {
		m.taskStatusFilter = ""
	} else {
		m.taskStatusFilter = filterType
	}
	m.taskFilterType = filterType
	
	m.refreshTaskView()
}

// refreshTaskView applies both area context and status filters
func (m *model) refreshTaskView() {
	if m.tasks == nil {
		return
	}
	
	// Start with all tasks
	filteredTasks := m.tasks
	
	// Apply filters in order: project -> area -> status
	
	// Apply project filter if set
	if m.taskFilterType == "project" && m.taskFilterProject != "" {
		filteredTasks = denote.FilterTasks(filteredTasks, "project", m.taskFilterProject)
	}
	
	// Apply area context if set
	if m.taskAreaContext != "" {
		filteredTasks = denote.FilterTasks(filteredTasks, "area", m.taskAreaContext)
	}
	
	// Then apply status filter if set
	if m.taskStatusFilter != "" && m.taskStatusFilter != "all" {
		filteredTasks = denote.FilterTasks(filteredTasks, m.taskStatusFilter, "")
	}
	
	// Apply current sort to the filtered tasks
	if m.taskSortBy != "" {
		denote.SortTasks(filteredTasks, m.taskSortBy, m.reversedSort)
	}
	
	// Update the file lists
	m.files = make([]string, len(filteredTasks))
	m.filtered = make([]string, len(filteredTasks))
	for i, task := range filteredTasks {
		m.files[i] = task.Path
		m.filtered[i] = task.Path
	}
	
	m.cursor = 0
	
	// Update the task formatter to use filteredTasks
	// We need to capture the filtered tasks for the formatter
	capturedFilteredTasks := filteredTasks
	m.ui.TaskFormatter = func(path string) string {
		// Look in the filtered tasks first
		for _, task := range capturedFilteredTasks {
			if task.Path == path {
				return m.formatTaskLine(task)
			}
		}
		// Fallback to looking in all tasks
		task := m.findTaskByPath(path)
		if task != nil {
			return m.formatTaskLine(task)
		}
		return filepath.Base(path)
	}
}

// showProjectsOnly switches to showing only project files
func (m *model) showProjectsOnly() {
	scanner := denote.NewScanner(m.getTasksDirectory())
	projects, err := scanner.FindProjects()
	
	if err == nil && len(projects) > 0 {
		// Store projects and clear tasks
		m.projects = projects
		m.tasks = nil
		
		// Update file lists with project paths
		m.files = make([]string, len(projects))
		m.filtered = make([]string, len(projects))
		for i, project := range projects {
			m.files[i] = project.Path
			m.filtered[i] = project.Path
		}
		
		m.taskFilterType = "projects"
		m.cursor = 0
		
		// Make sure we're in task mode
		m.ui.TaskModeActive = true
		
		// Create a closure that captures the projects slice
		capturedProjects := projects
		
		// Update formatter to show project info
		m.ui.TaskFormatter = func(path string) string {
			// Find the project by path
			for _, project := range capturedProjects {
				if project.Path == path {
					// Format project display
					status := ""
					switch project.Status {
					case denote.ProjectStatusActive:
						status = "● "
					case denote.ProjectStatusCompleted:
						status = "✓ "
					case denote.ProjectStatusPaused:
						status = "⏸ "
					case denote.ProjectStatusCancelled:
						status = "✗ "
					}
					
					priority := ""
					switch project.Priority {
					case denote.PriorityP1:
						priority = "[P1] "
					case denote.PriorityP2:
						priority = "[P2] "
					case denote.PriorityP3:
						priority = "[P3] "
					}
					
					due := ""
					if project.DueDate != "" {
						if denote.IsOverdue(project.DueDate) {
							due = " (overdue)"
						} else {
							days := denote.DaysUntilDue(project.DueDate)
							if days == 0 {
								due = " (today)"
							} else if days == 1 {
								due = " (tomorrow)"
							} else if days <= 7 {
								due = fmt.Sprintf(" (%dd)", days)
							}
						}
					}
					
					// Use title from metadata if available
					title := project.Note.Title
					if title == "" {
						// Fall back to just the filename without extension
						title = strings.TrimSuffix(filepath.Base(path), ".md")
					}
					
					// Add area prefix if available
					if project.Area != "" {
						title = fmt.Sprintf("%s / %s", project.Area, title)
					}
					
					return fmt.Sprintf("%s%s%s%s", status, priority, title, due)
				}
			}
			// If not found in projects, just return basename
			return strings.TrimSuffix(filepath.Base(path), ".md")
		}
	}
}

// showProjectTasks shows tasks filtered by project name
func (m *model) showProjectTasks(projectName string) tea.Cmd {
	// First reload all tasks
	scanner := denote.NewScanner(m.getTasksDirectory())
	tasks, err := scanner.FindTasks()
	if err != nil {
		return ui.ShowError("Failed to load tasks")
	}
	
	m.tasks = tasks
	
	// Apply default sorting
	if m.taskSortBy == "" {
		m.taskSortBy = "priority"
	}
	denote.SortTasks(m.tasks, m.taskSortBy, m.reversedSort)
	
	// Set the task formatter back
	m.ui.TaskFormatter = func(path string) string {
		task := m.findTaskByPath(path)
		if task != nil {
			return m.formatTaskLine(task)
		}
		return filepath.Base(path)
	}
	
	// Try different variations of the project name, including case variations
	// The project file might have "Oncall" as title but tasks have "oncall" as project
	
	baseVariations := []string{projectName}
	
	// Remove all hyphens
	withoutHyphens := strings.ReplaceAll(projectName, "-", "")
	if withoutHyphens != projectName {
		baseVariations = append(baseVariations, withoutHyphens)
	}
	
	// First word only
	if idx := strings.Index(projectName, "-"); idx > 0 {
		firstWord := projectName[:idx]
		baseVariations = append(baseVariations, firstWord)
	}
	
	// For each base variation, try both original case and lowercase
	variations := []string{}
	for _, base := range baseVariations {
		variations = append(variations, base)
		if lower := strings.ToLower(base); lower != base {
			variations = append(variations, lower)
		}
	}
	
	// Try each variation
	var filteredCount int
	var matchedVariation string
	foundMatch := false
	
	// Before trying variations, ensure we start with all tasks
	m.taskFilterType = ""
	m.taskStatusFilter = ""
	m.refreshTaskView()
	
	for _, variant := range variations {
		m.applyTaskFilter("project", variant)
		if len(m.filtered) > 0 {
			filteredCount = len(m.filtered)
			matchedVariation = variant
			m.taskFilterProject = variant
			foundMatch = true
			break
		}
	}
	
	if foundMatch && filteredCount > 0 {
		// Double-check that we actually have the right tasks
		// Count how many actually have this project
		actualCount := 0
		for _, taskPath := range m.filtered {
			if task := m.findTaskByPath(taskPath); task != nil && task.Project == matchedVariation {
				actualCount++
			}
		}
		
		if actualCount != filteredCount {
			return ui.ShowError(fmt.Sprintf("Filter mismatch: showing %d tasks but only %d have project '%s'", 
				filteredCount, actualCount, matchedVariation))
		}
		
		return ui.ShowInfo(fmt.Sprintf("Project '%s': %d tasks", matchedVariation, filteredCount))
	} else {
		// No matches - show what projects we have
		projectSet := make(map[string]bool)
		for _, task := range m.tasks {
			if task.Project != "" {
				projectSet[task.Project] = true
			}
		}
		var availableProjects []string
		for p := range projectSet {
			availableProjects = append(availableProjects, p)
		}
		sort.Strings(availableProjects)
		
		return ui.ShowError(fmt.Sprintf("No tasks for '%s'. Available projects: %v", projectName, availableProjects))
	}
}

// prepareProjectItems converts projects to display items with task counts
func (m *model) prepareProjectItems() []ui.ProjectItem {
	items := make([]ui.ProjectItem, 0, len(m.projects))
	
	// Count tasks per project
	taskCounts := m.countTasksPerProject()
	
	for _, project := range m.projects {
		// Apply area filter if set
		if m.taskAreaContext != "" && project.ProjectMetadata.Area != m.taskAreaContext {
			continue
		}
		
		projectKey := project.ProjectMetadata.Identifier
		if projectKey == "" {
			projectKey = project.ProjectMetadata.Title
		}
		
		openCount := 0
		doneCount := 0
		if counts, ok := taskCounts[projectKey]; ok {
			openCount = counts.open
			doneCount = counts.done
		}
		
		item := ui.ProjectItem{
			Project:   project,
			Title:     project.ProjectMetadata.Title,
			Status:    project.ProjectMetadata.Status,
			Priority:  project.GetPriorityInt(),
			StartDate: project.GetParsedStartDate(),
			DueDate:   project.GetParsedDueDate(),
			OpenTasks: openCount,
			DoneTasks: doneCount,
		}
		items = append(items, item)
	}
	
	return items
}

// countTasksPerProject counts open and done tasks for each project
func (m *model) countTasksPerProject() map[string]struct{ open, done int } {
	counts := make(map[string]struct{ open, done int })
	
	// Ensure tasks are loaded
	if m.tasks == nil {
		scanner := denote.NewScanner(m.getTasksDirectory())
		tasks, err := scanner.FindTasks()
		if err == nil {
			m.tasks = tasks
		}
	}
	
	// First, build a map of all possible project keys (title and identifier)
	projectKeys := make(map[string]string) // maps any variant to canonical key
	for _, project := range m.projects {
		canonicalKey := project.ProjectMetadata.Identifier
		if canonicalKey == "" {
			canonicalKey = project.ProjectMetadata.Title
		}
		
		// Map identifier to canonical
		if project.ProjectMetadata.Identifier != "" {
			projectKeys[project.ProjectMetadata.Identifier] = canonicalKey
		}
		// Map title to canonical
		if project.ProjectMetadata.Title != "" {
			projectKeys[project.ProjectMetadata.Title] = canonicalKey
		}
		// Also try lowercase variants
		projectKeys[strings.ToLower(project.ProjectMetadata.Title)] = canonicalKey
		if project.ProjectMetadata.Identifier != "" {
			projectKeys[strings.ToLower(project.ProjectMetadata.Identifier)] = canonicalKey
		}
	}
	
	// Count tasks by project
	for _, task := range m.tasks {
		if task.Project != "" {
			// Find the canonical project key
			canonicalKey := task.Project
			if mapped, ok := projectKeys[task.Project]; ok {
				canonicalKey = mapped
			} else if mapped, ok := projectKeys[strings.ToLower(task.Project)]; ok {
				canonicalKey = mapped
			}
			
			c := counts[canonicalKey]
			switch task.Status {
			case denote.TaskStatusDone:
				c.done++
			case denote.TaskStatusOpen, denote.TaskStatusPaused:
				c.open++
			}
			counts[canonicalKey] = c
		}
	}
	
	return counts
}

func main() {
	// Parse command line flags
	var tag = flag.String("tag", "", "Filter notes by tag (e.g., --tag=@mikeh)")
	var openID = flag.String("open-id", "", "Open note with specific Denote identifier (e.g., --open-id=20241225T093015)")
	var taskMode = flag.Bool("tasks", false, "Start in task mode (requires denote_tasks_support=true in config)")
	var area = flag.String("area", "", "Filter tasks by area when starting in task mode (e.g., -tasks -area work)")
	flag.Parse()

	// Load config first
	config := LoadConfig()
	
	// Validate area flag usage
	if *area != "" && !*taskMode {
		fmt.Fprintf(os.Stderr, "Error: -area flag requires -tasks flag\n")
		flag.Usage()
		os.Exit(1)
	}
	
	// Handle directory argument (remaining args after flags)
	args := flag.Args()
	if len(args) > 0 {
		dir := args[0]
		// Expand tilde in directory path
		if strings.HasPrefix(dir, "~/") {
			home, _ := os.UserHomeDir()
			dir = filepath.Join(home, dir[2:])
		}
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

	// Handle --open-id flag
	if *openID != "" {
		// Validate identifier format (should be exactly 15 chars: YYYYMMDDTHHMMSS)
		if len(*openID) != 15 {
			fmt.Printf("Invalid identifier format. Expected 15 characters (YYYYMMDDTHHMMSS), got %d\n", len(*openID))
			os.Exit(1)
		}
		
		// Basic format validation - should be digits with T in the middle
		if (*openID)[8] != 'T' {
			fmt.Printf("Invalid identifier format. Expected 'T' at position 9\n")
			os.Exit(1)
		}
		
		// Find all markdown files
		files, err := findMarkdownFiles(".")
		if err != nil {
			log.Fatal(err)
		}
		
		// Look for file with matching Denote identifier
		var matchedFile string
		for _, file := range files {
			filename := filepath.Base(file)
			// Check if filename starts with the exact identifier followed by - or _
			if len(filename) > 16 && strings.HasPrefix(filename, *openID) {
				// Make sure the identifier is followed by - or _ (not another digit)
				if filename[15] == '-' || filename[15] == '_' {
					matchedFile = file
					break
				}
			}
		}
		
		if matchedFile == "" {
			fmt.Printf("No file found with identifier: %s\n", *openID)
			os.Exit(1)
		}
		
		// Open the file in editor
		var editor string
		var editorArgs []string
		
		// Use configured editor first
		if config.Editor != "" {
			editor, editorArgs = parseCommand(config.Editor)
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
			fmt.Printf("No editor configured. Found file: %s\n", matchedFile)
			os.Exit(0)
		}
		
		// Add the filename to the arguments
		editorArgs = append(editorArgs, matchedFile)
		
		// Open the file in the editor
		cmd := exec.Command(editor, editorArgs...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error opening editor: %v\n", err)
			fmt.Printf("File: %s\n", matchedFile)
			os.Exit(1)
		}
		
		// Exit after opening file
		os.Exit(0)
	}

	p := tea.NewProgram(initialModel(config, *tag, *taskMode, *area), tea.WithAltScreen())
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

// processTaskFieldUpdate handles updating a task field with the current input value
func (m *model) processTaskFieldUpdate() tea.Cmd {
	if m.taskBeingEdited == nil || m.taskEditField == "" {
		return nil
	}
	
	value := strings.TrimSpace(m.taskEditInput.Value())
	var err error
	var successMsg string
	
	switch m.taskEditField {
	case "due":
		// Parse relative dates
		if value != "" {
			parsedDate, parseErr := parseRelativeDate(value)
			if parseErr != nil {
				return ui.ShowError(fmt.Sprintf("Invalid date: %v", parseErr))
			}
			value = parsedDate
		}
		err = denote.UpdateTaskDueDate(m.taskBeingEdited.Path, value)
		if err == nil {
			m.taskBeingEdited.DueDate = value
			successMsg = fmt.Sprintf("Updated due date for task #%d", m.taskBeingEdited.TaskID)
		}
		
	case "start":
		// Parse relative dates
		if value != "" {
			parsedDate, parseErr := parseRelativeDate(value)
			if parseErr != nil {
				return ui.ShowError(fmt.Sprintf("Invalid date: %v", parseErr))
			}
			value = parsedDate
		}
		err = denote.UpdateTaskStartDate(m.taskBeingEdited.Path, value)
		if err == nil {
			m.taskBeingEdited.StartDate = value
			successMsg = fmt.Sprintf("Updated start date for task #%d", m.taskBeingEdited.TaskID)
		}
		
	case "estimate":
		estimate := 0
		if value != "" {
			estimate, err = strconv.Atoi(value)
			if err != nil {
				return ui.ShowError("Estimate must be a number")
			}
		}
		err = denote.UpdateTaskEstimate(m.taskBeingEdited.Path, estimate)
		if err == nil {
			m.taskBeingEdited.Estimate = estimate
			successMsg = fmt.Sprintf("Updated estimate for task #%d", m.taskBeingEdited.TaskID)
		}
		
	case "priority":
		// Normalize priority input
		priority := value
		if priority == "1" {
			priority = "p1"
		} else if priority == "2" {
			priority = "p2"
		} else if priority == "3" {
			priority = "p3"
		}
		err = denote.UpdateTaskPriority(m.taskBeingEdited.Path, priority)
		if err == nil {
			m.taskBeingEdited.Priority = priority
			successMsg = fmt.Sprintf("Updated priority for task #%d", m.taskBeingEdited.TaskID)
		}
		
	case "project":
		err = denote.UpdateTaskProject(m.taskBeingEdited.Path, value)
		if err == nil {
			m.taskBeingEdited.Project = value
			successMsg = fmt.Sprintf("Updated project for task #%d", m.taskBeingEdited.TaskID)
		}
		
	case "area":
		err = denote.UpdateTaskArea(m.taskBeingEdited.Path, value)
		if err == nil {
			m.taskBeingEdited.Area = value
			successMsg = fmt.Sprintf("Updated area for task #%d", m.taskBeingEdited.TaskID)
		}
		
	case "tags":
		// Parse comma-delimited tags
		var tags []string
		if value != "" {
			// Split by comma and trim whitespace
			parts := strings.Split(value, ",")
			for _, tag := range parts {
				trimmed := strings.TrimSpace(tag)
				if trimmed != "" {
					tags = append(tags, trimmed)
				}
			}
		}
		err = denote.UpdateTaskTags(m.taskBeingEdited.Path, tags)
		if err == nil {
			m.taskBeingEdited.Tags = tags
			successMsg = fmt.Sprintf("Updated tags for task #%d", m.taskBeingEdited.TaskID)
		}
	}
	
	if err != nil {
		return ui.ShowError(fmt.Sprintf("Failed to update: %v", err))
	}
	
	// Reset to field selection (stay in edit mode)
	m.taskEditField = ""
	m.taskEditInput.SetValue("")
	m.taskEditInput.Placeholder = ""
	
	return ui.ShowSuccess(successMsg)
}

// parseRelativeDate parses relative date strings like "today", "tomorrow", "3d", "1w"
func parseRelativeDate(dateStr string) (string, error) {
	if dateStr == "" {
		return "", nil
	}
	
	now := time.Now()
	var targetDate time.Time
	
	lowerStr := strings.ToLower(dateStr)
	
	// Check for relative date keywords
	switch lowerStr {
	case "today":
		targetDate = now
	case "tomorrow":
		targetDate = now.AddDate(0, 0, 1)
	case "next week":
		targetDate = now.AddDate(0, 0, 7)
	case "next month":
		targetDate = now.AddDate(0, 1, 0)
	default:
		// Try parsing as day of week
		if weekday, ok := parseDayOfWeek(lowerStr); ok {
			targetDate = getNextWeekday(now, weekday)
		} else if duration, ok := parseRelativeDuration(lowerStr); ok {
			// Parse relative durations like "3d", "2w", "1m"
			targetDate = now.Add(duration)
		} else {
			// Try parsing as absolute date
			parsed, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				return "", fmt.Errorf("invalid date format: %s (use YYYY-MM-DD, day name, or relative like '3d', '2w', '1m')", dateStr)
			}
			targetDate = parsed
		}
	}
	
	return targetDate.Format("2006-01-02"), nil
}

// parseDayOfWeek parses day names
func parseDayOfWeek(day string) (time.Weekday, bool) {
	days := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"sun":       time.Sunday,
		"monday":    time.Monday,
		"mon":       time.Monday,
		"tuesday":   time.Tuesday,
		"tue":       time.Tuesday,
		"wednesday": time.Wednesday,
		"wed":       time.Wednesday,
		"thursday":  time.Thursday,
		"thu":       time.Thursday,
		"friday":    time.Friday,
		"fri":       time.Friday,
		"saturday":  time.Saturday,
		"sat":       time.Saturday,
	}
	
	weekday, ok := days[day]
	return weekday, ok
}

// getNextWeekday gets the next occurrence of a weekday
func getNextWeekday(from time.Time, weekday time.Weekday) time.Time {
	daysUntil := int(weekday - from.Weekday())
	if daysUntil <= 0 {
		daysUntil += 7
	}
	return from.AddDate(0, 0, daysUntil)
}

// parseRelativeDuration parses durations like "3d", "2w", "1m"
func parseRelativeDuration(s string) (time.Duration, bool) {
	if len(s) < 2 {
		return 0, false
	}
	
	// Extract number and unit
	numStr := s[:len(s)-1]
	unit := s[len(s)-1:]
	
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, false
	}
	
	switch unit {
	case "d": // days
		return time.Duration(num) * 24 * time.Hour, true
	case "w": // weeks
		return time.Duration(num) * 7 * 24 * time.Hour, true
	case "m": // months (approximate as 30 days)
		return time.Duration(num) * 30 * 24 * time.Hour, true
	default:
		return 0, false
	}
}