package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// CaptchaService 验证码服务（内存存储）。
type CaptchaService struct {
	mu     sync.RWMutex
	store  map[string]*captchaEntry
}

type captchaEntry struct {
	code      string
	expiresAt time.Time
}

func NewCaptchaService() *CaptchaService {
	svc := &CaptchaService{store: make(map[string]*captchaEntry)}
	go svc.cleanupExpired()
	return svc
}

func (s *CaptchaService) cleanupExpired() {
	for {
		time.Sleep(5 * time.Minute)
		s.mu.Lock()
		now := time.Now()
		for k, v := range s.store {
			if now.After(v.expiresAt) {
				delete(s.store, k)
			}
		}
		s.mu.Unlock()
	}
}

// GenerateCaptcha 生成数学验证码问题。
func (s *CaptchaService) GenerateCaptcha() (id, question, answer string, err error) {
	a, _ := rand.Int(rand.Reader, big.NewInt(50))
	b, _ := rand.Int(rand.Reader, big.NewInt(50))
	op := "+"
	if a.Int64()%2 == 0 {
		op = "-"
	}

	var result int64
	if op == "+" {
		result = a.Int64() + b.Int64()
	} else {
		if a.Int64() < b.Int64() {
			a, b = b, a
		}
		result = a.Int64() - b.Int64()
	}

	id = fmt.Sprintf("captcha_%d", time.Now().UnixNano())
	question = fmt.Sprintf("%d %s %d = ?", a.Int64(), op, b.Int64())
	answer = fmt.Sprintf("%d", result)

	s.mu.Lock()
	s.store[id] = &captchaEntry{code: answer, expiresAt: time.Now().Add(5 * time.Minute)}
	s.mu.Unlock()

	return id, question, "", nil
}

// VerifyCaptcha 验证验证码答案。
func (s *CaptchaService) VerifyCaptcha(id, answer string) bool {
	s.mu.RLock()
	entry, ok := s.store[id]
	s.mu.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		return false
	}
	return entry.code == answer
}

// GenerateSMSCode 生成6位短信验证码。
func (s *CaptchaService) GenerateSMSCode(phone string) (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	code := fmt.Sprintf("%06d", 100000+n.Int64())
	s.mu.Lock()
	s.store["sms_"+phone] = &captchaEntry{code: code, expiresAt: time.Now().Add(5 * time.Minute)}
	s.mu.Unlock()
	return code, nil
}

// VerifySMS 验证短信验证码。
func (s *CaptchaService) VerifySMS(phone, code string) bool {
	s.mu.RLock()
	entry, ok := s.store["sms_"+phone]
	s.mu.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		return false
	}
	return entry.code == code
}
