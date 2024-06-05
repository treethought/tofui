package ui

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/castr/api"
)

type navNameMsg struct {
	name string
}

func navNameCmd(name string) tea.Cmd {
	return func() tea.Msg {
		return navNameMsg{name: name}
	}
}

type FocusMsg struct {
	Name string
}

func focusCmd(name string) tea.Cmd {
	return func() tea.Msg {
		return FocusMsg{Name: name}
	}
}

type SelectCastMsg struct {
	cast *api.Cast
}

type App struct {
	models          map[string]tea.Model
	focusedModel    tea.Model
	focused         string
	navname         string
	height          int
	width           int
	sidebar         *Sidebar
	hideSidebar     bool
	prev            string
	quickSelect     *QuickSelect
	showQuickSelect bool
	publish         *PublishInput
	statusLine      *StatusLine
	help            *HelpView
}

func NewApp() *App {
	a := &App{
		models: make(map[string]tea.Model),
	}
	a.sidebar = NewSidebar(a)
	a.quickSelect = NewQuickSelect(a)
	a.publish = NewPublishInput(a)
	a.statusLine = NewStatusLine(a)
	a.help = NewHelpView()
	return a
}

func (a *App) Height() int {
	return a.height
}
func (a *App) Width() int {
	return a.width
}

func (a *App) GetModel(name string) tea.Model {
	m, ok := a.models[name]
	if !ok {
		log.Println("model not found: ", name)
	}
	return m
}

func (a *App) Register(name string, model tea.Model) {
	a.models[name] = model
}

func (a *App) SetNavName(name string) {
	a.navname = name
}

func (a *App) SetFocus(name string) tea.Cmd {
	if a.showQuickSelect {
		a.showQuickSelect = false
	}
	if a.publish.Active() {
		a.publish.SetActive(false)
		a.publish.SetFocus(false)
	}
	if name == "" || name == a.focused {
		return nil
	}
	// clear if we're back at feed
	a.prev = ""
	if name != "feed" {
		a.prev = a.focused
	}
	m, ok := a.models[name]
	if !ok {
		log.Println("model not found: ", name)
	}
	a.focusedModel = m
	a.focused = name
	a.sidebar.SetActive(false)
	return m.Init()
}

func (a *App) GetFocused() tea.Model {
	return a.focusedModel
}

func (a *App) FocusPrev() tea.Cmd {
	if a.help.IsFull() {
		a.help.SetFull(false)
	}
	prev := a.GetModel(a.prev)
	if a.prev == "" || prev == nil {
		return a.SetFocus("feed")
	}
	if m := a.GetModel(a.prev); m != nil {
		return a.SetFocus(a.prev)
	}
	return a.SetFocus("feed")
}

func (a *App) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	cmds = append(cmds, a.sidebar.Init(), a.quickSelect.Init(), a.publish.Init())
	focus := a.GetFocused()
	if focus != nil {
		cmds = append(cmds, focus.Init())
	}
	return tea.Batch(cmds...)
}

