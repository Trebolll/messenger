package handler

import (
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

type LinkPreviewResponse struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	SiteName    string `json:"site_name"`
	Favicon     string `json:"favicon"`
}

var (
	reOGTitle       = regexp.MustCompile(`(?i)<meta[^>]+property=["']og:title["'][^>]+content=["']([^"']+)["']`)
	reOGTitleAlt    = regexp.MustCompile(`(?i)<meta[^>]+content=["']([^"']+)["'][^>]+property=["']og:title["']`)
	reOGDesc        = regexp.MustCompile(`(?i)<meta[^>]+property=["']og:description["'][^>]+content=["']([^"']+)["']`)
	reOGDescAlt     = regexp.MustCompile(`(?i)<meta[^>]+content=["']([^"']+)["'][^>]+property=["']og:description["']`)
	reOGImage       = regexp.MustCompile(`(?i)<meta[^>]+property=["']og:image["'][^>]+content=["']([^"']+)["']`)
	reOGImageAlt    = regexp.MustCompile(`(?i)<meta[^>]+content=["']([^"']+)["'][^>]+property=["']og:image["']`)
	reOGSiteName    = regexp.MustCompile(`(?i)<meta[^>]+property=["']og:site_name["'][^>]+content=["']([^"']+)["']`)
	reOGSiteNameAlt = regexp.MustCompile(`(?i)<meta[^>]+content=["']([^"']+)["'][^>]+property=["']og:site_name["']`)
	reTitleTag      = regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)
	reMetaDesc      = regexp.MustCompile(`(?i)<meta[^>]+name=["']description["'][^>]+content=["']([^"']+)["']`)
	reMetaDescAlt   = regexp.MustCompile(`(?i)<meta[^>]+content=["']([^"']+)["'][^>]+name=["']description["']`)
	reTwTitle       = regexp.MustCompile(`(?i)<meta[^>]+name=["']twitter:title["'][^>]+content=["']([^"']+)["']`)
	reTwImage       = regexp.MustCompile(`(?i)<meta[^>]+name=["']twitter:image["'][^>]+content=["']([^"']+)["']`)
)

func matchFirst(body string, patterns ...*regexp.Regexp) string {
	for _, re := range patterns {
		if m := re.FindStringSubmatch(body); len(m) > 1 {
			v := strings.TrimSpace(m[1])
			if v != "" {
				return v
			}
		}
	}
	return ""
}

func GetLinkPreview(c *gin.Context) {
	rawURL := strings.TrimSpace(c.Query("url"))
	if rawURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url required"})
		return
	}

	// Убедимся что есть схема
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url"})
		return
	}

	client := &http.Client{
		Timeout: 6 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "request error"})
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; LinkPreviewBot/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "fetch error"})
		return
	}
	defer resp.Body.Close()

	// Читаем только первые 100KB — достаточно для мета-тегов
	limited := io.LimitReader(resp.Body, 100*1024)
	bodyBytes, err := io.ReadAll(limited)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "read error"})
		return
	}

	// Некоторые сайты отдают latin1 — заменяем невалидные байты
	body := string(bodyBytes)
	if !utf8.ValidString(body) {
		body = strings.ToValidUTF8(body, "?")
	}

	title := matchFirst(body, reOGTitle, reOGTitleAlt, reTwTitle, reTitleTag)
	desc := matchFirst(body, reOGDesc, reOGDescAlt, reMetaDesc, reMetaDescAlt)
	image := matchFirst(body, reOGImage, reOGImageAlt, reTwImage)
	siteName := matchFirst(body, reOGSiteName, reOGSiteNameAlt)

	// Резолвим относительные URL изображений
	if image != "" && !strings.HasPrefix(image, "http") {
		base := parsed.Scheme + "://" + parsed.Host
		if strings.HasPrefix(image, "/") {
			image = base + image
		} else {
			image = base + "/" + image
		}
	}

	// Фавикон
	favicon := parsed.Scheme + "://" + parsed.Host + "/favicon1.ico"

	// Название сайта из хоста если не нашли
	if siteName == "" {
		host := parsed.Host
		host = strings.TrimPrefix(host, "www.")
		parts := strings.SplitN(host, ".", 2)
		if len(parts) > 0 {
			siteName = strings.Title(parts[0])
		}
	}

	// Обрезаем длинные строки
	if len([]rune(title)) > 120 {
		runes := []rune(title)
		title = string(runes[:120]) + "…"
	}
	if len([]rune(desc)) > 200 {
		runes := []rune(desc)
		desc = string(runes[:200]) + "…"
	}

	c.JSON(http.StatusOK, LinkPreviewResponse{
		URL:         rawURL,
		Title:       title,
		Description: desc,
		Image:       image,
		SiteName:    siteName,
		Favicon:     favicon,
	})
}
