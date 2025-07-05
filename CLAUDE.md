# CLAUDE.md - Project Context for notes-tui

This file contains important context about the notes-tui project to help AI assistants understand the codebase, architecture decisions, and future direction.

## Project Overview

**notes-tui** is a terminal user interface (TUI) application for managing Markdown notes with Denote-style naming conventions. Built with Go and the Bubble Tea framework, it provides a fast, keyboard-driven interface for note management.

### Key Features
- Markdown note browsing and creation
- Denote-style filename support (YYYYMMDDTHHMMSS-title__tags.md)
- Multiple search/filter modes (text, tags, tasks, daily notes)
- Configurable sorting options
- TaskWarrior integration
- Internal markdown preview with syntax highlighting
- External editor/preview command support

## Architecture

### Technology Stack
- **Language**: Go 1.23+
- **TUI Framework**: Bubble Tea (Elm Architecture pattern)
- **Styling**: Lipgloss
- **Search**: Ripgrep integration for tag/task search
- **Config**: TOML format

### Code Organization

```
.
├── main.go                 # Main application (currently monolithic)
├── internal/ui/           # New UI package (recently refactored)
│   ├── components.go      # Reusable UI components
│   ├── theme.go          # Theme system
│   ├── layout.go         # Layout management
│   ├── views.go          # View composition
│   ├── integration.go    # Integration with main model
│   ├── messages.go       # Status message system
│   └── markdown.go       # Markdown rendering
├── config.example.toml    # Configuration example
└── test-notes/           # Test markdown files
```

### Recent Refactoring (2024-06-29)

We recently refactored the frontend architecture to improve modularity:

1. **Extracted UI components** from the monolithic main.go
2. **Created a theme system** with support for multiple themes
3. **Added status messages** for user feedback
4. **Improved help menu** to show all keyboard shortcuts
5. **Maintained backward compatibility** while improving architecture

The refactoring follows these principles:
- Separation of concerns (UI logic vs business logic)
- Reusable components
- Type-safe interfaces
- Minimal external dependencies

### Theme System (2024-06-30)

Implemented configurable theme selection with five built-in themes:

1. **default** - Balanced colors for most terminals
2. **dark** - Optimized for dark terminals  
3. **light** - Optimized for light terminals with dark text
4. **high-contrast** - Maximum contrast for accessibility
5. **minimal** - Monochrome with minimal color usage

The theme system is extensible and supports comprehensive styling of all UI components including:
- List views (cursor, items, empty messages)
- Modal dialogs (titles, prompts, help text)
- Headers (title, filters, sort info)
- Help bars (keys, descriptions, separators)
- Popovers (borders, titles, scroll bars)
- Status messages (info, success, warning, error)

## Keyboard Shortcuts

The application uses vim-like modal keybindings:

### Normal Mode
- `/` - Search files
- `Enter` - Preview file
- `e` - Edit in external editor
- `n` - Create new note
- `d` - Create/open daily note
- `D` - Show all daily notes
- `#` - Search by tag
- `t` - Show files with tasks
- `o` - Sort menu
- `O` - Filter by days old
- `R` - Rename to Denote format
- `X` - Delete file
- `Ctrl+K` - Create TaskWarrior task (when enabled)
- `q` - Quit

### Navigation
- `j`/`k` or arrows - Move up/down
- `gg` - Jump to top
- `G` - Jump to bottom

## Configuration

Configuration is stored in `~/.config/notes-tui/config.toml`:

```toml
notes_directory = "/path/to/notes"
editor = "nvim"                    # or "code --wait", etc.
preview_command = "glow"           # optional external preview
add_frontmatter = true             # YAML frontmatter
denote_filenames = true            # Use Denote naming
show_titles = true                 # Extract titles from files
prompt_for_tags = true             # Ask for tags when creating
taskwarrior_support = true         # Enable Ctrl+K
theme = "default"                  # Theme selection
```

## Denote Integration

The app supports Denote-style filenames:
- Format: `YYYYMMDDTHHMMSS-title__tag1_tag2.md`
- Automatic generation when creating notes
- Rename existing files to Denote format with `R`
- Sort by Denote identifier
- Parse and display human-readable titles

## Development Guidelines

### Code Style
- Follow standard Go conventions
- Keep functions focused and small
- Use meaningful variable names
- Add comments for complex logic

