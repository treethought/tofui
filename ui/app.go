package ui

import (
	"crypto/sha256"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"

	"github.com/treethought/tofui/api"
	"github.com/treethought/tofui/config"
)

// TODO provide to models
var renderer *lipgloss.Renderer = lipgloss.DefaultRenderer()

var activeColor = lipgloss.AdaptiveColor{Dark: "#874BFD", Light: "#874BFD"}

var (
	mainStyle = lipgloss.NewStyle().Margin(0).Padding(0).Border(lipgloss.RoundedBorder())
)

func NewStyle() lipgloss.Style {
	return renderer.NewStyle()
}

type UpdateSignerMsg struct {
	Signer *api.Signer
}

type navNameMsg struct {
	name string
}

func navNameCmd(name string) tea.Cmd {
	return func() tea.Msg {
		return navNameMsg{name: name}
	}
}

type SelectCastMsg struct {
	cast *api.Cast
}

type AppContext struct {
	s      ssh.Session
	signer *api.Signer
	pk     string
}

type App struct {
	ctx          *AppContext
	client       *api.Client
	cfg          *config.Config
	focusedModel tea.Model
	focused      string
	navname      string
	sidebar      *Sidebar
	showSidebar  bool
	prev         string
	prevName     string
	quickSelect  *QuickSelect
	publish      *PublishInput
	statusLine   *StatusLine
	// signinPrompt    *SigninPrompt
	splash *SplashView
	help   *HelpView

	feed    *FeedView
	channel *FeedView
	profile *Profile
	cast    *CastView
}

func (a *App) PublicKey() string {
	return a.ctx.pk
}

func NewSSHApp(cfg *config.Config, s ssh.Session, r *lipgloss.Renderer) (*App, error) {
	if r != nil {
		renderer = r
	}
	if s.PublicKey() == nil {
		return nil, fmt.Errorf("public key is nil")
	}
	// hash the pk so we can use it in auth flow
	h := sha256.New()
	pkBytes := s.PublicKey().Marshal()
	h.Write(pkBytes)
	pk := fmt.Sprintf("%x", h.Sum(nil))

	signer := api.GetSigner(pk)
	if signer != nil {
		log.Println("logged in as: ", signer.Username)
	}

	ctx := &AppContext{s: s, pk: pk, signer: signer}
	app := NewApp(cfg, ctx)
	return app, nil
}

func NewLocalApp(cfg *config.Config) *App {
	signer := api.GetSigner("local")
	if signer != nil {
		log.Println("logged in locally as: ", signer.Username)
	}
	ctx := &AppContext{signer: signer, pk: "local"}
	app := NewApp(cfg, ctx)
	return app
}

func NewApp(cfg *config.Config, ctx *AppContext) *App {
	if ctx == nil {
		ctx = &AppContext{}
	}
	a := &App{
		showSidebar: true,
		ctx:         ctx,
		client:      api.NewClient(cfg),
		cfg:         cfg,
	}
	a.feed = NewFeedView(a, feedTypeFollowing)
	a.focusedModel = a.feed

	a.profile = NewProfile(a)

	a.channel = NewFeedView(a, feedTypeChannel)

	a.cast = NewCastView(a, nil)

	a.sidebar = NewSidebar(a)
	a.quickSelect = NewQuickSelect(a)
	a.publish = NewPublishInput(a)
	a.statusLine = NewStatusLine(a)
	a.help = NewHelpView(a)
	a.splash = NewSplashView(a)
	a.splash.SetActive(true)
	if a.ctx.signer == nil {
		a.splash.ShowSignin(true)
	}
	a.SetNavName("feed")

	return a
}

func (a *App) SetNavName(name string) {
	a.prevName = a.navname
	a.navname = name
}

func (a *App) focusMain() {
	if a.quickSelect.Active() {
		a.quickSelect.SetActive(false)
	}
	if a.publish.Active() {
		a.publish.SetActive(false)
		a.publish.SetFocus(false)
	}
	a.sidebar.SetActive(false)
	if a.help.IsFull() {
		a.help.SetFull(false)
	}
}

func (a *App) FocusFeed() tea.Cmd {
	a.focusMain()
	a.SetNavName("feed")
	a.focusedModel = a.feed
	a.focused = "feed"
	return nil
}

func (a *App) FocusProfile() tea.Cmd {
	a.focusMain()
	a.SetNavName("profile")
	a.focusedModel = a.profile
	a.focused = "profile"
	return a.profile.Init()
}

func (a *App) FocusChannel() tea.Cmd {
	a.focusMain()
	a.SetNavName("channel")
	a.focusedModel = a.channel
	a.focused = "channel"
	return a.channel.Init()
}

func (a *App) FocusCast() tea.Cmd {
	a.focusMain()
	a.SetNavName("cast")
	a.focusedModel = a.cast
	a.focused = "cast"
	return a.cast.Init()
}

func (a *App) GetFocused() tea.Model {
	return a.focusedModel
}

func (a *App) FocusPrev() tea.Cmd {
	switch a.prev {
	case "feed":
		return a.FocusFeed()
	case "profile":
		return a.FocusProfile()
	case "channel":
		return a.FocusChannel()
	case "cast":
		return a.FocusCast()
	}
	return a.FocusFeed()
}

