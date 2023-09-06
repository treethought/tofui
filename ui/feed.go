package ui

import (
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/castr/api"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type castItem struct {
	cast api.Cast
}

func (i castItem) Title() string {
	return i.cast.Author.Username
}
func (i castItem) Description() string {
	return i.cast.Text
}

func (i castItem) FilterValue() string {
	return i.cast.Author.Username
}

type FeedView struct {
	client *api.Client
	list   list.Model
	cursor int
}

func newList() list.Model {
	d := list.NewDefaultDelegate()
	d.Styles.NormalTitle.Width(80)
	d.Styles.SelectedTitle.Width(80)
	d.Styles.NormalTitle.MaxHeight(2)
	d.Styles.SelectedTitle.MaxHeight(2)

	d.Styles.DimmedDesc.MaxHeight(10)
	d.Styles.NormalDesc.MaxHeight(10)
	d.Styles.DimmedDesc.Width(80)
	d.Styles.NormalDesc.Width(80)
	// d.SetHeight(8)
	list := list.New([]list.Item{}, castItemDelegate{}, 0, 0)
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
		ci, cmd := NewCastView(cast)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		items = append(items, ci)
		height += lipgloss.Height(cast.Text) + 1
	}
	cmd := m.list.SetItems(items)
	return tea.Batch(cmd, tea.Batch(cmds...))
}

func (m *FeedView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
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
		if i, ok := i.(*CastView); ok {
			_, cmd := i.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	log.Println("FeedView.Update: ")

	return m, tea.Batch(cmds...)
}

func (m *FeedView) View() string {
	return m.list.View()

}
