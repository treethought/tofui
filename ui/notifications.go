package ui

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/treethought/tofui/api"
)

type notificationsMsg struct {
	notifications []*api.Notification
}

func getNotificationsCmd(client *api.Client, signer *api.Signer) tea.Cmd {
	return func() tea.Msg {
		if signer == nil {
			return nil
		}
		resp, err := client.GetNotifications(signer.FID)
		if err != nil {
			log.Println("error getting notifications: ", err)
			return nil
		}
		return &notificationsMsg{notifications: resp.Notifications}
	}
}

type notifItem struct {
	*api.Notification
}

func (n *notifItem) FilterValue() string {
	return string(n.Type)
}

func buildUserList(users []api.User) string {
	s := ""
	for i, u := range users {
		if i > 3 {
			s += fmt.Sprintf(" and %d others", len(users)-i)
			return s
		}
		s += u.DisplayName
		if i < len(users)-1 {
			s += ", "
		}
	}
	return s
}

func (n *notifItem) Title() string {
	switch n.Type {
	case api.NotificationsTypeFollows:
		users := []api.User{}
		for _, f := range n.Follows {
			users = append(users, f.User)
		}
		userStr := buildUserList(users)
		return fmt.Sprintf("%s  %s followed you ", EmojiPerson, userStr)
	case api.NotificationsTypeLikes:
		users := []api.User{}
		for _, r := range n.Reactions {
			if r.Object == api.CastReactionObjTypeLikes {
				users = append(users, r.User)
			}
		}
		userStr := buildUserList(users)
		return fmt.Sprintf("%s  %s liked your post", EmojiLike, userStr)
	case api.NotificationsTypeRecasts:
		users := []api.User{}
		for _, r := range n.Reactions {
			if r.Object == api.CastReactionObjTypeRecasts {
				users = append(users, r.User)
			}
		}
		userStr := buildUserList(users)
		return fmt.Sprintf("%s  %s recasted your post", EmojiRecyle, userStr)
	case api.NotificationsTypeReply:
		return fmt.Sprintf("%s  %s replied to your post", EmojiComment, n.Cast.Author.DisplayName)

	default:
		return "unknown notification type: " + string(n.Type)

	}
}

func (i *notifItem) Description() string {
	switch i.Type {
	case api.NotificationsTypeLikes, api.NotificationsTypeRecasts:
		if i.Cast != nil {
			return i.Cast.Text
		}
		for _, r := range i.Reactions {
			if r.Object == api.CastReactionObjTypeLikes {
				if r.Cast.Object == "cast_dehydrated" {
					return r.Cast.Hash
				}
				return r.Cast.Text
			}
		}
		return "?"
	case api.NotificationsTypeReply:
		return i.Cast.Text

	}
	return ""

}

type NotificationsView struct {
	app    *App
	list   *list.Model
	w, h   int
	active bool
	items  []list.Item
}

func NewNotificationsView(app *App) *NotificationsView {
	d := list.NewDefaultDelegate()
	d.SetHeight(2)
	d.ShowDescription = true

	l := list.New([]list.Item{}, d, 100, 100)
	l.KeyMap.CursorUp.SetKeys("k", "up")
	l.KeyMap.CursorDown.SetKeys("j", "down")
	l.KeyMap.Quit.SetKeys("ctrl+c")
	l.Title = "notifications"
	l.SetShowTitle(true)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)
	l.SetShowHelp(true)
	l.SetShowStatusBar(true)
	l.SetShowPagination(true)

	return &NotificationsView{app: app, list: &l}
}

func (m *NotificationsView) SetSize(w, h int) {
	m.w, m.h = w, h
	m.list.SetWidth(w)
	m.list.SetHeight(h)
}
func (m *NotificationsView) Active() bool {
	return m.active
}
func (m *NotificationsView) SetActive(active bool) {
	m.active = active
}

func (m *NotificationsView) Init() tea.Cmd {
	return getNotificationsCmd(m.app.client, m.app.ctx.signer)
}

func (m *NotificationsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil

	case *notificationsMsg:
		items := []list.Item{}
		for _, n := range msg.notifications {
			items = append(items, &notifItem{n})
		}
		m.items = items
		m.list.SetItems(items)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "enter":
			item, ok := m.list.SelectedItem().(*notifItem)
			if !ok {
				return m, noOp()
			}
			d, _ := json.MarshalIndent(item, "", "  ")
			log.Println(string(d))
			return m, noOp()
		}
		l, cmd := m.list.Update(msg)
		m.list = &l
		return m, cmd
	}

	return m, nil
}

func (m *NotificationsView) View() string {
	return NewStyle().Width(m.w).Height(m.h).Render(m.list.View())
}
