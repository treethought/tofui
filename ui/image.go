// modified from https://github.com/mistakenelf/teacup/blob/main/image/image.go
package ui

import (
	"crypto/sha256"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/disintegration/imaging"
	"github.com/lucasb-eyer/go-colorful"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

type convertImageToStringMsg string
type errorMsg error

const (
	padding = 1
)

// ToString converts an image to a string representation of an image.
func ToString(width int, img image.Image) string {
	img = imaging.Resize(img, width, 0, imaging.Lanczos)
	b := img.Bounds()
	imageWidth := b.Max.X
	h := b.Max.Y
	str := strings.Builder{}

	for heightCounter := 0; heightCounter < h; heightCounter += 2 {
		for x := imageWidth; x < width; x += 2 {
			str.WriteString(" ")
		}

		for x := 0; x < imageWidth; x++ {
			c1, _ := colorful.MakeColor(img.At(x, heightCounter))
			color1 := lipgloss.Color(c1.Hex())
			c2, _ := colorful.MakeColor(img.At(x, heightCounter+1))
			color2 := lipgloss.Color(c2.Hex())
			str.WriteString(lipgloss.NewStyle().Foreground(color1).
				Background(color2).Render("▀"))
		}

		str.WriteString("\n")
	}

	return str.String()
}

type embedPreview struct {
	Title       string
	Description string
	ImageURL    string
}

func getEmbedPreview(url string) (*embedPreview, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	preview := &embedPreview{
		Title:       doc.Find("meta[property='og:title']").AttrOr("content", ""),
		Description: doc.Find("meta[property='og:description']").AttrOr("content", ""),
		ImageURL:    doc.Find("meta[property='og:image']").AttrOr("content", ""),
	}

	// If Open Graph tags aren't present, you can use fallbacks or other metadata, e.g., from Twitter cards.
	if preview.Title == "" {
		preview.Title = doc.Find("meta[name='twitter:title']").AttrOr("content", "")
	}
	if preview.Description == "" {
		preview.Description = doc.Find("meta[name='twitter:description']").AttrOr("content", "")
	}
	if preview.ImageURL == "" {
		preview.ImageURL = doc.Find("meta[name='twitter:image']").AttrOr("content", "")
	}

	return preview, nil
}

func downloadImage(width int, url string) tea.Cmd {
	return func() tea.Msg {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			log.Println("error getting cache dir : ", err)
			return errorMsg(err)
		}

		imgageCacheDir := filepath.Join(cacheDir, "castr", "img")
		os.MkdirAll(imgageCacheDir, os.ModePerm)

		// check for cached image in tmp dir
		// if it exists return instance
		// if not download image and cache instance

		// get hash of url
		ep, err := getEmbedPreview(url)
		if err == nil {
			url = ep.ImageURL
		}

		hash := sha256.Sum256([]byte(url))

		fileName := filepath.Join(imgageCacheDir, fmt.Sprintf("%x", hash))
		log.Println(fileName)

		// check if file exists
		if _, ferr := os.Stat(fileName); ferr == nil {
			return convertImageToStringCmd(width, fileName)
		}

		f, err := os.Create(fileName)
		if err != nil {
			return errorMsg(err)
		}
		defer f.Close()

		log.Println("downloading : ", url)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Println("error creating request : ", err)
			return errorMsg(err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("error downloading image : ", err)
			return errorMsg(err)
		}

		_, err = io.Copy(f, resp.Body)
		if err != nil {
			log.Println("error copying response body to file : ", err)
			return errorMsg(err)
		}
		return convertImageToStringCmd(width, f.Name())
	}
}

// convertImageToStringCmd redraws the image based on the width provided.
func convertImageToStringCmd(width int, filename string) tea.Cmd {
	return func() tea.Msg {
		imageContent, err := os.Open(filepath.Clean(filename))
		if err != nil {
			return errorMsg(err)
		}

		img, _, err := image.Decode(imageContent)
		if err != nil {
			return errorMsg(err)
		}

		imageString := ToString(width, img)

		return convertImageToStringMsg(imageString)
	}
}

