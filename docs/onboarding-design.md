# Onboarding Experience Design for notes-tui

## Overview
A guided first-run experience that helps users configure notes-tui based on their primary use case and experience level.

## Implementation Concept

### 1. First Launch Detection
```go
// Check for ~/.config/notes-tui/config.toml
// If not exists, launch onboarding
if !configExists() {
    return NewOnboardingModel()
}
```

### 2. Onboarding Flow

#### Screen 1: Welcome
```
┌─────────────────────────────────────────┐
│       Welcome to notes-tui! 🎉         │
│                                         │
│  A fast, keyboard-driven note manager   │
│  designed for terminal enthusiasts.     │
│                                         │
│  This quick setup will help configure   │
│  notes-tui for your workflow.          │
│                                         │
│  [Enter] Continue  [q] Skip setup      │
└─────────────────────────────────────────┘
```

#### Screen 2: Persona Selection
```
┌─────────────────────────────────────────┐
│    How do you primarily take notes?     │
│                                         │
│  > Power User (Terminal workflow)       │
│    Zettelkasten (Knowledge base)        │
│    Task Management (GTD/Projects)       │
│    Daily Journal (Personal logs)        │
│    Custom (Configure everything)        │
│                                         │
│  [↑↓] Navigate  [Enter] Select         │
└─────────────────────────────────────────┘
```

#### Screen 3: Directory Setup
```
┌─────────────────────────────────────────┐
│      Where should notes be stored?      │
│                                         │
│  Directory: ~/notes_                    │
│                                         │
│  Common locations:                      │
│  [1] ~/notes                           │
│  [2] ~/Documents/notes                 │
│  [3] ~/Dropbox/notes                   │
│  [4] Custom path...                    │
│                                         │
│  [Enter] Confirm  [Esc] Back           │
└─────────────────────────────────────────┘
```

#### Screen 4: Editor Configuration
```
┌─────────────────────────────────────────┐
│    Select your preferred editor:        │
│                                         │
│  > Neovim (nvim)                       │
│    Vim (vim)                           │
│    VS Code (code --wait)               │
│    Emacs (emacs)                       │
│    Nano (nano)                         │
│    Other...                            │
│                                         │
│  [↑↓] Navigate  [Enter] Select         │
└─────────────────────────────────────────┘
```

#### Screen 5: Theme Selection
```
┌─────────────────────────────────────────┐
│        Choose a color theme:            │
│                                         │
│  > Default (Balanced colors)            │
│    Dark (Night mode)                   │
│    Light (Day mode)                    │
│    High Contrast (Accessibility)       │
│    Minimal (Reduced colors)            │
│                                         │
│  Preview: [Sample text with theme]     │
│                                         │
│  [↑↓] Navigate  [Enter] Select         │
└─────────────────────────────────────────┘
```

#### Screen 6: Feature Toggle
```
┌─────────────────────────────────────────┐
│      Enable optional features:          │
│                                         │
│  [x] Denote-style filenames            │
│  [x] YAML frontmatter                  │
│  [ ] TaskWarrior integration           │
│  [x] Show note titles                  │
│  [ ] Prompt for tags                   │
│                                         │
│  [Space] Toggle  [Enter] Continue      │
└─────────────────────────────────────────┘
```

#### Screen 7: Interactive Tutorial
```
┌─────────────────────────────────────────┐
│         Quick Interactive Tour          │
│                                         │
│  Let's try the basics:                 │
│                                         │
│  1. Press 'n' to create a note         │
│                                         │
│  [Waiting for user to press 'n'...]    │
│                                         │
│  Great! Now type a title and press     │
│  Enter to create your first note.      │
│                                         │
│  [Skip] Skip tutorial                  │
└─────────────────────────────────────────┘
```

### 3. Progressive Disclosure

#### Beginner Mode (First Week)
- Show extended help bar with descriptions
- Display tooltips for actions
- Confirmation prompts for destructive actions

#### Intermediate Mode (After 50 operations)
- Condensed help bar
- Faster animations
- Skip some confirmations

#### Expert Mode (After 200 operations)
- Minimal UI
- No confirmations except delete
- Hidden help (toggle with '?')

### 4. Configuration Templates

Based on persona selection, apply appropriate template:

```go
func applyPersonaTemplate(persona string) Config {
    switch persona {
    case "power-user":
        return loadTemplate("power-user.toml")
    case "zettelkasten":
        return loadTemplate("zettelkasten.toml")
    case "task-focused":
        return loadTemplate("task-focused.toml")
    case "journaler":
        return loadTemplate("daily-journaler.toml")
    default:
        return defaultConfig()
    }
}
```

### 5. Help System Evolution

#### Dynamic Help Bar
```go
type HelpLevel int

const (
    HelpBeginner HelpLevel = iota
    HelpIntermediate  
    HelpExpert
)

func (m *Model) getHelpLevel() HelpLevel {
    operationCount := m.stats.totalOperations
    
    if operationCount < 50 {
        return HelpBeginner
    } else if operationCount < 200 {
        return HelpIntermediate
    }
    return HelpExpert
}
```

#### Contextual Tips
```
// Show tips based on user behavior
if m.searchCount == 0 && m.fileCount > 20 {
    showTip("Try '/' to search your notes")
}

if m.createCount > 10 && !m.usedTags {
    showTip("Use tags to organize notes: add #topic")
}
```

### 6. Onboarding Completion

#### Success Screen
```
┌─────────────────────────────────────────┐
│         Setup Complete! ✓               │
│                                         │
│  Your configuration has been saved to:  │
│  ~/.config/notes-tui/config.toml       │
│                                         │
│  Quick reference:                       │
│  • Press 'n' to create a note          │
│  • Press '/' to search                 │
│  • Press 'q' to quit                   │
│  • Press '?' for help anytime          │
│                                         │
│  [Enter] Start using notes-tui         │
└─────────────────────────────────────────┘
```

### 7. Re-engagement Features

#### Welcome Back Messages
```go
func getWelcomeMessage(lastUsed time.Time) string {
    daysSince := time.Since(lastUsed).Hours() / 24
    
    if daysSince > 30 {
        return "Welcome back! Press '?' to see what's new"
    } else if daysSince > 7 {
        return fmt.Sprintf("You have %d notes. Press 'D' to see recent daily notes", noteCount)
    }
    return ""
}
```

#### Feature Discovery
```go
// Track unused features and suggest them
unusedFeatures := m.getUnusedFeatures()
if len(unusedFeatures) > 0 && m.operationCount % 50 == 0 {
    feature := unusedFeatures[rand.Intn(len(unusedFeatures))]
    showTip(feature.description)
}
```

### 8. Analytics for Improvement

Track anonymous usage patterns:
- Most used features
- Common workflows  
- Error patterns
- Feature discovery rate

## Benefits

1. **Lower barrier to entry** - Guided setup reduces intimidation
2. **Persona-optimized defaults** - Users start with relevant config
3. **Progressive learning** - Features revealed as needed
4. **Higher retention** - Users understand value immediately
5. **Reduced support burden** - Self-service onboarding

## Implementation Priority

1. **Phase 1**: Basic wizard (directory, editor, theme)
2. **Phase 2**: Persona templates and interactive tutorial  
3. **Phase 3**: Progressive disclosure and tips
4. **Phase 4**: Analytics and re-engagement

## Success Metrics

- Time to first note creation: < 2 minutes
- Onboarding completion rate: > 80%
- Feature adoption rate: > 60% using 3+ features
- User retention: > 70% active after 1 week