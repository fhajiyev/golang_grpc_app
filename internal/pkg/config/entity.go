package config

// RequestIngredients holds request
type RequestIngredients struct {
	AdID         string
	AndroidID    string
	AppID        int64
	UnitID       int64
	AppVersion   int
	Carrier      string
	DeviceName   string
	DeviceID     string
	Gender       string
	HMAC         string
	Locale       string
	SdkVersion   int
	Manufacturer string
}

// Config entity definition
type Config struct {
	Key   string
	Value string
}

// Configs entity definition holds an array of configs
type Configs struct {
	Configs *[]Config
}
