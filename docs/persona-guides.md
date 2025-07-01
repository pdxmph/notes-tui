# Persona Quick-Start Guides for notes-tui

## 🧑‍💻 Terminal Power User Guide

### Your Profile
You live in the terminal and want lightning-fast note management without breaking your flow.

### Quick Setup
```bash
# Add to your shell config
export EDITOR="nvim"
alias nt="notes-tui"
alias nd="notes-tui --daily"  # Quick daily note
```

### Essential Keybindings
- `/` - Fuzzy search (uses ripgrep under the hood)
- `e` - Open in $EDITOR
- `n` → type → `Enter` - Create note in < 2 seconds
- `gg`/`G` - Vim navigation
- `q` - Quick exit

### Power User Config
```toml
# ~/.config/notes-tui/config.toml
notes_directory = "~/brain"
editor = "nvim"
denote_filenames = true
show_titles = false  # Show raw filenames
prompt_for_tags = false  # Skip tag prompt for speed
theme = "minimal"  # Reduce visual noise
```

### Workflow Tips
1. Use `d` for quick daily notes during meetings
2. Combine with `tmux` for split-pane workflows
3. Pipe to external tools: `notes-tui | grep pattern`
4. Use `R` to batch-rename to Denote format

---

## 📝 Zettelkasten Practitioner Guide

### Your Profile
You maintain an interconnected knowledge base using the Denote methodology.

### Quick Setup
```bash
# Create your Zettelkasten directory
mkdir -p ~/zettelkasten/{permanent,literature,fleeting}
```

### Essential Keybindings
- `n` → title → tags - Create atomic notes
- `#` - Search by tag across all notes
- `o` → `i` - Sort by Denote ID (chronological)
- `R` - Convert existing notes to Denote format

### Zettelkasten Config
```toml
# ~/.config/notes-tui/config.toml
notes_directory = "~/zettelkasten"
denote_filenames = true
add_frontmatter = true
prompt_for_tags = true
show_titles = true
theme = "light"  # Better for long reading sessions

[frontmatter]
template = """
---
title: {{title}}
date: {{date}}
tags: {{tags}}
type: permanent
---
"""
```

### Workflow Tips
1. Use tags like `#concept`, `#person`, `#project`
2. Create daily fleeting notes with `d`
3. Review and refactor into permanent notes weekly
4. Use `#` to find related notes by tag

---

## ✅ Task-Oriented User Guide

### Your Profile
You use notes to track projects and tasks, integrating with TaskWarrior.

### Quick Setup
```bash
# Install TaskWarrior
brew install task
# Enable integration
echo "taskwarrior_support = true" >> ~/.config/notes-tui/config.toml
```

### Essential Keybindings
- `t` - Show all notes with open tasks
- `Ctrl+K` - Create TaskWarrior task from note
- `d` - Daily note for today's tasks
- `D` - Review all daily notes

### Task-Focused Config
```toml
# ~/.config/notes-tui/config.toml
notes_directory = "~/work/notes"
taskwarrior_support = true
denote_filenames = true
add_frontmatter = true
theme = "high-contrast"  # Clear task visibility

[daily_template]
content = """
# {{date}}

## Tasks
- [ ] 

## Notes

## Tomorrow
"""
```

### Workflow Tips
1. Start each day with `d` to create daily note
2. Use `- [ ]` for tasks in markdown
3. Press `t` to review all open tasks
4. `Ctrl+K` links tasks to their context

---

## 📅 Daily Journaler Guide

### Your Profile
You maintain daily logs for work or personal reflection.

### Quick Setup
```bash
# Create journal structure
mkdir -p ~/journal/{daily,weekly,monthly}
```

### Essential Keybindings
- `d` - Today's entry (creates if missing)
- `D` - Browse all daily notes
- `o` → `d` - Sort by date
- `O` → `7` - Last week's entries

### Journal Config
```toml
# ~/.config/notes-tui/config.toml
notes_directory = "~/journal"
denote_filenames = true
add_frontmatter = true
show_titles = true
theme = "dark"  # Easy on eyes for evening journaling

[journal]
daily_template = """
# {{date}} - {{weekday}}

## Gratitude
- 

## Accomplished
- 

## Learned
- 

## Tomorrow
- 
"""
```

### Workflow Tips
1. Set a daily reminder to journal
2. Use `D` for weekly reviews
3. Tag entries with moods: `#happy`, `#stressed`
4. Search memories with `/` or by tag with `#`

---

## 🎨 Theme Customizer Guide

### Your Profile
You have specific visual needs or work in varying lighting conditions.

### Available Themes
- **default** - Balanced for most terminals
- **dark** - Night coding sessions
- **light** - Bright environments
- **high-contrast** - Accessibility needs
- **minimal** - Distraction-free

### Theme Testing
```bash
# Try each theme
for theme in default dark light high-contrast minimal; do
  echo "theme = \"$theme\"" > ~/.config/notes-tui/config.toml
  notes-tui
done
```

### Custom Theme Config
```toml
# ~/.config/notes-tui/config.toml
theme = "high-contrast"  # or your preference

# Future: Custom theme support
[custom_theme]
background = "#1a1b26"
foreground = "#a9b1d6"
cursor = "#7aa2f7"
accent = "#9ece6a"
```

### Tips for Each Theme
- **dark**: Best with terminal bg #000000-#1a1a1a
- **light**: Works well with bg #ffffff-#f5f5f5
- **high-contrast**: WCAG AAA compliant
- **minimal**: Focuses attention on content

---

## Getting Started Checklist

### First-Time Setup
1. [ ] Install notes-tui
2. [ ] Create config directory: `mkdir -p ~/.config/notes-tui`
3. [ ] Copy relevant persona config above
4. [ ] Set your notes directory
5. [ ] Choose your theme
6. [ ] Try basic operations: `n`, `d`, `/`, `#`

### Daily Workflow
1. [ ] Open with `notes-tui` or alias
2. [ ] Create/open daily note with `d`
3. [ ] Quick capture with `n`
4. [ ] Review and organize
5. [ ] Search as needed

### Weekly Maintenance
1. [ ] Review uncompleted tasks with `t`
2. [ ] Organize new notes with tags
3. [ ] Archive old daily notes
4. [ ] Update your workflow as needed