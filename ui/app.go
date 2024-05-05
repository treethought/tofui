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

type SelectCastMsg struct {
	cast *api.Cast
}

type App struct {
	models        map[string]tea.Model
	focusedModel  tea.Model
	focused       string
	height        int
	width         int
	sidebar       *Sidebar
	hideSidebar   bool
	sidebarActive bool
}

func NewApp() *App {
	a := &App{
		models: make(map[string]tea.Model),
	}
	a.sidebar = NewSidebar(a)
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
	if name == "" || name == a.focused {
		return nil
	}
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

func (a *App) Init() tea.Cmd {
	log.Println("a.Init()")
	cmds := []tea.Cmd{}
	cmds = append(cmds, a.sidebar.Init())
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
	case []*api.Channel:
		_, cmd := a.sidebar.Update(msg)
		return a, cmd
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

		sw := int(float64(wx) * 0.2)
		a.sidebar.list.SetSize(sw, wy)

		childMsg := tea.WindowSizeMsg{
			Width:  wx - sw,
			Height: wy,
		}

		for n, m := range a.models {
			um, cmd := m.Update(childMsg)
			a.models[n] = um
			cmds = append(cmds, cmd)
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "tab":
			a.sidebarActive = !a.sidebarActive
		case "esc":
			cmd := a.propagateEvent(msg)
			if cmd != nil {
				return a, cmd
			}
			return a, a.SetFocus("feed")
		}
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
