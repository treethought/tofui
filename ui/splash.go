package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
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

var splashStyle = NewStyle().Align(lipgloss.Center).Margin(2, 2)

type SplashView struct {
	vp      *viewport.Model
	info    *viewport.Model
	loading *Loading
	active  bool
}

func NewSplashView() *SplashView {
	x, y := lipgloss.Size(txt)
	vp := viewport.New(x, y)
	vp.SetContent(txt)
	l := NewLoading()
	l.SetActive(true)
	info := viewport.New(20, 6)
	info.SetContent("fetching feed...")
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = NewStyle().Foreground(lipgloss.Color("205"))
	return &SplashView{vp: &vp, loading: l, info: &info, active: true}
}

func (m *SplashView) Active() bool {
	return m.active
}
func (m *SplashView) SetActive(active bool) {
	m.loading.SetActive(active)
	m.active = active
}

func (m *SplashView) SetInfo(content string) {
	m.info.SetContent(content)
}

func (m *SplashView) SetSize(w, h int) {
	x, y := splashStyle.GetFrameSize()
	m.vp.Width = w - x
	m.vp.Height = h - y - 4
	m.info.Width = w - x
	m.info.Height = h - y - 8
	m.loading.SetSize((w-x)/2, h)
}

func (m *SplashView) Init() tea.Cmd {
	return tea.Batch(m.loading.Init())
}
func (m *SplashView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.active {
		return m, nil
	}
	_, cmd := m.loading.Update(msg)
	return m, cmd
}
func (m *SplashView) View() string {
	return splashStyle.Render(
		lipgloss.JoinVertical(lipgloss.Top,
			m.vp.View(),
			lipgloss.NewStyle().MarginTop(1).Render(m.loading.View()),
			m.info.View(),
		),
	)
}
