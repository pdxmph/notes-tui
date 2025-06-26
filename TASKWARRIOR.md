# TaskWarrior Integration for notes-tui

This integration enables bidirectional linking between your notes and TaskWarrior tasks.

## Setup

1. **Configure TaskWarrior UDA** (one-time setup):
```bash
task config uda.notesid.type string
task config uda.notesid.label "Notes ID"
```

2. **Install the wrapper script**:
```bash
# Simple version
cp task2notes ~/bin/  # or anywhere in your PATH
chmod +x ~/bin/task2notes

# Or use the enhanced version with more features
cp task2notes-enhanced ~/bin/task2notes
chmod +x ~/bin/task2notes
```

## Usage

### Creating tasks from notes (Note → Task)
1. Navigate to a note with a Denote identifier in notes-tui
2. Press `Ctrl+K`
3. Enter task description
4. Task is created with `notesid` UDA linking back to the note

### Opening notes from tasks (Task → Note)
```bash
# Basic usage
task2notes 42  # Opens the note linked to task 42

# Enhanced version features
task2notes list              # List all tasks with linked notes
task2notes find 20241225T093015  # Find all tasks for a specific note
```

## Example Workflow

1. **Taking meeting notes**:
   - Create a note: `20241225T093015-meeting-with-team.md`
   - During the meeting, press `Ctrl+K` to create action items
   - Each task automatically links back to the meeting notes

2. **Reviewing tasks**:
   - Run `task list` to see your tasks
   - Use `task2notes <id>` to jump to the context/notes for any task
   - The full context is always one command away

## Tips

- Use `task notesid.any: list` to see all tasks with linked notes
- Add `notesid` to your task reports to see the link:
  ```bash
  task config report.next.columns id,project,description,notesid
  ```
- Create a shell alias for quick access: `alias t2n=task2notes`

## Requirements

- TaskWarrior with the `notesid` UDA configured
- notes-tui with Denote filename support enabled
- Notes using Denote-style identifiers (YYYYMMDDTHHMMSS)
