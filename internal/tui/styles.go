package tui

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	// Primary colors
	colorPrimary   = lipgloss.Color("#00D4AA") // InvITSM teal
	colorSecondary = lipgloss.Color("#6B7280") // Gray
	colorAccent    = lipgloss.Color("#F59E0B") // Amber
	colorSuccess   = lipgloss.Color("#10B981") // Green
	colorError     = lipgloss.Color("#EF4444") // Red
	colorInfo      = lipgloss.Color("#3B82F6") // Blue

	// Background colors
	colorBg      = lipgloss.Color("#1F2937") // Dark gray
	colorSurface = lipgloss.Color("#374151") // Lighter gray
	colorBorder  = lipgloss.Color("#4B5563") // Border gray

	// Text colors
	colorText      = lipgloss.Color("#F9FAFB") // White
	colorTextMuted = lipgloss.Color("#9CA3AF") // Light gray
	colorTextDark  = lipgloss.Color("#1F2937") // Dark text
)

// Common styles
var (
	// Title style
	titleStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			MarginBottom(1)

	// Subtitle style
	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			MarginBottom(2)

	// Box style for sections
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2).
			MarginBottom(1)

	// Input field style
	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.Border{
			Top:         "─",
			Bottom:      "─",
			Left:        "│",
			Right:       "│",
			TopLeft:     "╭",
			TopRight:    "╮",
			BottomLeft:  "╰",
			BottomRight: "╯",
		}).
		BorderForeground(colorBorder).
		Padding(0, 1).
		Width(50)

	// Input focused style
	inputFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary).
				Padding(0, 1).
				Width(50)

	// Label style
	labelStyle = lipgloss.NewStyle().
			Foreground(colorTextMuted).
			MarginRight(1).
			Width(20).
			Align(lipgloss.Right)

	// Help text style
	helpStyle = lipgloss.NewStyle().
			Foreground(colorTextMuted).
			Italic(true).
			MarginTop(1)

	// Error style
	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true).
			MarginTop(1)

	// Success style
	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true).
			MarginTop(1)

	// Button style
	buttonStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Background(colorPrimary).
			Bold(true).
			Padding(0, 2).
			MarginTop(2)

	// Button disabled style
	buttonDisabledStyle = lipgloss.NewStyle().
				Foreground(colorTextMuted).
				Background(colorBorder).
				Padding(0, 2).
				MarginTop(2)

	// Progress bar style
	progressStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Width(50)

	// Footer style
	footerStyle = lipgloss.NewStyle().
			Foreground(colorTextMuted).
			Align(lipgloss.Center).
			MarginTop(2)
)

// Helper functions

// RenderLabel renders a label with consistent styling
func RenderLabel(text string) string {
	return labelStyle.Render(text + ":")
}

// RenderInput renders an input field
func RenderInput(value string, focused bool) string {
	if focused {
		return inputFocusedStyle.Render(value)
	}
	return inputStyle.Render(value)
}

// RenderBox renders a box with content
func RenderBox(content string) string {
	return boxStyle.Render(content)
}

// RenderHelp renders help text
func RenderHelp(text string) string {
	return helpStyle.Render(text)
}

// RenderError renders an error message
func RenderError(text string) string {
	return errorStyle.Render("⚠️  " + text)
}

// RenderSuccess renders a success message
func RenderSuccess(text string) string {
	return successStyle.Render("✅ " + text)
}
