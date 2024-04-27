package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/castr/api"
)

const (
	width       = 96
	columnWidth = 30
)

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	titleStyle = lipgloss.NewStyle().
			MarginRight(5).
			Foreground(highlight)

	imgStyle = lipgloss.NewStyle()

	infoStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(subtle)

	contentStyle = lipgloss.NewStyle()

	md, _ = glamour.NewTermRenderer(
		// detect background color and pick either the default dark or light theme
		glamour.WithAutoStyle(),
		// wrap output at specific width (default is 80)
		glamour.WithWordWrap(80),
	)
)

func CastHeader(cast *api.Cast, img ImageModel) string {
	return lipgloss.JoinHorizontal(lipgloss.Center,
		img.View(),
		lipgloss.JoinVertical(lipgloss.Top,
			titleStyle.Render(cast.Author.DisplayName),
			fmt.Sprintf("@%s", cast.Author.Username),
		),
	)
}

func CastContent(cast *api.Cast, maxHeight int, imgs ...ImageModel) string {
	m, err := md.Render(cast.Text)
	if err != nil {
		m = cast.Text
	}
	return contentStyle.MaxHeight(maxHeight).Render(m)
}

type CastFeedItem struct {
	cast    *api.Cast
	pfp     ImageModel
	compact bool
}

// NewCastFeedItem displays a cast in compact form within a list
// implements list.Item (and tea.Model only for updating image)
func NewCastFeedItem(cast *api.Cast, compact bool) (*CastFeedItem, tea.Cmd) {
	c := &CastFeedItem{
		cast:    cast,
		pfp:     NewImage(false, true, special),
		compact: compact,
	}

	cmds := []tea.Cmd{
		c.pfp.SetURL(cast.Author.PfpURL), c.pfp.SetSize(4, 4),
	}
	return c, tea.Batch(cmds...)
}

func (m *CastFeedItem) Init() tea.Cmd { return nil }

func (m *CastFeedItem) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{}

	pfp, cmd := m.pfp.Update(msg)
	m.pfp = pfp
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *CastFeedItem) View() string { return "" }

func (m *CastFeedItem) AsRow() []string {
	return []string{
		m.cast.Hash,
		m.pfp.View(),
		m.cast.Author.DisplayName,
		m.cast.Text,
	}
}

func (i *CastFeedItem) Title() string {
	return CastHeader(i.cast, i.pfp)
}

func (i *CastFeedItem) Description() string {
	return CastContent(i.cast, 3)
}

func (i *CastFeedItem) FilterValue() string {
	return i.cast.Author.Username
}
