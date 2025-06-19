package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
)

type model struct {
	files      []string        // all markdown files
	filtered   []string        // filtered results
	cursor     int             // which file is selected
	selected   string          // selected file
	searchMode bool            // are we in search mode?
	search     textinput.Model // search input
	cwd        string          // current working directory
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

	return model{
		files:    files,
		filtered: files, // Initially show all files
		search:   ti,
		cwd:      cwd,
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
func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
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
			return m, tea.Quit

		case "esc":
			if m.searchMode {
				// Exit search mode
				m.searchMode = false
				m.search.SetValue("")
				m.filtered = m.files
				m.cursor = 0
			}

		case "/":
			// Enter search mode
			m.searchMode = true
			m.search.Focus()
			return m, nil

		case "up", "k":
			if !m.searchMode && m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if !m.searchMode && m.cursor < len(m.filtered)-1 {
				m.cursor++
			}

		case "enter":
			if m.searchMode {
				// Exit search mode on enter
				m.searchMode = false
			} else if m.cursor < len(m.filtered) {
				m.selected = m.filtered[m.cursor]
				return m, tea.Quit
			}
		}
	}

	// Handle search input
	if m.searchMode {
		m.search, cmd = m.search.Update(msg)
		query := m.search.Value()
		m.filtered = filterFiles(m.files, query)
		m.cursor = 0 // Reset cursor when filtering
	}

	return m, cmd
}
func (m model) View() string {
	var s strings.Builder
	
	// Header
	s.WriteString(fmt.Sprintf("Notes (%d files)\n\n", len(m.filtered)))

	if len(m.filtered) == 0 && m.searchMode && m.search.Value() != "" {
		s.WriteString("No files match your search.\n\n")
	} else if len(m.files) == 0 {
		s.WriteString("No markdown files found.\n\n")
	} else {
		// Show file list
		maxVisible := 20 // Show max 20 files to leave room for search
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
			s.WriteString(fmt.Sprintf("%s%s\n", cursor, displayName))
		}
		
		if len(m.filtered) > maxVisible {
			s.WriteString(fmt.Sprintf("\n... and %d more files\n", len(m.filtered)-maxVisible))
		}
	}

	// Search field at bottom
	s.WriteString("\n")
	if m.searchMode {
		s.WriteString(fmt.Sprintf("Search: %s\n", m.search.View()))
		s.WriteString("[Esc] cancel [Enter] select")
	} else {
		s.WriteString("[/] search [Enter] open [q] quit")
	}

	return s.String()
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