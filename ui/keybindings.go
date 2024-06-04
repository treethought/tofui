package ui

import "github.com/charmbracelet/bubbles/key"

type globalKeyMap struct {
	Feed key.Binding
	Up   key.Binding
	Down key.Binding

	Publish     key.Binding
	QuickSelect key.Binding
	Help        key.Binding
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
		{k.Up}, {k.Down},
		{k.Feed}, {k.QuickSelect},
		{k.Publish},
		{k.Help},
	}
}

var DefaultKeyMap = globalKeyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),        // actual keybindings
		key.WithHelp("↑/k", "move up"), // corresponding help text
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("↓/j", "move down"),
	),
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
}
