package ui

import (
	"fmt"
	"log"

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
	help    *HelpView

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
		pfp:      NewImage(true, true, special),
		img:      NewImage(true, true, special),
		replies:  NewRepliesView(app),
		vp:       &vp,
		header:   &hp,
		pubReply: NewPublishInput(app),
		hasImg:   false,
		help:     NewHelpView(app, CastViewKeyMap),
	}
	c.pfp.SetSize(4, 4)
	c.help.SetFull(false)
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

func (m *CastView) LikeCast() tea.Cmd {
	if m.cast == nil {
		return nil
	}
	return likeCastCmd(m.app.client, m.app.ctx.signer, m.cast)
}

func (m *CastView) OpenCast() tea.Cmd {
	if m.cast == nil {
		return nil
	}
	return OpenURL(fmt.Sprintf("https://warpcast.com/%s/%s", m.cast.Author.Username, m.cast.Hash))
}

func (m *CastView) Reply() {
	if m.cast == nil {
		return
	}
	m.pubReply.SetActive(true)
	m.pubReply.SetFocus(true)
}

func (m *CastView) ViewProfile() tea.Cmd {
	if m.cast == nil {
		return nil
	}

	userFid := m.cast.Author.FID
	return tea.Sequence(
		m.app.FocusProfile(),
		getUserCmd(m.app.client, userFid, m.app.ctx.signer.FID),
		getUserFeedCmd(m.app.client, userFid, m.app.ctx.signer.FID),
	)
}
func (m *CastView) ViewChannel() tea.Cmd {
	if m.cast == nil || m.cast.ParentURL == "" {
		return nil
	}

	return tea.Batch(
		getChannelFeedCmd(m.app.client, m.cast.ParentURL),
		fetchChannelCmd(m.app.client, m.cast.ParentURL),
		m.app.FocusChannel(),
	)
}

func (m *CastView) ViewParent() tea.Cmd {
	if m.cast == nil || m.cast.ParentHash == "" {
		return nil
	}
	return func() tea.Msg {
		cast, err := m.app.client.GetCastWithReplies(m.app.ctx.signer, m.cast.ParentHash)
		if err != nil {
			log.Println("failed to get parent cast", err)
			return nil
		}

		return m.SetCast(cast)
	}
}

func (m *CastView) SetCast(cast *api.Cast) tea.Cmd {
	m.Clear()
	m.cast = cast
	m.pfp.SetURL(m.cast.Author.PfpURL, false)
	m.pfp.SetSize(4, 4)
	cmds := []tea.Cmd{
		m.replies.SetOpHash(m.cast.Hash),
		m.pubReply.SetContext(m.cast.Hash, m.cast.ParentURL, m.cast.Author.FID),
		m.pfp.Render(),
	}
	m.hasImg = false
	if len(m.cast.Embeds) > 0 {
		m.hasImg = true
		m.img.SetURL(m.cast.Embeds[0].URL, true)
		cmds = append(cmds, m.resize(), m.img.Render())
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

func (m *CastView) resize() tea.Cmd {
	cmds := []tea.Cmd{}
	fx, fy := style.GetFrameSize()
	w := min(m.w-fx, int(float64(GetWidth())*0.75))
	h := min(m.h-fy, GetHeight()-4)

	m.help.SetSize(m.w, 1)

	m.header.Height = min(10, int(float64(h)*0.2))

	hHeight := lipgloss.Height(m.header.View())

	cy := h - hHeight

	m.vp.Width = w - fx
	m.vp.Height = int(float64(cy) * 0.5) //- fy

	if m.hasImg {
		q := int(float64(cy) * 0.25)
		m.vp.Height = q
		m.img.SetSize(m.vp.Width/2, q)

		cmds = append(cmds, m.img.Render())
	} else {
		m.img.Clear()
	}
	m.replies.SetSize(w, int(float64(cy)*0.5))

	m.pubReply.SetSize(w, h)
	return tea.Batch(cmds...)
}

func (m *CastView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		return m, m.resize()

	case *ctxInfoMsg:
		_, cmd := m.pubReply.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if m.pubReply.Active() {
			_, cmd := m.pubReply.Update(msg)
			return m, cmd
		}
		if CastViewKeyMap.HandleMsg(m, msg) != nil {
			return m, CastViewKeyMap.HandleMsg(m, msg)
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
			m.help.View(),
			m.replies.View(),
		),
	)
}