func (a *App) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	cmds = append(cmds, a.splash.Init(), a.sidebar.Init(), a.quickSelect.Init(), a.publish.Init())
	focus := a.GetFocused()
	if focus != nil {
		cmds = append(cmds, focus.Init())
	}
	return tea.Batch(cmds...)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	// log.Println("received msg type: ", reflect.TypeOf(msg))
	var cmds []tea.Cmd
	_, sbcmd := a.statusLine.Update(msg)
	cmds = append(cmds, sbcmd)
	switch msg := msg.(type) {
	case *UpdateSignerMsg:
		a.ctx.signer = msg.Signer
		a.splash.ShowSignin(false)
		log.Println("updated signer for: ", msg.Signer.Username)
		return a, a.Init()
	case navNameMsg:
		a.SetNavName(msg.name)
		return a, nil
	case *postResponseMsg:
		_, cmd := a.publish.Update(msg)
		return a, tea.Sequence(cmd, a.FocusPrev())
	case *channelListMsg:
		if msg.activeOnly {
			_, cmd := a.sidebar.Update(msg)
			return a, cmd
		}
		_, qcmd := a.quickSelect.Update(msg.channels)
		_, pcmd := a.publish.Update(msg.channels)
		return a, tea.Batch(qcmd, pcmd)
	case *feedLoadedMsg:
		a.splash.SetActive(false)
	case *channelInfoMsg:
		a.splash.SetInfo(msg.channel.Name)
	case *api.FeedResponse:
		// for first load
		a.splash.SetInfo("loading channels...")
		// pas through to feed or profile
	// case SelectProfileMsg:
	case SelectCastMsg:
		nav := fmt.Sprintf("cast by @%s", msg.cast.Author.Username)
		if msg.cast.ParentHash != "" {
			nav = fmt.Sprintf("reply by @%s", msg.cast.Author.Username)
		}
		a.SetNavName(nav)
		return a, tea.Sequence(
			a.cast.SetCast(msg.cast),
			a.FocusCast(),
		)

	case tea.WindowSizeMsg:
		SetHeight(msg.Height)
		SetWidth(msg.Width)

		a.statusLine.SetSize(msg.Width, 1)
		_, statusHeight := lipgloss.Size(a.statusLine.View())

		wx, wy := msg.Width, msg.Height-statusHeight
		fx, fy := mainStyle.GetFrameSize()
		wx = wx - fx
		wy = wy - fy

		spx := min(80, wx-10)
		spy := min(80, wy-10)

		a.splash.SetSize(spx, spy)

		sx := min(30, int(float64(wx)*0.2))
		a.sidebar.SetSize(sx, wy-statusHeight)
		sideWidth, _ := lipgloss.Size(a.sidebar.View())

		mx := wx - sideWidth
		mx = min(mx, int(float64(wx)*0.8))

		my := min(wy, int(float64(wy)*0.9))

		dialogX, dialogY := int(float64(mx)*0.8), int(float64(my)*0.8)
		a.publish.SetSize(dialogX, dialogY)
		a.quickSelect.SetSize(dialogX, dialogY)
		a.help.SetSize(dialogX, dialogY)

		childMsg := tea.WindowSizeMsg{
			Width:  mx,
			Height: my,
		}

		_, fcmd := a.feed.Update(childMsg)
		_, pcmd := a.profile.Update(childMsg)
		_, ccmd := a.channel.Update(childMsg)
		_, cscmd := a.cast.Update(childMsg)

		cmds = append(cmds, fcmd, pcmd, ccmd, cscmd)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		}

		if a.splash.Active() {
			_, cmd := a.splash.Update(msg)
			return a, cmd
		}
		if a.publish.Active() {
			_, cmd := a.publish.Update(msg)
			return a, cmd
		}

		cmd := NavKeyMap.HandleMsg(a, msg)
		if cmd != nil {
			return a, cmd
		}

	case *currentAccountMsg:
		_, cmd := a.sidebar.Update(msg)
		a.splash.ShowSignin(false)
		return a, cmd
	}
	if a.publish.Active() {
		_, cmd := a.publish.Update(msg)
		return a, cmd
	}
	if a.quickSelect.Active() {
		q, cmd := a.quickSelect.Update(msg)
		a.quickSelect = q.(*QuickSelect)
		return a, cmd
	}

	if a.help.IsFull() {
		_, cmd := a.help.Update(msg)
		return a, cmd
	}

	if a.sidebar.Active() {
		_, cmd := a.sidebar.Update(msg)
		return a, cmd
	}
	_, scmd := a.sidebar.pfp.Update(msg)
	cmds = append(cmds, scmd)

	current := a.GetFocused()
	if current == nil {
		log.Println("no focused model")
		return Fallback, nil
	}

	_, cmd := current.Update(msg)
	cmds = append(cmds, cmd)
	return a, tea.Batch(cmds...)

}

func (a *App) View() string {
	focus := a.GetFocused()
	if focus == nil {
		return "no focused model"
	}
	main := focus.View()
	side := a.sidebar.View()
	if a.splash.Active() {
		main = lipgloss.Place(GetWidth(), GetHeight(), lipgloss.Center, lipgloss.Center, a.splash.View())
		return main
	}

	if a.publish.Active() {
		main = a.publish.View()
	}
	if a.quickSelect.Active() {
		main = a.quickSelect.View()
	}
	if !a.showSidebar {
		return NewStyle().Align(lipgloss.Center).Render(main)
	}

	if a.help.IsFull() {
		main = a.help.View()
	}

	ss := mainStyle
	if !a.sidebar.Active() {
		ss = ss.BorderForeground(activeColor)
	}
	main = ss.Render(main)

	return lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.JoinHorizontal(lipgloss.Center, side, main),
		a.statusLine.View(),
	)

}
