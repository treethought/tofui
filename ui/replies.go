package ui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/treethought/tofui/api"
)

type repliesMsg struct {
	castConvo *api.Cast
	err       error
}

type RepliesView struct {
	app    *App
	opHash string
	convo  *api.Cast
	items  []*CastFeedItem
	feed   *FeedView
}

func getConvoCmd(client *api.Client, signer *api.Signer, hash string) tea.Cmd {
	return func() tea.Msg {
		cc, err := client.GetCastWithReplies(signer, hash)
		if err != nil {
			return &repliesMsg{err: err}
		}
		return &repliesMsg{castConvo: cc}
	}
}

func NewRepliesView(app *App) *RepliesView {
	feed := NewFeedView(app)
	feed.SetShowChannel(false)
	feed.SetShowStats(false)
	return &RepliesView{
		feed: feed,
    app: app,
	}
}

func (m *RepliesView) Init() tea.Cmd {
	return nil
}

func (m *RepliesView) Clear() {
	m.feed.Clear()
	m.opHash = ""
	m.convo = nil
	m.items = nil
}

func (m *RepliesView) SetOpHash(hash string) tea.Cmd {
	m.Clear()
	m.opHash = hash
  if m.app == nil {
    log.Println("app is nil")
  }
	if m.app.ctx == nil {
		log.Println("app context is nil")
	}
	if m.app.ctx.signer == nil {
		log.Println("signer is nil")
	}

	return getConvoCmd(m.app.client, m.app.ctx.signer, hash)
}

func (m *RepliesView) SetSize(w, h int) {
	m.feed.SetSize(w, h)
}

func (m *RepliesView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *repliesMsg:
		if msg.err != nil {
			log.Println("error getting convo: ", msg.err)
			return m, nil
		}
		m.Clear()
		m.convo = msg.castConvo
		return m, m.feed.setItems(msg.castConvo.DirectReplies)
	}
	_, cmd := m.feed.Update(msg)
	return m, cmd
}

func (m *RepliesView) View() string {
	return m.feed.View()
}
