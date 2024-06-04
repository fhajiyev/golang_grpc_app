package location

// Location entity definition
type Location struct {
	Country   string  `json:"country"`
	ZipCode   string  `json:"zipCode"`
	State     string  `json:"state"`
	City      string  `json:"city"`
	TimeZone  string  `json:"timeZone"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	IPAddress string  `json:"ipAddress"`
}
