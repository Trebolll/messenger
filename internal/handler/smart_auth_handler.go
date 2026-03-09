package handler

import (
	"errors"
	"log"
	"messenger/internal/model"
	"messenger/internal/service"
	"messenger/internal/utils"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var reEmail = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type SmartAuthHandler struct {
	userService   *service.UserService
	notifyService *service.NotifyService
	otpStore      *service.OTPStore
}

func NewSmartAuthHandler(us *service.UserService, ns *service.NotifyService, otp *service.OTPStore) *SmartAuthHandler {
	return &SmartAuthHandler{userService: us, notifyService: ns, otpStore: otp}
}

// SendCode — POST /api/auth/send
// Шаг 1 регистрации: отправить код подтверждения на email или телефон
// { "login": "user@mail.com" } или { "login": "+79991234567" }
func (h *SmartAuthHandler) SendCode(c *gin.Context) {
	var req struct {
		Login string `json:"login" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "укажите email или телефон"})
		return
	}

	dest, method, err := parseLogin(req.Login)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем — уже зарегистрирован? Тогда нельзя регистрироваться повторно
	exists, err := h.userService.ExistsByLogin(dest, method)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка проверки"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "пользователь уже зарегистрирован, войдите", "exists": true})
		return
	}

	code, err := h.otpStore.Generate(dest)
	if err != nil {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}
	if err := h.notifyService.SendOTP(dest, code); err != nil {
		log.Printf("[SendCode] ошибка отправки на %s: %v", dest, err)
		h.otpStore.Delete(dest) // сбрасываем — чтобы повтор не давал 429
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"method": method, "login": dest})
}

// VerifyCode — POST /api/auth/verify-code
// Шаг 2 регистрации: проверить код, вернуть токен подтверждения для заполнения профиля
// { "login": "...", "code": "123456" }
func (h *SmartAuthHandler) VerifyCode(c *gin.Context) {
	var req struct {
		Login string `json:"login" binding:"required"`
		Code  string `json:"code"  binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "укажите логин и код"})
		return
	}

	dest, method, err := parseLogin(req.Login)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.otpStore.Verify(dest, strings.TrimSpace(req.Code)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Код верный — выдаём временный токен подтверждения (15 минут) для шага заполнения профиля
	confirmToken, err := utils.GenerateConfirmToken(dest, "confirm_secret")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"confirm_token": confirmToken,
		"method":        method,
		"login":         dest,
	})
}

// Register — POST /api/auth/register
// Шаг 3: заполнить профиль и создать аккаунт
// { "confirm_token": "...", "login": "...", "username": "ivan", "full_name": "Иван Иванов",
//
//	"birth_date": "1995-01-15", "location": "Москва", "password": "secret", "password2": "secret",
//	"extra_contact": "ivan@mail.com" или "+79991234567" }
func (h *SmartAuthHandler) Register(c *gin.Context) {
	var req struct {
		ConfirmToken string `json:"confirm_token" binding:"required"`
		Login        string `json:"login"         binding:"required"`
		Username     string `json:"username"      binding:"required"`
		FullName     string `json:"full_name"`
		BirthDate    string `json:"birth_date"`
		Location     string `json:"location"`
		Password     string `json:"password"      binding:"required"`
		Password2    string `json:"password2"     binding:"required"`
		ExtraContact string `json:"extra_contact"` // email если регался по phone и наоборот
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "заполните все обязательные поля"})
		return
	}

	// Проверяем confirm_token
	loginFromToken, err := utils.ParseConfirmToken(req.ConfirmToken, "confirm_secret")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "токен подтверждения недействителен или истёк"})
		return
	}

	dest, method, err := parseLogin(req.Login)
	if err != nil || dest != loginFromToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "несоответствие данных подтверждения"})
		return
	}

	if req.Password != req.Password2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "пароли не совпадают"})
		return
	}
	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "пароль должен быть не менее 6 символов"})
		return
	}

	u := &model.User{
		Username: req.Username,
		FullName: req.FullName,
		Location: req.Location,
		Password: req.Password,
	}
	if method == "phone" {
		u.Phone = dest
		u.Email = strings.ToLower(req.ExtraContact)
	} else {
		u.Email = dest
		u.Phone = req.ExtraContact
	}

	if err := h.userService.CreateUser(u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Разбираем дату рождения
	if req.BirthDate != "" {
		bd := req.BirthDate
		_ = h.userService.SetBirthDate(u.ID, bd)
	}

	token, err := utils.GenerateJWT(u.ID, "your_secret_key", 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось создать токен"})
		return
	}
	u.Password = ""
	c.JSON(http.StatusCreated, gin.H{"token": token, "user": u})
}

// Login — POST /api/auth/login
// { "login": "user@mail.com" или "+79991234567", "password": "secret" }
func (h *SmartAuthHandler) Login(c *gin.Context) {
	var req struct {
		Login    string `json:"login"    binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "укажите логин и пароль"})
		return
	}

	dest, method, err := parseLogin(req.Login)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user *model.User
	if method == "phone" {
		user, err = h.userService.LoginByPhone(dest, req.Password)
	} else {
		user, err = h.userService.LoginUser(dest, req.Password)
	}
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "неверный логин или пароль"})
		return
	}

	token, err := utils.GenerateJWT(user.ID, "your_secret_key", 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось создать токен"})
		return
	}
	user.Password = ""
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

// ── хелперы ───────────────────────────────────────────────────────────────────

func parseLogin(raw string) (dest, method string, err error) {
	s := strings.TrimSpace(raw)
	if isPhoneInput(s) {
		dest, err = normalizePhone(s)
		method = "phone"
		return
	}
	if reEmail.MatchString(s) {
		return strings.ToLower(s), "email", nil
	}
	// Всё остальное считаем username
	return "", "", errors.New("введите корректный email или номер телефона")
}

func isPhoneInput(s string) bool {
	clean := regexp.MustCompile(`[\s\-()+]`).ReplaceAllString(s, "")
	return regexp.MustCompile(`^[78]?\d{10}$|^\d{11,15}$`).MatchString(clean) && !strings.Contains(s, "@")
}

func normalizePhone(raw string) (string, error) {
	phone := regexp.MustCompile(`[\s\-()]+`).ReplaceAllString(strings.TrimSpace(raw), "")
	if strings.HasPrefix(phone, "8") && len(phone) == 11 {
		phone = "+7" + phone[1:]
	}
	if !strings.HasPrefix(phone, "+") {
		phone = "+" + phone
	}
	if ok, _ := regexp.MatchString(`^\+[1-9]\d{6,14}$`, phone); !ok {
		return "", errors.New("неверный формат телефона, используйте: +79991234567")
	}
	return phone, nil
}
