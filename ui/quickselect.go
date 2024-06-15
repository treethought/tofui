package ui

import (
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/tofui/api"
)

type QuickSelect struct {
	app         *App
	active      bool
	channelList *list.Model
	w, h        int
	onSelect    func(i *selectItem) tea.Cmd
}

type selectItem struct {
	name  string
	value string
	icon  string
	itype string
}

func (m *selectItem) FilterValue() string {
	return m.name
}
func (m *selectItem) View() string {
	return m.value
}
func (m *selectItem) Title() string {
	return m.name
}

func (i *selectItem) Description() string {
	return ""
}

func NewQuickSelect(app *App) *QuickSelect {
	d := list.NewDefaultDelegate()
	d.SetHeight(1)
	d.ShowDescription = false

	l := list.New([]list.Item{}, d, 100, 100)
	l.KeyMap.CursorUp.SetKeys("k", "up")
	l.KeyMap.CursorDown.SetKeys("j", "down")
	l.KeyMap.Quit.SetKeys("ctrl+c")
	l.Title = "select channel"
	l.SetShowTitle(true)
	l.SetFilteringEnabled(true)
	l.SetShowFilter(true)
	l.SetShowHelp(true)
	l.SetShowStatusBar(true)
	l.SetShowPagination(true)

	return &QuickSelect{app: app, channelList: &l}
}

type channelListMsg struct {
	channels   []*api.Channel
	activeOnly bool
}

func getUserChannels(client *api.Client, fid uint64, activeOnly bool) tea.Msg {
	channels, err := client.GetUserChannels(fid, activeOnly, api.WithLimit(100))
	if err != nil {
		log.Println("error getting user channels: ", err)
		return nil
	}
	return &channelListMsg{channels, activeOnly}
}

func getChannelsCmd(client *api.Client, activeOnly bool, fid uint64) tea.Cmd {
	return func() tea.Msg {
		if activeOnly && fid != 0 {
			return getUserChannels(client, fid, activeOnly)
		}
		msg := &channelListMsg{}
		ids, err := client.GetCachedChannelIds()
		if err != nil {
			log.Println("error getting channel names: ", err)
		}
		for _, id := range ids {
			channel, err := client.GetChannelById(id)
			if err != nil {
				log.Println("error getting channel: ", err)
				continue
			}
			msg.channels = append(msg.channels, channel)
		}
		return msg
	}
}
func (m *QuickSelect) SetOnSelect(f func(i *selectItem) tea.Cmd) {
	m.onSelect = f
}

func (m *QuickSelect) SetSize(w, h int) {
	m.w = w
	m.h = h
	m.channelList.SetSize(w, h)
}
func (m *QuickSelect) Active() bool {
	return m.active
}
func (m *QuickSelect) SetActive(active bool) {
	m.active = active
}

func (m *QuickSelect) Init() tea.Cmd {
	var fid uint64
	if m.app.ctx.signer != nil {
		fid = m.app.ctx.signer.FID
	}
	return tea.Batch(
		getChannelsCmd(m.app.client, false, fid), func() tea.Msg { return tea.KeyCtrlQuestionMark })
}

func (m *QuickSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case []*api.Channel:
		items := []list.Item{}
		for _, c := range msg {
			items = append(items, &selectItem{name: c.Name, value: c.ParentURL, itype: "channel"})
		}
		return m, m.channelList.SetItems(items)

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if !m.active {
			return m, nil
		}
		if msg.String() == "enter" {
			if m.onSelect != nil {
				return m, m.onSelect(m.channelList.SelectedItem().(*selectItem))
			}
			currentItem, ok := m.channelList.SelectedItem().(*selectItem)
			if !ok {
				log.Println("no item selected")
				return m, nil
			}
			if currentItem.name == "profile" {
				log.Println("profile selected")
				if m.app.ctx.signer == nil {
					return m, nil
				}
				return m, tea.Sequence(
					selectProfileCmd(m.app.ctx.signer.FID),
					m.app.FocusProfile(),
				)
			}
			if currentItem.name == "feed" {
				log.Println("feed selected")
				return m, tea.Sequence(m.app.FocusFeed(), getDefaultFeedCmd(m.app.client, m.app.ctx.signer))
			}
			if currentItem.itype == "channel" {
				log.Println("channel selected")
				return m, tea.Sequence(
					m.app.FocusChannel(),
					getFeedCmd(m.app.client, &api.FeedRequest{
						FeedType: "filter", FilterType: "parent_url",
						ParentURL: currentItem.value, Limit: 100,
					}),
				)
			}
		}
	}
	l, cmd := m.channelList.Update(msg)
	m.channelList = &l
	return m, cmd
}

var dialogBoxStyle = NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#874BFD")).
	Padding(1, 0).
	BorderTop(true).
	BorderLeft(true).
	BorderRight(true).
	BorderBottom(true)

func (m *QuickSelect) View() string {
	dialog := lipgloss.Place(10, 10,
		lipgloss.Center, lipgloss.Center,
		dialogBoxStyle.Width(m.w).Height(m.h).Render(m.channelList.View()),
		lipgloss.WithWhitespaceChars("~~"),
		lipgloss.WithWhitespaceForeground(subtle),
	)
	return dialog
}
