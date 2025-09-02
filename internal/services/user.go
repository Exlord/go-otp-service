package services

import (
	"Exlord/otpservice/internal/storage"
	"Exlord/otpservice/internal/types"
	"context"
)

type UserService struct{ repo *storage.InMemoryUserRepo }

func NewUserService(r *storage.InMemoryUserRepo) *UserService { return &UserService{repo: r} }

func (s *UserService) UpsertByPhone(ctx context.Context, phone string) (*types.User, error) {
	return s.repo.UpsertByPhone(ctx, phone)
}

func (s *UserService) Get(ctx context.Context, id string) (*types.User, bool) {
	return s.repo.Get(ctx, id)
}

func (s *UserService) List(ctx context.Context, search string, page, pageSize int) ([]*types.User, int) {
	return s.repo.List(ctx, search, page, pageSize)
}
