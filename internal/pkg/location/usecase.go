package location

import (
	"fmt"
	"net"
	"net/http"
)

// UseCase interface definition
type UseCase interface {
	GetClientLocation(httpRequest *http.Request, countryFromLocale string) *Location
	GetCountryFromIP(ip net.IP) (string, error)
}

type useCase struct {
	repo Repository
}

// GetClientLocation returns client location based on params
func (u *useCase) GetClientLocation(httpRequest *http.Request, countryFromLocale string) *Location {
	return u.repo.GetClientLocation(httpRequest, countryFromLocale)
}

// GetCountryFromIP returns Country based on ip
func (u *useCase) GetCountryFromIP(ip net.IP) (string, error) {
	return u.repo.GetCountryFromIP(ip)
}

// NewUseCase returns UseCase interface
func NewUseCase(repo Repository) UseCase {
	return &useCase{repo: repo}
}

func getCacheKeyUnit(unitID int64) string {
	return fmt.Sprintf("CACHE_GO_UNIT-%v", unitID)
}
