package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/tofui/api"
)

var style = NewStyle()
var statsStyle = NewStyle()
var castHeaderStyle = NewStyle().Margin(1, 1).Align(lipgloss.Top)

type CastView struct {
	app     *App
	cast    *api.Cast
	img     *ImageModel
	pfp     *ImageModel
	replies *RepliesView
	vp      *viewport.Model
	header  *viewport.Model
	hasImg  bool

	pubReply *PublishInput
	w, h     int
}

func NewCastView(app *App, cast *api.Cast) *CastView {
	vp := viewport.New(0, 0)
	hp := viewport.New(0, 0)
	hp.Style = NewStyle().BorderBottom(true).BorderStyle(lipgloss.RoundedBorder())
	c := &CastView{
		app:      app,
		cast:     cast,
		pfp:      NewImage(false, true, special),
		img:      NewImage(true, true, special),
		replies:  NewRepliesView(app),
		vp:       &vp,
		header:   &hp,
		pubReply: NewPublishInput(app),
		hasImg:   false,
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
	m.hasImg = false
	if len(m.cast.Embeds) > 0 {
		m.hasImg = true
		cmds = append(cmds, m.img.SetURL(m.cast.Embeds[0].URL, true))
	}
	return tea.Sequence(cmds...)
}

func (m *CastView) Init() tea.Cmd {
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m *CastView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmds := []tea.Cmd{}

		fx, fy := style.GetFrameSize()

		w, h := msg.Width-fx, msg.Height-fy
		m.w, m.h = w, h

		m.header.Width = w
		m.header.Height = min(10, int(float64(h)*0.2))

		hHeight := lipgloss.Height(m.header.View())

		cx, cy := w, h-hHeight

		m.vp.Width = cx
		m.vp.Height = int(float64(cy) * 0.5)

		m.img.SetSize(0, 0)

		if m.hasImg {
			m.img.SetSize(4, 4)
			m.vp.Height = int(float64(cy) * 0.25)
		}
		m.replies.SetSize(msg.Width, int(float64(cy)*0.5))

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
			return m, likeCastCmd(m.app.client, m.app.ctx.signer, m.cast)
		}
		if msg.String() == "C" {
			m.pubReply.SetActive(true)
			m.pubReply.SetFocus(true)
			return m, nil
		}
	}
	m.vp.SetContent(CastContent(m.cast, 10))
	m.header.SetContent(m.castHeader())
	cmds := []tea.Cmd{}

	_, rcmd := m.replies.Update(msg)
	cmds = append(cmds, rcmd)

	if m.img.Matches(msg) {
		_, icmd := m.img.Update(msg)
		_, pcmd := m.pfp.Update(msg)
		return m, tea.Batch(icmd, pcmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *CastView) castHeader() string {
	if m.cast == nil {
		return ""
	}
	return castHeaderStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			UsernameHeader(&m.cast.Author, m.pfp),
			CastStats(m.cast, 1),
		),
	)

}

func (m *CastView) View() string {
	if m.pubReply.Active() {
		return m.pubReply.View()
	}

	return style.Height(m.h).Render(
		lipgloss.JoinVertical(lipgloss.Center,
			m.header.View(),
			m.vp.View(),
			m.img.View(),
			m.replies.View(),
		),
	)
}