### Testing
- Test with the `test-notes/` directory
- Ensure ripgrep is installed for tag/task search
- Test all keyboard shortcuts
- Verify configuration loading

### Adding Features
1. Consider the modal interface design
2. Maintain backward compatibility
3. Update help text when adding shortcuts
4. Add configuration options when appropriate
5. Update this file with significant changes

## Future Enhancements

### Planned Features (GitHub Issues)
- **#30**: Status messages for task creation
- **#31**: Status messages for filter operations  
- **#32**: Theme selection support

### Potential Improvements
1. **Package Structure**: Further modularize main.go
2. **Testing**: Add comprehensive test suite
3. **Search**: Full-text search within files
4. **Templates**: Note templates for different types
5. **Sync**: Git integration for note syncing
6. **Export**: Various export formats

### Architecture Goals
- Extract file operations to separate package
- Create proper error types
- Implement state machine for modes
- Add plugin/extension system
- Support custom keybindings

## Common Tasks

### Adding a New Keyboard Shortcut
1. Add case in `Update()` method in main.go
2. Update help text in `internal/ui/views.go`
3. Document in README.md
4. Add to this file's shortcut list

### Adding a New Theme
1. Create theme function in `internal/ui/theme.go`
2. Add to `GetTheme()` switch statement
3. Document color choices
4. Test with various terminal backgrounds

### Adding a Status Message
1. Use `ui.ShowSuccess()`, `ui.ShowError()`, etc.
2. Return as command: `cmds = append(cmds, ui.ShowSuccess("message"))`
3. Message auto-dismisses after ~3 seconds

## Debugging Tips

- Use `fmt.Fprintf(os.Stderr, ...)` for debug output
- Check `~/.config/notes-tui/config.toml` for config issues
- Ensure ripgrep is in PATH for search features
- Terminal must support Unicode for some UI elements

## Contributing

When contributing:
1. Maintain the vim-like interface philosophy
2. Keep the UI responsive and fast
3. Preserve existing keyboard shortcuts
4. Update documentation for new features
5. Consider configuration options for new behaviors

## Questions/Design Decisions

### Why Bubble Tea?
- Excellent terminal handling
- Elm architecture is predictable
- Good ecosystem (Lipgloss, Bubbles)
- Active development

### Why Monolithic main.go?
- Started simple, grew organically
- Refactoring in progress
- Goal: maintain simplicity while improving structure

### Why Ripgrep for search?
- Extremely fast
- Handles large note collections
- Good regex support
- Available on all platforms

### Title Display Convention
When `show_titles = true` is configured, titles are extracted in the following priority order:
1. **YAML frontmatter `title:` field** (primary source)
2. **First level-1 Markdown heading** (`# Heading`) if no frontmatter title
3. **Denote filename parsing** (slug conversion) as last resort

The slug fallback (converting kebab-case to spaces) should be extremely rare and only occurs when:
- No YAML frontmatter with `title:` field exists
- No `# Heading` appears in the first ~20 lines of the file
- The file uses Denote naming convention

**Important**: Titles are displayed AS-IS from the source, preserving original capitalization and formatting.

## Task Management Implementation (2025-07-05)

### Current State
The task management layer has been significantly developed with the following features:

#### Core Task Features
1. **Task Mode** - Toggle with `t` key to enter dedicated task management interface
2. **Task Creation** - `n` key creates tasks with YAML frontmatter (uses Unix timestamp for task_id)
3. **Task Display** - Custom TaskListView component with color-coded statuses and priorities
4. **Quick Actions**:
   - `d` - Mark task as done
   - `p` - Toggle pause/unpause
   - `1`, `2`, `3` - Set priority levels
5. **Filtering System** (`f` key):
   - All tasks, Open only, Active (not done/dropped)
   - Overdue, Due this week
   - Filter by area (dynamic submenu)
   - Filter by project
   - Projects view
6. **Sorting Options** (`o` key):
   - Due date, Priority, Status, Modified
   - Reverse sort option
7. **Project Support** - View projects and navigate to their tasks

#### Technical Architecture
- **Denote Package** (`internal/denote/`):
  - `types.go` - Task/Project structs with metadata
  - `parser.go` - Parse Denote filenames and YAML frontmatter
  - `scanner.go` - Find and filter task files
  - `update.go` - Update task status and priority in frontmatter
