package main

import (
	"log"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
)

type FocusMsg struct {
	Name string
}

type App struct {
	models       map[string]tea.Model
	focusedModel tea.Model
	focused      string
	height       int
	width        int
}

func NewApp() *App {
	return &App{
		models: make(map[string]tea.Model),
	}
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
	return m.Init()
}

func (a *App) GetFocused() tea.Model {
	return a.focusedModel
}

func (a *App) Init() tea.Cmd {
	log.Println("a.Init()")
	focus := a.GetFocused()
	if focus != nil {
		return focus.Init()
	}
	return nil
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
	log.Println("received msg type: ", reflect.TypeOf(msg))
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case FocusMsg:
		cmd := a.SetFocus(msg.Name)
		if cmd != nil {
			return a, cmd
		}
	case tea.WindowSizeMsg:
		for n, m := range a.models {
			um, cmd := m.Update(msg)
			a.models[n] = um
			cmds = append(cmds, cmd)
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "esc":
			cmd := a.propagateEvent(msg)
			if cmd != nil {
				return a, cmd
			}
			return a, a.SetFocus("methods")
		}
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
	if focus != nil {
		return focus.View()
	}
	return ""
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
