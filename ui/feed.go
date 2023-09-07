package ui

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/castr/api"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)
)

type FeedView struct {
	client *api.Client
	list   list.Model
	cursor int
}

func newList() list.Model {
	d := list.NewDefaultDelegate()
	d.SetHeight(6)
	d.Styles.NormalTitle.Width(80)
	d.Styles.SelectedTitle.Width(80)
	d.Styles.NormalTitle.MaxHeight(2)
	d.Styles.SelectedTitle.MaxHeight(2)

	d.Styles.DimmedDesc.MaxHeight(10)
	d.Styles.NormalDesc.MaxHeight(10)
	d.Styles.DimmedDesc.Width(80)
	d.Styles.NormalDesc.Width(80)

	list := list.New([]list.Item{}, d, 0, 0)
	list.KeyMap.CursorUp.SetKeys("k", "up")
	list.KeyMap.CursorDown.SetKeys("j", "down")
	list.KeyMap.Quit.SetKeys("ctrl+c")
	list.Title = "Feed"
	list.SetShowTitle(true)
	list.SetFilteringEnabled(false)
	return list

}

func NewFeedView(client *api.Client) *FeedView {
	return &FeedView{
		client: client,
		list:   newList(),
	}
}

func (m *FeedView) Init() tea.Cmd {
	return func() tea.Msg {
		feed, err := m.client.GetFeed(api.FeedRequest{FID: 4964})
		if err != nil {
			log.Println("feedview error getting feed", err)
			return err
		}
		return feed
	}
}

func (m *FeedView) setItems(feed *api.FeedResponse) tea.Cmd {
	items := []list.Item{}
	height := 0
	cmds := []tea.Cmd{}
	for _, cast := range feed.Casts {
		ci, cmd := NewCastFeedItem(cast)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		items = append(items, ci)
		height += lipgloss.Height(cast.Text) + 1
	}
	cmd := m.list.SetItems(items)
	return tea.Batch(cmd, tea.Batch(cmds...))
}

func selectCast(cast *api.Cast) tea.Cmd {
	return func() tea.Msg {
		return SelectCastMsg{cast: cast}
	}
}

func (m *FeedView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			current := m.list.SelectedItem().(*CastFeedItem)
			return m, selectCast(current.cast)
		}

		if msg.String() == "o" {
			current := m.list.SelectedItem().(*CastFeedItem)
			cast := current.cast

			return m, OpenURL(fmt.Sprintf("https://warpcast.com/%s/%s", cast.Author.Username, cast.Hash))
		}
	case *api.FeedResponse:
		return m, m.setItems(msg)
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		return m, nil
	}

	l, cmd := m.list.Update(msg)
	cmds = append(cmds, cmd)
	m.list = l

	for _, i := range m.list.Items() {
		if i, ok := i.(*CastFeedItem); ok {
			_, cmd := i.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *FeedView) View() string {
	return m.list.View()

}
