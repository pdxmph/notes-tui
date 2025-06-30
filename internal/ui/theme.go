package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme defines the application's visual style
type Theme struct {
	// Base colors
	Primary     lipgloss.Color
	Secondary   lipgloss.Color
	Background  lipgloss.Color
	Foreground  lipgloss.Color
	
	// Semantic colors
	Success     lipgloss.Color
	Warning     lipgloss.Color
	Error       lipgloss.Color
	Info        lipgloss.Color
	
	// UI element colors
	Border      lipgloss.Color
	Cursor      lipgloss.Color
	Muted       lipgloss.Color
	Accent      lipgloss.Color
	
	// Component styles
	List        ListStyle
	Modal       ModalStyle
	Header      HeaderStyle
	Help        HelpStyle
	Popover     PopoverStyle
	Dialog      DialogStyle
	Status      StatusStyle
}

// DefaultTheme returns the default color theme
func DefaultTheme() Theme {
	// Base colors
	primary := lipgloss.Color("12")     // Blue
	secondary := lipgloss.Color("14")   // Light blue
	muted := lipgloss.Color("240")      // Gray
	accent := lipgloss.Color("214")     // Orange
	
	// Semantic colors
	success := lipgloss.Color("10")     // Green
	warning := lipgloss.Color("11")     // Yellow
	error := lipgloss.Color("9")        // Red
	info := lipgloss.Color("12")        // Blue
	
	return Theme{
		Primary:    primary,
		Secondary:  secondary,
		Background: lipgloss.Color("0"),
		Foreground: lipgloss.Color("7"),
		Success:    success,
		Warning:    warning,
		Error:      error,
		Info:       info,
		Border:     lipgloss.Color("62"),
		Cursor:     accent,
		Muted:      muted,
		Accent:     accent,
		
		// List styles
		List: ListStyle{
			Cursor:   lipgloss.NewStyle().Foreground(accent).Bold(true),
			Item:     lipgloss.NewStyle(),
			EmptyMsg: lipgloss.NewStyle().Foreground(muted).Italic(true),
		},
		
		// Modal styles
		Modal: ModalStyle{
			Title:  lipgloss.NewStyle().Bold(true).Foreground(primary),
			Prompt: lipgloss.NewStyle(),
			Help:   lipgloss.NewStyle().Foreground(muted),
			Border: lipgloss.NewStyle(),
		},
		
		// Header styles
		Header: HeaderStyle{
			Title:  lipgloss.NewStyle().Bold(true),
			Filter: lipgloss.NewStyle().Foreground(secondary),
			Sort:   lipgloss.NewStyle().Foreground(secondary),
		},
		
		// Help styles
		Help: HelpStyle{
			Key:       lipgloss.NewStyle().Foreground(accent).Bold(true),
			Desc:      lipgloss.NewStyle().Foreground(muted),
			Separator: lipgloss.NewStyle().Foreground(lipgloss.Color("236")),
		},
		
		// Popover styles
		Popover: PopoverStyle{
			Border:    lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")).Padding(1, 2),
			Title:     lipgloss.NewStyle().Bold(true).MarginBottom(1),
			ScrollBar: lipgloss.NewStyle().Foreground(muted),
			Help:      lipgloss.NewStyle().Foreground(muted).MarginTop(1),
		},
		
		// Dialog styles
		Dialog: DialogStyle{
			Title:   lipgloss.NewStyle().Bold(true).Foreground(primary),
			Message: lipgloss.NewStyle(),
			Option:  lipgloss.NewStyle().Foreground(accent),
			Border:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(error).Padding(1, 2),
		},
		
		// Status styles
		Status: StatusStyle{
			Info:    lipgloss.NewStyle().Background(info).Foreground(lipgloss.Color("0")).Padding(0, 1),
			Success: lipgloss.NewStyle().Background(success).Foreground(lipgloss.Color("0")).Padding(0, 1),
			Warning: lipgloss.NewStyle().Background(warning).Foreground(lipgloss.Color("0")).Padding(0, 1),
			Error:   lipgloss.NewStyle().Background(error).Foreground(lipgloss.Color("15")).Padding(0, 1),
		},
	}
}

// DarkTheme returns a dark color theme
func DarkTheme() Theme {
	theme := DefaultTheme()
	// Customize for dark terminals
	theme.Background = lipgloss.Color("235")
	theme.Foreground = lipgloss.Color("252")
	return theme
}

