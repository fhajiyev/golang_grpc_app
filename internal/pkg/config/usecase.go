package config

// UseCase interface definition
type UseCase interface {
	GetConfigs(configReq RequestIngredients) *[]Config
}

type useCase struct {
	repo Repository
}

// GetConfigs returns configs based on the request ingredients
func (u *useCase) GetConfigs(configReq RequestIngredients) *[]Config {
	return u.repo.GetConfigs(configReq)
}

// NewUseCase returns UseCase interface
func NewUseCase(repo Repository) UseCase {
	return &useCase{repo: repo}
}
