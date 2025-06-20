---
title: Contacts TUI - Hotkey Display Styling Options
type: note
permalink: basic-memory/contacts-tui-hotkey-display-styling-options
tags:
- '["contacts-tui"'
- '"ui-enhancement"'
- '"styling"]'
---

# Contacts TUI - Hotkey Display Styling Options

## Current Implementation
Currently using bracket notation for hotkeys: `[p]ing`, `[i]nvite`, etc.

## Alternative Styling Options

Instead of brackets, we can use Lipgloss styling for cleaner visual presentation:

### Bold
```go
display += lipgloss.NewStyle().Bold(true).Render(string(char))
```

### Underline
```go
display += lipgloss.NewStyle().Underline(true).Render(string(char))
```

### Bold + Underline
```go
display += lipgloss.NewStyle().Bold(true).Underline(true).Render(string(char))
```

### Color Highlighting
```go
display += lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true).Render(string(char))
```

## Implementation Notes
- Would need to update all hotkey display locations in `app.go`
- Affects state selection menu, relationship type filter menu
- Consider consistency across all menus
- Test terminal compatibility (some terminals handle underline differently)

## Related Issue
This enhancement came up during implementation of issue #19 (hotkey selection for menus)