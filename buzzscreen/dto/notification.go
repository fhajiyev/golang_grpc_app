package dto

type (
	// NotificationsRequest contains request information of notification list api.
	NotificationsRequest struct {
		AdsRequest
	}

	// Notification contains polling notification model information.
	Notification struct {
		ID           int               `json:"id"`
		Importance   string            `json:"importance"`
		Title        string            `json:"title"`
		Description  string            `json:"description"`
		IconURL      string            `json:"icon_url"`
		InboxSummary string            `json:"inbox_summary"`
		Link         string            `json:"link"`
		AutoCancel   bool              `json:"auto_cancel"`
		Payload      map[string]string `json:"payload"`
	}
)
