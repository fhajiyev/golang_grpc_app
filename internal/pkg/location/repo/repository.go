package repo

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/location"
)

// Repository struct definition
type Repository struct {
	geoDB GeoDB
}

// LocationResponse defines a format
type LocationResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Result  *struct {
		*location.Location `json:"location"`
	} `json:"result"`
}

// GetCountryFromIP returns country based on IP
func (r *Repository) GetCountryFromIP(ip net.IP) (string, error) {
	cou, err := r.geoDB.Country(ip)
	if err != nil {
		return "", err
	}
	return cou.Country.IsoCode, err
}

// GetClientLocation returns location based on locale param.
func (r *Repository) GetClientLocation(httpRequest *http.Request, countryFromLocale string) *location.Location {
	ip := strings.Split(httpRequest.RemoteAddr, ":")[0]
	if len(ip) < 5 {
		ip = "127.0.0.1"
	}

	locationRequest := network.Request{
		URL: env.Config.InsightURL + "/location",
		Params: &url.Values{
			"headerForwardedFor": {strings.Join(httpRequest.Header["X-Forwarded-For"], ",")},
			"remoteIp":           {ip},
			"defaultCountry":     {countryFromLocale},
		},
		Method:  "GET",
		Timeout: time.Second * 2,
	}
	var res LocationResponse
	statusCode, err := locationRequest.GetResponse(&res)

	if statusCode != 200 || res.Result == nil || res.Result.Location == nil {
		if err != nil {
			core.Logger.WithError(err).Warnf("GetClientLocation() - statusCode: %v, Header: %+v, Res: %v", statusCode, httpRequest.Header, res)
		} else {
			core.Logger.Warnf("GetClientLocation() - statusCode: %v, Header: %+v, Res: %v", statusCode, httpRequest.Header, res)
		}
		location := location.Location{IPAddress: ip}
		if len(countryFromLocale) > 1 {
			location.Country = countryFromLocale
		} else {
			location.Country = "ZZ"
		}
		return &location
	}

	if res.Result.Location.IPAddress == "" {
		res.Result.Location.IPAddress = ip
	}
	return res.Result.Location
}

// New returns Repository struct
func New(geoDB GeoDB) *Repository {
	return &Repository{geoDB}
}
