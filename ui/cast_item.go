package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
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

	displayNameStyle = NewStyle().
				MarginRight(5).
				Foreground(highlight)

	usernameStyle = NewStyle()

	imgStyle = NewStyle()

	headerStyle = NewStyle().BorderBottom(true)

	infoStyle = NewStyle().
			MarginLeft(5).MarginRight(5).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(subtle).AlignHorizontal(lipgloss.Center)

	contentStyle = NewStyle()

	md, _ = glamour.NewTermRenderer(
		// detect background color and pick either the default dark or light theme
		glamour.WithAutoStyle(),
		// wrap output at specific width (default is 80)
		glamour.WithWordWrap(80),
	)
)

func UsernameHeader(user *api.User, img *ImageModel) string {
	if user == nil {
		return spinner.New().View()
	}
	return headerStyle.Render(lipgloss.JoinHorizontal(lipgloss.Center,
		img.View(),
		lipgloss.JoinHorizontal(lipgloss.Top,
			displayNameStyle.Render(
				user.DisplayName,
			),
			usernameStyle.Render(
				fmt.Sprintf("@%s", user.Username),
			),
		),
	),
	)
}

func CastStats(cast *api.Cast, margin int) string {
	if cast == nil {
		return spinner.New().View()
	}
	liked := EmojiEmptyLike
	if cast.ViewerContext.Liked {
		liked = EmojiLike
	}
	stats := lipgloss.JoinHorizontal(lipgloss.Top,
		NewStyle().Render(fmt.Sprintf("%d ", cast.Replies.Count)),
		NewStyle().MarginRight(margin).Render(EmojiComment),
		NewStyle().Render(fmt.Sprintf("%d ", cast.Reactions.LikesCount)),
		NewStyle().MarginRight(margin).Render(liked),
		NewStyle().Render(fmt.Sprintf("%d ", cast.Reactions.RecastsCount)),
		NewStyle().MarginRight(margin).Render(EmojiRecyle),
	)
	return stats

}

func CastContent(cast *api.Cast, maxHeight int, imgs ...ImageModel) string {
	if cast == nil {
		return spinner.New().View()
	}
	m, err := md.Render(cast.Text)
	if err != nil {
		m = cast.Text
	}
	return contentStyle.MaxHeight(maxHeight).Render(m)
}

func getCastChannelCmd(client *api.Client, cast *api.Cast) tea.Cmd {
	return func() tea.Msg {
		if cast.ParentURL == "" {
			return nil
		}
		ch, err := client.GetChannelByParentUrl(cast.ParentURL)
		if err != nil {
			return channelInfoErrMsg{err, cast.Hash, cast.ParentURL}
		}
		return channelInfoMsg{ch, cast.Hash, cast.ParentURL}
	}
}

type channelInfoMsg struct {
	channel   *api.Channel
	cast      string
	parentURL string
}
type channelInfoErrMsg struct {
	err       error
	cast      string
	parentURL string
}

type CastFeedItem struct {
	app        *App
	cast       *api.Cast
	channel    string
	channelURL string
	pfp        *ImageModel
	compact    bool
}

// NewCastFeedItem displays a cast in compact form within a list
// implements list.Item (and tea.Model only for updating image)
func NewCastFeedItem(app *App, cast *api.Cast, compact bool) (*CastFeedItem, tea.Cmd) {
	c := &CastFeedItem{
		app:     app,
		cast:    cast,
		pfp:     NewImage(true, true, special),
		compact: compact,
	}

	cmds := []tea.Cmd{
		c.pfp.SetURL(cast.Author.PfpURL, false),
		getCastChannelCmd(app.client, cast),
	}

	if c.compact {
		cmds = append(cmds, c.pfp.SetSize(2, 1))
	} else {
		cmds = append(cmds, c.pfp.SetSize(4, 4))
	}
	return c, tea.Batch(cmds...)
}

func (m *CastFeedItem) Init() tea.Cmd { return nil }

func (m *CastFeedItem) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m = &CastFeedItem{
		app:     m.app,
		cast:    m.cast,
		channel: m.channel,
		pfp:     m.pfp,
		compact: m.compact,
	}
	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case channelInfoErrMsg:
		if msg.cast != m.cast.Hash {
			return m, nil
		}
	case channelInfoMsg:
		if msg.cast != m.cast.Hash {
			return m, nil
		}
		m.channel = msg.channel.Name
		m.channelURL = msg.parentURL
	}

	if m.pfp.Matches(msg) {
		pfp, cmd := m.pfp.Update(msg)
		m.pfp = pfp
		return m, cmd

	}

	return m, tea.Batch(cmds...)
}

func (m *CastFeedItem) View() string { return "" }

func (m *CastFeedItem) AsRow(ch, stats bool) []string {
	cols := []string{}
	if ch {
		cols = append(cols, fmt.Sprintf("/%s", m.channel))
	}
	if stats {
		cols = append(cols, CastStats(m.cast, 2))
	} else {
	}
	cols = append(cols, m.cast.Author.DisplayName, m.cast.Text)
	return cols
}

func (i *CastFeedItem) Title() string {
	return UsernameHeader(&i.cast.Author, i.pfp)
}

func (i *CastFeedItem) Description() string {
	return CastContent(i.cast, 3)
}

func (i *CastFeedItem) FilterValue() string {
	return i.cast.Author.Username
}
