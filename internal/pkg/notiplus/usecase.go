package notiplus

// UseCase interface definition
type UseCase interface {
	GetConfigsByUnitID(unitID int64) ([]Config, error)
}

type useCase struct {
	repo Repository
}

// GetConfigsByUnitID returns configs based on unit id
func (u *useCase) GetConfigsByUnitID(unitID int64) ([]Config, error) {
	return u.repo.GetConfigsByUnitID(unitID)
}

// NewUseCase returns UseCase interface
func NewUseCase(repo Repository) UseCase {
	return &useCase{repo: repo}
}
