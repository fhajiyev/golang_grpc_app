package repo

import (
	"net"

	"github.com/oschwald/geoip2-golang"
)

// GeoDB interface definition
type GeoDB interface {
	Country(ip net.IP) (*geoip2.Country, error)
}
