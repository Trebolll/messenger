package service

import (
	"messenger/internal/service"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Since operatorGateway is not exported, we can't test it directly unless we are in the same package
// or we export it. But wait, we can just use the internal/service package for our test.
// Actually, our test file is in 'service' package (the directory is tests/unit/service, but the package name is service).
// Wait, the package name in notify_service.go is 'service'.
// If I name my test package 'service', it will be in the same package IF the file is in the same directory.
// But it's in a different directory.

// I'll add a wrapper or just test via a public method if possible,
// but SendOTP has side effects.
// Let's check if I can make a test for SendOTP with dummy env.

func TestNotifyService_SendOTP_InvalidEmail2SMS(t *testing.T) {
	s := service.NewNotifyServiceFromEnv() // uses empty env

	// Testing unknown operator for email2sms
	err := s.SendOTP("+71110000000", "123456")
	// It should fail or try fallback. Default is "auto".
	// If provider is "auto", it tries email2sms first.
	// For +7111... it will return gateway == "", so sendEmail2SMS returns error.
	// Then it falls back to sendSMSC, which fails because SMSC_LOGIN is empty.
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "smsc: не задан SMSC_LOGIN")
}

func TestNotifyService_SendOTP_Email(t *testing.T) {
	s := service.NewNotifyServiceFromEnv()

	// Should try sendEmail. If no API key or SMTP host, returns error.
	err := s.SendOTP("test@example.com", "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email: задайте RESEND_API_KEY или SMTP_HOST")
}
