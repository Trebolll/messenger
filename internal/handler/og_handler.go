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

var ogTemplate = template.Must(template.New("og").Parse(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta property="og:type" content="article">
  <meta property="og:site_name" content="lambda">
  <meta property="og:title" content="{{.Title}}">
  <meta property="og:description" content="{{.Description}}">
  {{if .Image}}<meta property="og:image" content="{{.Image}}">{{end}}
  <meta property="og:url" content="{{.URL}}">
  <meta name="twitter:card" content="summary_large_image">
  <meta name="twitter:title" content="{{.Title}}">
  <meta name="twitter:description" content="{{.Description}}">
  {{if .Image}}<meta name="twitter:image" content="{{.Image}}">{{end}}
  <meta http-equiv="refresh" content="0; url={{.RedirectURL}}">
</head>
<body>
  <a href="{{.RedirectURL}}">Перейти к посту</a>
</body>
</html>`))

type ogData struct {
	Title       string
	Description string
	Image       string
	URL         string
	RedirectURL string
}

func (h *OGHandler) HandleOG(c *gin.Context) {
	postIDStr := c.Query("post")
	if postIDStr == "" {
		c.Redirect(http.StatusFound, "/")
		return
	}

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	post, err := h.wallRepo.GetPostByID(postID)
	if err != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Заголовок — имя автора
	title := fmt.Sprintf("%s в lambda", post.AuthorName)

	// Описание — текст поста (до 200 символов)
	desc := post.Content
	runes := []rune(desc)
	if len(runes) > 200 {
		desc = string(runes[:200]) + "…"
	}
	// Убираем переносы строк для мета-тегов
	desc = strings.ReplaceAll(desc, "\n", " ")

	// Картинка — первое изображение из вложений
	var image string
	for _, att := range post.Attachments {
		if strings.HasPrefix(att.MimeType, "image/") {
			image = att.Url
			break
		}
	}
	// Если нет картинки в посте — аватар автора
	if image == "" && post.AuthorAvatar != "" {
		image = post.AuthorAvatar
	}

	postURL := fmt.Sprintf("%s/?post=%s", h.baseURL, postIDStr)
	redirectURL := fmt.Sprintf("/?post=%s", postIDStr)

	data := ogData{
		Title:       title,
		Description: desc,
		Image:       image,
		URL:         postURL,
		RedirectURL: redirectURL,
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)
	ogTemplate.Execute(c.Writer, data)
}
