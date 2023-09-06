package ui

import tea "github.com/charmbracelet/bubbletea"

var Fallback = FallbackModel{}

type FallbackModel struct{}

func (m FallbackModel) Init() tea.Cmd {
	return nil
}
func (m FallbackModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
func (m FallbackModel) View() string {
	return "Something went wrong"
}
