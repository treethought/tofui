package ui

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/tofui/api"
)

type Sidebar struct {
	app     *App
	active  bool
	nav     *list.Model
	account *api.User
	pfp     *ImageModel
	w, h    int
}

type currentAccountMsg struct {
	account *api.User
}

func getCurrentAccount(client *api.Client, signer *api.Signer) tea.Cmd {
	return func() tea.Msg {
		if signer == nil {
			return nil
		}
		user, err := client.GetUserByFID(signer.FID, signer.FID)
		if err != nil {
			log.Println("error getting current account: ", err)
			return nil
		}
		return &currentAccountMsg{account: user}
	}
}

type sidebarItem struct {
	name  string
	value string
	icon  string
	itype string
}

func (m *sidebarItem) FilterValue() string {
	return m.name
}
func (m *sidebarItem) View() string {
	return m.value
}
func (m *sidebarItem) Title() string {
	return m.name
}

func (i *sidebarItem) Description() string {
	return ""
}

var navStyle = NewStyle().Margin(2, 2, 0, 0).BorderRight(true).BorderStyle(lipgloss.RoundedBorder())

func NewSidebar(app *App) *Sidebar {
	d := list.NewDefaultDelegate()
	d.SetHeight(1)
	d.ShowDescription = false

	l := list.New([]list.Item{}, d, 0, 0)
	l.KeyMap.CursorUp.SetKeys("k", "up")
	l.KeyMap.CursorDown.SetKeys("j", "down")
	l.KeyMap.Quit.SetKeys("ctrl+c")
	l.Title = "tofui"
	l.SetShowTitle(true)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)

	pfp := NewImage(true, true, special)
	pfp.SetSize(1, 1)

	return &Sidebar{app: app, nav: &l, pfp: pfp}
}

func (m *Sidebar) SetSize(w, h int) {
	x, y := navStyle.GetFrameSize()
	m.w, m.h = w-x, h-y
	m.nav.SetWidth(m.w)
	m.nav.SetHeight(m.h)
	m.pfp.SetSize(4, 4)
}

func (m *Sidebar) Active() bool {
	return m.active
}
func (m *Sidebar) SetActive(active bool) {
	m.active = active
}

func (m *Sidebar) navHeader() []list.Item {
	items := []list.Item{}
	if api.GetSigner(m.app.ctx.pk) != nil {
		items = append(items, &sidebarItem{name: "profile"})
	}
	items = append(items, &sidebarItem{name: "feed"})
	items = append(items, &sidebarItem{name: "--channels---", value: "--channels--", icon: "🏠"})
	return items
}

func (m *Sidebar) Init() tea.Cmd {
	log.Println("sidebar init")
	var fid uint64
	if m.app.ctx.signer != nil {
		fid = m.app.ctx.signer.FID
	}
	return tea.Batch(
		m.nav.SetItems(m.navHeader()),
		getChannelsCmd(m.app.client, true, fid),
		getCurrentAccount(m.app.client, m.app.ctx.signer),
		m.pfp.Init(),
	)
}

func selectProfileCmd(fid uint64) tea.Cmd {
	return func() tea.Msg {
		return SelectProfileMsg{fid}
	}
}

func (m *Sidebar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *channelListMsg:
		items := m.navHeader()
		for _, c := range msg.channels {
			items = append(items, &sidebarItem{name: c.Name, value: c.ParentURL, itype: "channel"})
		}
		return m, m.nav.SetItems(items)

	case tea.KeyMsg:
		if !m.active {
			return m, nil
		}
		if msg.String() == "enter" {
			currentItem := m.nav.SelectedItem().(*sidebarItem)
			if currentItem.name == "profile" {
				m.SetActive(false)
				log.Println("profile selected")
				fid := api.GetSigner(m.app.ctx.pk).FID
				if fid == 0 {
					return m, nil
				}
				return m, tea.Sequence(
					m.app.FocusProfile(),
					getUserCmd(m.app.client, fid, m.app.ctx.signer.FID),
					getUserFeedCmd(m.app.client, fid, m.app.ctx.signer.FID),
				)
			}
			if currentItem.name == "feed" {
				m.SetActive(false)
				log.Println("feed selected")
				return m, tea.Sequence(m.app.FocusFeed(), getDefaultFeedCmd(m.app.client, m.app.ctx.signer))
			}
			if currentItem.itype == "channel" {
				m.SetActive(false)
				m.app.SetNavName(fmt.Sprintf("channel: %s", currentItem.name))
				return m, tea.Sequence(
					m.app.FocusChannel(),
					getFeedCmd(m.app.client,
						&api.FeedRequest{FeedType: "filter", FilterType: "parent_url", ParentURL: currentItem.value, Limit: 100}),
				)
			}
		}
	case *currentAccountMsg:
		m.account = msg.account
		return m, m.pfp.SetURL(m.account.PfpURL, false)

	}

	//update list size if window size changes
	cmds := []tea.Cmd{}
	l, ncmd := m.nav.Update(msg)
	m.nav = &l
	cmds = append(cmds, ncmd)

	pfp, pcmd := m.pfp.Update(msg)
	m.pfp = pfp
	cmds = append(cmds, pcmd)
	return m, tea.Batch(cmds...)
}
func (m *Sidebar) View() string {
	if m.account == nil {
		return navStyle.Render(m.nav.View())
	}

	accountStyle := NewStyle().
		Border(lipgloss.RoundedBorder(), true, false, true).
		Width(m.w).
		MaxWidth(m.w).
		Align(lipgloss.Center, lipgloss.Center).Margin(0).Padding(0)

	account := accountStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Center,
			m.pfp.View(),
			lipgloss.JoinVertical(lipgloss.Left,
				m.account.DisplayName,
				fmt.Sprintf("@%s", m.account.Username),
			),
		),
	)

	return navStyle.Render(
		lipgloss.JoinVertical(lipgloss.Top,
			m.nav.View(),
			account,
		),
	)
}
