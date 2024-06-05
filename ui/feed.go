package ui

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/key"
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
	client  *api.Client
	table   table.Model
	items   []*CastFeedItem
	loading *Loading
	req     *api.FeedRequest

	showChannel bool
	showStats   bool
}

func newTable() table.Model {

	t := table.New(
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

func DefaultFeedParams() *api.FeedRequest {
	return &api.FeedRequest{FeedType: "following", Limit: 100}
}

func NewFeedView(client *api.Client, req *api.FeedRequest) *FeedView {
	p := progress.New()
	p.ShowPercentage = false
	return &FeedView{
		client:      client,
		table:       newTable(),
		items:       []*CastFeedItem{},
		loading:     NewLoading(),
		req:         req,
		showChannel: true,
		showStats:   true,
	}
}

func (m *FeedView) SetShowChannel(show bool) {
	m.showChannel = show
	m.setTableConfig()
}
func (m *FeedView) SetShowStats(show bool) {
	m.showStats = show
	m.setTableConfig()
}

func (m *FeedView) setTableConfig() {
	fw, _ := docStyle.GetFrameSize()
	w := m.table.Width() - fw

	if !m.showChannel && !m.showStats {
		m.table.SetColumns([]table.Column{
			{Title: "user", Width: int(float64(w) * 0.2)},
			{Title: "cast", Width: int(float64(w) * 0.8)},
		})
		return
	}
	m.table.SetColumns([]table.Column{
		{Title: "channel", Width: int(float64(w) * 0.2)},
		{Title: "", Width: int(float64(w) * 0.1)},
		{Title: "user", Width: int(float64(w) * 0.2)},
		{Title: "cast", Width: int(float64(w) * 0.5)},
	})
	return

}

func (m *FeedView) Init() tea.Cmd {
	m.setTableConfig()
	cmds := []tea.Cmd{}
	if len(m.items) == 0 {
		m.loading.SetActive(true)
		cmds = append(cmds, m.loading.Init())
	}

	if m.req != nil {
		cmds = append(cmds, getFeedCmd(m.req))
	}
	return tea.Batch(cmds...)
}

func (m *FeedView) Clear() {
	m.items = nil
	m.req = nil
	m.table.SetRows([]table.Row{})
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

func (m *FeedView) setItems(casts []*api.Cast) tea.Cmd {
	rows := []table.Row{}
	cmds := []tea.Cmd{}
	for _, cast := range casts {
		ci, cmd := NewCastFeedItem(cast, true)
		m.items = append(m.items, ci)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		rows = append(rows, ci.AsRow(m.showChannel, m.showStats))
	}
	m.table.SetRows(rows)
	m.loading.SetActive(false)
	return tea.Batch(cmds...)
}

func (m *FeedView) populateItems() tea.Cmd {
	rows := []table.Row{}
	for _, i := range m.items {
		rows = append(rows, i.AsRow(m.showChannel, m.showStats))
	}
	if len(rows) > 0 {
		m.loading.SetActive(false)
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
	row := m.table.Cursor()
	if row >= len(m.items) {
		return nil
	}
	return m.items[row]
}
func (m *FeedView) SetSize(w, h int) {
	docStyle = docStyle.MaxWidth(w)
	x, y := docStyle.GetFrameSize()
	m.table.SetWidth(w - x)
	m.table.SetHeight(h - y)
	m.setTableConfig()

	lw := int(float64(w) * 0.2)
	m.loading.SetSize(lw, h)
}

func (m *FeedView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	_, cmd := m.loading.Update(msg)

	cmds = append(cmds, cmd)
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
		return m, m.setItems(msg.Casts)
	case reactMsg:
		current := m.getCurrentItem()
		if current.cast.Hash != msg.hash {
			return m, m.SetDefaultParams()
		}
		if msg.rtype == "like" && msg.state {
			current.cast.ViewerContext.Liked = true
		}

	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
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

	// update table with updated items
	m.populateItems()

	t, cmd := m.table.Update(msg)
	cmds = append(cmds, cmd)
	m.table = t
	return m, tea.Batch(cmds...)
}

func (m *FeedView) View() string {
	if m.loading.IsActive() {
		return m.loading.View()
	}
	return docStyle.Render(m.table.View())

}
