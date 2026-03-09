package service

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"net/url"
	"os"
	"strings"
	"time"

	"encoding/json"
	"io"
	"net/http"

	"github.com/resend/resend-go/v3"
)

// NotifyService — отправка OTP
//
// Порядок попыток для SMS:
//  1. email-to-SMS через SMTP (бесплатно, работает для РФ-операторов)
//  2. SMSC.ru (платно, fallback если email-to-SMS не сработал)
//
// SMS_PROVIDER=smsc      → только SMSC
// SMS_PROVIDER=email2sms → только email-to-SMS (без fallback)
// SMS_PROVIDER=auto      → email-to-SMS → SMSC fallback (рекомендуется)
type NotifyService struct {
	provider string

	smscLogin    string
	smscPassword string
	smscSender   string

	tgBotToken string

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

// SendOTP — dest с @ → email, иначе SMS
func (s *NotifyService) SendOTP(dest, code string) error {
	if strings.Contains(dest, "@") {
		return s.sendEmail(dest, code)
	}
	switch s.provider {
	case "smsc":
		return s.sendSMSC(dest, code)
	case "telegram":
		return s.sendTelegram(dest, code)
	case "email2sms":
		return s.sendEmail2SMS(dest, code)
	default: // "auto" или пусто — email-to-SMS → SMSC fallback
		if err := s.sendEmail2SMS(dest, code); err != nil {
			log.Printf("[notify] email2sms failed (%v), fallback → SMSC", err)
			return s.sendSMSC(dest, code)
		}
		return nil
	}
}

// ── email-to-SMS ──────────────────────────────────────────────────────────────
// Операторы РФ держат шлюзы: <10 цифр без +7>@sms.<оператор>
// Определяем оператора по DEF-коду (первые 3 цифры после +7).

func operatorGateway(phone string) string {
	digits := strings.TrimPrefix(phone, "+7")
	if len(digits) < 10 {
		return ""
	}
	def := digits[:3]
	number := digits // 10 цифр

	// МТС: 910-919, 980-985, 936, 939
	mts := []string{"910", "911", "912", "913", "914", "915", "916", "917", "918", "919",
		"980", "981", "982", "983", "984", "985", "936", "939"}
	// Билайн: 903, 905, 906, 960-969
	beeline := []string{"903", "905", "906", "960", "961", "962", "963", "964", "965", "966", "967", "968", "969"}
	// Мегафон: 920-934, 937, 938, 950, 951
	megafon := []string{"920", "921", "922", "923", "924", "925", "926", "927", "928", "929",
		"930", "931", "932", "933", "934", "937", "938", "950", "951"}
	// Теле2: 900-902, 904, 908, 952, 953, 958
	tele2 := []string{"900", "901", "902", "904", "908", "952", "953", "958"}

	contains := func(list []string, s string) bool {
		for _, v := range list {
			if v == s {
				return true
			}
		}
		return false
	}

	switch {
	case contains(mts, def):
		return number + "@sms.mts.ru"
	case contains(beeline, def):
		return number + "@sms.beeline.ru"
	case contains(megafon, def):
		return number + "@sms.megafon.ru"
	case contains(tele2, def):
		return number + "@sms.tele2.ru"
	default:
		return ""
	}
}

func (s *NotifyService) sendEmail2SMS(phone, code string) error {
	if !strings.HasPrefix(phone, "+7") {
		return fmt.Errorf("email2sms: только номера +7 (РФ), получен: %s", phone)
	}
	gateway := operatorGateway(phone)
	if gateway == "" {
		return fmt.Errorf("email2sms: неизвестный оператор для %s", phone)
	}
	log.Printf("[notify] email2sms → %s", gateway)

	text := fmt.Sprintf("Код: %s", code)

	// Resend — приоритет (уже настроен для домена lambdahub.ru)
	if s.resendAPIKey != "" {
		type result struct{ err error }
		ch := make(chan result, 1)
		go func() {
			client := resend.NewClient(s.resendAPIKey)
			_, err := client.Emails.Send(&resend.SendEmailRequest{
				From:    s.smtpFrom,
				To:      []string{gateway},
				Subject: text,
				Text:    text,
			})
			ch <- result{err}
		}()
		select {
		case r := <-ch:
			if r.err != nil {
				return fmt.Errorf("email2sms via resend: %w", r.err)
			}
			return nil
		case <-time.After(5 * time.Second):
			return fmt.Errorf("email2sms: таймаут 5с")
		}
	}

	// Fallback — SMTP
	if s.smtpHost == "" {
		return fmt.Errorf("email2sms: задайте RESEND_API_KEY или SMTP_HOST")
	}
	type result struct{ err error }
	ch := make(chan result, 1)
	go func() { ch <- result{s.sendSMTP(gateway, text)} }()
	select {
	case r := <-ch:
		return r.err
	case <-time.After(5 * time.Second):
		return fmt.Errorf("email2sms: таймаут 5с")
	}
}

// ── SMSC.ru ───────────────────────────────────────────────────────────────────

func (s *NotifyService) sendSMSC(phone, code string) error {
	if s.smscLogin == "" {
		return fmt.Errorf("smsc: не задан SMSC_LOGIN")
	}
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
	if s.smtpHost == "" {
		return fmt.Errorf("email: задайте RESEND_API_KEY или SMTP_HOST в .env")
	}
	return s.sendSMTP(to, fmt.Sprintf("Ваш код подтверждения: %s\n\nДействителен 5 минут.", code))
}

// ── SMTP (общий транспорт) ────────────────────────────────────────────────────

func (s *NotifyService) sendSMTP(to, text string) error {
	subject := "Код подтверждения"
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
