package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/castr/api"
)

var style = NewStyle().Margin(2,2).BorderStyle(lipgloss.RoundedBorder()).Border(lipgloss.RoundedBorder(), true)

type CastView struct {
	cast    *api.Cast
	img     *ImageModel
	pfp     *ImageModel
	replies *RepliesView
	vp      *viewport.Model

	pubReply *PublishInput
}

func NewCastView(cast *api.Cast) *CastView {
	vp := viewport.New(0, 0)
	c := &CastView{
		cast:     cast,
		pfp:      NewImage(false, true, special),
		img:      NewImage(true, true, special),
		replies:  NewRepliesView(),
		vp:       &vp,
		pubReply: NewPublishInput(nil),
	}
	return c
}

func (m *CastView) Clear() {
	m.cast = nil
	m.pubReply.SetFocus(false)
	m.pubReply.SetActive(false)
	m.replies.Clear()
	m.img.Clear()
	m.pfp.Clear()
}

func (m *CastView) SetCast(cast *api.Cast) tea.Cmd {
	m.Clear()
	m.cast = cast
	cmds := []tea.Cmd{
		m.replies.SetOpHash(m.cast.Hash),
		m.pfp.SetURL(m.cast.Author.PfpURL, false),
		m.pfp.SetSize(4, 4),
		m.pubReply.SetContext(m.cast.Hash, m.cast.ParentURL, m.cast.Author.FID),
	}
	if len(m.cast.Embeds) > 0 {
		cmds = append(cmds, m.img.SetURL(m.cast.Embeds[0].URL, true))
	}
	return tea.Sequence(cmds...)
}

func (m *CastView) Init() tea.Cmd {
	return nil
}

func (m *CastView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmds := []tea.Cmd{}

		fx, fy := style.GetFrameSize()

		w, h := msg.Width-fx, msg.Height-fy-6 // room for stats/header

		cx, cy := w, h/2

		m.vp.Width = cx
		m.vp.Height = cy / 2

		m.img.SetSize(cx/2, cy/2)

		m.replies.SetSize(msg.Width, h/2)

		m.pubReply.SetSize(msg.Width, msg.Height-10)
		return m, tea.Batch(cmds...)

	case *ctxInfoMsg:
		_, cmd := m.pubReply.Update(msg)
		return m, cmd

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
	m.vp.SetContent(CastContent(m.cast, 10))
	cmds := []tea.Cmd{}

	_, rcmd := m.replies.Update(msg)
	cmds = append(cmds, rcmd)

	// v, vcmd := m.vp.Update(msg)
	// m.vp = &v
	// cmds = append(cmds, vcmd)

	if m.img.Matches(msg) {
		_, icmd := m.img.Update(msg)
		_, pcmd := m.pfp.Update(msg)
		return m, tea.Batch(icmd, pcmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *CastView) View() string {
	if m.pubReply.Active() {
		return m.pubReply.View()
	}

	return style.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			UsernameHeader(&m.cast.Author, m.pfp),
			NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).Padding(0).Render(CastStats(m.cast, 10)),
			// CastContent(m.cast, 10),
			m.vp.View(),
			m.img.View(),
			m.replies.View(),
		),
	)
}
