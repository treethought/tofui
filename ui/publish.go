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

	"github.com/treethought/castr/api"
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

func postCastCmd(text, parent, channel string, parentAuthor uint64) tea.Cmd {
	return func() tea.Msg {
		resp, err := api.GetClient().PostCast(text, parent, channel, parentAuthor)
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
	return []key.Binding{k.Cast, k.Back}
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
}

type ctx struct {
	channel      string
	parent       string
	parentAuthor uint64
	parentUser   *api.User
}

type PublishInput struct {
	keys        keyMap
	help        help.Model
	ta          *textarea.Model
	vp          *viewport.Model
	showConfirm bool
	active      bool
	w, h        int
	ctx         ctx
}

func NewPublishInput(app *App) *PublishInput {
	ta := textarea.New()
	ta.Placeholder = "publish cast..."
	ta.CharLimit = 320
	ta.ShowLineNumbers = false
	ta.Prompt = ""

	vp := viewport.New(0, 0)
	vp.SetContent(ta.View())

	return &PublishInput{ta: &ta, vp: &vp, keys: keys, help: help.New()}
}

func (m *PublishInput) Init() tea.Cmd {
	return nil
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
}

func (m *PublishInput) SetContext(parent, channel string, parentAuthor uint64) tea.Cmd {
	return func() tea.Msg {
		m.ctx.channel = channel
		m.ctx.parent = parent
		m.ctx.parentAuthor = parentAuthor
		m.ctx.parentUser = nil
		parentUser, err := api.GetClient().GetUserByFID(parentAuthor)
		if err != nil {
			log.Println("error getting parent author: ", err)
			return nil
		}
		channel, err := api.GetClient().GetChannelByParentUrl(channel)
		if err != nil {
			log.Println("error getting channel: ", err)
			return nil
		}
		log.Println("got ctx: ", channel.Name, parentUser.Username)
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
}

func (m *PublishInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *ctxInfoMsg:
		log.Println("got ctx msg: ", msg.channel.Name, msg.user.Username)
		m.ctx.parentUser = msg.user
		m.ctx.channel = msg.channel.Name
		return m, nil
	case *postResponseMsg:
		if msg.err != nil {
			log.Println("error posting cast: ", msg.err)
			m.vp.SetContent(lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render("error posting cast!"))
			return m, nil
		}
		if msg.resp == nil || !msg.resp.Success {
			m.vp.SetContent(lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render("error posting cast!"))
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

		switch {
		case key.Matches(msg, m.keys.Cast):
			m.showConfirm = true
			return m, nil
		case key.Matches(msg, m.keys.Back):
			m.active = false
			return nil, nil
		}

		if m.showConfirm {
			if msg.String() == "y" || msg.String() == "Y" {
				return m, postCastCmd(m.ta.Value(), m.ctx.parent, m.ctx.channel, m.ctx.parentAuthor)
			} else if msg.String() == "n" || msg.String() == "N" || msg.String() == "esc" {
				m.showConfirm = false
				return m, nil
			}
		}
	}

	ta, cmd := m.ta.Update(msg)
	m.ta = &ta
	return m, cmd
}

func (m *PublishInput) viewConfirm() string {
	header := lipgloss.NewStyle().BorderBottom(true).BorderStyle(lipgloss.NormalBorder()).Render(confirmPrefix)
	return lipgloss.JoinVertical(lipgloss.Top,
		header, m.ta.View())
}

func (m *PublishInput) View() string {
	content := m.ta.View()
	if m.showConfirm {
		content = m.viewConfirm()
	} else {
		content = lipgloss.JoinVertical(lipgloss.Top,
			content,
			m.help.View(m.keys),
		)
	}

	titleText := "publish cast"
	if m.ctx.parentUser != nil {
		titleText = fmt.Sprintf("reply to @%s", m.ctx.parentUser.Username)
	} else if m.ctx.channel != "" {
		titleText = fmt.Sprintf("publish cast to channel: /%s", m.ctx.channel)
	}

	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#874BFD")).BorderBottom(true).BorderStyle(lipgloss.NormalBorder())
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
