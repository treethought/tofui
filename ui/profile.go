package ui

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/castr/api"
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

	style := NewStyle().BorderStyle(lipgloss.ThickBorder()).BorderBottom(true).Padding(2)

	return style.Render(lipgloss.JoinVertical(lipgloss.Top,
		NewStyle().MarginTop(0).MarginBottom(0).Padding(0).Render(user.Profile.Bio.Text),
		stats,
	))

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
	return &Profile{
		app:  app,
		pfp:  NewImage(false, true, special),
		feed: NewFeedView(app),
	}
}

func getUserCmd(client *api.Client, fid, viewer uint64) tea.Cmd {
	return func() tea.Msg {
		log.Println("get user by fid cmd", fid)
		user, err := client.GetUserByFID(fid, viewer)
		return ProfileMsg{fid, user, err}
	}
}

func (m *Profile) SetFID(fid uint64) tea.Cmd {
	var viewer uint64
	if m.app.ctx.signer != nil {
		viewer = m.app.ctx.signer.FID
	}
	return tea.Batch(
		getUserCmd(m.app.client, fid, viewer),
		getFeedCmd(m.app.client, &api.FeedRequest{
			FeedType: "filter", FilterType: "fids", Limit: 100,
			FIDs: []uint64{fid}, ViewerFID: viewer, FID: viewer,
		}))
}

func (m *Profile) Init() tea.Cmd {
	return m.feed.Init()
}

func (m *Profile) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		x, y := msg.Width, msg.Height
		m.pfp.SetSize(4, 4)

		// TODO use size of header/stats
		fx := x
		fy := int(float64(y) * 0.6)
		m.feed.SetSize(fx, fy)
		return m, nil

	case ProfileMsg:
		if msg.user != nil {
			m.user = msg.user
			return m, tea.Batch(
				m.pfp.SetURL(m.user.PfpURL, false),
				m.pfp.SetSize(4, 4),
				navNameCmd(fmt.Sprintf("profile: @%s", m.user.Username)),
			)
		}
		return m, nil
		// cmd := m.pfp.SetURL(m.user.PfpURL, false)
		// case *api.FeedResponse:
		//    log.Println("got feed", len(msg.Casts))
		// 	return m, m.feed.setItems(msg)
	}
	_, fcmd := m.feed.Update(msg)
	_, pcmd := m.pfp.Update(msg)
	return m, tea.Batch(fcmd, pcmd)
}
func (m *Profile) View() string {

	// profile := NewStyle().MaxHeight(2).Render(lipgloss.JoinHorizontal(lipgloss.Left,
	//   UsernameHeader(m.user, m.pfp),
	//   UserBio(m.user),
	//   )

	return lipgloss.JoinVertical(lipgloss.Center,
		UsernameHeader(m.user, m.pfp),
		UserBio(m.user),
		m.feed.View(),
	)
}
