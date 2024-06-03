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
	Cast key.Binding
	Back key.Binding
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
}

type PublishInput struct {
	keys        keyMap
	help        help.Model
	ta          textarea.Model
	vp          viewport.Model
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

	return &PublishInput{ta: ta, vp: vp, keys: keys, help: help.New()}
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

func (m *PublishInput) SetContext(parent, channel string, parentAuthor uint64) {
	m.ctx.channel = channel
	m.ctx.parent = parent
	m.ctx.parentAuthor = parentAuthor
}

func (m *PublishInput) SetFocus(focus bool) {
	if focus {
		m.ta.Focus()
		return
	}
	m.ta.Blur()
}

func (m *PublishInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
		m.showConfirm = false
		m.SetFocus(false)
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

	var cmd tea.Cmd
	m.ta, cmd = m.ta.Update(msg)
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

	dialog := lipgloss.Place(10, 10,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Top,
			dialogBoxStyle.Width(m.w).Height(m.h).Render(content),
			fmt.Sprintf("channel: %s", m.ctx.channel),
			fmt.Sprintf("parent: %s", m.ctx.parent),
			fmt.Sprintf("parent_author: %s", m.ctx.parent),
		),
		// lipgloss.WithWhitespaceChars("猫咪"),
		lipgloss.WithWhitespaceChars("~~"),
		lipgloss.WithWhitespaceForeground(subtle),
	)
	return dialog
}