- **UI Components** (`internal/ui/`):
  - `task_list.go` - Specialized list view for tasks
  - Integration with main model via `TaskFormatter` and `TaskModeActive`
- **Main App** (`main.go`):
  - Task mode state management
  - Keyboard shortcut handling
  - Task creation/update logic

#### Key Differences from notes-cli
1. **ID System**: Uses Unix timestamp vs sequential counter
2. **Interface**: Modal/keyboard-driven vs command-line flags
3. **Missing Features**:
   - No task metadata edit dialog (can't update due dates, estimates, etc.)
   - No log entry functionality
   - No bulk operations (ranges like `3-5`)
   - No shell completions

### Areas Needing Attention
1. **Code Organization**: Task logic spread across main.go (2800+ lines)
2. **ID Management**: Consider implementing sequential IDs with counter file
3. **Feature Parity**: Several notes-cli features could enhance the TUI
4. **Testing**: No comprehensive tests for task functionality

### Recommended Next Steps
1. Create GitHub issues for specific enhancements
2. Consider extracting task logic into dedicated package
3. Implement missing high-value features (metadata editing, log entries)
4. Add tests for critical task operations
5. Document task mode features in README

## Recent Fixes and Architectural Decisions (2025-07-05 continued)

### Task Management Enhancements

#### 1. Tasks Directory Configuration
- Added `tasks_directory` configuration option to support separate task storage
- Tasks can now be kept separate from regular notes (important for Denote formatting requirements)
- Falls back to notes_directory if not specified
- Helper method `getTasksDirectory()` ensures consistent directory usage

#### 2. Sequential Task ID System
- Implemented proper sequential task IDs (not Unix timestamps) for easier CLI referencing
- Created `internal/denote/id_counter.go` with mutex-protected counter
- Counter file `.notes-cli-id-counter.json` stored in tasks directory
- Automatically scans existing tasks on first run to find highest ID
- Compatible with notes-cli format

#### 3. Task Metadata Editing
- Added comprehensive task metadata editing with `u` key in task mode
- Editable fields: due date, start date, estimate, priority, project, area, tags
- Smart date parsing supports:
  - Relative dates: "today", "tomorrow", "3d", "1w", "next week"
  - Day names: "monday", "friday" (finds next occurrence)
  - Standard format: "YYYY-MM-DD"
- Tag editing shows existing tags as comma-delimited list
- UI stays in edit mode after changes for multiple edits
- Fixed "Cancel" to "Done" for clarity (changes are saved immediately)

#### 4. Area Context Filtering
- **Major architectural change**: Areas are now persistent contexts, not filters
- Area selection persists while applying other filters (open, active, overdue, etc.)
- Backspace key has intelligent behavior:
  - With both area and status filter: clears status only
  - With just area: clears area
- Filter mode shows current area context and option to clear with 'x'
- Header displays both area context and status filters separately

#### 5. Robust Frontmatter Parsing
- Fixed issue where `---` horizontal rules in content corrupted frontmatter updates
- Implemented `looksLikeYAML()` validation to distinguish YAML from content
- Parser now only treats `---` as frontmatter boundary if preceded by valid YAML
- Handles TaskWarrior imports that included `---` separators in content

#### 6. Project Filtering Fixes
- Fixed project name matching to handle case differences (e.g., "Oncall" vs "oncall")
- Added `identifier` field to ProjectMetadata for explicit project keys
- Fixed bug where project filter was incorrectly treated as status filter
- Project filtering now properly filters tasks instead of showing all

### Important Principles

1. **No Data Transformation**: Display data exactly as stored, never apply cosmetic transformations
2. **Filter Architecture**: Distinguish between contexts (area), filters (status), and queries (project)
3. **Frontmatter Safety**: Always validate YAML structure before treating `---` as boundaries
4. **User Feedback**: Clear status messages for all operations

### Known Issues and Technical Debt

1. **Data Inconsistencies**: TaskWarrior imports created mismatches between project titles and task project fields
2. **Code Organization**: Task logic still heavily concentrated in main.go
3. **Filter Complexity**: Multiple filter dimensions (area, status, project) need clearer architecture

---

Last updated: 2025-07-05
Version: 0.5.2