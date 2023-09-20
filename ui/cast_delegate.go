package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	divider = lipgloss.NewStyle().
		SetString("•").
		Padding(0, 1).
		Foreground(subtle).
		String()

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
		// BorderForeground(lipgloss.Color("#874BFD")).
		Padding(0, 0).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true)

	boxSelectedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#874BFD"))
)

// custom delegate for cast items
// not in use, but may be used for more complete rendering in feed
type castItemDelegate struct{}

func (d castItemDelegate) Height() int {
	return 4
}

func (d castItemDelegate) Spacing() int {
	return 0
}
func (d castItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d castItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(*CastFeedItem)
	if !ok {
		return
	}

	rf := boxStyle
	if index == m.Index() {
		rf = boxSelectedStyle
	}

	s := rf.Render(lipgloss.JoinVertical(lipgloss.Top,
		i.Title(),
		i.Description(),
	))

	fmt.Fprintln(w, s)

}
