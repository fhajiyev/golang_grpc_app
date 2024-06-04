package profilerequest

// UseCase interface definition
type UseCase interface {
	PopulateProfile(account Account) error
}

type useCase struct {
	repo Repository
}

// PopulateProfile populates profile UID table based on ID types received inside pixel log data.
func (u *useCase) PopulateProfile(account Account) error {
	return u.repo.PopulateProfile(account)
}

// NewUseCase profilerequest 생성
func NewUseCase(repo Repository) UseCase {
	return &useCase{repo: repo}
}
