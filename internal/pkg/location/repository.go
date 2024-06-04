package location

import (
	"net"
	"net/http"
)

// Repository defines an interface for location repository
type Repository interface {
	GetClientLocation(httpRequest *http.Request, countryFromLocale string) *Location
	GetCountryFromIP(ip net.IP) (string, error)
}
