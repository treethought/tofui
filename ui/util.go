package ui

import (
	"log"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	EmojiLike      = "❤️"
	EmojiEmptyLike = "🤍"
	EmojiRecyle    = "♻️"
	EmojiComment   = "💬"
)

func OpenURL(url string) tea.Cmd {
	return func() tea.Msg {
		log.Println("Opening URL: ", url)
		var cmd string
		var args []string

		switch runtime.GOOS {
		case "windows":
			cmd = "cmd"
			args = []string{"/c", "start"}
		case "darwin":
			cmd = "open"
		default: // "linux", "freebsd", "openbsd", "netbsd"
			cmd = "xdg-open"
		}
		args = append(args, url)

		_ = exec.Command(cmd, args...).Start()
		return nil
	}
}
