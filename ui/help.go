package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type keymap interface {
	ShortHelp() []key.Binding
	FullHelp() [][]key.Binding
}

type HelpView struct {
	app  *App
	h    help.Model
	vp   viewport.Model
	full bool
	km   keymap
}

func NewHelpView(app *App, km keymap) *HelpView {
	return &HelpView{
		app: app,
		h:   help.New(),
		vp:  viewport.Model{},
		km:  km,
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
		m.vp.SetContent(m.h.FullHelpView(m.km.FullHelp()))
		return
	}
	hv := m.km.ShortHelp()
	m.vp.SetContent(m.h.ShortHelpView(hv))
}

func (m *HelpView) Init() tea.Cmd {
	m.SetFull(false)
	return nil
}

func (m *HelpView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	vp, cmd := m.vp.Update(msg)
	m.vp = vp
	return m, cmd
}

func (m *HelpView) ShortView() string {
	hv := m.km.ShortHelp()
	return m.h.ShortHelpView(hv)
}

func (m *HelpView) View() string {
	return m.vp.View()
}
