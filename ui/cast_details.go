package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/castr/api"
)

type CastView struct {
	cast    *api.Cast
	img     *ImageModel
	pfp     *ImageModel
	replies *RepliesView
	vp      viewport.Model

	pubReply *PublishInput
}

func NewCastView(cast *api.Cast) *CastView {
	c := &CastView{
		cast:     cast,
		pfp:      NewImage(false, true, special),
		img:      NewImage(true, true, special),
		replies:  NewRepliesView(),
		vp:       viewport.New(0, 0),
		pubReply: NewPublishInput(nil),
	}
	return c
}

func (m *CastView) SetCast(cast *api.Cast) tea.Cmd {
	m.img.Clear()
	m.pfp.Clear()
	m.cast = cast
	m.pubReply.SetContext(cast.Hash, cast.ParentURL, cast.Author.FID)
	return m.Init()
}

func (m *CastView) Init() tea.Cmd {
	if m.cast == nil {
		return nil
	}
	m.vp.SetContent(CastContent(m.cast, 10))

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
			cmds = append(cmds, m.img.SetSize(msg.Width/4, msg.Height/4))
		}
		m.replies.SetSize(msg.Width, msg.Height/2)
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		if m.pubReply.Active() {
			_, cmd := m.pubReply.Update(msg)
			return m, cmd
		}
		if msg.String() == "o" {
			return m, OpenURL(fmt.Sprintf("https://warpcast.com/%s/%s", m.cast.Author.Username, m.cast.Hash))
		}
		if msg.String() == "l" {
			return m, likeCastCmd(m.cast)
		}
		if msg.String() == "C" {
			m.pubReply.SetActive(true)
			m.pubReply.SetFocus(true)
			return m, nil
		}
	}
	cmds := []tea.Cmd{}
	v, vcmd := m.vp.Update(msg)
	m.vp = v
	cmds = append(cmds, vcmd)

	if m.img.Matches(msg) {
		_, icmd := m.img.Update(msg)
		_, pcmd := m.pfp.Update(msg)
		return m, tea.Batch(icmd, pcmd)
	}

	_, rcmd := m.replies.Update(msg)
	cmds = append(cmds, rcmd)

	return m, tea.Batch(cmds...)
}

func (m *CastView) View() string {
	if m.pubReply.Active() {
		return m.pubReply.View()
	}
	return lipgloss.JoinVertical(lipgloss.Top,
		UsernameHeader(&m.cast.Author, m.pfp),
		lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).Padding(0).Render(CastStats(m.cast, 10)),
		CastContent(m.cast, 10),
		m.vp.View(),
		m.img.View(),
		m.replies.View(),
	)
}
