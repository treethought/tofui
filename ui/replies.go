package ui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/treethought/castr/api"
)

type repliesMsg struct {
	castConvo *api.Cast
	err       error
}

type RepliesView struct {
	opHash string
	convo  *api.Cast
	items  []*CastFeedItem
	feed   *FeedView
}

func getConvoCmd(hash string) tea.Cmd {
	return func() tea.Msg {
		cc, err := api.GetClient().GetCastWithReplies(hash)
		if err != nil {
			return &repliesMsg{err: err}
		}
		return &repliesMsg{castConvo: cc}
	}
}

func NewRepliesView() *RepliesView {
	feed := NewFeedView(api.GetClient(), nil)
	feed.SetShowChannel(false)
	feed.SetShowStats(false)
	return &RepliesView{
		feed: feed,
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
	return getConvoCmd(hash)
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
