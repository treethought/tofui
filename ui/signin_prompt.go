package ui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type SigninPrompt struct {
	vp     *viewport.Model
	active bool
}

func NewSigninPrompt() *SigninPrompt {
	vp := viewport.New(0, 0)
	return &SigninPrompt{vp: &vp}
}
func (m *SigninPrompt) Active() bool {
	return m.active
}
func (m *SigninPrompt) SetActive(active bool) {
	m.active = active
}
func (m *SigninPrompt) SetContent(content string) {
	m.vp.SetContent(content)
}

func (m *SigninPrompt) SetSize(w, h int) {
	m.vp.Width = w
	m.vp.Height = h
}
func (m *SigninPrompt) Init() tea.Cmd {
	return nil
}
func (m *SigninPrompt) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	vp, cmd := m.vp.Update(msg)
	m.vp = &vp
	return m, cmd
}
func (m *SigninPrompt) View() string {
	return m.vp.View()
}
