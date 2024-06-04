package dto

import (
	"net/http"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/impressiondata"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/trackingdata"
)

// GetContentImpressionRequest is request parameters for content impression
type GetContentImpressionRequest struct {
	ImpressionDataStr string `query:"data"`
	ImpressionData    *impressiondata.ImpressionData

	TrackingDataStr *string `query:"tracking_data"`
	TrackingData    *trackingdata.TrackingData

	Place     *string `query:"place"`
	Position  *string `query:"position"`
	SessionID *string `query:"session_id"`

	Request *http.Request `form:"-" query:"-"`
}
