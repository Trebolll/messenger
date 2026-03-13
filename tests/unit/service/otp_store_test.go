package service

import (
	"messenger/internal/service"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOTPStore_GenerateAndVerify_Success(t *testing.T) {
	s := service.NewOTPStore()
	phone := "+79991234567"

	code, err := s.Generate(phone)
	assert.NoError(t, err)
	assert.Len(t, code, 6)

	err = s.Verify(phone, code)
	assert.NoError(t, err)
}

func TestOTPStore_Verify_InvalidCode(t *testing.T) {
	s := service.NewOTPStore()
	phone := "+79991234567"

	_, err := s.Generate(phone)
	assert.NoError(t, err)

	err = s.Verify(phone, "000000")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "неверный код")
}

func TestOTPStore_ResendDelay(t *testing.T) {
	s := service.NewOTPStore()
	phone := "+79991234567"

	_, err := s.Generate(phone)
	assert.NoError(t, err)

	_, err = s.Generate(phone)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "повторная отправка возможна через")
}

func TestOTPStore_MaxAttempts(t *testing.T) {
	s := service.NewOTPStore()
	phone := "+79991234567"

	_, err := s.Generate(phone)
	assert.NoError(t, err)

	for i := 0; i < 5; i++ {
		err = s.Verify(phone, "000000")
		assert.Error(t, err)
	}

	// 6th attempt should fail with "превышено количество попыток"
	err = s.Verify(phone, "000000")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "превышено количество попыток")
}

func TestOTPStore_ResetToken(t *testing.T) {
	s := service.NewOTPStore()
	token := "test-token"
	dest := "test@example.com"

	s.StoreReset(token, dest)

	d, ok := s.ConsumeReset(token)
	assert.True(t, ok)
	assert.Equal(t, dest, d)

	// Second consume should fail
	_, ok = s.ConsumeReset(token)
	assert.False(t, ok)
}
