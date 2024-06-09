package ui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var txt = `
████████╗ ██████╗ ███████╗██╗   ██╗██╗
╚══██╔══╝██╔═══██╗██╔════╝██║   ██║██║
   ██║   ██║   ██║█████╗  ██║   ██║██║
   ██║   ██║   ██║██╔══╝  ██║   ██║██║
   ██║   ╚██████╔╝██║     ╚██████╔╝██║
   ╚═╝    ╚═════╝ ╚═╝      ╚═════╝ ╚═╝

  Terminally On Farcaster User Interface
`

type SplashView struct {
	vp     *viewport.Model
	active bool
}

func NewSplashView() *SplashView {
	x, y := lipgloss.Size(txt)
	vp := viewport.New(x, y)
	vp.SetContent(txt)
	return &SplashView{vp: &vp}
}

func (m *SplashView) Active() bool {
	return m.active
}
func (m *SplashView) SetActive(active bool) {
	m.active = active
}

func (m *SplashView) SetSize(w, h int) {
	m.vp.Width = w
	m.vp.Height = h
}

func (m *SplashView) Init() tea.Cmd {
	return nil
}
func (m *SplashView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	vp, cmd := m.vp.Update(msg)
	m.vp = &vp
	return m, cmd
}
func (m *SplashView) View() string {
	return m.vp.View()
}
