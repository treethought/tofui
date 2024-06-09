package ui

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/tofui/api"
)

var (
	docStyle = NewStyle().Margin(2, 2).Align(lipgloss.Center)
)

type feedLoadedMsg struct{}

type apiErrorMsg struct {
	err error
}

type reactMsg struct {
	hash  string
	rtype string
	state bool
}

type FeedView struct {
	app     *App
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
	s.Header = NewStyle().Bold(true).Padding(0, 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = NewStyle().Bold(true).
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	s.Cell = NewStyle().Padding(0, 1)
	t.SetStyles(s)
	return t
}

func NewFeedView(app *App) *FeedView {
	p := progress.New()
	p.ShowPercentage = false
	return &FeedView{
		app:         app,
		table:       newTable(),
		items:       []*CastFeedItem{},
		loading:     NewLoading(),
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
		cmds = append(cmds, m.SetDefaultParams(), getFeedCmd(m.app.client, m.req))
	} else {
		cmds = append(cmds, getDefaultFeedCmd(m.app.client, m.app.ctx.signer))
	}
	return tea.Sequence(cmds...)
}

func (m *FeedView) Clear() {
	m.loading.SetActive(true)
	m.items = nil
	m.req = nil
	m.table.SetRows([]table.Row{})
	m.setItems(nil)
}

func likeCastCmd(client *api.Client, signer *api.Signer, cast *api.Cast) tea.Cmd {
	return func() tea.Msg {
		log.Println("liking cast", cast.Hash)
		if err := client.React(signer, cast.Hash, "like"); err != nil {
			return apiErrorMsg{err}
		}
		return reactMsg{hash: cast.Hash, rtype: "like", state: true}
	}
}

func getDefaultFeedCmd(client *api.Client, signer *api.Signer) tea.Cmd {
	req := &api.FeedRequest{FeedType: "following", Limit: 100}
	if signer != nil {
		req.FID = signer.FID
		req.ViewerFID = signer.FID
	}
	return getFeedCmd(client, req)
}

func getFeedCmd(client *api.Client, req *api.FeedRequest) tea.Cmd {
	return func() tea.Msg {
		if req.Limit == 0 {
			req.Limit = 100
		}
		feed, err := client.GetFeed(req)
		if err != nil {
			log.Println("feedview error getting feed", err)
			return err
		}
		return feed
	}
}

func (m *FeedView) SetDefaultParams() tea.Cmd {
	var fid uint64
	if m.app.ctx.signer != nil {
		fid = m.app.ctx.signer.FID
	}
	return tea.Sequence(
		m.setItems(nil),
		getFeedCmd(m.app.client, &api.FeedRequest{
			FeedType: "following", Limit: 100,
			FID: fid, ViewerFID: fid,
		}),
	)
}
func (m *FeedView) SetParams(req *api.FeedRequest) tea.Cmd {
	return tea.Sequence(
		m.setItems(nil),
		getFeedCmd(m.app.client, req),
	)
}

func (m *FeedView) setItems(casts []*api.Cast) tea.Cmd {
	rows := []table.Row{}
	cmds := []tea.Cmd{}
	for _, cast := range casts {
		ci, cmd := NewCastFeedItem(m.app, cast, true)
		m.items = append(m.items, ci)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		rows = append(rows, ci.AsRow(m.showChannel, m.showStats))
	}
	m.table.SetRows(rows)
	m.loading.SetActive(false)

	done := func() tea.Msg {
		return &feedLoadedMsg{}
	}
	cmds = append(cmds, done)

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
	if row < 0 || row >= len(m.items) {
		return nil
	}
	return m.items[row]
}
func (m *FeedView) SetSize(w, h int) {
	docStyle = docStyle.MaxWidth(w).MaxHeight(h)
	x, y := docStyle.GetFrameSize()
	m.table.SetWidth(w - x)
	m.table.SetHeight(h - y)
	m.setTableConfig()

	lw := int(float64(w) * 0.2)
	m.loading.SetSize(lw, h)
}

func (m *FeedView) SelectCurrentItem() tea.Cmd {
	current := m.getCurrentItem()
	if current == nil {
		return nil
	}
	return selectCast(current.cast)
}

func (m *FeedView) OpenCurrentItem() tea.Cmd {
	current := m.getCurrentItem()
	if current == nil {
		return nil
	}
	return OpenURL(fmt.Sprintf("https://warpcast.com/%s/%s", current.cast.Author.Username, current.cast.Hash))
}
func (m *FeedView) ViewCurrentProfile() tea.Cmd {
	current := m.getCurrentItem()
	if current == nil {
		return nil
	}
	userFid := current.cast.Author.FID
	return func() tea.Msg {
		return SelectProfileMsg{fid: userFid}
	}
}

func (m *FeedView) ViewCurrentChannel() tea.Cmd {
	current := m.getCurrentItem()
	if current == nil {
		return nil
	}
	if current.cast.ParentURL == "" {
		return nil
	}
	m.Clear()

	cmds := []tea.Cmd{}
	if c, err := m.client.GetChannelByParentUrl(current.cast.ParentURL); err == nil {
		cmds = append(cmds, navNameCmd(fmt.Sprintf("channel: %s", c.Name)))
	}
	cmds = append(cmds,
		focusCmd("feed"),
		getFeedCmd(m.client, &api.FeedRequest{
			FeedType: "filter", FilterType: "parent_url",
			ParentURL: current.cast.ParentURL, Limit: 100},
		),
	)

	return tea.Sequence(cmds...)
}

func (m *FeedView) LikeCurrentItem() tea.Cmd {
	current := m.getCurrentItem()
	if current == nil {
		return nil
	}
	if current.cast.Hash == "" {
		return nil
	}
	return likeCastCmd(m.app.client, m.app.ctx.signer, current.cast)
}

func (m *FeedView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	_, cmd := m.loading.Update(msg)

	cmds = append(cmds, cmd)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if cmd := FeedKeyMap.HandleMsg(m, msg); cmd != nil {
			return m, cmd
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
