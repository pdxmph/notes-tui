# CLAUDE.md - Project Context for notes-tui

This file contains important context about the notes-tui project to help AI assistants understand the codebase, architecture decisions, and future direction.

## Project Overview

**notes-tui** is a terminal user interface (TUI) application for managing Markdown notes with Denote-style naming conventions. Built with Go and the Bubble Tea framework, it provides a fast, keyboard-driven interface for note management.

### Key Features
- Markdown note browsing and creation
- Denote-style filename support (YYYYMMDDTHHMMSS--title__tags.md)
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
- Format: `YYYYMMDDTHHMMSS--title__tag1_tag2.md`
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

---

Last updated: 2024-06-29
Version: 0.5.1