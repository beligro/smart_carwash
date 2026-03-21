package webcodes

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

const (
	codeLength   = 6
	codeTTL      = 15 * time.Minute
	emailRateLimit = 60 * time.Second
)

// CodeType тип кода (регистрация или смена пароля)
type CodeType string

const (
	CodeTypeRegister      CodeType = "register"
	CodeTypePasswordChange CodeType = "password_change"
)

// EmailSender интерфейс отправки писем (RuSender)
type EmailSender interface {
	Send(ctx context.Context, toEmail, subject, htmlBody, textBody string) error
}

// Store хранит коды подтверждения по email и время последней отправки (rate limit)
type Store struct {
	mu       sync.Mutex
	codes    map[string]codeEntry
	lastSent map[string]time.Time
}

type codeEntry struct {
	code   string
	expiry time.Time
}

// NewStore создаёт хранилище кодов
func NewStore() *Store {
	s := &Store{
		codes:    make(map[string]codeEntry),
		lastSent: make(map[string]time.Time),
	}
	go s.cleanup()
	return s
}

func storeKey(email string, typ CodeType) string {
	return string(typ) + ":" + strings.ToLower(strings.TrimSpace(email))
}

func (s *Store) cleanup() {
	tick := time.NewTicker(1 * time.Minute)
	defer tick.Stop()
	for range tick.C {
		s.mu.Lock()
		now := time.Now()
		for key, e := range s.codes {
			if now.After(e.expiry) {
				delete(s.codes, key)
			}
		}
		s.mu.Unlock()
	}
}

// SetCode сохраняет код для email и типа на codeTTL
func (s *Store) SetCode(email string, typ CodeType, code string) {
	key := storeKey(email, typ)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.codes[key] = codeEntry{code: code, expiry: time.Now().Add(codeTTL)}
	s.lastSent[key] = time.Now()
}

// Verify проверяет код (constant-time) и при успехе удаляет его
func (s *Store) Verify(email string, typ CodeType, code string) bool {
	key := storeKey(email, typ)
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.codes[key]
	if !ok || time.Now().After(e.expiry) {
		return false
	}
	eq := subtleCompare(e.code, strings.TrimSpace(code))
	if eq {
		delete(s.codes, key)
		return true
	}
	return false
}

func subtleCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

// CanSend возвращает false и время ожидания, если для email+type ещё действует лимит
func (s *Store) CanSend(email string, typ CodeType) (allowed bool, retryAfter time.Duration) {
	key := storeKey(email, typ)
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.lastSent[key]
	if !ok {
		return true, 0
	}
	elapsed := time.Since(t)
	if elapsed >= emailRateLimit {
		return true, 0
	}
	return false, emailRateLimit - elapsed
}

// GenerateCode возвращает случайный 6-значный код
func GenerateCode() string {
	const digits = "0123456789"
	b := make([]byte, codeLength)
	for i := range b {
		b[i] = digits[rand.Intn(len(digits))]
	}
	return string(b)
}

// WebEmailCodeSender отправляет коды на email (регистрация и смена пароля)
type WebEmailCodeSender struct {
	store *Store
	email EmailSender
}

// NewWebEmailCodeSender создаёт отправителя кодов на email (email может быть nil — тогда только сохранение в store, для тестов)
func NewWebEmailCodeSender(store *Store, email EmailSender) *WebEmailCodeSender {
	return &WebEmailCodeSender{store: store, email: email}
}

const (
	subjectRegister      = "Код подтверждения регистрации — H2O"
	subjectPasswordChange = "Код для смены пароля — H2O"
	bodyTemplate         = "Ваш код для H2O: %s\n\nКод действителен 15 минут."
	htmlTemplate         = "<p>Ваш код для H2O: <strong>%s</strong></p><p>Код действителен 15 минут.</p>"
)

// SendRegisterCode генерирует код, сохраняет и отправляет на email
func (w *WebEmailCodeSender) SendRegisterCode(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	allowed, retryAfter := w.store.CanSend(email, CodeTypeRegister)
	if !allowed {
		return fmt.Errorf("повторите отправку через %d сек", int(retryAfter.Seconds()))
	}
	code := GenerateCode()
	w.store.SetCode(email, CodeTypeRegister, code)
	if w.email != nil {
		body := fmt.Sprintf(bodyTemplate, code)
		html := fmt.Sprintf(htmlTemplate, code)
		if err := w.email.Send(ctx, email, subjectRegister, html, body); err != nil {
			return fmt.Errorf("не удалось отправить письмо: %w", err)
		}
	}
	return nil
}

// VerifyRegisterCode проверяет код регистрации
func (w *WebEmailCodeSender) VerifyRegisterCode(email, code string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	return w.store.Verify(email, CodeTypeRegister, code)
}

// SendPasswordChangeCode генерирует код, сохраняет и отправляет на email
func (w *WebEmailCodeSender) SendPasswordChangeCode(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	allowed, retryAfter := w.store.CanSend(email, CodeTypePasswordChange)
	if !allowed {
		return fmt.Errorf("повторите отправку через %d сек", int(retryAfter.Seconds()))
	}
	code := GenerateCode()
	w.store.SetCode(email, CodeTypePasswordChange, code)
	if w.email != nil {
		body := fmt.Sprintf(bodyTemplate, code)
		html := fmt.Sprintf(htmlTemplate, code)
		if err := w.email.Send(ctx, email, subjectPasswordChange, html, body); err != nil {
			return fmt.Errorf("не удалось отправить письмо: %w", err)
		}
	}
	return nil
}

// VerifyPasswordChangeCode проверяет код смены пароля
func (w *WebEmailCodeSender) VerifyPasswordChangeCode(email, code string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	return w.store.Verify(email, CodeTypePasswordChange, code)
}