// ImageModel represents the properties of a code bubble.
type ImageModel struct {
	Viewport    viewport.Model
	BorderColor lipgloss.AdaptiveColor
	Active      bool
	Borderless  bool
	FileName    string
	ImageString string
}

// New creates a new instance of code.
func NewImage(active, borderless bool, borderColor lipgloss.AdaptiveColor) ImageModel {
	viewPort := viewport.New(0, 0)
	border := lipgloss.NormalBorder()

	if borderless {
		border = lipgloss.HiddenBorder()
	}

	viewPort.Style = lipgloss.NewStyle().
		PaddingLeft(padding).
		PaddingRight(padding).
		Border(border).
		BorderForeground(borderColor)

	return ImageModel{
		Viewport:    viewPort,
		Active:      active,
		Borderless:  borderless,
		BorderColor: borderColor,
	}
}

// Init initializes the code bubble.
func (m ImageModel) Init() tea.Cmd {
	return nil
}

func (m *ImageModel) SetURL(url string) tea.Cmd {
	return downloadImage(m.Viewport.Width, url)
}

// SetFileName sets current file to highlight, this
// returns a cmd which will highlight the text.
func (m *ImageModel) SetFileName(filename string) tea.Cmd {
	m.FileName = filename

	return convertImageToStringCmd(m.Viewport.Width, filename)
}

// SetBorderColor sets the current color of the border.
func (m *ImageModel) SetBorderColor(color lipgloss.AdaptiveColor) {
	m.BorderColor = color
}

// SetSize sets the size of the bubble.
func (m *ImageModel) SetSize(w, h int) tea.Cmd {
	m.Viewport.Width = w
	m.Viewport.Height = h

	border := lipgloss.NormalBorder()

	if m.Borderless {
		border = lipgloss.HiddenBorder()
	}

	m.Viewport.Style = lipgloss.NewStyle().
		PaddingLeft(padding).
		PaddingRight(padding).
		Border(border).
		BorderForeground(m.BorderColor)

	if m.FileName != "" {
		return convertImageToStringCmd(m.Viewport.Width, m.FileName)
	}

	return nil
}

// SetIsActive sets if the bubble is currently active
func (m *ImageModel) SetIsActive(active bool) {
	m.Active = active
}

// GotoTop jumps to the top of the viewport.
func (m *ImageModel) GotoTop() {
	m.Viewport.GotoTop()
}

// SetBorderless sets weather or not to show the border.
func (m *ImageModel) SetBorderless(borderless bool) {
	m.Borderless = borderless
}

// Update handles updating the UI of a code bubble.
func (m ImageModel) Update(msg tea.Msg) (ImageModel, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case convertImageToStringMsg:
		m.ImageString = lipgloss.NewStyle().
			Width(m.Viewport.Width).
			Height(m.Viewport.Height).
			Render(string(msg))

		m.Viewport.SetContent(m.ImageString)

		return m, nil
	case errorMsg:
		log.Println("error : ", msg.Error())
		m.FileName = ""
		m.ImageString = lipgloss.NewStyle().
			Width(m.Viewport.Width).
			Height(m.Viewport.Height).
			Render("Error: " + msg.Error())

		m.Viewport.SetContent(m.ImageString)

		return m, nil
	}

	if m.Active {
		m.Viewport, cmd = m.Viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View returns a string representation of the code bubble.
func (m ImageModel) View() string {
	border := lipgloss.NormalBorder()

	if m.Borderless {
		border = lipgloss.HiddenBorder()
	}

	m.Viewport.Style = lipgloss.NewStyle().
		PaddingLeft(padding).
		PaddingRight(padding).
		Border(border).
		BorderForeground(m.BorderColor)

	return m.Viewport.View()
}
