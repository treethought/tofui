package ui

import (
	"fmt"

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
	app     *App
	vp      *viewport.Model
	info    *viewport.Model
	loading *Loading
	active  bool
	signin  bool
}

func NewSplashView(app *App) *SplashView {
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
	return &SplashView{
		vp: &vp, loading: l,
		info: &info, active: true,
		app: app,
	}
}

func (m *SplashView) Active() bool {
	return m.active
}
func (m *SplashView) SetActive(active bool) {
	m.loading.SetActive(active)
	m.active = active
}
func (m *SplashView) ShowSignin(v bool) {
	m.loading.SetActive(!v)
	m.signin = v
	if v {
		m.info.SetContent("Press Enter to sign in")
	}
}
func (m *SplashView) SetInfo(content string) {
	if m.signin {
		return
	}
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
	if m.signin {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				portPart := fmt.Sprintf(":%d", m.app.cfg.Server.HTTPPort)
				if portPart == ":443" {
					portPart = ""
				}
				u := fmt.Sprintf("%s/signin?pk=%s", m.app.cfg.BaseURL(), m.app.ctx.pk)
				m.info.SetContent(fmt.Sprintf("Please sign in at %s", u))
				return m, OpenURL(u)
			}
		}
	}

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
