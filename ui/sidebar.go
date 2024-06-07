package ui

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/castr/api"
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

func getCurrentAccount() tea.Cmd {
	return func() tea.Msg {
		signer := api.GetSigner()
		if signer == nil {
			return nil
		}
		user, err := api.GetClient().GetUserByFID(signer.FID)
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

var navStyle = lipgloss.NewStyle().Margin(2, 2, 0, 2).BorderRight(true).BorderStyle(lipgloss.RoundedBorder())

func NewSidebar(app *App) *Sidebar {
	d := list.NewDefaultDelegate()
	d.SetHeight(1)
	d.ShowDescription = false

	l := list.New([]list.Item{}, d, 0, 0)
	l.KeyMap.CursorUp.SetKeys("k", "up")
	l.KeyMap.CursorDown.SetKeys("j", "down")
	l.KeyMap.Quit.SetKeys("ctrl+c")
	l.Title = "castr"
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
	m.nav.SetWidth(w - x)
	m.nav.SetHeight(h - y - 4)
	m.pfp.SetSize(4, 4)
	m.w, m.h = w, h
}

func (m *Sidebar) Active() bool {
	return m.active
}
func (m *Sidebar) SetActive(active bool) {
	m.active = active
}

func (m *Sidebar) navHeader() []list.Item {
	items := []list.Item{}
	items = append(items, &sidebarItem{name: "profile", value: fmt.Sprintf("%d", api.GetSigner().FID)})
	items = append(items, &sidebarItem{name: "feed", value: fmt.Sprintf("%d", api.GetSigner().FID)})
	items = append(items, &sidebarItem{name: "--channels---", value: "--channels--", icon: "🏠"})
	return items
}

func (m *Sidebar) Init() tea.Cmd {
	log.Println("sidebar init")
	return tea.Batch(
		m.nav.SetItems(m.navHeader()),
		getChannelsCmd(true),
		getCurrentAccount(),
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
				fid := api.GetSigner().FID
				if fid == 0 {
					return m, nil
				}
				return m, tea.Sequence(m.app.SetFocus("profile"), selectProfileCmd(fid))
			}
			if currentItem.name == "feed" {
				m.SetActive(false)
				log.Println("feed selected")
				return m, tea.Sequence(m.app.SetFocus("feed"), getFeedCmd(DefaultFeedParams()))
			}
			if currentItem.itype == "channel" {
				m.SetActive(false)
				m.app.SetNavName(fmt.Sprintf("channel: %s", currentItem.name))
				return m, tea.Sequence(m.app.SetFocus("feed"), getFeedCmd(&api.FeedRequest{FeedType: "filter", FilterType: "parent_url", ParentURL: currentItem.value, Limit: 100}))
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

	accountStyle := lipgloss.NewStyle().
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
