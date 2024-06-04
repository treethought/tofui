package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

type loadTickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return loadTickMsg(t)
	})
}

type Loading struct {
	prog   progress.Model
	active bool
	pct    float64
}

func NewLoading() *Loading {
	p := progress.New()
	p.ShowPercentage = false
	return &Loading{
		active: true,
		prog:   p,
	}
}

func (m *Loading) IsActive() bool {
	return m.active
}

func (m *Loading) SetActive(v bool) {
	m.active = v
	if !m.active {
		m.pct = 0
	}
	return
}

func (m *Loading) Init() tea.Cmd {
	if m.active {
		return tickCmd()
	}
	return nil
}

func (m *Loading) SetSize(w, h int) {
	m.prog.Width = w
}

func (m *Loading) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.prog.Width = msg.Width - 2
		return m, nil
	case loadTickMsg:
		if !m.active {
			return m, nil
		}
		m.pct = m.pct + 0.25
		if m.pct > 1 {
			m.pct = 0
		}
		return m, tickCmd()
	}

	return m, nil
}

func (m *Loading) View() string {
	if !m.active {
		return ""
	}
	return m.prog.ViewAs(m.pct)
}
