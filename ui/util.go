package ui

import (
	"os/exec"
	"regexp"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	regexPattern = `https?://[^\s<>{}[\](),]+`
	re           = regexp.MustCompile(regexPattern)
)

func OpenURL(url string) tea.Cmd {
	return func() tea.Msg {
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

func extractLinks(input string) []string {
	return re.FindAllString(input, -1)
}
