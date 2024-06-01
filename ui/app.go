package ui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/castr/api"
)

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
	height          int
	width           int
	sidebar         *Sidebar
	hideSidebar     bool
	sidebarActive   bool
	prev            string
	QuickSelect     *QuickSelect
	showQuickSelect bool
	publish         *PublishInput
	showPublish     bool
}

func NewApp() *App {
	a := &App{
		models: make(map[string]tea.Model),
	}
	a.sidebar = NewSidebar(a)
	a.QuickSelect = NewQuickSelect(a)
	a.publish = NewPublishInput(a)
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
	log.Println("registered", name)
}

func (a *App) SetFocus(name string) tea.Cmd {
	if a.showQuickSelect {
		a.showQuickSelect = false
	}
	if a.showPublish {
		a.showPublish = false
		a.publish.SetFocus(false)
	}
	if name == "" || name == a.focused {
		return nil
	}
	a.prev = a.focused
	m, ok := a.models[name]
	if !ok {
		log.Println("model not found: ", name)
	}
	a.focusedModel = m
	a.focused = name
	if a.sidebarActive {
		a.sidebarActive = false
	}
	return m.Init()
}

func (a *App) GetFocused() tea.Model {
	return a.focusedModel
}

func (a *App) FocusPrev() tea.Cmd {
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
	log.Println("a.Init()")
	cmds := []tea.Cmd{}
	cmds = append(cmds, a.sidebar.Init(), a.QuickSelect.Init(), a.publish.Init())
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
	switch msg := msg.(type) {
	case FocusMsg:
		cmd := a.SetFocus(msg.Name)
		if cmd != nil {
			return a, cmd
		}
	case *channelListMsg:
		if msg.activeOnly {
			_, cmd := a.sidebar.Update(msg.channels)
			return a, cmd
		} else {
			_, cmd := a.QuickSelect.Update(msg.channels)
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
		return a, tea.Batch(feed.setItems(msg), focusCmd)
	case SelectProfileMsg:
		focusCmd := a.SetFocus("profile")
		cmd := a.GetModel("profile").(*Profile).SetFID(msg.fid)
		return a, tea.Batch(focusCmd, cmd)
	case SelectCastMsg:
		focusCmd := a.SetFocus("cast")
		cmd := a.GetModel("cast").(*CastView).SetCast(msg.cast)
		return a, tea.Batch(cmd, focusCmd)

	case tea.WindowSizeMsg:
		SetHeight(msg.Height)
		SetWidth(msg.Width)
		// substract the sidebar width from the window width
		wx, wy := msg.Width, msg.Height

		sw := wx / 6
		a.sidebar.nav.SetSize(sw, wy)

		qw := wx - sw
		qh := wy / 3
		a.QuickSelect.SetSize(qw, qh)

		pw := wx - sw
		py := wy / 3
		a.publish.SetSize(pw, py)

		childMsg := tea.WindowSizeMsg{
			Width:  wx - pw - 1,
			Height: wy - py - 1,
		}

		for n, m := range a.models {
			um, cmd := m.Update(childMsg)
			a.models[n] = um
			cmds = append(cmds, cmd)
		}
	case tea.KeyMsg:
		if a.showPublish {
			_, cmd := a.publish.Update(msg)
			return a, cmd
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "tab":
			a.sidebarActive = !a.sidebarActive
		case "esc":
			return a, a.FocusPrev()
		case "ctrl+k":
			a.showQuickSelect = true
		case "P":
			a.showPublish = true
			a.publish.SetFocus(true)
			return a, nil
		}
	}
	if a.showPublish {
		_, cmd := a.publish.Update(msg)
		return a, cmd
	}
	if a.showQuickSelect {
		q, cmd := a.QuickSelect.Update(msg)
		a.QuickSelect = q.(*QuickSelect)
		return a, cmd
	}
	if a.sidebarActive {
		_, cmd := a.sidebar.Update(msg)
		return a, cmd
	}

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
	if a.showPublish {
		return a.publish.View()
	}
	if a.showQuickSelect {
		return a.QuickSelect.View()
	}
	if a.hideSidebar {
		return focus.View()
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, a.sidebar.View(), focus.View())
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
