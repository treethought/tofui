package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/castr/api"
)

type CastView struct {
	cast           *api.Cast
	img            *ImageModel
	pfp            *ImageModel
	replies        *RepliesView
	vp             viewport.Model
}

func NewCastView(cast *api.Cast) *CastView {
	c := &CastView{
		cast:    cast,
		pfp:     NewImage(false, true, special),
		img:     NewImage(true, true, special),
		replies: NewRepliesView(),
		vp:      viewport.New(0, 0),
	}
	return c
}

func (m *CastView) SetCast(cast *api.Cast) tea.Cmd {
	m.cast = cast
	m.img = NewImage(true, true, special)
	m.pfp = NewImage(true, true, special)
	m.vp.SetContent(CastContent(m.cast, 10))
	return m.Init()
}

func (m *CastView) Init() tea.Cmd {
	if m.cast == nil {
		return nil
	}
	cmds := []tea.Cmd{
		m.pfp.SetURL(m.cast.Author.PfpURL, false),
		m.pfp.SetSize(4, 4),
		m.replies.SetOpHash(m.cast.Hash),
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
		v, cmd := m.vp.Update(msg)
		m.vp = v
		cmds = append(cmds, cmd)
		if m.img != nil {
			cmds = append(cmds, m.img.SetSize(msg.Width/2, msg.Height/2))
		}
		m.replies.SetSize(msg.Width, msg.Height/2)
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		if msg.String() == "o" {
			return m, OpenURL(fmt.Sprintf("https://warpcast.com/%s/%s", m.cast.Author.Username, m.cast.Hash))
		}
		if msg.String() == "l" {
			return m, likeCastCmd(m.cast)
		}
	}
	cmds := []tea.Cmd{}
	v, vcmd := m.vp.Update(msg)
	m.vp = v
	cmds = append(cmds, vcmd)

	img, icmd := m.img.Update(msg)
	m.img = img
	cmds = append(cmds, icmd)

	pfp, pcmd := m.pfp.Update(msg)
	m.pfp = pfp
	cmds = append(cmds, pcmd)

	_, rcmd := m.replies.Update(msg)
	cmds = append(cmds, rcmd)

	return m, tea.Batch(cmds...)
}

func (m *CastView) View() string {
	return lipgloss.JoinVertical(lipgloss.Top,
		UsernameHeader(&m.cast.Author, m.pfp),
		lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).Padding(0).Render(CastStats(m.cast, 10)),
		CastContent(m.cast, 10),
		m.img.View(),
		m.replies.View(),
	)
}
