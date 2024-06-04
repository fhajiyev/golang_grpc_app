package auth

import (
	"time"
)

// UseCase contains auth use cases.
type UseCase interface {
	CreateAuth(identifier Identifier) (string, error)
	GetAuth(token string) (*Auth, error)
}

const tokenExpiration time.Duration = 90 * 24 * time.Hour

type useCase struct {
	repo Repository
}

// NewUseCase creates new instance for auth use cases.
func NewUseCase(r Repository) UseCase {
	return &useCase{repo: r}
}

// CreateAuth func definition
func (u *useCase) CreateAuth(identifier Identifier) (string, error) {
	return u.repo.CreateAuth(identifier)
}

// GetAuth func definition
func (u *useCase) GetAuth(token string) (*Auth, error) {
	return u.repo.GetAuth(token)
}
