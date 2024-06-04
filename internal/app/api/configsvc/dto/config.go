package dto

// GetConfigsRequest struct definition
type GetConfigsRequest struct {
	UnitID       int64  `query:"unit_id"`
	Manufacturer string `query:"manufacturer"`

	Package    string `query:"package"`
	SDKVersion int    `query:"sdk_version"`
}

// Config struct definition
type Config struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
