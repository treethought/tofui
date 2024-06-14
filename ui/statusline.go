package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakenelf/teacup/statusbar"
)

var statusStyle = NewStyle().BorderTop(true).BorderStyle(lipgloss.RoundedBorder())

type StatusLine struct {
	app  *App
	sb   statusbar.Model
	help *HelpView
	full bool
}

func NewStatusLine(app *App) *StatusLine {
	sb := statusbar.New(
		statusbar.ColorConfig{
			Foreground: lipgloss.AdaptiveColor{Dark: "#ffffff", Light: "#ffffff"},
			Background: lipgloss.AdaptiveColor{Light: "#F25D94", Dark: "#483285"},
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
		help: NewHelpView(app),
		full: false,
	}
}

func (m *StatusLine) SetSize(width, height int) {
	fx, _ := statusStyle.GetFrameSize()
	m.sb.SetSize(width - fx)
	m.sb.Height = 1 
}

func (m *StatusLine) Init() tea.Cmd {
	return nil
}

func (m *StatusLine) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.sb.SetContent(m.app.navname, "", "", m.help.ShortView())
	_, cmd := m.sb.Update(msg)
	return m, cmd
}

func (m *StatusLine) View() string {
	return statusStyle.Render(m.sb.View())
}
