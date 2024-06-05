package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type globalKeyMap struct {
	Feed key.Binding

	Publish     key.Binding
	QuickSelect key.Binding
	Help        key.Binding
	ToggleBar   key.Binding
	Previous    key.Binding
}

func (k globalKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Feed,
		k.QuickSelect,
		k.Help,
	}
}

func (k globalKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Feed}, {k.QuickSelect},
		{k.Publish},
		{k.Previous},
		{k.Help}, {k.ToggleBar},
	}
}

var GlobalKeyMap = globalKeyMap{
	Feed: key.NewBinding(
		key.WithKeys("F", "1"),
		key.WithHelp("F/1", "feed"),
	),
	Publish: key.NewBinding(
		key.WithKeys("P"),
		key.WithHelp("P", "publish cast"),
	),
	QuickSelect: key.NewBinding(
		key.WithKeys("ctrl+k"),
		key.WithHelp("ctrl+k", "quick select"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	ToggleBar: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "toggle sidebar"),
	),
	Previous: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "focus previous"),
	),
}

func noOp() tea.Cmd {
	return func() tea.Msg {
		return nil
	}
}

func (k globalKeyMap) HandleMsg(a *App, msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, k.Feed):
		a.sidebar.SetActive(false)
		return a.SetFocus("feed")
	case key.Matches(msg, k.Publish):
		a.publish.SetActive(true)
		a.publish.SetFocus(true)
		return noOp()
	case key.Matches(msg, k.QuickSelect):
		a.showQuickSelect = true
		return nil
	case key.Matches(msg, k.Help):
		a.help.SetFull(!a.help.IsFull())
	case key.Matches(msg, k.Previous):
		return a.FocusPrev()
	case key.Matches(msg, k.ToggleBar):
		if a.showQuickSelect {
			_, cmd := a.quickSelect.Update(msg)
			return cmd
		}
		a.sidebar.SetActive(!a.sidebar.Active())
	}

	return nil
}
