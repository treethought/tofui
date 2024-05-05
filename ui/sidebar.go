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
	app    *App
	active bool
	list   list.Model
}

type sidebarItem struct {
	name  string
	value string
	icon  string
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

	return &Sidebar{app: app, list: l}
}

func (m *Sidebar) Init() tea.Cmd {
	log.Println("sidebar init")
	items := []list.Item{}
	items = append(items, &sidebarItem{name: "profile", value: fmt.Sprintf("%d", api.GetSigner().FID)})
	items = append(items, &sidebarItem{name: "Home", value: "Home", icon: "🏠"})

	log.Println("sidebar set items")
	return m.list.SetItems(items)
}

func (m *Sidebar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			currentItem := m.list.SelectedItem().(*sidebarItem)
			if currentItem.name == "profile" {
				log.Println("profile selected")
				fid := api.GetSigner().FID
				if fid == 0 {
					return m, nil
				}
				return m, func() tea.Msg {
					return SelectProfileMsg{fid}
				}
			}
		}
	}

	//update list size if window size changes
	l, cmd := m.list.Update(msg)
	m.list = l
	return m, cmd
}
func (m *Sidebar) View() string {
	return lipgloss.NewStyle().BorderRight(true).BorderStyle(lipgloss.DoubleBorder()).Render(m.list.View())
	// return m.list.View()
}
