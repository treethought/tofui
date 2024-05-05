package ui

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/castr/api"
)

func UserBio(user *api.User) string {
	if user == nil {
		return spinner.New().View()
	}
	stats := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%d", user.FollowingCount)),
		lipgloss.NewStyle().MarginRight(10).Render(" following"),
		lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%d", user.FollowerCount)),
		lipgloss.NewStyle().Render(" followers"),
	)

	style := lipgloss.NewStyle().BorderStyle(lipgloss.ThickBorder()).BorderBottom(true).Padding(2)

	return style.Render(lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.NewStyle().MarginTop(0).MarginBottom(0).Padding(0).Render(user.Profile.Bio.Text),
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
	user *api.User
	pfp  *ImageModel
	feed *FeedView
}

func NewProfile() *Profile {
	return &Profile{
		pfp:  NewImage(false, true, special),
		feed: NewFeedView(api.GetClient(), nil),
	}
}

func getUserCmd(fid uint64) tea.Cmd {
	return func() tea.Msg {
		log.Println("get user by fid cmd", fid)
		user, err := api.GetClient().GetUserByFID(fid)
		return ProfileMsg{fid, user, err}
	}
}

func (m *Profile) SetFID(fid uint64) tea.Cmd {
	return tea.Batch(getUserCmd(fid), getFeedCmd(&api.FeedRequest{FeedType: "filter", FilterType: "fids", Limit: 100, FIDs: []uint64{fid}}))
}

func (m *Profile) Init() tea.Cmd {
	return nil
}

func (m *Profile) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		x, y := msg.Width, msg.Height
		fx := int(float64(x) * 0.1)
		fy := int(float64(y) * 0.1)
		m.pfp.SetSize(fx, fy)
		return m, nil

	case ProfileMsg:
		log.Println("profile msg", msg)
		if msg.user != nil {
			log.Println("got user by fid: ", msg.fid, msg.user.Username)
			m.user = msg.user
			return m, tea.Batch(m.pfp.SetURL(m.user.PfpURL, false), m.pfp.SetSize(4, 4))
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

	// profile := lipgloss.NewStyle().MaxHeight(2).Render(lipgloss.JoinHorizontal(lipgloss.Left,
	//   UsernameHeader(m.user, m.pfp),
	//   UserBio(m.user),
	//   )

	return lipgloss.JoinVertical(lipgloss.Left,
		UsernameHeader(m.user, m.pfp),
		UserBio(m.user),
		m.feed.View(),
	)
}