func (a *App) propagateEvent(msg tea.Msg) tea.Cmd {
	for name, m := range a.models {
		if name == a.focused {
			um, cmd := m.Update(msg)
			a.models[name] = um
			return cmd
		}
	}
	return nil
}
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// log.Println("received msg type: ", reflect.TypeOf(msg))
	var cmds []tea.Cmd
	_, sbcmd := a.statusLine.Update(msg)
	cmds = append(cmds, sbcmd)
	switch msg := msg.(type) {
	case navNameMsg:
		a.SetNavName(msg.name)
		return a, nil
	case FocusMsg:
		cmd := a.SetFocus(msg.Name)
		if cmd != nil {
			return a, cmd
		}
	case *channelListMsg:
		if msg.activeOnly {
			_, cmd := a.sidebar.Update(msg)
			return a, cmd
		} else {
			_, cmd := a.quickSelect.Update(msg.channels)
			return a, cmd
		}
	case *api.FeedResponse:
		// allow msg to pass through to profile's embedded feed
		if a.focused == "profile" {
			_, cmd := a.GetFocused().Update(msg)
			return a, cmd
		}
		feed := a.GetModel("feed").(*FeedView)
		feed.Clear()
		focusCmd := a.SetFocus("feed")
		return a, tea.Batch(feed.setItems(msg.Casts), focusCmd)
	case SelectProfileMsg:
		a.SetNavName("profile")
		focusCmd := a.SetFocus("profile")
		cmd := a.GetModel("profile").(*Profile).SetFID(msg.fid)
		return a, tea.Batch(focusCmd, cmd)
	case SelectCastMsg:
		nav := fmt.Sprintf("cast by @%s", msg.cast.Author.Username)
		if msg.cast.ParentHash != "" {
			nav = fmt.Sprintf("reply by @%s", msg.cast.Author.Username)
		}

		a.SetNavName(nav)
		focusCmd := a.SetFocus("cast")
		cmd := a.GetModel("cast").(*CastView).SetCast(msg.cast)
		return a, tea.Sequence(cmd, focusCmd)

	case tea.WindowSizeMsg:
		SetHeight(msg.Height)
		SetWidth(msg.Width)

		a.statusLine.SetSize(msg.Width, msg.Height)

		// substract the sidebar width from the window width
		wx, wy := msg.Width, msg.Height-1
		a.width = wx
		a.height = wy

		sw := int(float64(wx) * 0.2)
		a.sidebar.SetSize(sw, wy)

		qw := wx - sw
		qh := wy - 10
		a.quickSelect.SetSize(qw, qh)

		pw := wx - sw
		py := wy - 10
		a.publish.SetSize(pw, py)

		hw := wx - sw
		hy := wy - 10
		a.help.SetSize(hw, hy)

		childMsg := tea.WindowSizeMsg{
			Width:  int(float64(wx) * 0.8),
			Height: wy,
		}

		for n, m := range a.models {
			um, cmd := m.Update(childMsg)
			a.models[n] = um
			cmds = append(cmds, cmd)
		}
	case tea.KeyMsg:
		if a.publish.Active() {
			_, cmd := a.publish.Update(msg)
			return a, cmd
		}

		cmd := NavKeyMap.HandleMsg(a, msg)
		if cmd != nil {
			return a, cmd
		}

		if a.sidebar.Active() {
			_, cmd := a.sidebar.Update(msg)
			return a, cmd
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		}
	case *currentAccountMsg:
		_, cmd := a.sidebar.Update(msg)
		return a, cmd
	}
	if a.publish.Active() {
		_, cmd := a.publish.Update(msg)
		return a, cmd
	}
	if a.showQuickSelect {
		q, cmd := a.quickSelect.Update(msg)
		a.quickSelect = q.(*QuickSelect)
		return a, cmd
	}

	if a.help.IsFull() {
		_, cmd := a.help.Update(msg)
		return a, cmd
	}

	_, scmd := a.sidebar.Update(msg)
	cmds = append(cmds, scmd)

	current := a.GetFocused()
	if current == nil {
		log.Println("no focused model")
		return Fallback, nil
	}

	_, cmd := current.Update(msg)
	cmds = append(cmds, cmd)
	return a, tea.Batch(cmds...)

}

func (a *App) View() string {
	focus := a.GetFocused()
	if focus == nil {
		return "no focused model"
	}
	main := focus.View()
	side := a.sidebar.View()

	if a.publish.Active() {
		main = a.publish.View()
	}
	if a.showQuickSelect {
		main = a.quickSelect.View()
	}
	if a.hideSidebar {
		return focus.View()
	}
	if a.help.IsFull() {
		main = a.help.View()
	}
	return lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.JoinHorizontal(lipgloss.Center, side, main),
		a.statusLine.View(),
	)

}

func UpdateChildren(msg tea.Msg, models ...tea.Model) tea.Cmd {
	cmds := make([]tea.Cmd, len(models))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range models {
		models[i], cmds[i] = models[i].Update(msg)
	}

	return tea.Batch(cmds...)
}
