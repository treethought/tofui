package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
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

	divider = lipgloss.NewStyle().
		SetString("•").
		Padding(0, 1).
		Foreground(subtle).
		String()

	titleStyle = lipgloss.NewStyle().
			MarginRight(5).
			// Padding(0, 1).
			// Italic(true).
			Foreground(highlight)

	imgStyle = lipgloss.NewStyle() //.MarginTop(1)

	infoStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(subtle)

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
				BorderForeground(lipgloss.Color("#874BFD")).
				Padding(0, 0).
				BorderTop(true).
				BorderLeft(true).
				BorderRight(true).
				BorderBottom(true)

	userNameStyle = lipgloss.NewStyle().
			Background(highlight).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(subtle).
			MarginRight(2)

	contentStyle = lipgloss.NewStyle()
	// Align(lipgloss.Left).
	// Foreground(lipgloss.Color("#FAFAFA")).
	// Background(highlight).
	// Margin(1, 3, 0, 0).
	// Padding(1, 2)
	// Height(19).
	// Width(columnWidth)

	md, _ = glamour.NewTermRenderer(
		// detect background color and pick either the default dark or light theme
		glamour.WithAutoStyle(),
		// wrap output at specific width (default is 80)
		glamour.WithWordWrap(80),
	)
)

type castItemDelegate struct{}

func (d castItemDelegate) Height() int {
	return 10
}

func (d castItemDelegate) Spacing() int {
	return 0
}
func (d castItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d castItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(*CastView)
	if !ok {
		return
	}

	rf := boxStyle
	if index == m.Cursor() {
		rf = boxSelectedStyle
	}

	s := rf.Render(i.View())

	fmt.Fprintln(w, s)

}

type CastView struct {
	cast api.Cast
	img  ImageModel
	pfp  ImageModel
}

func NewCastView(cast api.Cast) (*CastView, tea.Cmd) {
	c := &CastView{
		cast: cast,
		pfp:  NewImage(true, true, special),
	}
	img := NewImage(true, true, special)
	c.img = img

	cmds := []tea.Cmd{
		c.pfp.SetURL(cast.Author.PfpURL), c.pfp.SetSize(4, 4),
	}
	if len(cast.Embeds) > 0 {
		cmds = append(cmds, c.img.SetURL(cast.Embeds[0].URL), c.img.SetSize(10, 10))
	}
	return c, tea.Batch(cmds...)
}

func (m *CastView) Init() tea.Cmd {
	return nil
}

func (m *CastView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{}
	img, cmd := m.img.Update(msg)
	m.img = img
	cmds = append(cmds, cmd)

	pfp, cmd := m.pfp.Update(msg)
	m.pfp = pfp
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *CastView) View() string {
	return lipgloss.JoinVertical(lipgloss.Top,
		m.Title(),
		m.Description(),
	)
}

func (m *CastView) String() string {
	return m.View()
}

func (i *CastView) Title() string {

	return lipgloss.JoinHorizontal(lipgloss.Center,
		i.pfp.View(),
		lipgloss.JoinVertical(lipgloss.Top,
			titleStyle.Render(i.cast.Author.DisplayName),
			fmt.Sprintf("@%s", i.cast.Author.Username),
		),
	)
}

func (i *CastView) Description() string {
	m, err := md.Render(i.cast.Text)
	if err != nil {
		m = i.cast.Text
	}

	content := contentStyle.Render(m)
	stats := infoStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top,
		infoStyle.
			Render(fmt.Sprintf("❤️%d ", len(i.cast.Reactions.Likes))),
		infoStyle.
			Render(i.cast.HumanTime()),
	))

	img := ""
	if i.img.FileName != "" {
		img = i.img.View()
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		content,
		imgStyle.Render(img),
		stats,
	)
}

func (i *CastView) FilterValue() string {
	return i.cast.Author.Username
}
