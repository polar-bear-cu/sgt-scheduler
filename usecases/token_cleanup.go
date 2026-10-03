package usecases

import "context"

type RefreshTokenStore interface {
	DeleteExpiredRefreshTokens(ctx context.Context) (int64, error)
}

type TokenCleanupUsecase struct {
	tokens RefreshTokenStore
}

func NewTokenCleanup(tokens RefreshTokenStore) *TokenCleanupUsecase {
	return &TokenCleanupUsecase{tokens: tokens}
}

func (u *TokenCleanupUsecase) Run(ctx context.Context) (int64, error) {
	return u.tokens.DeleteExpiredRefreshTokens(ctx)
}
