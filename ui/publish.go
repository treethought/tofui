package ui

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/tofui/api"
)

const confirmPrefix = "Publish cast? (y/n)"

type postResponseMsg struct {
	err  error
	resp *api.PostCastResponse
}

type ctxInfoMsg struct {
	user    *api.User
	channel *api.Channel
}

func postCastCmd(client *api.Client, signer *api.Signer, text, parent, channel string, parentAuthor uint64) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.PostCast(signer, text, parent, channel, parentAuthor)
		if err != nil {
			return &postResponseMsg{err: err}
		}
		return &postResponseMsg{resp: resp}
	}
}

type keyMap struct {
	Cast          key.Binding
	Back          key.Binding
	ChooseChannel key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Cast, k.Back, k.ChooseChannel}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Cast},
		{k.Back},
	}
}

var keys = keyMap{
	Cast: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "publish cast"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back to feed"),
	),
	ChooseChannel: key.NewBinding(
		key.WithKeys("ctrl+w"),
		key.WithHelp("ctrl+w", "choose channel"),
	),
}

type castContext struct {
	channel      string
	parent       string
	parentAuthor uint64
	parentUser   *api.User
}

type PublishInput struct {
	app         *App
	keys        keyMap
	help        help.Model
	ta          *textarea.Model
	vp          *viewport.Model
	showConfirm bool
	active      bool
	w, h        int
	castCtx     castContext
	qs          *QuickSelect
}

func NewPublishInput(app *App) *PublishInput {
	ta := textarea.New()
	if app.ctx.signer == nil {
		ta.Placeholder = "please sign in to post"
	} else {
		ta.Placeholder = "publish cast..."
	}
	ta.CharLimit = 1024
	ta.ShowLineNumbers = false
	ta.Prompt = ""

	vp := viewport.New(0, 0)
	vp.SetContent(ta.View())

	qs := NewQuickSelect(app)

	return &PublishInput{ta: &ta, vp: &vp, keys: keys, help: help.New(), app: app, qs: qs}
}

func (m *PublishInput) Init() tea.Cmd {
	m.qs.SetOnSelect(func(i *selectItem) tea.Cmd {
		m.qs.SetActive(false)
		return m.SetContext("", i.name, 0)
	})
	return m.qs.Init()
}

func (m *PublishInput) Active() bool {
	return m.active
}
func (m *PublishInput) SetActive(active bool) {
	m.active = active
}

func (m *PublishInput) SetSize(w, h int) {
	m.w = w
	m.h = h
	m.ta.SetWidth(w)
	m.ta.SetHeight(h)
	m.vp.Width = w
	m.vp.Height = h
	m.qs.SetSize(w, h)
}

func (m *PublishInput) SetContext(parent, channelParentUrl string, parentAuthor uint64) tea.Cmd {
	return func() tea.Msg {
		m.castCtx.channel = channelParentUrl
		m.castCtx.parent = parent
		m.castCtx.parentAuthor = parentAuthor
		m.castCtx.parentUser = nil
		var viewer uint64
		if m.app.ctx.signer != nil {
			viewer = m.app.ctx.signer.FID
		}
		var parentUser *api.User
		var channel *api.Channel
		var err error
		if parentAuthor > 0 {
			parentUser, err = m.app.client.GetUserByFID(parentAuthor, viewer)
			if err != nil {
				log.Println("error getting parent author: ", err)
				return nil
			}
		}
		if channelParentUrl != "" {
			channel, err = m.app.client.GetChannelByParentUrl(channelParentUrl)
			if err != nil {
				log.Println("error getting channel by parent url, trying channel id: ", err)
				channel, err = m.app.client.GetChannelById(channelParentUrl)
				if err != nil {
					log.Println("error getting channel by id: ", err)
					return nil
				}
			}
		}
		return &ctxInfoMsg{user: parentUser, channel: channel}
	}
}

