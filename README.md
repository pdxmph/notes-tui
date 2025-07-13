# notes-tui

A lightweight TUI for Markdown notes built with Go and Bubble Tea.

## Important consideration before using this code or interacting with this codebase

This application is an experiment in using Claude Code as the primary driver the development of a small, focused app that concerns itself with the owner's particular point of view on the task it is accomplishing.

As such, this is not meant to be what people think of as "an open source project," because I don't have a commitment to building a community around it and don't have the bandwidth to maintain it beyond "fix bugs I find in the process of pushing it in a direction that works for me."

It's important to understand this for a few reasons:

1. If you use this code, you'll be using something largely written by an LLM with all the things we know this entails in 2025: Potential inefficiency, security risks, and the risk of data loss.

2. If you use this code, you'll be using something that works for me the way I would like it to work. If it doesn't do what you want it to do, or if it fails in some way particular to your preferred environment, tools, or use cases, your best option is to take advantage of its very liberal license and fork it.

3. I'll make a best effort to only tag the codebase when it is in a working state with no bugs that functional testing has revealed.

While I appreciate and applaud assorted efforts to certify code and projects AI-free, I think it's also helpful to post commentary like this up front: Yes, this was largely written by an LLM so treat it accordingly. Don't think of it like code you can engage with, think of it like someone's take on how to do a task or solve a problem.

That said:

## Features

- **Fast file browsing** with fuzzy search
- **Instant note creation** with title-to-filename conversion
- **Daily notes** with special handling for `yyyy-mm-dd-daily.md` files
- **Tag search** using `#tag` syntax in content and YAML front matter
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
- **`theme`**: Color theme selection (default: "default"). Available themes:
  - `"default"` - Balanced colors for most terminals
  - `"dark"` - Optimized for dark terminals
  - `"light"` - Optimized for light terminals with dark text
  - `"high-contrast"` - Maximum contrast for accessibility
  - `"minimal"` - Monochrome with minimal color usage
- **`filtered_tags`**: Array of tags to exclude from the UI (default: []). Example: `["archived", "private", "app-data"]`

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
- **`o`**: Open sort menu
- **`O`**: Filter notes by age (e.g., last 7 days)
- **`R`**: Rename file to Denote format
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

## Features in Detail

### Search Modes

- **File search** (`/`): Fuzzy search by filename
- **Tag search** (`#`): Find files containing hashtags in content or YAML front matter

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

## Requirements

- Go 1.23+
- `ripgrep` for tag search
