package ui

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/tofui/api"
)

var (
	feedStyle          = NewStyle().Margin(2, 2).Align(lipgloss.Center)
	channelHeaderStyle = NewStyle().Margin(1, 1).Align(lipgloss.Top).Border(lipgloss.RoundedBorder())
)

type feedType string

var (
	feedTypeFollowing feedType = "following"
	feedTypeChannel   feedType = "channel"
	feedTypeProfile   feedType = "profile"
	feedTypeReplies   feedType = "replies"
)

type feedLoadedMsg struct{}

type apiErrorMsg struct {
	err error
}

type fetchChannelMsg struct {
	parentURL string
	channel   *api.Channel
	err       error
}

type channelFeedMsg struct {
	casts []*api.Cast
	err   error
}

type reactMsg struct {
	hash  string
	rtype string
	state bool
}

type FeedView struct {
	app     *App
	table   table.Model
	items   []*CastFeedItem
	loading *Loading
	req     *api.FeedRequest

	showChannel bool
	showStats   bool
	description string
	descVp      *viewport.Model
	headerImg   *ImageModel
	feedType    feedType
	w, h        int
}

func getTableStyles() table.Styles {
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
	return s

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
	s := getTableStyles()
	t.SetStyles(s)
	return t
}

func NewFeedView(app *App, ft feedType) *FeedView {
	p := progress.New()
	p.ShowPercentage = false
	dvp := viewport.New(0, 0)

	return &FeedView{
		app:         app,
		table:       newTable(),
		items:       []*CastFeedItem{},
		loading:     NewLoading(),
		showChannel: true,
		showStats:   true,
		descVp:      &dvp,
		headerImg:   NewImage(true, true, special),
		feedType:    ft,
	}
}

func (m *FeedView) SetDescription(desc string) {
	if m.feedType != feedTypeChannel {
		log.Println("not setting description: type: ", m.feedType)
		return
	}
	m.description = desc
	m.descVp.SetContent(desc)
	m.SetSize(m.w, m.h)
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
	fx, _ := feedStyle.GetFrameSize()
	w := m.table.Width() - fx //- 10

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
		{Title: "cast", Width: int(float64(w)*0.5) - 4},
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
	} else if m.feedType == feedTypeFollowing {
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
	if signer == nil {
		return nil
	}
	req := &api.FeedRequest{Limit: 100}
	req.FeedType = "following"
	req.FID = signer.FID
	req.ViewerFID = signer.FID
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

func getChannelFeedCmd(client *api.Client, pu string) tea.Cmd {
	return func() tea.Msg {
		log.Println("getting channel feed")
		req := &api.FeedRequest{
			FeedType: "filter", FilterType: "parent_url",
			ParentURL: pu, Limit: 100,
		}
		feed, err := client.GetFeed(req)
		return &channelFeedMsg{feed.Casts, err}
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

func (m *FeedView) hideDescription() {
	m.descVp.Width = 0
	m.descVp.Height = 0
	m.headerImg.SetSize(0, 0)

}

func (m *FeedView) SetSize(w, h int) {
	m.w, m.h = w, h

	m.hideDescription()
	if m.description != "" {
		m.descVp.SetContent(m.description)
		dmin := 8
		dPct := int(float64(h) * 0.2)
		dy := dPct
		if dmin > dPct {
			dy = dmin
		}
		m.headerImg.SetSize(4, 4)
		fx, fy := channelHeaderStyle.GetFrameSize()
		m.descVp.Width = w - fx - 4
		m.descVp.Height = dy - fy
	}

	_, dy := lipgloss.Size(channelHeaderStyle.Render(m.descVp.View()))
	fx, fy := feedStyle.GetFrameSize()
	x := min(w-fx, int(float64(GetWidth())*0.75))
	m.table.SetWidth(x)
	m.table.SetHeight(h - fy - dy)

	// m.table.SetWidth(w -fx)
	m.setTableConfig()

	lw := int(float64(w) * 0.75)
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
	m.loading.SetActive(true)
	return tea.Sequence(
		m.app.FocusProfile(),
		getUserCmd(m.app.client, userFid, m.app.ctx.signer.FID),
		getUserFeedCmd(m.app.client, userFid, m.app.ctx.signer.FID),
	)
}

func fetchChannelCmd(client *api.Client, pu string) tea.Cmd {
	return func() tea.Msg {
		log.Println("fetching channel obj")
		c, err := client.GetChannelByParentUrl(pu)
		return &fetchChannelMsg{pu, c, err}
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
	if m.feedType == feedTypeChannel {
		log.Println("already viewing channel")
		return nil
	}
	m.loading.SetActive(true)
	return tea.Batch(
		getChannelFeedCmd(m.app.client, current.cast.ParentURL),
		fetchChannelCmd(m.app.client, current.cast.ParentURL),
		m.app.FocusChannel(),
	)
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
	case *channelFeedMsg:
		if msg.err != nil {
			log.Println("channel feed error", msg.err)
			return m, nil
		}
		m.Clear()
		return m, tea.Batch(m.setItems(msg.casts))
	case *profileFeedMsg:
		if m.feedType != feedTypeProfile {
			return m, nil
		}
		m.Clear()
		return m, m.setItems(msg.casts)
	case *fetchChannelMsg:
		if msg.err != nil {
			return m, nil
		}
		m.SetDescription(channelDescription(msg.channel, m.headerImg))
		return m, m.headerImg.SetURL(msg.channel.ImageURL, false)

	case reactMsg:
		current := m.getCurrentItem()
		if current == nil || current.cast == nil {
			return m, nil
		}

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

	_, icmd := m.headerImg.Update(msg)
	cmds = append(cmds, icmd)

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

	_, lcmd := m.loading.Update(msg)
	cmds = append(cmds, lcmd)

	t, cmd := m.table.Update(msg)
	cmds = append(cmds, cmd)
	m.table = t
	return m, tea.Batch(cmds...)
}

func (m *FeedView) View() string {
	if m.loading.IsActive() {
		return m.loading.View()
	}
	if m.feedType == feedTypeChannel {
		return lipgloss.JoinVertical(lipgloss.Top,
			channelHeaderStyle.Render(m.descVp.View()),
			feedStyle.Render(m.table.View()),
		)
	}

	return feedStyle.Render(m.table.View())

}

func channelStats(c *api.Channel, margin int) string {
	if c == nil {
		return spinner.New().View()
	}
	stats := lipgloss.JoinHorizontal(lipgloss.Top,
		NewStyle().Render(fmt.Sprintf("/%s ", c.ID)),
		NewStyle().MarginRight(margin).Render(EmojiPerson),
		NewStyle().Render(fmt.Sprintf("%d ", c.FollowerCount)),
		NewStyle().MarginRight(margin).Render("followers"),
		// NewStyle().Render(fmt.Sprintf("%d ", c.Object
		// NewStyle().MarginRight(margin).Render(EmojiRecyle),
	)
	return stats
}

func channelHeader(c *api.Channel, img *ImageModel) string {
	return headerStyle.Render(lipgloss.JoinHorizontal(lipgloss.Center,
		img.View(),
		lipgloss.JoinVertical(lipgloss.Top,
			displayNameStyle.Render(c.Name),
			c.Description,
		),
	),
	)
}

func channelDescription(c *api.Channel, img *ImageModel) string {
	return lipgloss.JoinVertical(lipgloss.Bottom,
		channelHeader(c, img),
		NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).Padding(0).Render(channelStats(c, 1)),
	)
}