func (m *PublishInput) SetFocus(focus bool) {
	if focus {
		m.ta.Focus()
		return
	}
	m.ta.Blur()
}
func (m *PublishInput) Clear() {
	m.ta.Reset()
	m.vp.SetContent(m.ta.View())
	m.showConfirm = false
	m.SetFocus(false)
	m.SetContext("", "", 0)
}

func (m *PublishInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *ctxInfoMsg:
		m.castCtx.parentUser = msg.user
		if msg.channel != nil {
			m.castCtx.channel = msg.channel.Name
		}
		return m, nil
	case *postResponseMsg:
		if msg.err != nil {
			log.Println("error posting cast: ", msg.err)
			m.vp.SetContent(NewStyle().Foreground(lipgloss.Color("#ff0000")).Render("error posting cast!"))
			return m, nil
		}
		if msg.resp == nil || !msg.resp.Success {
			m.vp.SetContent(NewStyle().Foreground(lipgloss.Color("#ff0000")).Render("error posting cast!"))
			return m, nil
		}
		log.Println("cast posted: ", msg.resp.Cast.Hash)
		m.Clear()
		m.SetActive(false)
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return nil, tea.Quit
		}
		if m.qs.Active() {
			_, cmd := m.qs.Update(msg)
			return m, cmd
		}

		switch {
		case key.Matches(msg, m.keys.Cast):
			m.showConfirm = true
			return m, nil
		case key.Matches(msg, m.keys.Back):
			m.active = false
			return nil, nil
		case key.Matches(msg, m.keys.ChooseChannel):
			m.qs.SetActive(true)
		}

		if m.showConfirm {
			if msg.String() == "y" || msg.String() == "Y" {
				return m, postCastCmd(
					m.app.client, m.app.ctx.signer,
					m.ta.Value(), m.castCtx.parent,
					m.castCtx.channel, m.castCtx.parentAuthor,
				)
			} else if msg.String() == "n" || msg.String() == "N" || msg.String() == "esc" {
				m.showConfirm = false
				return m, nil
			}
		}
	}
	if m.app.ctx.signer == nil {
		m.ta.Blur()
		m.ta.SetValue("please sign in to post")
		return m, nil
	}

	var cmds []tea.Cmd
	_, cmd := m.qs.Update(msg)
	cmds = append(cmds, cmd)

	ta, tcmd := m.ta.Update(msg)
	m.ta = &ta
	cmds = append(cmds, tcmd)
	return m, tea.Batch(cmds...)
}

func (m *PublishInput) viewConfirm() string {
	header := NewStyle().BorderBottom(true).BorderStyle(lipgloss.NormalBorder()).Render(confirmPrefix)
	return lipgloss.JoinVertical(lipgloss.Top,
		header, m.ta.View())
}

func (m *PublishInput) View() string {
	content := m.ta.View()
	if m.showConfirm {
		content = m.viewConfirm()
	} else if m.qs.Active() {
		content = m.qs.View()

	} else {
		content = lipgloss.JoinVertical(lipgloss.Top,
			content,
			m.help.View(m.keys),
		)
	}

	titleText := "publish cast"
	if m.castCtx.parentUser != nil {
		titleText = fmt.Sprintf("reply to @%s", m.castCtx.parentUser.Username)
	} else if m.castCtx.channel != "" {
		titleText = fmt.Sprintf("publish cast to channel: /%s", m.castCtx.channel)
	}

	titleStyle := NewStyle().Foreground(lipgloss.Color("#874BFD")).BorderBottom(true).BorderStyle(lipgloss.NormalBorder())
	title := titleStyle.Render(titleText)

	dialog := lipgloss.Place(m.w/2, m.h/2,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Top,
			title,
			dialogBoxStyle.Width(m.w).Height(m.h).Render(content),
		),
		// lipgloss.WithWhitespaceChars("猫咪"),
		lipgloss.WithWhitespaceChars("~~"),
		lipgloss.WithWhitespaceForeground(subtle),
	)
	return dialog
}
