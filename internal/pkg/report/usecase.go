package report

// UseCase interface definition
type UseCase interface {
	ReportContent(camp Request) error
	ReportAd(camp Request) error
}

type useCase struct {
	repo Repository
}

// ReportContent func definition
func (u *useCase) ReportContent(req Request) error {
	return u.repo.SaveContentReport(req)
}

// ReportAd func definition
func (u *useCase) ReportAd(req Request) error {
	return u.repo.SaveAdReport(req)
}

// NewUseCase returns UseCase interface
func NewUseCase(repo Repository) UseCase {
	return &useCase{repo: repo}
}
