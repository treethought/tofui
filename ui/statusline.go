package ui

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakenelf/teacup/statusbar"
)

type StatusLine struct {
	app  *App
	sb   statusbar.Model
	help help.Model
}

func NewStatusLine(app *App) *StatusLine {
	sb := statusbar.New(
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Dark: "#ffffff", Light: "#ffffff"},
			Background: lipgloss.AdaptiveColor{Light: "#F25D94", Dark: "#F25D94"},
		},
		statusbar.ColorConfig{
			// Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
			// Background: lipgloss.AdaptiveColor{Light: "#3c3836", Dark: "#3c3836"},
		},
		statusbar.ColorConfig{
			// Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
			// Background: lipgloss.AdaptiveColor{Light: "#A550DF", Dark: "#A550DF"},
		},
		statusbar.ColorConfig{
			// Foreground: lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"},
			// Background: lipgloss.AdaptiveColor{Light: "#6124DF", Dark: "#6124DF"},
		},
	)
	return &StatusLine{
		sb:   sb,
		app:  app,
		help: help.New(),
	}
}

func (m *StatusLine) SetSize(width, height int) {
	m.sb.SetSize(width)
	m.sb.Height = 1
}

func (m *StatusLine) Init() tea.Cmd {
	return nil
}

func (m *StatusLine) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.sb.SetContent(m.app.focused, m.help.ShortHelpView(DefaultKeyMap.ShortHelp()), "", "")
	_, cmd := m.sb.Update(msg)
	return m, cmd
}

func (m *StatusLine) View() string {
	return m.sb.View()
}
