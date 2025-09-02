package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"
)

type OTPStore interface {
	Save(phone, code string, ttl time.Duration)
	Verify(phone, code string) bool
}

type RateLimiter interface {
	Allow(key string) error
}

type OTPService struct {
	store OTPStore
	rate RateLimiter
	ttl time.Duration
}

func NewOTPService(store OTPStore, rate RateLimiter) *OTPService {
	return &OTPService{store: store, rate: rate, ttl: 2 * time.Minute}
}

func (s *OTPService) Generate(ctx context.Context, phone string) (string, error) {
	if err := s.rate.Allow("otp:"+phone); err != nil { return "", err }
	code := rand6()
	s.store.Save(phone, code, s.ttl)
	return code, nil
}

func (s *OTPService) Verify(ctx context.Context, phone, code string) bool {
	return s.store.Verify(phone, code)
}

func rand6() string {
	// secure 6-digit numeric
	var n [3]byte
	_, _ = rand.Read(n[:])
	v := (uint32(n[0])<<16 | uint32(n[1])<<8 | uint32(n[2])) % 1000000
	return fmt.Sprintf("%06d", v)
}
