package ui

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/tofui/api"
)

func UserBio(user *api.User) string {
	if user == nil {
		l := NewLoading()
		l.SetActive(true)
		return l.View()
	}
	stats := lipgloss.JoinHorizontal(lipgloss.Top,
		NewStyle().Bold(true).Render(fmt.Sprintf("%d", user.FollowingCount)),
		NewStyle().MarginRight(10).Render(" following"),
		NewStyle().Bold(true).Render(fmt.Sprintf("%d", user.FollowerCount)),
		NewStyle().Render(" followers"),
	)

	style := NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderBottom(true).Padding(2)

	return style.Render(lipgloss.JoinVertical(lipgloss.Top,
		NewStyle().MarginTop(0).MarginBottom(0).Padding(0).Render(user.Profile.Bio.Text),
		stats,
	))

}

type profileFeedMsg struct {
	fid   uint64
	casts []*api.Cast
}

type SelectProfileMsg struct {
	fid uint64
}

type ProfileMsg struct {
	fid  uint64
	user *api.User
	err  error
}

type Profile struct {
	app  *App
	user *api.User
	pfp  *ImageModel
	feed *FeedView
}

func NewProfile(app *App) *Profile {
	f := NewFeedView(app, feedTypeProfile)
	return &Profile{
		app:  app,
		pfp:  NewImage(false, true, special),
		feed: f,
	}
}

func getUserCmd(client *api.Client, fid, viewer uint64) tea.Cmd {
	return func() tea.Msg {
		log.Println("get user by fid cmd", fid)
		user, err := client.GetUserByFID(fid, viewer)
		return ProfileMsg{fid, user, err}
	}
}

func getUserFeedCmd(client *api.Client, fid, viewer uint64) tea.Cmd {
	return func() tea.Msg {
		req := &api.FeedRequest{
			FeedType: "filter", FilterType: "fids", Limit: 100,
			FIDs: []uint64{fid}, ViewerFID: viewer, FID: viewer,
		}
		feed, err := client.GetFeed(req)
		if err != nil {
			log.Println("feedview error getting feed", err)
			return err
		}
		return &profileFeedMsg{fid, feed.Casts}
	}
}

func (m *Profile) SetFID(fid uint64) tea.Cmd {
	var viewer uint64
	if m.app.ctx.signer != nil {
		viewer = m.app.ctx.signer.FID
	}
	return tea.Batch(
		getUserCmd(m.app.client, fid, viewer),
		getUserFeedCmd(m.app.client, fid, viewer),
	)
}

func (m *Profile) Init() tea.Cmd {
	return m.feed.Init()
}

func (m *Profile) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		x, y := msg.Width, msg.Height
		m.pfp.SetSize(4, 4)

		hy := lipgloss.Height(UsernameHeader(m.user, m.pfp))
		by := lipgloss.Height(UserBio(m.user))

		fy := y - hy - by

		// TODO use size of header/stats
		m.feed.SetSize(x, fy)
		return m, nil

	case *SelectProfileMsg:
		return m, m.SetFID(msg.fid)

	case ProfileMsg:
		if msg.user != nil {
			m.user = msg.user
			m.pfp.SetURL(m.user.PfpURL, false)
			m.pfp.SetSize(4, 4)
			return m, tea.Batch(
				m.pfp.Render(),
				navNameCmd(fmt.Sprintf("profile: @%s", m.user.Username)),
			)
		}
		return m, nil
	}
	_, fcmd := m.feed.Update(msg)
	_, pcmd := m.pfp.Update(msg)
	return m, tea.Batch(fcmd, pcmd)
}
func (m *Profile) View() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		UsernameHeader(m.user, m.pfp),
		UserBio(m.user),
		m.feed.View(),
	)
}
