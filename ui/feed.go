package ui

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/castr/api"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)
)

type apiErrorMsg struct {
	err error
}

type reactMsg struct {
	hash  string
	rtype string
	state bool
}

type FeedView struct {
	compact bool
	client  *api.Client
	list    list.Model
	table   table.Model
	cursor  int
	items   []*CastFeedItem
	loading *Loading
	req     *api.FeedRequest
}

func newTable() table.Model {
	columns := []table.Column{
		{Title: "channel", Width: 20},
		{Title: "", Width: 2},
		{Title: "user", Width: 20},
		{Title: "cast", Width: 200},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithKeyMap(table.KeyMap{
			LineUp:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
			LineDown:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
			PageUp:     key.NewBinding(key.WithKeys("pageup", "K"), key.WithHelp("PgUp/K", "page up")),
			PageDown:   key.NewBinding(key.WithKeys("pagedown", "J"), key.WithHelp("PgDn/J", "page down")),
			GotoTop:    key.NewBinding(key.WithKeys("home", "g"), key.WithHelp("Home/g", "go to top")),
			GotoBottom: key.NewBinding(key.WithKeys("end", "G"), key.WithHelp("End/G", "go to bottom")),
		}),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
	return t
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

func DefaultFeedParams() *api.FeedRequest {
	return &api.FeedRequest{FeedType: "following", Limit: 100}
}

func NewFeedView(client *api.Client, req *api.FeedRequest) *FeedView {
	p := progress.New()
	p.ShowPercentage = false
	return &FeedView{
		client:  client,
		list:    newList(),
		table:   newTable(),
		items:   []*CastFeedItem{},
		loading: NewLoading(),
		compact: true,
		req:     req,
	}
}

func (m *FeedView) Init() tea.Cmd {
	if len(m.items) > 0 {
		return nil
	}
	m.loading.SetActive(true)
	cmds := []tea.Cmd{
		m.loading.Init(),
	}

	if m.req != nil {
		cmds = append(cmds, getFeedCmd(m.req))
	}
	return tea.Batch(cmds...)
}

func (m *FeedView) Clear() {
	m.items = []*CastFeedItem{}
	m.req = nil
}

func likeCastCmd(cast *api.Cast) tea.Cmd {
	return func() tea.Msg {
		log.Println("liking cast", cast.Hash)
		if err := api.GetClient().React(cast.Hash, "like"); err != nil {
			return apiErrorMsg{err}
		}
		return reactMsg{hash: cast.Hash, rtype: "like", state: true}
	}
}
func getFeedCmd(req *api.FeedRequest) tea.Cmd {
	return func() tea.Msg {
		if req.Limit == 0 {
			req.Limit = 100
		}
		feed, err := api.GetClient().GetFeed(req)
		if err != nil {
			log.Println("feedview error getting feed", err)
			return err
		}
		return feed
	}
}

func (m *FeedView) SetDefaultParams() tea.Cmd {
	return tea.Sequence(
		m.setItems(nil),
		getFeedCmd(&api.FeedRequest{FeedType: "following", Limit: 100}),
	)
}
func (m *FeedView) SetParams(req *api.FeedRequest) tea.Cmd {
	return tea.Sequence(
		m.setItems(nil),
		getFeedCmd(req),
	)
}

func (m *FeedView) setItems(feed *api.FeedResponse) tea.Cmd {
	height := 0
	items := []list.Item{}
	rows := []table.Row{}
	cmds := []tea.Cmd{}
	for _, cast := range feed.Casts {
		ci, cmd := NewCastFeedItem(cast, m.compact)
		m.items = append(m.items, ci)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		if m.compact {
			rows = append(rows, ci.AsRow())
			continue
		}
		items = append(items, ci)
		height += lipgloss.Height(cast.Text) + 1
	}
	if m.compact {
		m.table.SetRows(rows)
		return tea.Batch(cmds...)
	}

	cmd := m.list.SetItems(items)
	return tea.Batch(cmd, tea.Batch(cmds...))
}

func (m *FeedView) populateItems() tea.Cmd {
	m.loading.SetActive(false)
	if !m.compact {
		return m.list.SetItems([]list.Item{})
	}
	rows := []table.Row{}
	for _, i := range m.items {
		rows = append(rows, i.AsRow())
	}
	m.table.SetRows(rows)
	return nil
}

func selectCast(cast *api.Cast) tea.Cmd {
	return func() tea.Msg {
		return SelectCastMsg{cast: cast}
	}
}

func (m *FeedView) getCurrentItem() *CastFeedItem {
	if !m.compact {
		return m.list.SelectedItem().(*CastFeedItem)
	}

	row := m.table.Cursor()
	if row >= len(m.items) {
		return nil
	}
	return m.items[row]
}
func (m *FeedView) SetSize(w, h int) {
	m.list.SetSize(w, h)
	m.table.SetWidth(w)
	m.table.SetHeight(h)
}

func (m *FeedView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			current := m.getCurrentItem()
			return m, selectCast(current.cast)
		}

		if msg.String() == "o" {
			current := m.getCurrentItem()
			cast := current.cast
			return m, OpenURL(fmt.Sprintf("https://warpcast.com/%s/%s", cast.Author.Username, cast.Hash))
		}
		if msg.String() == "p" {
			current := m.getCurrentItem()
			userFid := current.cast.Author.FID
			return m, func() tea.Msg {
				return SelectProfileMsg{fid: userFid}
			}
		}
		if msg.String() == "c" {
			current := m.getCurrentItem()
			if current.cast.ParentURL == "" {
				return m, nil
			}
			m.Clear()
			return m, tea.Sequence(focusCmd("feed"), getFeedCmd(&api.FeedRequest{FeedType: "filter", FilterType: "parent_url", ParentURL: current.cast.ParentURL, Limit: 100}))
		}
		if msg.String() == "l" {
			current := m.getCurrentItem()
			if current.cast.Hash == "" {
				return m, nil
			}
			return m, likeCastCmd(current.cast)
		}

	case loadTickMsg:
		_, cmd := m.loading.Update(msg)
		return m, cmd

	case *api.FeedResponse:
		m.loading.SetActive(false)
		return m, m.setItems(msg)
	case reactMsg:
		current := m.getCurrentItem()
		if current.cast.Hash != msg.hash {
			return m, m.SetDefaultParams()
		}
		if msg.rtype == "like" && msg.state {
			current.cast.ViewerContext.Liked = true
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		m.table.SetWidth(msg.Width - h)
		m.table.SetHeight(msg.Height - v)
		return m, nil
	}

	newItems := []*CastFeedItem{}
	for _, c := range m.items {
		ni, cmd := c.Update(msg)
		ci, ok := ni.(*CastFeedItem)
		if !ok {
			log.Println("failed to cast to CastFeedItem")
		}
		newItems = append(newItems, ci)

		cmds = append(cmds, cmd)
	}
	m.items = newItems

	// update list/table with updated items
	m.populateItems()

	if m.compact {
		t, cmd := m.table.Update(msg)
		cmds = append(cmds, cmd)
		m.table = t
	} else {
		l, cmd := m.list.Update(msg)
		cmds = append(cmds, cmd)
		m.list = l
	}

	return m, tea.Batch(cmds...)
}

func (m *FeedView) View() string {
	if m.loading.IsActive() {
		return m.loading.View()
	}
	if m.compact {
		return m.table.View()
	}
	return m.list.View()

}
