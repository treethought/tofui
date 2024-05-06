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
	app      *App
	active   bool
	nav      *list.Model
	channels *list.Model
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

func NewSidebar(app *App) *Sidebar {
	d := list.NewDefaultDelegate()
	d.SetHeight(2)
	d.ShowDescription = false
	// d.Styles.NormalTitle.Border(lipgloss.NormalBorder(), true, true, true, true).Margin(0).Padding(0).AlignHorizontal(lipgloss.Left)
	// d.Styles.SelectedTitle.Border(lipgloss.NormalBorder(), true, true, true, true).Margin(0).Padding(0).AlignHorizontal(lipgloss.Left)
	// d.Styles.NormalTitle.MaxHeight(2)
	// d.Styles.SelectedTitle.MaxHeight(2)
	// d.Styles.NormalTitle.Border(lipgloss.NormalBorder(), true, true, true, true)

	// d.Styles.DimmedDesc.MaxHeight(10)
	// d.Styles.NormalDesc.MaxHeight(10)
	// d.Styles.DimmedDesc.Width(80)
	// d.Styles.NormalDesc.Border(lipgloss.NormalBorder(), true, true, true, true).Margin(0).Padding(0).AlignHorizontal(lipgloss.Left)

	l := list.New([]list.Item{}, d, 0, 0)
	l.KeyMap.CursorUp.SetKeys("k", "up")
	l.KeyMap.CursorDown.SetKeys("j", "down")
	l.KeyMap.Quit.SetKeys("ctrl+c")
	l.Title = "castr"
	l.SetShowTitle(true)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)

	return &Sidebar{app: app, nav: &l}
}

func getUserChannelsCmd() tea.Cmd {
	return func() tea.Msg {
		fid := api.GetSigner().FID
		channels, err := api.GetClient().GetUserChannels(fid, true, api.WithLimit(10))
		if err != nil {
			log.Println("error getting user channels: ", err)
			return nil
		}
		return channels
	}
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
	log.Println("sidebar set items")
	return tea.Batch(m.nav.SetItems(m.navHeader()), getUserChannelsCmd())
}

func selectProfileCmd(fid uint64) tea.Cmd {
	return func() tea.Msg {
		return SelectProfileMsg{fid}
	}
}

func (m *Sidebar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case []*api.Channel:
		items := m.navHeader()
		for _, c := range msg {
			items = append(items, &sidebarItem{name: c.Name, value: c.ParentURL, itype: "channel"})
		}
		return m, m.nav.SetItems(items)

	case tea.KeyMsg:
		if msg.String() == "enter" {
			currentItem := m.nav.SelectedItem().(*sidebarItem)
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
				return m, tea.Sequence(m.app.SetFocus("feed"), getFeedCmd(&api.FeedRequest{FeedType: "filter", FilterType: "parent_url", ParentURL: currentItem.value, Limit: 100}))
			}
		}
	}

	//update list size if window size changes
	l, cmd := m.nav.Update(msg)
	m.nav = &l
	return m, cmd
}
func (m *Sidebar) View() string {
	return lipgloss.NewStyle().BorderRight(true).BorderStyle(lipgloss.DoubleBorder()).Render(m.nav.View())
	// return m.list.View()
}
