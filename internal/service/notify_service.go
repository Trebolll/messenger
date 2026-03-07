package service

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"net/url"
	"os"
	"strings"

	"encoding/json"
	"io"
	"net/http"

	"github.com/resend/resend-go/v3"
)

// NotifyService — отправка OTP через SMSC.ru, Telegram или Email
// SMS_PROVIDER=smsc|telegram
// Email: если задан RESEND_API_KEY — через Resend, иначе через SMTP
type NotifyService struct {
	provider string

	// SMSC
	smscLogin    string
	smscPassword string
	smscSender   string

	// Telegram
	tgBotToken string

	// Email
	resendAPIKey string
	smtpHost     string
	smtpPort     string
	smtpUser     string
	smtpPassword string
	smtpFrom     string
}

func NewNotifyServiceFromEnv() *NotifyService {
	return &NotifyService{
		provider:     os.Getenv("SMS_PROVIDER"),
		smscLogin:    os.Getenv("SMSC_LOGIN"),
		smscPassword: os.Getenv("SMSC_PASSWORD"),
		smscSender:   os.Getenv("SMSC_SENDER"),
		tgBotToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
		resendAPIKey: os.Getenv("RESEND_API_KEY"),
		smtpHost:     os.Getenv("SMTP_HOST"),
		smtpPort:     os.Getenv("SMTP_PORT"),
		smtpUser:     os.Getenv("SMTP_USER"),
		smtpPassword: os.Getenv("SMTP_PASSWORD"),
		smtpFrom:     os.Getenv("SMTP_FROM"),
	}
}

// SendOTP — если адрес содержит @ идёт email, иначе SMS
func (s *NotifyService) SendOTP(dest, code string) error {
	if strings.Contains(dest, "@") {
		return s.sendEmail(dest, code)
	}
	switch s.provider {
	case "telegram":
		return s.sendTelegram(dest, code)
	default:
		return s.sendSMSC(dest, code)
	}
}

// ── SMSC.ru ───────────────────────────────────────────────────────────────────

func (s *NotifyService) sendSMSC(phone, code string) error {
	params := url.Values{}
	params.Set("login", s.smscLogin)
	params.Set("psw", s.smscPassword)
	params.Set("phones", phone)
	params.Set("mes", fmt.Sprintf("Ваш код: %s. Действителен 5 минут.", code))
	params.Set("fmt", "3")
	params.Set("charset", "utf-8")
	if s.smscSender != "" {
		params.Set("sender", s.smscSender)
	}
	resp, err := http.Get("https://smsc.ru/sys/send.php?" + params.Encode())
	if err != nil {
		return fmt.Errorf("smsc: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Error string `json:"error"`
		Code  int    `json:"error_code"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("smsc: parse error: %w", err)
	}
	if result.Error != "" {
		return fmt.Errorf("smsc: ошибка %d: %s", result.Code, result.Error)
	}
	return nil
}

// ── Telegram ──────────────────────────────────────────────────────────────────

func (s *NotifyService) sendTelegram(chatID, code string) error {
	text := fmt.Sprintf("🔐 Ваш код подтверждения: *%s*\n\nДействителен 5 минут\\. Никому не сообщайте\\.", code)
	params := url.Values{}
	params.Set("chat_id", chatID)
	params.Set("text", text)
	params.Set("parse_mode", "MarkdownV2")
	resp, err := http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?%s", s.tgBotToken, params.Encode()))
	if err != nil {
		return fmt.Errorf("telegram: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("telegram: parse error: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("telegram: %s", result.Description)
	}
	return nil
}

// ── Email ─────────────────────────────────────────────────────────────────────

func (s *NotifyService) sendEmail(to, code string) error {
	// Resend SDK — приоритет
	if s.resendAPIKey != "" {
		client := resend.NewClient(s.resendAPIKey)
		_, err := client.Emails.Send(&resend.SendEmailRequest{
			From:    s.smtpFrom,
			To:      []string{to},
			Subject: "Код подтверждения",
			Text:    fmt.Sprintf("Ваш код: %s\n\nДействителен 5 минут. Никому не сообщайте.", code),
		})
		if err != nil {
			return fmt.Errorf("resend: %w", err)
		}
		return nil
	}
	// Fallback — SMTP
	if s.smtpHost == "" {
		return fmt.Errorf("email: задайте RESEND_API_KEY или SMTP_HOST в .env")
	}
	return s.sendSMTP(to, code)
}

func (s *NotifyService) sendSMTP(to, code string) error {
	subject := "Код подтверждения"
	text := fmt.Sprintf("Ваш код подтверждения: %s\r\n\r\nДействителен 5 минут.", code)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
		s.smtpFrom, to, subject, text)
	addr := s.smtpHost + ":" + s.smtpPort
	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPassword, s.smtpHost)
	if s.smtpPort == "465" {
		tlsConf := &tls.Config{ServerName: s.smtpHost}
		conn, err := tls.Dial("tcp", addr, tlsConf)
		if err != nil {
			return fmt.Errorf("smtp TLS: %w", err)
		}
		client, err := smtp.NewClient(conn, s.smtpHost)
		if err != nil {
			return fmt.Errorf("smtp client: %w", err)
		}
		defer client.Close()
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
		if err = client.Mail(s.smtpFrom); err != nil {
			return fmt.Errorf("smtp MAIL: %w", err)
		}
		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("smtp RCPT: %w", err)
		}
		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("smtp DATA: %w", err)
		}
		defer w.Close()
		_, err = fmt.Fprint(w, msg)
		return err
	}
	return smtp.SendMail(addr, auth, s.smtpFrom, []string{to}, []byte(msg))
}
