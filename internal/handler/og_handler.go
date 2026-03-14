package handler

import (
	"fmt"
	"html/template"
	"messenger/internal/repository"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OGHandler struct {
	wallRepo *repository.WallRepository
	baseURL  string
}

func NewOGHandler(wallRepo *repository.WallRepository, baseURL string) *OGHandler {
	return &OGHandler{wallRepo: wallRepo, baseURL: baseURL}
}

// ogTemplate — страница с OG-мета-тегами для ботов (Telegram, VK, etc.)
// Для видео-постов прописывает iframe-плеер (/player?post=UUID),
// потому что Telegram воспроизводит только og:video с type=text/html.
var ogTemplate = template.Must(template.New("og").Parse(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta property="og:type" content="{{if .Video}}video.other{{else}}article{{end}}">
  <meta property="og:site_name" content="lambda">
  <meta property="og:title" content="{{.Title}}">
  <meta property="og:description" content="{{.Description}}">
  <meta property="og:url" content="{{.URL}}">
  {{if .Video}}
  <meta property="og:video" content="{{.PlayerURL}}">
  <meta property="og:video:secure_url" content="{{.PlayerURL}}">
  <meta property="og:video:type" content="text/html">
  <meta property="og:video:width" content="1280">
  <meta property="og:video:height" content="720">
  {{if .Image}}<meta property="og:image" content="{{.Image}}">{{end}}
  {{if .Image}}<meta property="og:image:secure_url" content="{{.Image}}">{{end}}
  <meta name="twitter:card" content="player">
  <meta name="twitter:player" content="{{.PlayerURL}}">
  <meta name="twitter:player:width" content="1280">
  <meta name="twitter:player:height" content="720">
  <meta name="twitter:title" content="{{.Title}}">
  <meta name="twitter:description" content="{{.Description}}">
  {{if .Image}}<meta name="twitter:image" content="{{.Image}}">{{end}}
  {{else}}
  {{if .Image}}<meta property="og:image" content="{{.Image}}">{{end}}
  {{if .Image}}<meta property="og:image:secure_url" content="{{.Image}}">{{end}}
  <meta name="twitter:card" content="summary_large_image">
  <meta name="twitter:title" content="{{.Title}}">
  <meta name="twitter:description" content="{{.Description}}">
  {{if .Image}}<meta name="twitter:image" content="{{.Image}}">{{end}}
  {{end}}
  <meta http-equiv="refresh" content="0; url={{.RedirectURL}}">
</head>
<body>
  <a href="{{.RedirectURL}}">Перейти к посту</a>
</body>
</html>`))

// playerTemplate — минималистичный HTML-плеер.
// Telegram открывает og:video (type=text/html) в iframe прямо в чате.
// autoplay + playsinline + controls — всё что нужно для воспроизведения со звуком.
var playerTemplate = template.Must(template.New("player").Parse(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Title}}</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    html, body {
      width: 100%; height: 100%;
      background: #000;
      overflow: hidden;
    }
    video {
      width: 100%; height: 100%;
      object-fit: contain;
      display: block;
    }
  </style>
</head>
<body>
  <video
    src="{{.Video}}"
    autoplay
    controls
    playsinline
    preload="auto"
  ></video>
</body>
</html>`))

type ogData struct {
	Title       string
	Description string
	Image       string
	Video       string
	VideoMime   string
	PlayerURL   string // URL iframe-плеера: /player?post=UUID
	URL         string
	RedirectURL string
}

// isBotUserAgent проверяет, является ли User-Agent ботом для превью ссылок.
func isBotUserAgent(ua string) bool {
	ua = strings.ToLower(ua)
	bots := []string{
		"telegrambot",
		"twitterbot",
		"facebookexternalhit",
		"vkshare",
		"whatsapp",
		"slackbot",
		"discordbot",
		"linkedinbot",
		"applebot",
	}
	for _, bot := range bots {
		if strings.Contains(ua, bot) {
			return true
		}
	}
	return false
}

func (h *OGHandler) buildOGData(postIDStr string) (*ogData, error) {
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		return nil, err
	}

	post, err := h.wallRepo.GetPostByID(postID)
	if err != nil {
		return nil, err
	}

	title := fmt.Sprintf("%s в lambda", post.AuthorName)

	desc := post.Content
	runes := []rune(desc)
	if len(runes) > 200 {
		desc = string(runes[:200]) + "…"
	}
	desc = strings.ReplaceAll(desc, "\n", " ")

	var image string
	var video string
	var videoMime string
	for _, att := range post.Attachments {
		if strings.HasPrefix(att.MimeType, "video/") && video == "" {
			video = att.Url
			videoMime = att.MimeType
		}
		if strings.HasPrefix(att.MimeType, "image/") && image == "" {
			image = att.Url
		}
	}
	if image == "" && post.AuthorAvatar != "" {
		image = post.AuthorAvatar
	}

	postURL := fmt.Sprintf("%s/?post=%s", h.baseURL, postIDStr)
	redirectURL := fmt.Sprintf("/?post=%s", postIDStr)

	// playerURL — абсолютный URL iframe-плеера, нужен Telegram
	playerURL := ""
	if video != "" {
		playerURL = fmt.Sprintf("%s/player?post=%s", h.baseURL, postIDStr)
	}

	_ = videoMime // оставляем поле для совместимости, но Telegram игнорирует тип файла

	return &ogData{
		Title:       title,
		Description: desc,
		Image:       image,
		Video:       video,
		VideoMime:   videoMime,
		PlayerURL:   playerURL,
		URL:         postURL,
		RedirectURL: redirectURL,
	}, nil
}

// HandleOG обрабатывает маршрут /og?post=UUID (прямой запрос OG-страницы).
func (h *OGHandler) HandleOG(c *gin.Context) {
	postIDStr := c.Query("post")
	if postIDStr == "" {
		c.Redirect(http.StatusFound, "/")
		return
	}

	data, err := h.buildOGData(postIDStr)
	if err != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)
	ogTemplate.Execute(c.Writer, data)
}

// HandlePlayer обрабатывает маршрут /player?post=UUID.
// Telegram открывает этот URL в iframe внутри чата как встроенный плеер.
func (h *OGHandler) HandlePlayer(c *gin.Context) {
	postIDStr := c.Query("post")
	if postIDStr == "" {
		c.Status(http.StatusNotFound)
		return
	}

	data, err := h.buildOGData(postIDStr)
	if err != nil || data.Video == "" {
		c.Status(http.StatusNotFound)
		return
	}

	// Разрешаем встраивание в iframe (Telegram требует отсутствие X-Frame-Options: DENY)
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("X-Frame-Options", "ALLOWALL")
	c.Header("Content-Security-Policy", "frame-ancestors *")
	c.Status(http.StatusOK)
	playerTemplate.Execute(c.Writer, data)
}

// HandleIndex обрабатывает маршрут /?post=UUID.
// Если запрос от бота — отдаёт OG-страницу с мета-тегами.
// Иначе — отдаёт обычный index.html (SPA).
func (h *OGHandler) HandleIndex(c *gin.Context) {
	postIDStr := c.Query("post")
	ua := c.GetHeader("User-Agent")

	if postIDStr != "" && isBotUserAgent(ua) {
		data, err := h.buildOGData(postIDStr)
		if err == nil {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.Status(http.StatusOK)
			ogTemplate.Execute(c.Writer, data)
			return
		}
	}

	c.HTML(http.StatusOK, "index.html", nil)
}
