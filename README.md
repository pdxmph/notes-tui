# notes-tui

A lightweight TUI for Markdown notes built with Go and Bubble Tea.

## Features

- **Fast file browsing** with fuzzy search
- **Instant note creation** with title-to-filename conversion
- **Daily notes** with special handling for `yyyy-mm-dd-daily.md` files
- **Tag search** using `#tag` syntax in content and YAML front matter
- **Task search** to find open `- [ ]` checkboxes
- **Dual preview modes**: internal popover and external command
- **Configurable** directory, editor, and preview commands

## Installation

```bash
go install github.com/pdxmph/notes-tui@latest
```

Or clone and build:

```bash
git clone https://github.com/pdxmph/notes-tui.git
cd notes-tui
go build
```

## Usage

```bash
# Use current directory
notes-tui

# Use specific directory
notes-tui /path/to/notes
```

## Configuration

Create a config file at `~/.config/notes-tui/config.toml`:

```toml
# Default directory for notes
notes_directory = "/Users/mph/notes"

# Editor command with arguments (supports spaces)
editor = "emacsclient --create-frame --no-wait"

# External preview command (optional)
preview_command = "glow --style dark --pager"
```

See `config.example.toml` for more examples.

### Configuration Options

- **`notes_directory`**: Default directory for notes (overridden by command line argument)
- **`editor`**: Editor command with arguments. Supports commands with spaces. Falls back to `$EDITOR` if not set.
- **`preview_command`**: External preview command (optional). When set, `Enter` key uses external preview instead of internal popover.
- **`add_frontmatter`**: Add YAML frontmatter to new notes (default: false). When true, notes get frontmatter with title and date.
- **`prompt_for_tags`**: Prompt for tags when creating notes (default: false). Only works when `add_frontmatter` is true. Tags are stored as YAML array.
- **`denote_filenames`**: Use Denote-style filenames (default: false). Format: `YYYYMMDDTHHMMSS-title.md`
- **`show_titles`**: Show extracted titles instead of filenames in list (default: false)
- **`taskwarrior_support`**: Enable Ctrl+K to create TaskWarrior tasks (default: false)
- **`denote_tasks_support`**: Enable task management mode (default: false) - experimental feature
- **`tasks_directory`**: Directory for task files (defaults to `notes_directory`)
- **`theme`**: Color theme selection (default: "default"). Available themes:
  - `"default"` - Balanced colors for most terminals
  - `"dark"` - Optimized for dark terminals
  - `"light"` - Optimized for light terminals with dark text
  - `"high-contrast"` - Maximum contrast for accessibility
  - `"minimal"` - Monochrome with minimal color usage

### Editor Examples

```toml
editor = "emacsclient -cn"  # Emacs
editor = "code --wait"                            # VS Code
editor = "vim"                                    # Vim
editor = "subl --wait"                            # Sublime Text
```

### Preview Examples

```toml
# External preview replaces internal preview when configured
preview_command = "glow -p"                                # Glow with pager
preview_command = "bat --style=plain --color=always"       # Bat with color
preview_command = "mdcat"                                  # mdcat viewer

# Leave unset for internal preview popover (default)
# preview_command = ""
```

## Key Bindings

- **`/`**: Search files
- **`Enter`**: Preview (internal popover or external command if configured)
- **`e`**: Edit in configured editor
- **`X`**: Delete file (requires `y` to confirm)
- **`n`**: Create new note
- **`d`**: Create/open daily note
- **`D`**: Show only daily notes
- **`#`**: Search by tag
- **`t`**: Show files with open tasks
- **`o`**: Open sort menu
- **`O`**: Filter notes by age (e.g., last 7 days)
- **`R`**: Rename file to Denote format
- **`T`**: Toggle task mode (requires `denote_tasks_support = true`)
- **`g`** then **`g`**: Jump to top of list
- **`G`**: Jump to bottom of list
- **`q`**: Quit

### In Preview Mode

- **`Esc`** or **`q`**: Close preview
- **`e`**: Edit file from preview
- **`↑↓`** or **`j/k`**: Scroll
- **`PgUp/PgDn`** or **`Space`**: Page up/down

### In Sort Menu (`o`)

- **`d`**: Sort by date (newest first)
- **`m`**: Sort by modified time (newest first)
- **`t`**: Sort by title (alphabetical)
- **`i`**: Sort by Denote identifier (newest first)
- **`r`**: Reverse current sort order
- **`Esc`**: Exit sort menu

### In Task Mode (`T` with `denote_tasks_support = true`)

**Navigation & Viewing:**
- **`Enter`**: View task details
- **`e`**: Edit task file
- **`/`**: Search tasks
- **`T`**: Return to notes mode
- **`Backspace`**: Clear filters/go back

**Task Operations:**
- **`n`**: Create new task
- **`u`**: Update task metadata (due date, priority, project, etc.)
- **`d`**: Mark task as done
- **`p`**: Toggle pause/unpause
- **`1`/`2`/`3`**: Set priority (high/medium/low)
- **`X`**: Delete task

**Filtering & Sorting:**
- **`f`**: Filter tasks menu
  - **`a`**: All tasks
  - **`o`**: Open tasks only
  - **`c`**: Active tasks (not done/dropped)
  - **`v`**: Overdue tasks
  - **`w`**: Due this week
  - **`A`**: Filter by area
  - **`p`**: Filter by project
  - **`P`**: Show projects
  - **`x`**: Clear area context
- **`o`**: Sort tasks menu
  - **`d`**: Sort by due date
  - **`p`**: Sort by priority
  - **`s`**: Sort by status
  - **`m`**: Sort by modified
  - **`r`**: Reverse sort

**Task Metadata Editor (`u`):**
- Enter dates as: `YYYY-MM-DD`, `today`, `tomorrow`, `3d`, `1w`, `monday`
- Priority: `1` (high), `2` (medium), `3` (low)
- Tags: comma-separated list
- Press `Escape` when done editing

## Features in Detail

### Search Modes

- **File search** (`/`): Fuzzy search by filename
- **Tag search** (`#`): Find files containing hashtags in content or YAML front matter
- **Task search** (`t`): Find files with open `- [ ]` checkboxes

### Note Creation

- **Regular notes** (`n`): Creates `title-in-kebab-case.md`
  - With `add_frontmatter = true`: Adds YAML frontmatter with title and date
  - With `prompt_for_tags = true`: Prompts for comma-separated tags after title
- **Daily notes** (`d`): Creates `YYYY-MM-DD-daily.md` with template

### Tag Support

Finds tags in multiple formats:

- Inline: `#tag`
- YAML frontmatter: `tags: [tag1, tag2]`
- YAML lists:
  ```yaml
  tags:
    - tag1
    - tag2
  ```

### Task Management (Experimental)

When `denote_tasks_support = true` is set in your config, task mode provides a complete task management system:

- **Task Files**: Uses Denote naming with `__task__` tag (e.g., `20250105T093045-project-setup__task__.md`)
- **Sequential IDs**: Each task gets a unique numeric ID for easy reference
- **Metadata**: Tasks support due dates, priorities, projects, areas, and tags
- **Status Tracking**: Open, Done, Paused, or Dropped
- **Smart Filtering**: By status, area, project, due date
- **Area Context**: Areas act as persistent filters while browsing tasks

Task files use YAML frontmatter:
```yaml
---
title: "Implement user authentication"
task_id: 42
status: "open"
priority: 1
due: "2025-01-10"
project: "webapp"
area: "Development"
tags: [backend, security]
---
```

## Requirements

- Go 1.23+
- `ripgrep` for tag and task search
