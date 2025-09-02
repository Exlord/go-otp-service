package storage

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"Exlord/otpservice/internal/types"

	"github.com/google/uuid"
)

// ----- Users -----
type InMemoryUserRepo struct {
	mu      sync.RWMutex
	byID    map[uuid.UUID]*types.User
	byPhone map[string]uuid.UUID
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{byID: map[uuid.UUID]*types.User{}, byPhone: map[string]uuid.UUID{}}
}

func (r *InMemoryUserRepo) UpsertByPhone(ctx context.Context, phone string) (*types.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if id, ok := r.byPhone[phone]; ok {
		return r.byID[id], nil
	}
	u := &types.User{ID: uuid.New(), Phone: phone, RegisteredAt: time.Now().UTC()}
	r.byID[u.ID] = u
	r.byPhone[phone] = u.ID
	return u, nil
}

func (r *InMemoryUserRepo) Get(ctx context.Context, id string) (*types.User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, false
	}
	u, ok := r.byID[uid]
	return u, ok
}

func (r *InMemoryUserRepo) List(ctx context.Context, search string, page, pageSize int) ([]*types.User, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var all []*types.User
	for _, u := range r.byID {
		all = append(all, u)
	}
	if s := strings.TrimSpace(strings.ToLower(search)); s != "" {
		filtered := all[:0]
		for _, u := range all {
			if strings.Contains(strings.ToLower(u.Phone), s) || strings.Contains(strings.ToLower(u.ID.String()), s) {
				filtered = append(filtered, u)
			}
		}
		all = filtered
	}
	sort.Slice(all, func(i, j int) bool { return all[i].RegisteredAt.After(all[j].RegisteredAt) })
	total := len(all)
	start := (page - 1) * pageSize
	if start > total {
		return []*types.User{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return all[start:end], total
}

// ----- OTP store -----
type otpRecord struct {
	Code      string
	ExpiresAt time.Time
}

type InMemoryOTPStore struct {
	mu   sync.RWMutex
	data map[string]otpRecord // by phone
}

func NewInMemoryOTPStore() *InMemoryOTPStore {
	return &InMemoryOTPStore{data: map[string]otpRecord{}}
}

func (s *InMemoryOTPStore) Save(phone, code string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[phone] = otpRecord{Code: code, ExpiresAt: time.Now().Add(ttl)}
}

func (s *InMemoryOTPStore) Verify(phone, code string) bool {
	s.mu.RLock()
	rec, ok := s.data[phone]
	s.mu.RUnlock()
	if !ok {
		return false
	}
	if time.Now().After(rec.ExpiresAt) {
		return false
	}
	return rec.Code == code
}
