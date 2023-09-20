package ui

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/treethought/castr/api"
)

type setCastMsg struct {
	cast *api.Cast
}

type CastView struct {
	cast  *api.Cast
	imgs  []ImageModel
	pfp   ImageModel
	links []string
}

func NewCastView(cast *api.Cast) *CastView {
	c := &CastView{
		cast: cast,
		pfp:  NewImage(true, true, special),
	}
	return c
}

func (m *CastView) SetCast(cast *api.Cast) tea.Cmd {
	return func() tea.Msg {
		return setCastMsg{cast: cast}
	}
}

// m.cast = cast
// m.links = extractLinks(cast.Text)
// return m.Init()
// }

func (m *CastView) Init() tea.Cmd {
	if m.cast == nil {
		return nil
	}
	cmds := []tea.Cmd{
		m.pfp.SetURL(m.cast.Author.PfpURL), m.pfp.SetSize(4, 4),
	}
	// for _, link := range m.links {
	// 	log.Println("setting link: ", link)
	// 	im := NewImage(true, true, special)
	// 	m.imgs = append(m.imgs, im)
	// 	cmds = append(cmds, im.SetURL(link), im.SetSize(4, 4))
	// }
	for _, embed := range m.cast.Embeds {
		log.Println("setting embed: ", embed.URL)
		im := NewImage(true, true, special)
		m.imgs = append(m.imgs, im)
		cmds = append(cmds, im.SetURL(embed.URL), im.SetSize(4, 4))
	}

	return tea.Batch(cmds...)
}

func (m *CastView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case setCastMsg:
		log.Println("castview: setting cast: ", msg.cast)
		m.cast = msg.cast
		return m, m.Init()
	}

	log.Println("updating cast view")
	cmds := []tea.Cmd{}
	pfp, cmd := m.pfp.Update(msg)
	m.pfp = pfp
	cmds = append(cmds, cmd)

	for i, _ := range m.imgs {
		img, cmd := m.imgs[i].Update(msg)
		img.SetIsActive(true)
		m.imgs[i] = img
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *CastView) Images() string {
	imgStrs := []string{}
	for i, _ := range m.imgs {
		log.Println("rendering image: ", fmt.Sprintf("%+v", i))
		m.imgs[i].SetIsActive(true)
		imgStrs = append(imgStrs, m.imgs[i].View())
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, imgStrs...)
}

func (m *CastView) View() string {
	if m.cast == nil {
		return "loading"
	}
	return docStyle.Render(lipgloss.JoinVertical(lipgloss.Bottom,
		CastHeader(m.cast, m.pfp),
		CastContent(m.cast, 10, true),
		imgStyle.Align(lipgloss.Center).Render(m.Images()),
	),
	)
}
