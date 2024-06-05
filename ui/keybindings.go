package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type feedKeymap struct {
	ViewCast    key.Binding
	LikeCast    key.Binding
	ViewProfile key.Binding
	ViewChannel key.Binding
	OpenCast    key.Binding
}

func (k feedKeymap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.LikeCast,
		k.ViewProfile,
		k.ViewChannel,
	}
}
func (k feedKeymap) All() []key.Binding {
	return []key.Binding{
		k.ViewCast,
		k.LikeCast,
		k.ViewProfile,
		k.ViewChannel,
		k.OpenCast,
	}
}

func (k feedKeymap) HandleMsg(f *FeedView, msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, k.ViewCast):
		log.Println("ViewCast")
		return f.SelectCurrentItem()
	case key.Matches(msg, k.LikeCast):
		log.Println("LikeCast")
		return f.LikeCurrentItem()
	case key.Matches(msg, k.ViewProfile):
		log.Println("ViewProfile")
		return f.ViewCurrentProfile()
	case key.Matches(msg, k.ViewChannel):
		log.Println("ViewChannel")
		return f.ViewCurrentChannel()
	case key.Matches(msg, k.OpenCast):
		log.Println("OpenCast")
		return f.OpenCurrentItem()
	}
	return nil
}

var FeedKeyMap = feedKeymap{
	ViewCast: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "view cast"),
	),
	LikeCast: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "like cast"),
	),
	ViewProfile: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "view profile"),
	),
	ViewChannel: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "view channel"),
	),
	OpenCast: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "open in browser "),
	),
}

type navKeymap struct {
	Feed key.Binding

	Publish     key.Binding
	QuickSelect key.Binding
	Help        key.Binding
	ToggleBar   key.Binding
	Previous    key.Binding
}

func (k navKeymap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Feed,
		k.QuickSelect,
		k.Help,
	}
}

func (k navKeymap) All() []key.Binding {
	return []key.Binding{
		k.Feed, k.QuickSelect,
		k.Publish,
		k.Previous,
		k.Help, k.ToggleBar,
	}
}

var NavKeyMap = navKeymap{
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

func (k navKeymap) HandleMsg(a *App, msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, k.Feed):
		// TODO cleanup
		// reset params for user's feed
		var cmd tea.Cmd
		if f, ok := a.GetModel("feed").(*FeedView); ok {
			f.Clear()
			cmd = f.SetDefaultParams()
		}
		a.SetNavName("feed")
		a.sidebar.SetActive(false)
		return tea.Sequence(cmd, a.SetFocus("feed"))

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

type kmap struct {
	nav  navKeymap
	feed feedKeymap
}

var GlobalKeyMap = kmap{
	nav:  NavKeyMap,
	feed: FeedKeyMap,
}

func (k kmap) ShortHelp() []key.Binding {
	return append(k.nav.ShortHelp(), k.feed.ShortHelp()...)
}

func (k kmap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.nav.All(),
		k.feed.All(),
	}
}

func noOp() tea.Cmd {
	return func() tea.Msg {
		return nil
	}
}