// LightTheme returns a light color theme optimized for light terminals
func LightTheme() Theme {
	// Light theme with darker text on light background
	primary := lipgloss.Color("18")     // Dark blue
	secondary := lipgloss.Color("20")   // Darker blue
	muted := lipgloss.Color("243")      // Medium gray
	accent := lipgloss.Color("166")     // Dark orange
	
	// Semantic colors (darker versions)
	success := lipgloss.Color("22")     // Dark green
	warning := lipgloss.Color("136")    // Dark yellow
	error := lipgloss.Color("124")      // Dark red
	info := lipgloss.Color("25")        // Dark blue
	
	return Theme{
		Primary:    primary,
		Secondary:  secondary,
		Background: lipgloss.Color("255"),  // Near white
		Foreground: lipgloss.Color("235"),  // Near black
		Success:    success,
		Warning:    warning,
		Error:      error,
		Info:       info,
		Border:     lipgloss.Color("247"),  // Light gray
		Cursor:     accent,
		Muted:      muted,
		Accent:     accent,
		
		// List styles
		List: ListStyle{
			Cursor:   lipgloss.NewStyle().Foreground(accent).Bold(true),
			Item:     lipgloss.NewStyle().Foreground(lipgloss.Color("235")),
			EmptyMsg: lipgloss.NewStyle().Foreground(muted).Italic(true),
		},
		
		// Modal styles
		Modal: ModalStyle{
			Title:  lipgloss.NewStyle().Bold(true).Foreground(primary),
			Prompt: lipgloss.NewStyle().Foreground(lipgloss.Color("235")),
			Help:   lipgloss.NewStyle().Foreground(muted),
			Border: lipgloss.NewStyle(),
		},
		
		// Header styles
		Header: HeaderStyle{
			Title:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("235")),
			Filter: lipgloss.NewStyle().Foreground(secondary),
			Sort:   lipgloss.NewStyle().Foreground(secondary),
		},
		
		// Help styles
		Help: HelpStyle{
			Key:       lipgloss.NewStyle().Foreground(accent).Bold(true),
			Desc:      lipgloss.NewStyle().Foreground(muted),
			Separator: lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
		},
		
		// Popover styles
		Popover: PopoverStyle{
			Border:    lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("247")).Padding(1, 2),
			Title:     lipgloss.NewStyle().Bold(true).MarginBottom(1).Foreground(primary),
			ScrollBar: lipgloss.NewStyle().Foreground(muted),
			Help:      lipgloss.NewStyle().Foreground(muted).MarginTop(1),
		},
		
		// Dialog styles
		Dialog: DialogStyle{
			Title:   lipgloss.NewStyle().Bold(true).Foreground(primary),
			Message: lipgloss.NewStyle().Foreground(lipgloss.Color("235")),
			Option:  lipgloss.NewStyle().Foreground(accent),
			Border:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(error).Padding(1, 2),
		},
		
		// Status styles
		Status: StatusStyle{
			Info:    lipgloss.NewStyle().Background(info).Foreground(lipgloss.Color("255")).Padding(0, 1),
			Success: lipgloss.NewStyle().Background(success).Foreground(lipgloss.Color("255")).Padding(0, 1),
			Warning: lipgloss.NewStyle().Background(warning).Foreground(lipgloss.Color("255")).Padding(0, 1),
			Error:   lipgloss.NewStyle().Background(error).Foreground(lipgloss.Color("255")).Padding(0, 1),
		},
	}
}

