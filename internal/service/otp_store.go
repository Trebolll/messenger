package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	otpTTL         = 5 * time.Minute
	otpResendDelay = 60 * time.Second // повторная отправка не раньше чем через 60с
	otpMaxAttempts = 5                // максимум попыток ввода кода
)

type otpEntry struct {
	code      string
	expiresAt time.Time
	sentAt    time.Time
	attempts  int
}

// OTPStore хранит коды в памяти с TTL — без таблицы в БД
type OTPStore struct {
	mu      sync.Mutex
	entries map[string]*otpEntry // key = нормализованный номер телефона
}

func NewOTPStore() *OTPStore {
	s := &OTPStore{entries: make(map[string]*otpEntry)}
	go s.cleanupLoop()
	return s
}

// Generate создаёт новый код для номера.
// Возвращает ошибку если повторная отправка слишком частая.
func (s *OTPStore) Generate(phone string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if e, ok := s.entries[phone]; ok {
		if time.Since(e.sentAt) < otpResendDelay {
			remaining := otpResendDelay - time.Since(e.sentAt)
			return "", fmt.Errorf("повторная отправка возможна через %.0f секунд", remaining.Seconds())
		}
	}

	code, err := generateCode()
	if err != nil {
		return "", err
	}

	s.entries[phone] = &otpEntry{
		code:      code,
		expiresAt: time.Now().Add(otpTTL),
		sentAt:    time.Now(),
	}
	return code, nil
}

// Verify проверяет код. Возвращает nil если код верный.
func (s *OTPStore) Verify(phone, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.entries[phone]
	if !ok {
		return errors.New("код не найден — запросите новый")
	}
	if time.Now().After(e.expiresAt) {
		delete(s.entries, phone)
		return errors.New("код истёк — запросите новый")
	}
	e.attempts++
	if e.attempts > otpMaxAttempts {
		delete(s.entries, phone)
		return errors.New("превышено количество попыток — запросите новый код")
	}
	if e.code != code {
		return fmt.Errorf("неверный код, осталось попыток: %d", otpMaxAttempts-e.attempts)
	}

	// Код верный — удаляем чтобы нельзя было использовать повторно
	delete(s.entries, phone)
	return nil
}

func generateCode() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// 3 случайных байта → число 0-999999 → 6-значный код
	n := (int(b[0])<<16 | int(b[1])<<8 | int(b[2])) % 1_000_000
	return fmt.Sprintf("%06d", n), nil
}

// Delete удаляет код — например при ошибке отправки чтобы следующая попытка не получила 429
func (s *OTPStore) Delete(phone string) {
	s.mu.Lock()
	delete(s.entries, phone)
	s.mu.Unlock()
}

// cleanupLoop раз в минуту удаляет истёкшие записи
func (s *OTPStore) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for phone, e := range s.entries {
			if now.After(e.expiresAt) {
				delete(s.entries, phone)
			}
		}
		s.mu.Unlock()
	}
}
