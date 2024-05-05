// modified from https://github.com/mistakenelf/teacup/blob/main/image/image.go
package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/disintegration/imaging"
	"github.com/lucasb-eyer/go-colorful"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"github.com/treethought/castr/db"
)

type imageDownloadMsg struct {
	url      string
	filename string
}
type convertImageToStringMsg struct {
	url string
	str string
}

type downloadError struct {
	err error
	url string
}

type decodeError struct {
	err error
	url string
}

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
	if cached, err := db.GetDB().Get([]byte(fmt.Sprintf("embed:%s", url))); err == nil {
		p := &embedPreview{}
		if err := json.Unmarshal(cached, p); err != nil {
			return p, nil
		}
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println("failed getting document", url, err)
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
	if preview.ImageURL != "" {
		if d, err := json.Marshal(preview); err == nil {
			if err := db.GetDB().Set([]byte(fmt.Sprintf("embed:%s", url)), d); err != nil {
				log.Println("error caching embed", err)
				return preview, nil
			}
		}
	}

	return preview, nil
}

func getImageCmd(width int, url string, embed bool) tea.Cmd {
	return func() tea.Msg {
		data, err := getImage(width, url, embed)
		if err != nil {
			return downloadError{err: err, url: url}
		}
		imgString, err := convertImageToString(width, data)
		if err != nil {
			return decodeError{err: err, url: url}
		}
		return convertImageToStringMsg{url: url, str: imgString}
	}
}

func getImage(width int, url string, embed bool) ([]byte, error) {
	if strings.HasSuffix(url, ".gif") || strings.HasSuffix(url, ".svg") {
		return nil, fmt.Errorf("gif not supported")
	}

	if embed {
		ep, err := getEmbedPreview(url)
		if err == nil && ep.ImageURL != "" {
			url = ep.ImageURL
		}
	}

	cached, err := db.GetDB().Get([]byte(fmt.Sprintf("img:%s", url)))
	if err == nil {
		return cached, nil
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	// set headers that would be sent in browser like user agent so we don't get ratelimited
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
	req.Header.Set("Accept", "image/png,image/jpeg,image/*,*/*;q=0.8")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	// req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("Cache-Control", "max-age=0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}
	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := db.GetDB().Set([]byte(fmt.Sprintf("img:%s", url)), d); err != nil {
		log.Println("error saving image", err)
	}
	return d, nil
}

func convertImageToString(width int, ib []byte) (string, error) {
	ir := bytes.NewReader(ib)

	img, _, err := image.Decode(ir)
	if err != nil {
		return "", err
	}
	// Check if the decoded image is of type NRGBA (non-alpha-premultiplied color)
	// If it's not, convert it to NRGBA
	// needed for bubbletea
	if _, ok := img.(*image.NRGBA); !ok {
		rgba := image.NewNRGBA(img.Bounds())
		draw.Draw(rgba, rgba.Bounds(), img, img.Bounds().Min, draw.Src)
		img = rgba
	}

	return ToString(width, img), nil
}

// ImageModel represents the properties of a code bubble.
type ImageModel struct {
	Viewport    viewport.Model
	BorderColor lipgloss.AdaptiveColor
	Active      bool
	Borderless  bool
	URL         string
	ImageString string
}

// New creates a new instance of code.
func NewImage(active, borderless bool, borderColor lipgloss.AdaptiveColor) *ImageModel {
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

	return &ImageModel{
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

func (m *ImageModel) SetURL(url string, embed bool) tea.Cmd {
	m.URL = url
	return getImageCmd(m.Viewport.Width, url, embed)
}

// SetFileName sets current file to highlight, this
// returns a cmd which will highlight the text.
// func (m *ImageModel) SetFileName(filename string) tea.Cmd {
// 	m.FileName = filename
// 	return convertImageToStringCmd(m.Viewport.Width, filename)
// }

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

	if m.ImageString != "" {
		return getImageCmd(w, m.URL, false)
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
func (m *ImageModel) Update(msg tea.Msg) (*ImageModel, tea.Cmd) {
	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case convertImageToStringMsg:
		if msg.url == m.URL && msg.str != "" {
			m.ImageString = lipgloss.NewStyle().
				Width(m.Viewport.Width).
				Height(m.Viewport.Height).
				Render(msg.str)
			m.Viewport.SetContent(m.ImageString)
		}
	case downloadError:
		if msg.url == m.URL {
			m.ImageString = lipgloss.NewStyle().
				Width(m.Viewport.Width).
				Height(m.Viewport.Height).
				Render("Error: " + msg.err.Error())
		}
	case decodeError:
		if msg.url == m.URL {
			log.Println("decode error: ", msg.url, msg.err.Error())
			m.ImageString = lipgloss.NewStyle().
				Width(m.Viewport.Width).
				Height(m.Viewport.Height).
				Render("Error: " + msg.err.Error())
		}
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
