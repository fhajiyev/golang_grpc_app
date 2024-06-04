package trackingdata

// TrackingData struct definition
type TrackingData struct {
	ModelArtifact string `json:"ma"` // ML model 정보가 저장됨
}

const (
	trackingDataAESKey  = "4i3jfkdie0998754"
	trackingDataURLSafe = true
)