// MinimalTheme returns a minimal theme with reduced color usage
func MinimalTheme() Theme {
	// Minimal theme - mostly monochrome
	fg := lipgloss.Color("7")          // Default foreground
	bg := lipgloss.Color("0")          // Default background
	muted := lipgloss.Color("8")       // Gray
	accent := lipgloss.Color("15")     // Bright white for emphasis
	
	return Theme{
		Primary:    fg,
		Secondary:  fg,
		Background: bg,
		Foreground: fg,
		Success:    fg,
		Warning:    fg,
		Error:      fg,
		Info:       fg,
		Border:     muted,
		Cursor:     accent,
		Muted:      muted,
		Accent:     accent,
		
		// List styles - use bold/italic for emphasis
		List: ListStyle{
			Cursor:   lipgloss.NewStyle().Bold(true),
			Item:     lipgloss.NewStyle(),
			EmptyMsg: lipgloss.NewStyle().Foreground(muted).Italic(true),
		},
		
		// Modal styles
		Modal: ModalStyle{
			Title:  lipgloss.NewStyle().Bold(true),
			Prompt: lipgloss.NewStyle(),
			Help:   lipgloss.NewStyle().Foreground(muted),
			Border: lipgloss.NewStyle(),
		},
		
		// Header styles
		Header: HeaderStyle{
			Title:  lipgloss.NewStyle().Bold(true),
			Filter: lipgloss.NewStyle().Italic(true),
			Sort:   lipgloss.NewStyle().Italic(true),
		},
		
		// Help styles - minimal coloring
		Help: HelpStyle{
			Key:       lipgloss.NewStyle().Bold(true),
			Desc:      lipgloss.NewStyle().Foreground(muted),
			Separator: lipgloss.NewStyle().Foreground(muted),
		},
		
		// Popover styles
		Popover: PopoverStyle{
			Border:    lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(muted).Padding(1, 2),
			Title:     lipgloss.NewStyle().Bold(true).MarginBottom(1),
			ScrollBar: lipgloss.NewStyle().Foreground(muted),
			Help:      lipgloss.NewStyle().Foreground(muted).MarginTop(1),
		},
		
		// Dialog styles
		Dialog: DialogStyle{
			Title:   lipgloss.NewStyle().Bold(true),
			Message: lipgloss.NewStyle(),
			Option:  lipgloss.NewStyle().Bold(true),
			Border:  lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(muted).Padding(1, 2),
		},
		
		// Status styles - use reverse video for visibility
		Status: StatusStyle{
			Info:    lipgloss.NewStyle().Reverse(true).Padding(0, 1),
			Success: lipgloss.NewStyle().Reverse(true).Padding(0, 1),
			Warning: lipgloss.NewStyle().Reverse(true).Padding(0, 1),
			Error:   lipgloss.NewStyle().Reverse(true).Bold(true).Padding(0, 1),
		},
	}
}

// GetTheme returns a theme by name
func GetTheme(name string) Theme {
	switch name {
	case "dark":
		return DarkTheme()
	case "light":
		return LightTheme()
	case "high-contrast":
		return HighContrastTheme()
	case "minimal":
		return MinimalTheme()
	default:
		return DefaultTheme()
	}
}

// HighContrastTheme returns a high contrast theme
func HighContrastTheme() Theme {
	return Theme{
		Primary:    lipgloss.Color("15"),  // White
		Secondary:  lipgloss.Color("11"),  // Yellow
		Background: lipgloss.Color("0"),   // Black
		Foreground: lipgloss.Color("15"),  // White
		Success:    lipgloss.Color("10"),  // Green
		Warning:    lipgloss.Color("11"),  // Yellow
		Error:      lipgloss.Color("9"),   // Red
		Info:       lipgloss.Color("12"),  // Blue
		Border:     lipgloss.Color("15"),  // White
		Cursor:     lipgloss.Color("11"),  // Yellow
		Muted:      lipgloss.Color("8"),   // Gray
		Accent:     lipgloss.Color("11"),  // Yellow
		
		// High contrast component styles
		List: ListStyle{
			Cursor:   lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true),
			Item:     lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
			EmptyMsg: lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		},
		
		Modal: ModalStyle{
			Title:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")),
			Prompt: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")),
			Help:   lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
			Border: lipgloss.NewStyle(),
		},
		
		Header: HeaderStyle{
			Title:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")),
			Filter: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")),
			Sort:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")),
		},
		
		Help: HelpStyle{
			Key:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")),
			Desc:      lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
			Separator: lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		},
		
		Popover: PopoverStyle{
			Border:    lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("15")).Padding(1, 2),
			Title:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")),
			ScrollBar: lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
			Help:      lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		},
		
		Dialog: DialogStyle{
			Title:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")),
			Message: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")),
			Option:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")),
			Border:  lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("15")).Padding(1, 2),
		},
		
		Status: StatusStyle{
			Info:    lipgloss.NewStyle().Background(lipgloss.Color("12")).Foreground(lipgloss.Color("0")).Padding(0, 1),
			Success: lipgloss.NewStyle().Background(lipgloss.Color("10")).Foreground(lipgloss.Color("0")).Padding(0, 1),
			Warning: lipgloss.NewStyle().Background(lipgloss.Color("11")).Foreground(lipgloss.Color("0")).Padding(0, 1),
			Error:   lipgloss.NewStyle().Background(lipgloss.Color("9")).Foreground(lipgloss.Color("15")).Padding(0, 1),
		},
	}
}