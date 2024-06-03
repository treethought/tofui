package ui

import (
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/castr/api"
)

type QuickSelect struct {
	app      *App
	active   bool
	channels *list.Model
	w, h     int
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
	l.Title = "channel switcher"
	l.SetShowTitle(true)
	l.SetFilteringEnabled(true)
	l.SetShowFilter(true)
	l.SetShowHelp(true)
	l.SetShowStatusBar(true)
	l.SetShowPagination(true)

	return &QuickSelect{app: app, channels: &l}
}

type channelListMsg struct {
	channels   []*api.Channel
	activeOnly bool
}

func getUserChannels(activeOnly bool) tea.Msg {
	fid := api.GetSigner().FID
	channels, err := api.GetClient().GetUserChannels(fid, activeOnly, api.WithLimit(100))
	if err != nil {
		log.Println("error getting user channels: ", err)
		return nil
	}
	return &channelListMsg{channels, activeOnly}
}

func getChannelsCmd(activeOnly bool) tea.Cmd {
	return func() tea.Msg {
		if activeOnly {
			return getUserChannels(activeOnly)
		}
		msg := &channelListMsg{}
		ids, err := api.GetClient().GetCachedChannelIds()
		if err != nil {
			log.Println("error getting channel names: ", err)
		}
		for _, id := range ids {
			channel, err := api.GetClient().GetChannelById(id)
			if err != nil {
				log.Println("error getting channel: ", err)
				continue
			}
			msg.channels = append(msg.channels, channel)
		}
		return msg
	}
}

func (m *QuickSelect) SetSize(w, h int) {
	m.w = w
	m.h = h
	m.channels.SetSize(w, h)
}

func (m *QuickSelect) Init() tea.Cmd {
	return tea.Batch(getChannelsCmd(false), func() tea.Msg { return tea.KeyCtrlQuestionMark })
}

func (m *QuickSelect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case []*api.Channel:
		items := []list.Item{}
		for _, c := range msg {
			items = append(items, &selectItem{name: c.Name, value: c.ParentURL, itype: "channel"})
		}
		return m, m.channels.SetItems(items)

	case tea.KeyMsg:
		if msg.String() == "enter" {
			currentItem := m.channels.SelectedItem().(*selectItem)
			if currentItem.name == "profile" {
				log.Println("profile selected")
				fid := api.GetSigner().FID
				if fid == 0 {
					return m, nil
				}
				return m, tea.Sequence(m.app.SetFocus("profile"), selectProfileCmd(fid))
			}
			if currentItem.name == "feed" {
				log.Println("feed selected")
				return m, tea.Sequence(m.app.SetFocus("feed"), getFeedCmd(DefaultFeedParams()))
			}
			if currentItem.itype == "channel" {
				log.Println("channel selected")
				return m, tea.Sequence(m.app.SetFocus("feed"), getFeedCmd(&api.FeedRequest{FeedType: "filter", FilterType: "parent_url", ParentURL: currentItem.value, Limit: 100}))
			}
		}
	}
	l, cmd := m.channels.Update(msg)
	m.channels = &l
	return m, cmd
}

var dialogBoxStyle = lipgloss.NewStyle().
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
		dialogBoxStyle.Width(m.w).Height(m.h).Render(m.channels.View()),
		lipgloss.WithWhitespaceChars("~~"),
		lipgloss.WithWhitespaceForeground(subtle),
	)
	return dialog
}
