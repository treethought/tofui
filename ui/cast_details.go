package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/castr/api"
)

type CastView struct {
	cast *api.Cast
	img  *ImageModel
	pfp  *ImageModel
}

func NewCastView(cast *api.Cast) *CastView {
	c := &CastView{
		cast: cast,
		pfp:  NewImage(false, true, special),
		img:  NewImage(true, true, special),
	}
	return c
}

func (m *CastView) SetCast(cast *api.Cast) tea.Cmd {
	m.cast = cast
	return m.Init()
}

func (m *CastView) Init() tea.Cmd {
	if m.cast == nil {
		return nil
	}
	cmds := []tea.Cmd{
		m.pfp.SetURL(m.cast.Author.PfpURL, false),
		m.pfp.SetSize(4, 4),
	}
	if len(m.cast.Embeds) > 0 {
		cmds = append(cmds, m.img.SetURL(m.cast.Embeds[0].URL, true))
	}
	return tea.Batch(cmds...)
}

func (m *CastView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmds := []tea.Cmd{}
		if m.img != nil {
			cmds = append(cmds, m.img.SetSize(msg.Width/2, msg.Height/2))
		}
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		if msg.String() == "o" {
			return m, OpenURL(fmt.Sprintf("https://warpcast.com/%s/%s", m.cast.Author.Username, m.cast.Hash))
		}
	}
	cmds := []tea.Cmd{}
	img, cmd := m.img.Update(msg)
	m.img = img
	cmds = append(cmds, cmd)

	pfp, cmd := m.pfp.Update(msg)
	m.pfp = pfp
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *CastView) View() string {
	return lipgloss.JoinVertical(lipgloss.Top,
		UsernameHeader(&m.cast.Author, m.pfp),
		CastContent(m.cast, 10),
		m.img.View(),
	)
}
