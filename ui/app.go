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

type FocusMsg struct {
	Name string
}

func focusCmd(name string) tea.Cmd {
	return func() tea.Msg {
		return FocusMsg{Name: name}
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
	models          map[string]tea.Model
	ctx             *AppContext
	client          *api.Client
	cfg             *config.Config
	focusedModel    tea.Model
	focused         string
	navname         string
	sidebar         *Sidebar
	showSidebar     bool
	prev            string
	prevName        string
	quickSelect     *QuickSelect
	showQuickSelect bool
	publish         *PublishInput
	statusLine      *StatusLine
	// signinPrompt    *SigninPrompt
	splash *SplashView
	help   *HelpView
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
		models:      make(map[string]tea.Model),
		showSidebar: true,
		ctx:         ctx,
		client:      api.NewClient(cfg),
		cfg:         cfg,
	}
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

	feed := NewFeedView(a)
	a.Register("feed", feed)
	a.SetFocus("feed")

	castDetails := NewCastView(a, nil)
	a.Register("cast", castDetails)

	profile := NewProfile(a)
	a.Register("profile", profile)
	return a
}

func (a *App) GetModel(name string) tea.Model {
	m, ok := a.models[name]
	if !ok {
		log.Println("model not found: ", name)
	}
	return m
}

func (a *App) Register(name string, model tea.Model) {
	a.models[name] = model
}

func (a *App) SetNavName(name string) {
	a.prevName = a.navname
	a.navname = name
}

func (a *App) SetFocus(name string) tea.Cmd {
	if a.showQuickSelect {
		a.showQuickSelect = false
	}
	if a.publish.Active() {
		a.publish.SetActive(false)
		a.publish.SetFocus(false)
	}
	if name == "" || name == a.focused {
		return nil
	}
	// clear if we're back at feed
	a.prev = ""
	if name != "feed" {
		a.prev = a.focused
	}
	m, ok := a.models[name]
	if !ok {
		log.Println("model not found: ", name)
	}
	a.focusedModel = m
	a.focused = name
	a.sidebar.SetActive(false)
	return m.Init()
}

func (a *App) GetFocused() tea.Model {
	return a.focusedModel
}

func (a *App) FocusPrev() tea.Cmd {
	if a.help.IsFull() {
		a.help.SetFull(false)
	}
	prev := a.GetModel(a.prev)
	if a.prev == "" || prev == nil {
		return nil
	}

	if m := a.GetModel(a.prev); m != nil {
		a.SetNavName(a.prevName)
		return a.SetFocus(a.prev)
	}
	a.SetNavName("feed")
	return a.SetFocus("feed")
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

func (a *App) propagateEvent(msg tea.Msg) tea.Cmd {
	for name, m := range a.models {
		if name == a.focused {
			um, cmd := m.Update(msg)
			a.models[name] = um
			return cmd
		}
	}
	return nil
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
	case FocusMsg:
		cmd := a.SetFocus(msg.Name)
		if cmd != nil {
			return a, cmd
		}
	case *postResponseMsg:
		_, cmd := a.publish.Update(msg)
		return a, tea.Sequence(cmd, a.FocusPrev())
	case *channelListMsg:
		if msg.activeOnly {
			_, cmd := a.sidebar.Update(msg)
			return a, cmd
		} else {
			_, cmd := a.quickSelect.Update(msg.channels)
			return a, cmd
		}
	case *feedLoadedMsg:
		a.splash.SetActive(false)
	case *channelInfoMsg:
		a.splash.SetInfo(msg.channel.Name)
	case *api.FeedResponse:
		// allow msg to pass through to profile's embedded feed
		if a.focused == "profile" {
			_, cmd := a.GetFocused().Update(msg)
			return a, cmd
		}
		feed := a.GetModel("feed").(*FeedView)
		if a.focused == "feed" {
			feed.Clear()
		}
		focusCmd := a.SetFocus("feed")
		a.splash.SetInfo("loading channels...")
		return a, tea.Batch(feed.setItems(msg.Casts), focusCmd)
	case SelectProfileMsg:
		focusCmd := a.SetFocus("profile")
		cmd := a.GetModel("profile").(*Profile).SetFID(msg.fid)
		return a, tea.Batch(focusCmd, cmd)
	case SelectCastMsg:
		nav := fmt.Sprintf("cast by @%s", msg.cast.Author.Username)
		if msg.cast.ParentHash != "" {
			nav = fmt.Sprintf("reply by @%s", msg.cast.Author.Username)
		}

		a.SetNavName(nav)
		focusCmd := a.SetFocus("cast")
		cmd := a.GetModel("cast").(*CastView).SetCast(msg.cast)
		return a, tea.Sequence(cmd, focusCmd)

	case tea.WindowSizeMsg:
		SetHeight(msg.Height)
		SetWidth(msg.Width)

		a.statusLine.SetSize(msg.Width, 1)

		// set the height of the statusLine
		wx, wy := msg.Width, msg.Height-1

		sideMax := 30
		sidePct := int(float64(wx) * 0.2)
		sx := sidePct
		if sideMax < sidePct {
			sx = sideMax
		}
		a.sidebar.SetSize(sx, wy)

		qw := wx - sx
		qh := wy - 10
		a.quickSelect.SetSize(qw, qh)

		pw := wx - sx
		py := wy - 10
		a.publish.SetSize(pw, py)
		a.splash.SetSize(pw, py)

		hw := wx - sx
		hy := wy - 10
		a.help.SetSize(hw, hy)

		// substract the sidebar width from the window width
		mx, my := wx-sx, wy
		childMsg := tea.WindowSizeMsg{
			Width:  mx,
			Height: my,
		}

		for n, m := range a.models {
			um, cmd := m.Update(childMsg)
			a.models[n] = um
			cmds = append(cmds, cmd)
		}
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
	if a.showQuickSelect {
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
		main = a.splash.View()
	}

	if a.publish.Active() {
		main = a.publish.View()
	}
	if a.showQuickSelect {
		main = a.quickSelect.View()
	}
	if !a.showSidebar {
		return NewStyle().Align(lipgloss.Center).Render(main)
	}

	if a.help.IsFull() {
		main = a.help.View()
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.JoinHorizontal(lipgloss.Center, side, main),
		a.statusLine.View(),
	)

}

func UpdateChildren(msg tea.Msg, models ...tea.Model) tea.Cmd {
	cmds := make([]tea.Cmd, len(models))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range models {
		models[i], cmds[i] = models[i].Update(msg)
	}

	return tea.Batch(cmds...)
}
