package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type HelpView struct {
	h    help.Model
	vp   viewport.Model
	full bool
}

func NewHelpView() *HelpView {
	return &HelpView{
		h:  help.New(),
		vp: viewport.Model{},
	}
}

func (m *HelpView) SetSize(w, h int) {
	m.vp.Width = w
	m.vp.Height = h
}

func (m *HelpView) IsFull() bool {
	return m.full
}

func (m *HelpView) SetFull(full bool) {
	m.full = full
	if m.full {
		m.vp.SetContent(m.h.FullHelpView(GlobalKeyMap.FullHelp()))
		return
	}
	m.vp.SetContent(m.h.ShortHelpView(GlobalKeyMap.ShortHelp()))
}

func (m *HelpView) Init() tea.Cmd {
	m.vp.SetContent(m.h.ShortHelpView(GlobalKeyMap.ShortHelp()))
	return nil
}

func (m *HelpView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	vp, cmd := m.vp.Update(msg)
	m.vp = vp
	return m, cmd
}

func (m *HelpView) ShortView() string {
	return m.h.ShortHelpView(GlobalKeyMap.ShortHelp())
}

func (m *HelpView) View() string {
	return m.vp.View()
}
